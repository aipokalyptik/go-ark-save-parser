package testfixtures

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"os"
	"sort"
	"testing"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type SaveOptions struct {
	Header      []byte
	Objects     map[uuid.UUID][]byte
	Custom      map[string][]byte
	EmptyTables bool
}

func WriteSave(tb testing.TB, path string, opts SaveOptions) {
	tb.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		tb.Fatalf("open sqlite fixture: %v", err)
	}
	defer db.Close()
	MustExec(tb, db, `create table custom (key text primary key, value blob)`)
	MustExec(tb, db, `create table game (key blob primary key, value blob)`)
	header := opts.Header
	if header == nil {
		header = Header("Valguero_WP", map[uint32]string{1: "Blueprint'/Game/Test.Test_C'"})
	}
	MustExec(tb, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", header)
	for key, value := range opts.Custom {
		MustExec(tb, db, `insert into custom (key, value) values (?, ?)`, key, value)
	}
	if opts.EmptyTables {
		return
	}
	for id, raw := range opts.Objects {
		MustExec(tb, db, `insert into game (key, value) values (?, ?)`, id[:], raw)
	}
}

func WriteArchive(tb testing.TB, path string, className string) {
	tb.Helper()
	WriteArchiveWithProperties(tb, path, className, nil)
}

func WriteArchiveWithProperties(tb testing.TB, path string, className string, properties []byte) {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	WriteArchivePrefix(&buf, id, className, []string{"Object_0"})
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	if len(properties) > 0 {
		propertiesOffset := int32(buf.Len() - 1)
		binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
		buf.Write(properties)
	} else {
		binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(128))
	}
	writeFile(tb, path, buf.Bytes(), "archive fixture")
}

func WritePlayerArchive(tb testing.TB, path string) {
	tb.Helper()
	WritePlayerArchiveWithOptions(tb, path, PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
		TribeID:       777,
	})
}

type PlayerArchiveOptions struct {
	PlayerDataID        int32
	CharacterName       string
	PlayerName          string
	UniqueID            string
	TribeID             int32
	NumDeaths           int32
	ExtraCharacterLevel int32
	ExperiencePoints    float32
	TotalEngramPoints   int32
}

