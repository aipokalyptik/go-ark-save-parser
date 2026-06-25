package arkmutation

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
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

func TestRemoveObjectsByClassContainsWritesCopyAndReopens(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	removeID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	keepID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	createSyntheticSaveWithObjects(t, input, map[uuid.UUID][]byte{
		removeID: testfixtures.GenericObjectBytes(0x10000001, 0x10000003),
		keepID:   testfixtures.GenericObjectBytes(0x10000002, 0x10000003),
	}, map[uint32]string{
		0x10000001: "Blueprint'/Game/Test/SuperSpyglass.SuperSpyglass_C'",
		0x10000002: "Blueprint'/Game/Test/StorageBox.StorageBox_C'",
		0x10000003: "None",
	})

	removed, err := RemoveObjectsByClassContains(input, output, "SuperSpyglass")
	if err != nil {
		t.Fatalf("RemoveObjectsByClassContains() error = %v", err)
	}
	if removed != 1 {
		t.Fatalf("RemoveObjectsByClassContains() removed = %d, want 1", removed)
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
	if len(ids) != 1 || ids[0] != keepID {
		t.Fatalf("mutated ObjectIDs = %v, want only %s", ids, keepID)
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
	if len(ids) != 2 {
		t.Fatalf("input ObjectIDs length = %d, want original two objects", len(ids))
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

func TestImportBaseBinaryWritesCopyAndReopens(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	exportDir := filepath.Join(dir, "base-export", "base_aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	objectID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	createSyntheticSave(t, input, uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"), testfixtures.GenericObjectBytes(0x10000001, 0x10000003))
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	want := testfixtures.GenericObjectBytes(0x10000002, 0x10000003)
	if err := os.WriteFile(filepath.Join(exportDir, "str_"+objectID.String()+".bin"), want, 0o600); err != nil {
		t.Fatalf("write base export row: %v", err)
	}

	imported, err := ImportBaseBinary(input, output, filepath.Join(dir, "base-export"))
	if err != nil {
		t.Fatalf("ImportBaseBinary() error = %v", err)
	}
	if imported != 1 {
		t.Fatalf("ImportBaseBinary() imported = %d, want 1", imported)
	}
	mutated, err := arksave.Open(output)
	if err != nil {
		t.Fatalf("Open(output) error = %v", err)
	}
	got, err := mutated.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(imported) error = %v", err)
	}
	_ = mutated.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("imported ObjectBinary = % x, want % x", got, want)
	}

	if _, err := os.Stat(input); err != nil {
		t.Fatalf("input save missing after import: %v", err)
	}
}

func TestImportStructureBinaryWritesCopyAndReopens(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	exportDir := filepath.Join(dir, "structure-export")
	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	createSyntheticSave(t, input, uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"), testfixtures.GenericObjectBytes(0x10000001, 0x10000003))
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(exportDir, "manifest.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write structure export manifest: %v", err)
	}
	want := testfixtures.GenericObjectBytes(0x10000002, 0x10000003)
	if err := os.WriteFile(filepath.Join(exportDir, "str_"+structureID.String()+".bin"), want, 0o600); err != nil {
		t.Fatalf("write structure export row: %v", err)
	}

	imported, err := ImportStructureBinary(input, output, exportDir)
	if err != nil {
		t.Fatalf("ImportStructureBinary() error = %v", err)
	}
	if imported != 1 {
		t.Fatalf("ImportStructureBinary() imported = %d, want 1", imported)
	}
	mutated, err := arksave.Open(output)
	if err != nil {
		t.Fatalf("Open(output) error = %v", err)
	}
	got, err := mutated.ObjectBinary(structureID)
	if err != nil {
		t.Fatalf("ObjectBinary(imported) error = %v", err)
	}
	_ = mutated.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("imported ObjectBinary = % x, want % x", got, want)
	}
}

