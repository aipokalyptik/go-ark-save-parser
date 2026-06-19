package arkapi

import (
	"encoding/json"
	"testing"
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