func WritePlayerArchiveWithOptions(tb testing.TB, path string, opts PlayerArchiveOptions) {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var myData bytes.Buffer
	WriteNameIntProperty(&myData, "PlayerDataID", opts.PlayerDataID)
	WriteNameStringProperty(&myData, "PlayerCharacterName", opts.CharacterName)
	WriteNameStringProperty(&myData, "PlayerName", opts.PlayerName)
	if opts.UniqueID != "" {
		WriteNameStringProperty(&myData, "UniqueID", opts.UniqueID)
	}
	WriteNameIntProperty(&myData, "TribeID", opts.TribeID)
	if opts.NumDeaths != 0 {
		WriteNameIntProperty(&myData, "NumOfDeaths", opts.NumDeaths)
	}
	WriteArkString(&myData, "None")

	var buf bytes.Buffer
	WriteArchivePrefix(&buf, id, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C", []string{"PlayerData_0"})
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	WriteNameIntProperty(&buf, "SavedPlayerDataVersion", 17)
	if opts.ExtraCharacterLevel != 0 {
		WriteNameIntProperty(&buf, "CharacterStatusComponent_ExtraCharacterLevel", opts.ExtraCharacterLevel)
	}
	if opts.ExperiencePoints != 0 {
		WriteNameFloatProperty(&buf, "CharacterStatusComponent_ExperiencePoints", opts.ExperiencePoints)
	}
	if opts.TotalEngramPoints != 0 {
		WriteNameIntProperty(&buf, "PlayerState_TotalEngramPoints", opts.TotalEngramPoints)
	}
	WriteNameStructProperty(&buf, "MyData", "PlayerDataStruct", myData.Bytes())
	WriteArkString(&buf, "None")
	writeFile(tb, path, buf.Bytes(), "player archive fixture")
}

func WriteTribeArchive(tb testing.TB, path string) {
	tb.Helper()
	WriteTribeArchiveWithOptions(tb, path, TribeArchiveOptions{
		Name:     "Porters",
		TribeID:  12345,
		OwnerID:  42,
		NumDinos: 7,
	})
}

type TribeArchiveOptions struct {
	Name      string
	TribeID   int32
	OwnerID   int32
	NumDinos  int32
	Members   []string
	MemberIDs []int32
	TribeLog  []string
}

func WriteTribeArchiveWithOptions(tb testing.TB, path string, opts TribeArchiveOptions) {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var tribeData bytes.Buffer
	WriteNameStringProperty(&tribeData, "TribeName", opts.Name)
	WriteNameIntProperty(&tribeData, "TribeID", opts.TribeID)
	WriteNameIntProperty(&tribeData, "OwnerPlayerDataId", opts.OwnerID)
	WriteNameIntProperty(&tribeData, "NumTribeDinos", opts.NumDinos)
	if len(opts.Members) > 0 {
		WriteNameStringArrayProperty(&tribeData, "MembersPlayerName", opts.Members)
	}
	if len(opts.MemberIDs) > 0 {
		WriteNameIntArrayProperty(&tribeData, "MembersPlayerDataID", opts.MemberIDs)
	}
	if len(opts.TribeLog) > 0 {
		WriteNameStringArrayProperty(&tribeData, "TribeLog", opts.TribeLog)
	}
	WriteArkString(&tribeData, "None")

	var buf bytes.Buffer
	WriteArchivePrefix(&buf, id, "/Script/ShooterGame.PrimalTribeData", []string{"TribeData_0"})
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	WriteNameStructProperty(&buf, "TribeData", "TribeDataStruct", tribeData.Bytes())
	WriteArkString(&buf, "None")
	writeFile(tb, path, buf.Bytes(), "tribe archive fixture")
}

func WriteTributeFile(tb testing.TB, path string, playerIDs []uint64, tribeIDs []uint64) {
	tb.Helper()
	writeFile(tb, path, TributeBytes(tb, playerIDs, tribeIDs), "tribute fixture")
}

func WriteSparseFile(tb testing.TB, path string, size int64) {
	tb.Helper()
	file, err := os.Create(path)
	if err != nil {
		tb.Fatalf("create sparse file: %v", err)
	}
	if err := file.Truncate(size); err != nil {
		_ = file.Close()
		tb.Fatalf("truncate sparse file: %v", err)
	}
	if err := file.Close(); err != nil {
		tb.Fatalf("close sparse file: %v", err)
	}
}

func TributeBytes(tb testing.TB, playerIDs []uint64, tribeIDs []uint64) []byte {
	tb.Helper()
	var buf bytes.Buffer
	WriteIDList(tb, &buf, playerIDs)
	WriteIDList(tb, &buf, tribeIDs)
	return buf.Bytes()
}

func WriteIDList(tb testing.TB, buf *bytes.Buffer, ids []uint64) {
	tb.Helper()
	if err := binary.Write(buf, binary.LittleEndian, int32(len(ids))); err != nil {
		tb.Fatalf("write ID list count: %v", err)
	}
	for _, id := range ids {
		if err := binary.Write(buf, binary.LittleEndian, id); err != nil {
			tb.Fatalf("write ID list value: %v", err)
		}
	}
}

func Header(mapName string, names map[uint32]string) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int16(12))
	nameOffsetPosition := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, float64(1234.5))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(77))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	for buf.Len() < 30 {
		buf.WriteByte(0)
	}
	WriteArkString(&buf, mapName)
	nameOffset := int32(buf.Len())
	binary.LittleEndian.PutUint32(buf.Bytes()[nameOffsetPosition:nameOffsetPosition+4], uint32(nameOffset))
	_ = binary.Write(&buf, binary.LittleEndian, int32(len(names)))
	keys := make([]uint32, 0, len(names))
	for key := range names {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, key := range keys {
		_ = binary.Write(&buf, binary.LittleEndian, key)
		WriteArkString(&buf, names[key])
	}
	return buf.Bytes()
}

func GenericObjectBytes(classNameID uint32, noneNameID uint32) []byte {
	return ObjectBytesWithProperties(classNameID, noneNameID, nil)
}

func ObjectBytesWithProperties(classNameID uint32, noneNameID uint32, properties []byte) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, classNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	buf.Write(properties)
	_ = binary.Write(&buf, binary.LittleEndian, noneNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func WriteArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}

func WriteStringArray(buf *bytes.Buffer, values []string) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		WriteArkString(buf, value)
	}
}

