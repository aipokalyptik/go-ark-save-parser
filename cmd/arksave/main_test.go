package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"path/filepath"
	"strings"
	"testing"

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
