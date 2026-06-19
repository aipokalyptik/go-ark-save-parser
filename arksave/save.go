package arksave

import (
	"database/sql"
	"fmt"
	"path/filepath"

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

func Open(path string) (*Save, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?mode=ro", abs))
	if err != nil {
		return nil, err
	}
	save := &Save{
		path:    abs,
		db:      db,
		Context: &Context{Names: map[uint32]string{}},
		names:   arkbinary.NewContext(),
	}
	if err := save.readHeader(); err != nil {
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

func (s *Save) Object(id uuid.UUID) (*arkobject.GameObject, error) {
	raw, err := s.ObjectBinary(id)
	if err != nil {
		return nil, err
	}
	sections := make([]string, len(s.Context.Sections))
	for i, section := range s.Context.Sections {
		sections[i] = section.Raw
	}
	return arkobject.ParseGameObject(id, raw, s.names, sections)
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
	var out []HeaderLocation
	for {
		part, err := r.ReadString()
		if err != nil {
			return nil, err
		}
		if part == nil || *part == "" {
			return out, nil
		}
		out = append(out, HeaderLocation{Raw: *part})
	}
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