func TestImportStructureBinaryRequiresManifestAtExportRoot(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	exportDir := filepath.Join(dir, "structure-export")
	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	createSyntheticSave(t, input, uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"), testfixtures.GenericObjectBytes(0x10000001, 0x10000003))
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(exportDir, "str_"+structureID.String()+".bin"), testfixtures.GenericObjectBytes(0x10000002, 0x10000003), 0o600); err != nil {
		t.Fatalf("write structure export row: %v", err)
	}

	if _, err := ImportStructureBinary(input, output, exportDir); err == nil || !strings.Contains(err.Error(), "manifest.json") {
		t.Fatalf("ImportStructureBinary(missing manifest) error = %v, want manifest error", err)
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("output stat error = %v, want os.ErrNotExist", err)
	}
}

func TestImportDinoBinaryWritesCopyAndReopens(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	exportDir := filepath.Join(dir, "dino-export", "dino_aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	statusID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	createSyntheticSave(t, input, uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"), testfixtures.GenericObjectBytes(0x10000001, 0x10000003))
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	dinoBytes := testfixtures.GenericObjectBytes(0x10000002, 0x10000003)
	statusBytes := testfixtures.GenericObjectBytes(0x10000004, 0x10000003)
	if err := os.WriteFile(filepath.Join(exportDir, "dino_"+dinoID.String()+".bin"), dinoBytes, 0o600); err != nil {
		t.Fatalf("write dino export row: %v", err)
	}
	if err := os.WriteFile(filepath.Join(exportDir, "status_"+statusID.String()+".bin"), statusBytes, 0o600); err != nil {
		t.Fatalf("write status export row: %v", err)
	}

	imported, err := ImportDinoBinary(input, output, filepath.Join(dir, "dino-export"))
	if err != nil {
		t.Fatalf("ImportDinoBinary() error = %v", err)
	}
	if imported != 2 {
		t.Fatalf("ImportDinoBinary() imported = %d, want 2", imported)
	}
	mutated, err := arksave.Open(output)
	if err != nil {
		t.Fatalf("Open(output) error = %v", err)
	}
	for id, want := range map[uuid.UUID][]byte{dinoID: dinoBytes, statusID: statusBytes} {
		got, err := mutated.ObjectBinary(id)
		if err != nil {
			t.Fatalf("ObjectBinary(%s) error = %v", id, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("imported ObjectBinary(%s) = % x, want % x", id, got, want)
		}
	}
	_ = mutated.Close()
}

func TestImportEquipmentBinaryWritesCopyAndReopens(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "input.ark")
	output := filepath.Join(dir, "output.ark")
	exportDir := filepath.Join(dir, "equipment-export")
	itemID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	createSyntheticSave(t, input, uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"), testfixtures.GenericObjectBytes(0x10000001, 0x10000003))
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	want := testfixtures.GenericObjectBytes(0x10000002, 0x10000003)
	if err := os.WriteFile(filepath.Join(exportDir, "item_"+itemID.String()+".bin"), want, 0o600); err != nil {
		t.Fatalf("write equipment export row: %v", err)
	}

	imported, err := ImportEquipmentBinary(input, output, exportDir)
	if err != nil {
		t.Fatalf("ImportEquipmentBinary() error = %v", err)
	}
	if imported != 1 {
		t.Fatalf("ImportEquipmentBinary() imported = %d, want 1", imported)
	}
	mutated, err := arksave.Open(output)
	if err != nil {
		t.Fatalf("Open(output) error = %v", err)
	}
	got, err := mutated.ObjectBinary(itemID)
	if err != nil {
		t.Fatalf("ObjectBinary(imported) error = %v", err)
	}
	_ = mutated.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("imported ObjectBinary = % x, want % x", got, want)
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
	createSyntheticSaveWithObjects(t, path, map[uuid.UUID][]byte{objectID: objectBytes}, map[uint32]string{0x10000000: "None"})
}

func createSyntheticSaveWithObjects(t *testing.T, path string, objects map[uuid.UUID][]byte, names map[uint32]string) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header:  testfixtures.Header("Valguero_WP", names),
		Objects: objects,
	})
}
