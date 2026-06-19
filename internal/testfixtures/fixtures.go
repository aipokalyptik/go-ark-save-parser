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
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int32(7))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(1))
	buf.Write(id[:])
	WriteArkString(&buf, className)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	WriteStringArray(&buf, []string{"Object_0"})
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(-1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(128))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		tb.Fatalf("write archive fixture: %v", err)
	}
}

func WritePlayerArchive(tb testing.TB, path string) {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var myData bytes.Buffer
	WriteNameIntProperty(&myData, "PlayerDataID", 42)
	WriteNameStringProperty(&myData, "PlayerCharacterName", "Survivor")
	WriteNameStringProperty(&myData, "PlayerName", "PlatformName")
	WriteNameIntProperty(&myData, "TribeID", 777)
	WriteArkString(&myData, "None")

	var buf bytes.Buffer
	writeArchivePrefix(&buf, id, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C", []string{"PlayerData_0"})
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	WriteNameIntProperty(&buf, "SavedPlayerDataVersion", 17)
	WriteNameStructProperty(&buf, "MyData", "PlayerDataStruct", myData.Bytes())
	WriteArkString(&buf, "None")
	writeFile(tb, path, buf.Bytes(), "player archive fixture")
}

func WriteTribeArchive(tb testing.TB, path string) {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var tribeData bytes.Buffer
	WriteNameStringProperty(&tribeData, "TribeName", "Porters")
	WriteNameIntProperty(&tribeData, "TribeID", 12345)
	WriteNameIntProperty(&tribeData, "OwnerPlayerDataId", 42)
	WriteNameIntProperty(&tribeData, "NumTribeDinos", 7)
	WriteArkString(&tribeData, "None")

	var buf bytes.Buffer
	writeArchivePrefix(&buf, id, "/Script/ShooterGame.PrimalTribeData", []string{"TribeData_0"})
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	WriteNameStructProperty(&buf, "TribeData", "TribeDataStruct", tribeData.Bytes())
	WriteArkString(&buf, "None")
	writeFile(tb, path, buf.Bytes(), "tribe archive fixture")
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
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, classNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
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

func WriteNameIntProperty(buf *bytes.Buffer, name string, value int32) {
	WriteArkString(buf, name)
	WriteArkString(buf, "IntProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameStringProperty(buf *bytes.Buffer, name string, value string) {
	WriteArkString(buf, name)
	WriteArkString(buf, "StrProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+5))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	WriteArkString(buf, value)
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

func MustExec(tb testing.TB, db *sql.DB, query string, args ...any) {
	tb.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		tb.Fatalf("exec %q: %v", query, err)
	}
}

func writeArchivePrefix(buf *bytes.Buffer, id uuid.UUID, className string, names []string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(7))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	buf.Write(id[:])
	WriteArkString(buf, className)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	WriteStringArray(buf, names)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(-1))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
}

func writeFile(tb testing.TB, path string, data []byte, label string) {
	tb.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		tb.Fatalf("write %s: %v", label, err)
	}
}
