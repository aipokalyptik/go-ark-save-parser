package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestInspectCommandPrintsOfflineSaveSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)

	var out bytes.Buffer
	err := run([]string{"inspect", path}, &out)
	if err != nil {
		t.Fatalf("run(inspect) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Map: Valguero_WP",
		"Save version: 12",
		"Objects: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("inspect output %q does not contain %q", got, want)
		}
	}
}

func TestRunRejectsNetworkCommands(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"rcon"}, &out)
	if err == nil {
		t.Fatalf("run(rcon) error = nil, want unsupported command")
	}
	if !strings.Contains(err.Error(), "offline-only") {
		t.Fatalf("run(rcon) error = %v, want offline-only message", err)
	}
}

func TestPlayersCommandPrintsLocalProfileSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	createSyntheticArchive(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

	var out bytes.Buffer
	err := run([]string{"players", path}, &out)
	if err != nil {
		t.Fatalf("run(players) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player profile:",
		"Archive version: 7",
		"Objects: 1",
		"/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("players output %q does not contain %q", got, want)
		}
	}
}

func TestTribesCommandPrintsLocalTribeSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	createSyntheticTribeArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"tribes", path}, &out)
	if err != nil {
		t.Fatalf("run(tribes) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save:",
		"Archive version: 7",
		"Objects: 1",
		"/Script/ShooterGame.PrimalTribeData",
		"Tribe name: Porters",
		"Tribe ID: 12345",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes output %q does not contain %q", got, want)
		}
	}
}

func TestTribesCommandKeepsMetadataWhenSummaryMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	createSyntheticArchive(t, path, "/Script/ShooterGame.PrimalTribeData")

	var out bytes.Buffer
	err := run([]string{"tribes", path}, &out)
	if err != nil {
		t.Fatalf("run(tribes) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save:",
		"Archive version: 7",
		"Objects: 1",
		"/Script/ShooterGame.PrimalTribeData",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "Tribe name:") {
		t.Fatalf("tribes output %q includes summary despite missing TribeData", got)
	}
}

func TestExportJSONWritesSaveInfoToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "save_info.json")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"export-json", savePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-json output %q does not mention %q", out.String(), outPath)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(exported json) error = %v", err)
	}
	var decoded struct {
		MapName     string `json:"map_name"`
		SaveVersion int16  `json:"save_version"`
		ObjectCount int    `json:"object_count"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, data)
	}
	if decoded.MapName != "Valguero_WP" || decoded.SaveVersion != 12 || decoded.ObjectCount != 1 {
		t.Fatalf("exported json = %#v", decoded)
	}
}

func createSyntheticSave(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	defer db.Close()
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}, []byte{1, 0, 0, 0, 0, 0, 0, 0})
}

func createSyntheticArchive(t *testing.T, path string, className string) {
	t.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int32(7))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(1))
	buf.Write(id[:])
	writeArkString(&buf, className)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	writeStringArray(&buf, []string{"Object_0"})
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(-1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(128))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write archive fixture: %v", err)
	}
}

func createSyntheticTribeArchive(t *testing.T, path string) {
	t.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	var tribeData bytes.Buffer
	writeNameStringProperty(&tribeData, "TribeName", "Porters")
	writeNameIntProperty(&tribeData, "TribeID", 12345)
	writeArkString(&tribeData, "None")

	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int32(7))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(1))
	buf.Write(id[:])
	writeArkString(&buf, "/Script/ShooterGame.PrimalTribeData")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	writeStringArray(&buf, []string{"TribeData_0"})
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(-1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	writeNameStructProperty(&buf, "TribeData", "TribeDataStruct", tribeData.Bytes())
	writeArkString(&buf, "None")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatalf("write tribe archive fixture: %v", err)
	}
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}

func syntheticHeader() []byte {
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
	writeArkString(&buf, "Valguero_WP")
	nameOffset := int32(buf.Len())
	binary.LittleEndian.PutUint32(buf.Bytes()[nameOffsetPosition:nameOffsetPosition+4], uint32(nameOffset))
	_ = binary.Write(&buf, binary.LittleEndian, int32(1))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(1))
	writeArkString(&buf, "Blueprint'/Game/Test.Test_C'")
	return buf.Bytes()
}

func writeArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}

func writeStringArray(buf *bytes.Buffer, values []string) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		writeArkString(buf, value)
	}
}

func writeNameIntProperty(buf *bytes.Buffer, name string, value int32) {
	writeArkString(buf, name)
	writeArkString(buf, "IntProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeNameStringProperty(buf *bytes.Buffer, name string, value string) {
	writeArkString(buf, name)
	writeArkString(buf, "StrProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+5))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	writeArkString(buf, value)
}

func writeNameStructProperty(buf *bytes.Buffer, name string, structType string, body []byte) {
	writeArkString(buf, name)
	writeArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	writeArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	writeArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
}
