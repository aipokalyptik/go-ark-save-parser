package arkmutation

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
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

func TestMutationFailureRemovesOutputCopy(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	if err := os.WriteFile(input, []byte("not sqlite"), 0o600); err != nil {
		t.Fatalf("write invalid input: %v", err)
	}

	if err := PutCustomValue(input, output, "Extra", []byte{1, 2, 3}); err == nil {
		t.Fatalf("PutCustomValue(invalid input) error = nil, want error")
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("mutated output stat error = %v, want os.ErrNotExist", err)
	}
}

func createSyntheticSave(t *testing.T, path string, objectID uuid.UUID, objectBytes []byte) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header:  testfixtures.Header("Valguero_WP", map[uint32]string{0x10000000: "None"}),
		Objects: map[uuid.UUID][]byte{objectID: objectBytes},
	})
}
