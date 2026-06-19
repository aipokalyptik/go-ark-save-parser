package arkmutation

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestRemoveObjectWritesCopyAndReopens(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	createSyntheticSave(t, input, objectID, []byte{1, 2, 3})

	if err := RemoveObject(input, output, objectID); err != nil {
		t.Fatalf("RemoveObject() error = %v", err)
	}
	mutated, err := arksave.Open(output)
	if err != nil {
		t.Fatalf("Open(output) error = %v", err)
	}
	ids, err := mutated.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs(output) error = %v", err)
	}
	_ = mutated.Close()
	if len(ids) != 0 {
		t.Fatalf("mutated ObjectIDs length = %d, want 0", len(ids))
	}

	original, err := arksave.Open(input)
	if err != nil {
		t.Fatalf("Open(input) error = %v", err)
	}
	ids, err = original.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs(input) error = %v", err)
	}
	_ = original.Close()
	if len(ids) != 1 || ids[0] != objectID {
		t.Fatalf("input ObjectIDs = %v, want original object", ids)
	}
}

func TestPutCustomValueWritesCopyAndReopens(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	createSyntheticSave(t, input, objectID, []byte{1, 2, 3})

	want := []byte{9, 8, 7}
	if err := PutCustomValue(input, output, "Extra", want); err != nil {
		t.Fatalf("PutCustomValue() error = %v", err)
	}
	mutated, err := arksave.Open(output)
	if err != nil {
		t.Fatalf("Open(output) error = %v", err)
	}
	got, err := mutated.CustomValue("Extra")
	if err != nil {
		t.Fatalf("CustomValue(Extra) error = %v", err)
	}
	_ = mutated.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("CustomValue(Extra) = % x, want % x", got, want)
	}
}

func TestCopySaveRequiresDistinctNewOutputPath(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	createSyntheticSave(t, input, objectID, []byte{1, 2, 3})

	if err := CopySave(input, ""); err == nil {
		t.Fatalf("CopySave(empty output) error = nil, want error")
	}
	if err := CopySave(input, input); err == nil {
		t.Fatalf("CopySave(input, input) error = nil, want error")
	}
	if err := os.WriteFile(output, []byte("exists"), 0o600); err != nil {
		t.Fatalf("write existing output: %v", err)
	}
	if err := CopySave(input, output); !errors.Is(err, ErrOutputExists) {
		t.Fatalf("CopySave(existing output) error = %v, want ErrOutputExists", err)
	}
}

func createSyntheticSave(t *testing.T, path string, objectID uuid.UUID, objectBytes []byte) {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	defer db.Close()
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, objectID[:], objectBytes)
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
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000000))
	writeArkString(&buf, "None")
	return buf.Bytes()
}

func writeArkString(buf *bytes.Buffer, s string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(s)+1))
	buf.WriteString(s)
	buf.WriteByte(0)
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
