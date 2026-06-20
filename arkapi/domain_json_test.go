package arkapi

import (
	"encoding/json"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

func TestJSONAPIExportDinosSummarizesDinoAPI(t *testing.T) {
	save := openSyntheticDinoSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportDinos()
	if err != nil {
		t.Fatalf("ExportDinos() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportDinos() length = %d, want 1", len(items))
	}
	if items[0].ID1 != 1001 || !items[0].IsTamed || items[0].Location == nil || items[0].Location.X != 11 {
		t.Fatalf("DinoInfo = %#v", items[0])
	}
}

func TestJSONAPIExportDinosIncludesTamedAndBabyDetails(t *testing.T) {
	save := openSyntheticDinoDetailSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportDinos()
	if err != nil {
		t.Fatalf("ExportDinos() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportDinos() length = %d, want 1", len(items))
	}
	if items[0].TamedName != "Blue" || !items[0].IsNeutered {
		t.Fatalf("DinoInfo tamed fields = %#v", items[0])
	}
	if items[0].InventoryUUID != "99999999-aaaa-bbbb-cccc-ddddeeeeffff" {
		t.Fatalf("DinoInfo inventory UUID = %q", items[0].InventoryUUID)
	}
	if items[0].Owner.TribeName != "Porters" || items[0].Owner.TamerTribeID != 555 || items[0].Owner.PlayerID != 42 {
		t.Fatalf("DinoInfo owner = %#v", items[0].Owner)
	}
	if items[0].UploadedFromServerName != "TheIsland" {
		t.Fatalf("DinoInfo uploaded server = %q", items[0].UploadedFromServerName)
	}
	if len(items[0].ColorSetIndices) != 6 || items[0].ColorSetIndices[0] != 11 || items[0].ColorSetIndices[3] != 44 {
		t.Fatalf("DinoInfo color indices = %#v", items[0].ColorSetIndices)
	}
	if len(items[0].ColorSetNames) != 6 || items[0].ColorSetNames[1] != "Blue" || items[0].ColorSetNames[4] != "Black" {
		t.Fatalf("DinoInfo color names = %#v", items[0].ColorSetNames)
	}

	babySave := openSyntheticDinoBabyStageSave(t)
	defer babySave.Close()
	babies, err := NewJSON(babySave).ExportDinos()
	if err != nil {
		t.Fatalf("ExportDinos(baby) error = %v", err)
	}
	if len(babies) != 1 || babies[0].MaturationPercent != 75 || babies[0].BabyStage != arkobject.BabyStageAdolescent {
		t.Fatalf("DinoInfo baby fields = %#v", babies)
	}
}

func TestJSONAPIExportStructuresSummarizesStructureAPI(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportStructures()
	if err != nil {
		t.Fatalf("ExportStructures() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportStructures() length = %d, want 1", len(items))
	}
	if items[0].ID != 123 || items[0].Owner.TribeID != 555 || items[0].Location == nil || items[0].Location.X != 11 {
		t.Fatalf("StructureInfo = %#v", items[0])
	}
}

func TestJSONAPIExportStructuresIncludesLinkedStructureMetadata(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportStructures()
	if err != nil {
		t.Fatalf("ExportStructures() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ExportStructures() length = %d, want 2", len(items))
	}
	first := items[0]
	if first.UUID != "aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff" || len(first.LinkedStructureUUIDs) != 1 {
		t.Fatalf("StructureInfo = %#v", first)
	}
	if first.LinkedStructureUUIDs[0] != "bbbbbbbb-cccc-dddd-eeee-ffffffffffff" {
		t.Fatalf("LinkedStructureUUIDs = %#v", first.LinkedStructureUUIDs)
	}
}

func TestJSONAPIExportEquipmentSummarizesEquipmentAPI(t *testing.T) {
	save := openSyntheticEquipmentSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportEquipment()
	if err != nil {
		t.Fatalf("ExportEquipment() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportEquipment() length = %d, want 1", len(items))
	}
	if items[0].Kind != "weapon" || items[0].Rating != 7.5 || items[0].Crafter == nil || items[0].Crafter.TribeName != "Porters" {
		t.Fatalf("EquipmentInfo = %#v", items[0])
	}
}

func TestJSONAPIExportStackablesSummarizesStackableAPI(t *testing.T) {
	save := openSyntheticStackableSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportStackables()
	if err != nil {
		t.Fatalf("ExportStackables() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportStackables() length = %d, want 1", len(items))
	}
	if items[0].Quantity != 100 {
		t.Fatalf("StackableInfo = %#v", items[0])
	}
}

func TestJSONAPIExportBasesSummarizesBaseAPI(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportBases()
	if err != nil {
		t.Fatalf("ExportBases() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportBases() length = %d, want 1", len(items))
	}
	if items[0].StructureCount != 2 || items[0].Owner.TribeID != 555 || items[0].AverageLocation == nil {
		t.Fatalf("BaseInfo = %#v", items[0])
	}
	if len(items[0].StructureUUIDs) != 2 || items[0].StructureUUIDs[0] != "aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff" {
		t.Fatalf("StructureUUIDs = %#v", items[0].StructureUUIDs)
	}
}

func TestJSONAPIExportDomainJSONIsDeterministic(t *testing.T) {
	save := openSyntheticStackableSave(t)
	defer save.Close()

	raw, err := NewJSON(save).ExportDomainJSON("stackables")
	if err != nil {
		t.Fatalf("ExportDomainJSON(stackables) error = %v", err)
	}
	var decoded DomainExport
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.Domain != "stackables" || decoded.Count != 1 {
		t.Fatalf("DomainExport = %#v", decoded)
	}
	if _, err := NewJSON(save).ExportDomainJSON("unknown"); err == nil {
		t.Fatalf("ExportDomainJSON(unknown) error = nil, want unsupported domain")
	}
}
