package arkapi

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkmutation"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

func TestBaseBinaryExportRoundTripsThroughMutationImport(t *testing.T) {
	save := openSyntheticBaseSave(t)
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	originalRows := originalObjectRows(t, save, firstID, secondID)
	inputPath := save.Path()

	dir := t.TempDir()
	exportDir := filepath.Join(dir, "base-export")
	if _, err := NewBase(save, "Valguero").ExportBinary(exportDir); err != nil {
		t.Fatalf("ExportBinary() error = %v", err)
	}
	if err := save.Close(); err != nil {
		t.Fatalf("Close(save) error = %v", err)
	}

	outputPath := filepath.Join(dir, "base-import.ark")
	imported, err := arkmutation.ImportBaseBinary(inputPath, outputPath, exportDir)
	if err != nil {
		t.Fatalf("ImportBaseBinary() error = %v", err)
	}
	if imported != len(originalRows) {
		t.Fatalf("ImportBaseBinary() imported = %d, want %d", imported, len(originalRows))
	}
	assertImportedRows(t, outputPath, originalRows)
}

func TestStructureBinaryExportRoundTripsThroughMutationImport(t *testing.T) {
	save := openSyntheticStructureSave(t)
	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	originalRows := originalObjectRows(t, save, structureID)
	inputPath := save.Path()

	dir := t.TempDir()
	exportDir := filepath.Join(dir, "structure-export")
	if _, err := NewStructure(save).ExportBinary(exportDir); err != nil {
		t.Fatalf("ExportBinary() error = %v", err)
	}
	if err := save.Close(); err != nil {
		t.Fatalf("Close(save) error = %v", err)
	}

	outputPath := filepath.Join(dir, "structure-import.ark")
	imported, err := arkmutation.ImportStructureBinary(inputPath, outputPath, exportDir)
	if err != nil {
		t.Fatalf("ImportStructureBinary() error = %v", err)
	}
	if imported != len(originalRows) {
		t.Fatalf("ImportStructureBinary() imported = %d, want %d", imported, len(originalRows))
	}
	assertImportedRows(t, outputPath, originalRows)
}

func TestDinoBinaryExportRoundTripsThroughMutationImport(t *testing.T) {
	save := openSyntheticDinoStatsSave(t)
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	statusID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	originalRows := originalObjectRows(t, save, dinoID, statusID)
	inputPath := save.Path()

	dir := t.TempDir()
	exportDir := filepath.Join(dir, "dino-export")
	if _, err := NewDino(save).ExportBinary(exportDir); err != nil {
		t.Fatalf("ExportBinary() error = %v", err)
	}
	if err := save.Close(); err != nil {
		t.Fatalf("Close(save) error = %v", err)
	}

	outputPath := filepath.Join(dir, "dino-import.ark")
	imported, err := arkmutation.ImportDinoBinary(inputPath, outputPath, exportDir)
	if err != nil {
		t.Fatalf("ImportDinoBinary() error = %v", err)
	}
	if imported != len(originalRows) {
		t.Fatalf("ImportDinoBinary() imported = %d, want %d", imported, len(originalRows))
	}
	assertImportedRows(t, outputPath, originalRows)
}

func TestEquipmentBinaryExportRoundTripsThroughMutationImport(t *testing.T) {
	save := openSyntheticEquipmentSave(t)
	itemID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	originalRows := originalObjectRows(t, save, itemID)
	inputPath := save.Path()

	dir := t.TempDir()
	exportDir := filepath.Join(dir, "equipment-export")
	if _, err := NewEquipment(save).ExportBinary(exportDir); err != nil {
		t.Fatalf("ExportBinary() error = %v", err)
	}
	if err := save.Close(); err != nil {
		t.Fatalf("Close(save) error = %v", err)
	}

	outputPath := filepath.Join(dir, "equipment-import.ark")
	imported, err := arkmutation.ImportEquipmentBinary(inputPath, outputPath, exportDir)
	if err != nil {
		t.Fatalf("ImportEquipmentBinary() error = %v", err)
	}
	if imported != len(originalRows) {
		t.Fatalf("ImportEquipmentBinary() imported = %d, want %d", imported, len(originalRows))
	}
	assertImportedRows(t, outputPath, originalRows)
}

func originalObjectRows(t *testing.T, save *arksave.Save, ids ...uuid.UUID) map[uuid.UUID][]byte {
	t.Helper()
	rows := make(map[uuid.UUID][]byte, len(ids))
	for _, id := range ids {
		raw, err := save.ObjectBinary(id)
		if err != nil {
			t.Fatalf("ObjectBinary(%s) error = %v", id, err)
		}
		rows[id] = raw
	}
	return rows
}

func assertImportedRows(t *testing.T, outputPath string, rows map[uuid.UUID][]byte) {
	t.Helper()
	mutated, err := arksave.Open(outputPath)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", outputPath, err)
	}
	defer mutated.Close()

	for id, want := range rows {
		got, err := mutated.ObjectBinary(id)
		if err != nil {
			t.Fatalf("ObjectBinary(%s) from imported copy error = %v", id, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("ObjectBinary(%s) from imported copy differs from exported source row", id)
		}
	}
}
