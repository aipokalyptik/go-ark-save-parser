package arksave

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type Save struct {
	path string
	db   *sql.DB

	Context *Context
	names   *arkbinary.Context
}

type ObjectClassInfo struct {
	UUID      uuid.UUID
	ClassName string
}

type ParsedObjectInfo struct {
	UUID      uuid.UUID
	ClassName string
	Object    *arkobject.GameObject
}

type FaultyObjectInfo struct {
	UUID      uuid.UUID
	ClassName string
	Err       error
}

func Open(path string) (*Save, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", abs)
	if err != nil {
		return nil, err
	}
	save := &Save{
		path: abs,
		db:   db,
		Context: &Context{
			Names:                   map[uint32]string{},
			ActorTransforms:         map[uuid.UUID]arkobject.ActorTransform{},
			ActorTransformPositions: map[uuid.UUID]int{},
		},
		names: arkbinary.NewContext(),
	}
	if err := save.readHeader(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := save.readActorTransforms(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return save, nil
}

func (s *Save) Close() error {
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

func (s *Save) CustomValue(key string) ([]byte, error) {
	var value []byte
	err := s.db.QueryRow(`select value from custom where key = ? limit 1`, key).Scan(&value)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(value))
	copy(out, value)
	return out, nil
}

func (s *Save) ObjectIDs() ([]uuid.UUID, error) {
	rows, err := s.db.Query(`select key from game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		id, err := uuid.FromBytes(raw)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Save) ObjectBinary(id uuid.UUID) ([]byte, error) {
	var value []byte
	err := s.db.QueryRow(`select value from game where key = ?`, id[:]).Scan(&value)
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(value))
	copy(out, value)
	return out, nil
}

func (s *Save) ClassOf(id uuid.UUID) (string, error) {
	raw, err := s.ObjectBinary(id)
	if err != nil {
		return "", err
	}
	r := arkbinary.NewReader(raw, s.names)
	return r.ReadName("")
}

func (s *Save) Classes() ([]string, error) {
	rows, err := s.db.Query(`select value from game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := map[string]struct{}{}
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		seen[className] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	classes := make([]string, 0, len(seen))
	for className := range seen {
		classes = append(classes, className)
	}
	sort.Strings(classes)
	return classes, nil
}

func (s *Save) ObjectClassInfos() ([]ObjectClassInfo, error) {
	rows, err := s.db.Query(`select key, value from game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var infos []ObjectClassInfo
	for rows.Next() {
		var key []byte
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		id, err := uuid.FromBytes(key)
		if err != nil {
			return nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		infos = append(infos, ObjectClassInfo{UUID: id, ClassName: className})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(infos, func(i int, j int) bool {
		return infos[i].UUID.String() < infos[j].UUID.String()
	})
	return infos, nil
}

func (s *Save) ParsedObjects(match func(ObjectClassInfo) bool) ([]ParsedObjectInfo, error) {
	rows, err := s.db.Query(`select key, value from game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := s.objectSections()
	var infos []ParsedObjectInfo
	for rows.Next() {
		var key []byte
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		id, err := uuid.FromBytes(key)
		if err != nil {
			return nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		classInfo := ObjectClassInfo{UUID: id, ClassName: className}
		if match != nil && !match(classInfo) {
			continue
		}
		object, err := arkobject.ParseGameObject(id, raw, s.names, sections)
		if err != nil {
			return nil, err
		}
		infos = append(infos, ParsedObjectInfo{UUID: id, ClassName: className, Object: object})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortParsedObjects(infos)
	return infos, nil
}

func (s *Save) ParsedObjectsWithFaults(match func(ObjectClassInfo) bool) ([]ParsedObjectInfo, []FaultyObjectInfo, error) {
	rows, err := s.db.Query(`select key, value from game`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	sections := s.objectSections()
	var infos []ParsedObjectInfo
	var faults []FaultyObjectInfo
	for rows.Next() {
		var key []byte
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, nil, err
		}
		id, err := uuid.FromBytes(key)
		if err != nil {
			return nil, nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			faults = append(faults, FaultyObjectInfo{UUID: id, Err: err})
			continue
		}
		classInfo := ObjectClassInfo{UUID: id, ClassName: className}
		if match != nil && !match(classInfo) {
			continue
		}
		object, err := arkobject.ParseGameObject(id, raw, s.names, sections)
		if err != nil {
			faults = append(faults, FaultyObjectInfo{UUID: id, ClassName: className, Err: err})
			continue
		}
		infos = append(infos, ParsedObjectInfo{UUID: id, ClassName: className, Object: object})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	sortParsedObjects(infos)
	sort.Slice(faults, func(i int, j int) bool {
		return faults[i].UUID.String() < faults[j].UUID.String()
	})
	return infos, faults, nil
}

func (s *Save) ParsedObject(id uuid.UUID) (*arkobject.GameObject, error) {
	raw, err := s.ObjectBinary(id)
	if err != nil {
		return nil, err
	}
	return arkobject.ParseGameObject(id, raw, s.names, s.objectSections())
}

func (s *Save) ParsedObjectPartial(id uuid.UUID) (*arkobject.GameObject, error) {
	raw, err := s.ObjectBinary(id)
	if err != nil {
		return nil, err
	}
	return arkobject.ParseGameObjectPartial(id, raw, s.names, s.objectSections())
}

func (s *Save) ParsedObjectsByClassContains(substr string) ([]ParsedObjectInfo, error) {
	return s.ParsedObjects(func(info ObjectClassInfo) bool {
		return strings.Contains(info.ClassName, substr)
	})
}

func (s *Save) ParsedObjectsWithAnyProperty(names []string) ([]ParsedObjectInfo, error) {
	patterns := s.propertyNamePatterns(names)
	if len(patterns) == 0 {
		return nil, nil
	}
	return s.parsedObjectsMatchingRaw(func(raw []byte) bool {
		return containsAnyPattern(raw, patterns)
	})
}

func (s *Save) ParsedObjectsWithAnyPropertyWithFaults(names []string) ([]ParsedObjectInfo, []FaultyObjectInfo, error) {
	patterns := s.propertyNamePatterns(names)
	if len(patterns) == 0 {
		return nil, nil, nil
	}
	return s.parsedObjectsMatchingRawWithFaults(func(raw []byte) bool {
		return containsAnyPattern(raw, patterns)
	})
}

func (s *Save) ObjectClassInfosWithAnyProperty(names []string) ([]ObjectClassInfo, error) {
	patterns := s.propertyNamePatterns(names)
	if len(patterns) == 0 {
		return nil, nil
	}
	rows, err := s.db.Query(`select key, value from game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var infos []ObjectClassInfo
	for rows.Next() {
		var key []byte
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		if !containsAnyPattern(raw, patterns) {
			continue
		}
		id, err := uuid.FromBytes(key)
		if err != nil {
			return nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		infos = append(infos, ObjectClassInfo{UUID: id, ClassName: className})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(infos, func(i int, j int) bool {
		return infos[i].UUID.String() < infos[j].UUID.String()
	})
	return infos, nil
}

func (s *Save) parsedObjectsMatchingRaw(matchRaw func([]byte) bool) ([]ParsedObjectInfo, error) {
	rows, err := s.db.Query(`select key, value from game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sections := s.objectSections()
	var infos []ParsedObjectInfo
	for rows.Next() {
		var key []byte
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		if matchRaw != nil && !matchRaw(raw) {
			continue
		}
		id, err := uuid.FromBytes(key)
		if err != nil {
			return nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		object, err := arkobject.ParseGameObject(id, raw, s.names, sections)
		if err != nil {
			return nil, err
		}
		infos = append(infos, ParsedObjectInfo{UUID: id, ClassName: className, Object: object})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sortParsedObjects(infos)
	return infos, nil
}

func (s *Save) parsedObjectsMatchingRawWithFaults(matchRaw func([]byte) bool) ([]ParsedObjectInfo, []FaultyObjectInfo, error) {
	rows, err := s.db.Query(`select key, value from game`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	sections := s.objectSections()
	var infos []ParsedObjectInfo
	var faults []FaultyObjectInfo
	for rows.Next() {
		var key []byte
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, nil, err
		}
		if matchRaw != nil && !matchRaw(raw) {
			continue
		}
		id, err := uuid.FromBytes(key)
		if err != nil {
			return nil, nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			faults = append(faults, FaultyObjectInfo{UUID: id, Err: err})
			continue
		}
		object, err := arkobject.ParseGameObject(id, raw, s.names, sections)
		if err != nil {
			faults = append(faults, FaultyObjectInfo{UUID: id, ClassName: className, Err: err})
			continue
		}
		infos = append(infos, ParsedObjectInfo{UUID: id, ClassName: className, Object: object})
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	sortParsedObjects(infos)
	sort.Slice(faults, func(i int, j int) bool {
		return faults[i].UUID.String() < faults[j].UUID.String()
	})
	return infos, faults, nil
}

func (s *Save) propertyNamePatterns(names []string) [][]byte {
	if s == nil || s.Context == nil {
		return nil
	}
	wanted := map[string]struct{}{}
	for _, name := range names {
		if name != "" {
			wanted[name] = struct{}{}
		}
	}
	if len(wanted) == 0 {
		return nil
	}
	patterns := make([][]byte, 0, len(wanted))
	for id, name := range s.Context.Names {
		if _, ok := wanted[name]; !ok {
			continue
		}
		pattern := make([]byte, 8)
		binary.LittleEndian.PutUint32(pattern[:4], id)
		patterns = append(patterns, pattern)
	}
	return patterns
}

func containsAnyPattern(raw []byte, patterns [][]byte) bool {
	for _, pattern := range patterns {
		if bytes.Contains(raw, pattern) {
			return true
		}
	}
	return false
}

func (s *Save) ObjectIDsByClassContains(substr string) ([]uuid.UUID, error) {
	rows, err := s.db.Query(`select key, value from game`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var key []byte
		var raw []byte
		if err := rows.Scan(&key, &raw); err != nil {
			return nil, err
		}
		r := arkbinary.NewReader(raw, s.names)
		className, err := r.ReadName("")
		if err != nil {
			return nil, err
		}
		if !strings.Contains(className, substr) {
			continue
		}
		id, err := uuid.FromBytes(key)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (s *Save) Object(id uuid.UUID) (*arkobject.GameObject, error) {
	raw, err := s.ObjectBinary(id)
	if err != nil {
		return nil, err
	}
	return arkobject.ParseGameObject(id, raw, s.names, s.objectSections())
}

func (s *Save) objectSections() []string {
	sections := make([]string, len(s.Context.Sections))
	for i, section := range s.Context.Sections {
		sections[i] = section.Raw
	}
	return sections
}

func sortParsedObjects(infos []ParsedObjectInfo) {
	sort.Slice(infos, func(i int, j int) bool {
		return infos[i].UUID.String() < infos[j].UUID.String()
	})
}

func (s *Save) ActorTransform(id uuid.UUID) (arkobject.ActorTransform, bool) {
	value, ok := s.Context.ActorTransforms[id]
	return value, ok
}

func (s *Save) readActorTransforms() error {
	raw, err := s.CustomValue("ActorTransforms")
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	transforms, positions, err := parseActorTransforms(raw)
	if err != nil {
		return fmt.Errorf("read ActorTransforms: %w", err)
	}
	s.Context.ActorTransforms = transforms
	s.Context.ActorTransformPositions = positions
	return nil
}

func parseActorTransforms(raw []byte) (map[uuid.UUID]arkobject.ActorTransform, map[uuid.UUID]int, error) {
	r := arkbinary.NewReader(raw, nil)
	transforms := map[uuid.UUID]arkobject.ActorTransform{}
	positions := map[uuid.UUID]int{}
	for r.HasMore() {
		position := r.Position()
		id, err := r.ReadUUID()
		if err != nil {
			return nil, nil, err
		}
		if id == uuid.Nil {
			break
		}
		transform, err := arkobject.ReadActorTransform(r)
		if err != nil {
			return nil, nil, err
		}
		transforms[id] = transform
		positions[id] = position
	}
	return transforms, positions, nil
}

func (s *Save) readHeader() error {
	raw, err := s.CustomValue("SaveHeader")
	if err != nil {
		return fmt.Errorf("read SaveHeader: %w", err)
	}
	r := arkbinary.NewReader(raw, nil)
	version, err := r.ReadInt16()
	if err != nil {
		return err
	}
	s.Context.SaveVersion = version

	if version >= 14 {
		if _, err := r.ReadUInt32(); err != nil {
			return err
		}
		if _, err := r.ReadUInt32(); err != nil {
			return err
		}
	}

	nameOffset, err := r.ReadInt32()
	if err != nil {
		return err
	}
	gameTime, err := r.ReadFloat64()
	if err != nil {
		return err
	}
	s.Context.GameTime = gameTime
	if version >= 12 {
		unknown, err := r.ReadUInt32()
		if err != nil {
			return err
		}
		s.Context.UnknownValue = unknown
	}

	sections, err := readLocations(r)
	if err != nil {
		return err
	}
	s.Context.Sections = sections

	if err := r.SetPosition(30); err != nil {
		return err
	}
	mapName, err := r.ReadString()
	if err != nil {
		return err
	}
	if mapName != nil {
		s.Context.MapName = *mapName
	}

	if err := r.SetPosition(int(nameOffset)); err != nil {
		return err
	}
	names, err := readNameTable(r)
	if err != nil {
		return err
	}
	s.Context.Names = names
	s.names.SetNames(names)
	return nil
}

func readLocations(r *arkbinary.Reader) ([]HeaderLocation, error) {
	count, err := r.ReadUInt32()
	if err != nil {
		return nil, err
	}
	out := make([]HeaderLocation, 0, count)
	for i := uint32(0); i < count; i++ {
		part, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		if part != nil && *part != "" && !isWorldPartitionName(*part) {
			out = append(out, HeaderLocation{Raw: *part})
		}
		sentinel, err := r.ReadUInt32()
		if err != nil {
			return nil, err
		}
		if sentinel != 0xffffffff {
			return nil, fmt.Errorf("invalid header location sentinel %#x", sentinel)
		}
	}
	return out, nil
}

func isWorldPartitionName(part string) bool {
	return len(part) >= 3 && part[len(part)-3:] == "_WP"
}

func readNameTable(r *arkbinary.Reader) (map[uint32]string, error) {
	count, err := r.ReadInt32()
	if err != nil {
		return nil, err
	}
	if count < 0 {
		return nil, fmt.Errorf("negative name table count %d", count)
	}
	names := make(map[uint32]string, count)
	for i := int32(0); i < count; i++ {
		key, err := r.ReadUInt32()
		if err != nil {
			return nil, err
		}
		value, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		if value != nil {
			names[key] = *value
		}
	}
	return names, nil
}