func WriteInt32(buf *bytes.Buffer, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteUInt32(buf *bytes.Buffer, value uint32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameIntProperty(buf *bytes.Buffer, name string, value int32) {
	WriteArkString(buf, name)
	WriteArkString(buf, "IntProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteIntPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value int32) {
	writePropertyIDHeader(buf, name, propertyType, 4, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteFloatPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value float32) {
	writePropertyIDHeader(buf, name, propertyType, 4, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteDoublePropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value float64) {
	writePropertyIDHeader(buf, name, propertyType, 8, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteBoolPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value bool) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, propertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	if value {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
}

func WriteObjectReferencePropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, id uuid.UUID) {
	writePropertyIDHeader(buf, name, propertyType, 18, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, int16(0))
	buf.Write(id[:])
}

func WritePositionedIntPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, value int32) {
	writePropertyIDHeader(buf, name, propertyType, 4, position, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WritePositionedFloatPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, value float32) {
	writePropertyIDHeader(buf, name, propertyType, 4, position, true)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WritePositionedInt8PropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, value int8) {
	writePropertyIDHeader(buf, name, propertyType, 1, position, true)
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(byte(value))
}

func WritePositionedNamePropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, valueName uint32) {
	writePropertyIDHeader(buf, name, propertyType, 8, position, true)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, valueName)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
}

func writePropertyIDHeader(buf *bytes.Buffer, name uint32, propertyType uint32, size int32, position int32, positioned bool) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, propertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, size)
	_ = binary.Write(buf, binary.LittleEndian, position)
	if positioned {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
}

func WriteNameStringProperty(buf *bytes.Buffer, name string, value string) {
	WriteArkString(buf, name)
	WriteArkString(buf, "StrProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+5))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	WriteArkString(buf, value)
}

func WriteNameFloatProperty(buf *bytes.Buffer, name string, value float32) {
	WriteArkString(buf, name)
	WriteArkString(buf, "FloatProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameDoubleProperty(buf *bytes.Buffer, name string, value float64) {
	WriteArkString(buf, name)
	WriteArkString(buf, "DoubleProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameObjectPathProperty(buf *bytes.Buffer, name string, value string) {
	var body bytes.Buffer
	_ = binary.Write(&body, binary.LittleEndian, int32(1))
	WriteArkString(&body, value)

	WriteArkString(buf, name)
	WriteArkString(buf, "ObjectProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(body.Len()))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	buf.Write(body.Bytes())
}

func WriteNameByteArrayProperty(buf *bytes.Buffer, name string, values []byte) {
	WriteArkString(buf, name)
	WriteArkString(buf, "ArrayProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4+len(values)))
	WriteArkString(buf, "ByteProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	_ = binary.Write(buf, binary.LittleEndian, uint32(4+len(values)))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	buf.Write(values)
}

func WriteNameStringArrayProperty(buf *bytes.Buffer, name string, values []string) {
	bodySize := 4
	for _, value := range values {
		bodySize += 4 + len(value) + 1
	}
	writeNameArrayHeader(buf, name, "StrProperty", bodySize)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		WriteArkString(buf, value)
	}
}

func WriteNameIntArrayProperty(buf *bytes.Buffer, name string, values []int32) {
	bodySize := 4 + len(values)*4
	writeNameArrayHeader(buf, name, "IntProperty", bodySize)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		_ = binary.Write(buf, binary.LittleEndian, value)
	}
}

func writeNameArrayHeader(buf *bytes.Buffer, name string, elementType string, bodySize int) {
	WriteArkString(buf, name)
	WriteArkString(buf, "ArrayProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	WriteArkString(buf, elementType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	buf.WriteByte(0)
}

func WriteNameStructProperty(buf *bytes.Buffer, name string, structType string, body []byte) {
	WriteArkString(buf, name)
	WriteArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
}

func WriteNameStructArrayProperty(buf *bytes.Buffer, name string, structType string, elements [][]byte) {
	bodySize := 4
	for _, element := range elements {
		bodySize += len(element)
	}
	WriteArkString(buf, name)
	WriteArkString(buf, "ArrayProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	WriteArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(bodySize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(elements)))
	for _, element := range elements {
		buf.Write(element)
	}
}

func MustExec(tb testing.TB, db *sql.DB, query string, args ...any) {
	tb.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		tb.Fatalf("exec %q: %v", query, err)
	}
}

func WriteArchivePrefix(buf *bytes.Buffer, id uuid.UUID, className string, names []string) {
	WriteInt32(buf, 7)
	WriteInt32(buf, 0)
	WriteInt32(buf, 0)
	WriteInt32(buf, 1)
	buf.Write(id[:])
	WriteArkString(buf, className)
	WriteUInt32(buf, 0)
	WriteStringArray(buf, names)
	WriteUInt32(buf, 0)
	WriteInt32(buf, -1)
	WriteUInt32(buf, 0)
}

func writeFile(tb testing.TB, path string, data []byte, label string) {
	tb.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		tb.Fatalf("write %s: %v", label, err)
	}
}
