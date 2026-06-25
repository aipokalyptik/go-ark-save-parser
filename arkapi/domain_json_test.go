package arkapi

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"math"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
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

func TestJSONAPIExportDinosSkipsMalformedCryopodPayloads(t *testing.T) {
	save := openSyntheticMalformedCryopodSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportDinos()
	if err != nil {
		t.Fatalf("ExportDinos() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("ExportDinos() length = %d, want 0 valid dinos", len(items))
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
	if items[0].Generation != 1 {
		t.Fatalf("DinoInfo generation = %d, want 1", items[0].Generation)
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
	statsSave := openSyntheticDinoStatsSave(t)
	defer statsSave.Close()
	statsItems, err := NewJSON(statsSave).ExportDinos()
	if err != nil {
		t.Fatalf("ExportDinos(stats) error = %v", err)
	}
	if len(statsItems) != 1 || statsItems[0].Stats == nil {
		t.Fatalf("DinoInfo stats = %#v", statsItems)
	}
	if statsItems[0].Stats.BaseLevel != 12 || statsItems[0].Stats.CurrentLevel != 12 {
		t.Fatalf("DinoInfo stats levels = %#v", statsItems[0].Stats)
	}
	if statsItems[0].Stats.BaseStatPoints.Health != 5 || statsItems[0].Stats.AddedStatPoints.MeleeDamage != 2 {
		t.Fatalf("DinoInfo stats points = %#v", statsItems[0].Stats)
	}
	if statsItems[0].Stats.ImprintingPercent != 87.5 {
		t.Fatalf("DinoInfo imprinting = %f", statsItems[0].Stats.ImprintingPercent)
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

func TestJSONAPIExportDinosIncludesPedigreeUUIDs(t *testing.T) {
	save := openSyntheticDinoPedigreeSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportDinos()
	if err != nil {
		t.Fatalf("ExportDinos() error = %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("ExportDinos() length = %d, want 3: %#v", len(items), items)
	}
	byUUID := map[string]DinoInfo{}
	for _, item := range items {
		byUUID[item.UUID] = item
	}
	parentID := "aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"
	childID := "bbbbbbbb-cccc-dddd-eeee-ffffffffffff"
	grandchildID := "cccccccc-dddd-eeee-ffff-000000000000"

	parent := byUUID[parentID]
	if len(parent.ChildUUIDs) != 1 || parent.ChildUUIDs[0] != childID {
		t.Fatalf("parent ChildUUIDs = %#v, want [%s]", parent.ChildUUIDs, childID)
	}
	if len(parent.DescendantUUIDs) != 2 || parent.DescendantUUIDs[0] != childID || parent.DescendantUUIDs[1] != grandchildID {
		t.Fatalf("parent DescendantUUIDs = %#v, want child and grandchild", parent.DescendantUUIDs)
	}
	child := byUUID[childID]
	if len(child.ChildUUIDs) != 1 || child.ChildUUIDs[0] != grandchildID {
		t.Fatalf("child ChildUUIDs = %#v, want [%s]", child.ChildUUIDs, grandchildID)
	}
	if len(child.DescendantUUIDs) != 1 || child.DescendantUUIDs[0] != grandchildID {
		t.Fatalf("child DescendantUUIDs = %#v, want [%s]", child.DescendantUUIDs, grandchildID)
	}
}

func TestGeneTraitInfosMapsParsedGeneTraits(t *testing.T) {
	items := geneTraitInfos([]arkobject.GeneTrait{
		{Raw: "MutableMelee[2]", Name: "MutableMelee", Level: 2},
		{Raw: "Robust", Name: "Robust"},
	})

	if len(items) != 2 || items[0].Name != "MutableMelee" || items[0].Level != 2 {
		t.Fatalf("GeneTraitInfo = %#v", items)
	}
	if items[1].Raw != "Robust" || items[1].Name != "Robust" || items[1].Level != 0 {
		t.Fatalf("GeneTraitInfo fallback = %#v", items[1])
	}
}

func TestDinoIDInfosMapsAncestorIDs(t *testing.T) {
	items := dinoIDInfos([]arkobject.DinoID{
		{ID1: 11, ID2: 12},
		{ID1: 21, ID2: 22},
	})

	if len(items) != 2 || items[0].ID1 != 11 || items[0].ID2 != 12 {
		t.Fatalf("DinoIDInfo = %#v", items)
	}
	if items[1].ID1 != 21 || items[1].ID2 != 22 {
		t.Fatalf("DinoIDInfo second = %#v", items[1])
	}
}

func openSyntheticDinoPedigreeSave(t *testing.T) *arksave.Save {
	t.Helper()

	parentID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	childID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	grandchildID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	path := filepath.Join(t.TempDir(), "pedigree.ark")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			0x10000015: "DinoID1",
			0x10000016: "DinoID2",
			0x10000018: "TamedTimeStamp",
			0x10000019: "DoubleProperty",
			0x1000001e: "ArrayProperty",
			0x10000049: "StructProperty",
			0x10000052: "DinoAncestor",
			0x10000053: "FemaleDinoID1",
			0x10000054: "FemaleDinoID2",
			0x10000055: "DinoAncestors",
		}),
		Objects: map[uuid.UUID][]byte{
			parentID:     pedigreeDinoObjectBytes(11, 12),
			childID:      pedigreeDinoObjectBytes(21, 22, arkobject.DinoID{ID1: 11, ID2: 12}),
			grandchildID: pedigreeDinoObjectBytes(31, 32, arkobject.DinoID{ID1: 21, ID2: 22}),
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open(pedigree) error = %v", err)
	}
	return save
}

func pedigreeDinoObjectBytes(id1 int32, id2 int32, ancestors ...arkobject.DinoID) []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x10000015, id1)
	writeIntProperty(&props, 0x10000016, id2)
	writeDoubleProperty(&props, 0x10000018, 42)
	if len(ancestors) > 0 {
		elements := make([][]byte, 0, len(ancestors))
		for _, ancestor := range ancestors {
			var element bytes.Buffer
			writeIntProperty(&element, 0x10000053, int32(ancestor.ID1))
			writeIntProperty(&element, 0x10000054, int32(ancestor.ID2))
			_ = binary.Write(&element, binary.LittleEndian, uint32(0x10000004))
			_ = binary.Write(&element, binary.LittleEndian, int32(0))
			elements = append(elements, element.Bytes())
		}
		writeStructArrayProperty(&props, 0x10000055, 0x10000049, 0x10000052, elements)
	}
	return testfixtures.ObjectBytesWithProperties(0x10000014, 0x10000004, props.Bytes())
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

func TestJSONAPIExportStructuresSkipsFaultyMatchingRows(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()

	items, err := NewJSON(save).ExportStructures()
	if err != nil {
		t.Fatalf("ExportStructures() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportStructures() length = %d, want 1", len(items))
	}
	if items[0].UUID != "aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff" || items[0].ID != 123 {
		t.Fatalf("StructureInfo = %#v", items[0])
	}
}

func TestJSONAPIExportDomainReportsFaultCount(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()

	export, err := NewJSON(save).ExportDomain("structures")
	if err != nil {
		t.Fatalf("ExportDomain(structures) error = %v", err)
	}
	if export.Domain != "structures" || export.Count != 1 || export.FaultCount != 1 {
		t.Fatalf("DomainExport = %#v, want one structure and one fault", export)
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

func TestJSONAPIExportStructuresIncludesInventoryMetadata(t *testing.T) {
	save := openSyntheticStructureWithInventorySave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportStructures()
	if err != nil {
		t.Fatalf("ExportStructures() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportStructures() length = %d, want 1", len(items))
	}
	item := items[0]
	if item.InventoryUUID != "99999999-aaaa-bbbb-cccc-ddddeeeeffff" {
		t.Fatalf("StructureInfo inventory UUID = %q", item.InventoryUUID)
	}
	if item.ItemCount != 12 || item.MaxItemCount != 300 || item.OpenSlots != 288 || item.IsEmpty {
		t.Fatalf("StructureInfo inventory counts = current %d max %d open %d empty %v", item.ItemCount, item.MaxItemCount, item.OpenSlots, item.IsEmpty)
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
	if !items[0].IsCrafted || items[0].AverageStat != 1117 {
		t.Fatalf("EquipmentInfo ranking fields = crafted %v average %f, want crafted true average 1117", items[0].IsCrafted, items[0].AverageStat)
	}
	if len(items[0].ImplementedStats) != 2 || items[0].ImplementedStats[0] != "durability" || items[0].ImplementedStats[1] != "damage" {
		t.Fatalf("EquipmentInfo implemented stats = %#v, want durability and damage", items[0].ImplementedStats)
	}
	if items[0].Stats == nil || items[0].Stats.Damage != 112.3 || items[0].Stats.Durability != 62.5 {
		t.Fatalf("EquipmentInfo stats = %#v", items[0].Stats)
	}

	armorSave := openSyntheticArmorEquipmentSave(t)
	defer armorSave.Close()
	armorItems, err := NewJSON(armorSave).ExportEquipment()
	if err != nil {
		t.Fatalf("ExportEquipment(armor) error = %v", err)
	}
	if len(armorItems) != 1 || armorItems[0].Stats == nil {
		t.Fatalf("EquipmentInfo armor stats missing = %#v", armorItems)
	}
	if armorItems[0].Stats.Armor != 12 || armorItems[0].Stats.HypothermalResistance != 8.8 || armorItems[0].Stats.HyperthermalResistance != 15.6 {
		t.Fatalf("EquipmentInfo armor stats = %#v", armorItems[0].Stats)
	}
	if armorItems[0].AverageStat != 425 {
		t.Fatalf("EquipmentInfo armor average stat = %f, want 425", armorItems[0].AverageStat)
	}
	if len(armorItems[0].ImplementedStats) != 4 || armorItems[0].ImplementedStats[2] != "hypothermal_resistance" {
		t.Fatalf("EquipmentInfo armor implemented stats = %#v", armorItems[0].ImplementedStats)
	}
}

func TestJSONAPIExportEquipmentIncludesOwnerInventory(t *testing.T) {
	save := openSyntheticEquipmentOwnedByStructureSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportEquipment()
	if err != nil {
		t.Fatalf("ExportEquipment() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ExportEquipment() length = %d, want 2", len(items))
	}
	if items[0].OwnerInventoryUUID != "99999999-aaaa-bbbb-cccc-ddddeeeeffff" {
		t.Fatalf("EquipmentInfo owner inventory = %q", items[0].OwnerInventoryUUID)
	}
	if items[1].OwnerInventoryUUID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("EquipmentInfo second owner inventory = %q", items[1].OwnerInventoryUUID)
	}
}

func TestJSONAPIExportEquipmentIncludesModernCryopodSaddles(t *testing.T) {
	save := openSyntheticCryopoddedDinoSaveWithSaddle(t)
	defer save.Close()

	items, err := NewJSON(save).ExportEquipment()
	if err != nil {
		t.Fatalf("ExportEquipment() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ExportEquipment() length = %d, want 1 cryopod saddle: %#v", len(items), items)
	}
	if !items[0].InCryopod {
		t.Fatalf("EquipmentInfo.InCryopod = false, want true")
	}
	if items[0].Kind != string(arkobject.EquipmentSaddle) {
		t.Fatalf("EquipmentInfo.Kind = %q, want saddle", items[0].Kind)
	}
	if items[0].UUID != "dddddddd-eeee-ffff-0000-111111111111" {
		t.Fatalf("EquipmentInfo.UUID = %q, want containing cryopod UUID", items[0].UUID)
	}
	wantBlueprint := "/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C"
	if items[0].Blueprint != wantBlueprint {
		t.Fatalf("EquipmentInfo.Blueprint = %q, want %q", items[0].Blueprint, wantBlueprint)
	}
}

func TestEquipmentStatsInfoOmitsNonFiniteValues(t *testing.T) {
	info := EquipmentInfo{
		Rating:            math.NaN(),
		CurrentDurability: math.Inf(1),
		Stats: equipmentStatsInfo(arkobject.EquipmentStats{
			Internal:               map[arkobject.EquipmentStat]uint16{arkobject.EquipmentStatDamage: 123},
			Damage:                 math.NaN(),
			Durability:             math.Inf(1),
			Armor:                  math.Inf(-1),
			HypothermalResistance:  8.8,
			HyperthermalResistance: 15.6,
		}),
	}
	info.sanitize()
	if info.Rating != 0 || info.CurrentDurability != 0 {
		t.Fatalf("non-finite top-level equipment floats were not sanitized: %#v", info)
	}
	if _, err := json.Marshal(info); err != nil {
		t.Fatalf("json.Marshal(equipment info) error = %v", err)
	}

	stats := equipmentStatsInfo(arkobject.EquipmentStats{
		Internal:               map[arkobject.EquipmentStat]uint16{arkobject.EquipmentStatDamage: 123},
		Damage:                 math.NaN(),
		Durability:             math.Inf(1),
		Armor:                  math.Inf(-1),
		HypothermalResistance:  8.8,
		HyperthermalResistance: 15.6,
	})
	if stats == nil {
		t.Fatalf("equipmentStatsInfo() = nil, want stats with finite fields")
	}
	if stats.Damage != 0 || stats.Durability != 0 || stats.Armor != 0 {
		t.Fatalf("non-finite stats were not sanitized: %#v", stats)
	}
	if _, err := json.Marshal(stats); err != nil {
		t.Fatalf("json.Marshal(equipment stats) error = %v", err)
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

func TestJSONAPIExportStackablesIncludesOwnerInventory(t *testing.T) {
	save := openSyntheticStackableOwnedByStructureSave(t)
	defer save.Close()

	items, err := NewJSON(save).ExportStackables()
	if err != nil {
		t.Fatalf("ExportStackables() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ExportStackables() length = %d, want 2", len(items))
	}
	if items[0].OwnerInventoryUUID != "99999999-aaaa-bbbb-cccc-ddddeeeeffff" {
		t.Fatalf("StackableInfo owner inventory = %q", items[0].OwnerInventoryUUID)
	}
	if items[1].OwnerInventoryUUID != "11111111-2222-3333-4444-555555555555" {
		t.Fatalf("StackableInfo second owner inventory = %q", items[1].OwnerInventoryUUID)
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
	api := NewBase(save, "")
	bases, err := api.All()
	if err != nil {
		t.Fatalf("BaseAPI.All() error = %v", err)
	}
	wantLocation := bases[0].Location.AsMapCoords(api.mapName)
	wantAverageLocation := bases[0].AverageLocation.AsMapCoords(api.mapName)
	if items[0].MapLocation == nil || items[0].AverageMapLocation == nil {
		t.Fatalf("BaseInfo map locations = %#v", items[0])
	}
	if items[0].MapLocation.Lat != wantLocation.Lat || items[0].MapLocation.Lon != wantLocation.Long {
		t.Fatalf("MapLocation = %#v, want lat=%f lon=%f", items[0].MapLocation, wantLocation.Lat, wantLocation.Long)
	}
	if items[0].AverageMapLocation.Lat != wantAverageLocation.Lat || items[0].AverageMapLocation.Lon != wantAverageLocation.Long {
		t.Fatalf("AverageMapLocation = %#v, want lat=%f lon=%f", items[0].AverageMapLocation, wantAverageLocation.Lat, wantAverageLocation.Long)
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
