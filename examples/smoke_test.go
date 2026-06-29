package examples_test

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/e2etest"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestEquipmentExamplesUseTypedPathHelpers(t *testing.T) {
	for _, path := range []string{
		filepath.Join("equipment_ascendant_weapon_bps", "main.go"),
		filepath.Join("equipment_best", "main.go"),
		filepath.Join("equipment_rank", "main.go"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", path, err)
		}
		if strings.Contains(string(data), "NewEquipmentFromPath") {
			t.Fatalf("%s still owns EquipmentAPI lifecycle; use typed arkapi path helper", path)
		}
	}
}

func TestStructureAggregateExamplesUseTypedPathHelpers(t *testing.T) {
	for _, path := range []string{
		filepath.Join("structure_health", "main.go"),
		filepath.Join("structure_at_location", "main.go"),
		filepath.Join("structure_owner_count", "main.go"),
		filepath.Join("structure_owners", "main.go"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", path, err)
		}
		if strings.Contains(string(data), "NewStructureFromPath") {
			t.Fatalf("%s still owns StructureAPI lifecycle; use typed arkapi path helper", path)
		}
	}
}

func TestGeneralAggregateExamplesUseTypedPathHelpers(t *testing.T) {
	for _, path := range []string{
		filepath.Join("class_lookup", "main.go"),
		filepath.Join("class_property_summary", "main.go"),
		filepath.Join("object_classes", "main.go"),
		filepath.Join("object_summary", "main.go"),
		filepath.Join("parse_all", "main.go"),
		filepath.Join("property_filter", "main.go"),
		filepath.Join("property_positions", "main.go"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", path, err)
		}
		if strings.Contains(string(data), "NewGeneralFromPath") {
			t.Fatalf("%s still owns GeneralAPI lifecycle; use typed arkapi path helper", path)
		}
	}
}

func TestPlayerAggregateExamplesUseTypedPathHelpers(t *testing.T) {
	for _, path := range []string{
		filepath.Join("local_profiles", "main.go"),
		filepath.Join("player_unlocked_engrams", "main.go"),
	} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("ReadFile(%s) error = %v", path, err)
		}
		if strings.Contains(string(data), "NewPlayerFromPath") || strings.Contains(string(data), "NewPlayerFromDirectory") {
			t.Fatalf("%s still owns PlayerAPI lifecycle; use typed arkapi path helper", path)
		}
	}
}

func TestExamplesRunAgainstLocalSyntheticFixtures(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	copyPath := filepath.Join(dir, "copy.ark")
	classRemoveCopyPath := filepath.Join(dir, "class-remove-copy.ark")
	baseImportCopyPath := filepath.Join(dir, "base-import-copy.ark")
	objectCopyPath := filepath.Join(dir, "object-copy.ark")
	propertyCopyPath := filepath.Join(dir, "property-copy.ark")
	customCopyPath := filepath.Join(dir, "custom-copy.ark")
	dinoHeatmapPath := filepath.Join(dir, "dino-heatmap.json")
	dinoExportPath := filepath.Join(dir, "dino-export")
	equipmentExportPath := filepath.Join(dir, "equipment-export")
	heatmapPath := filepath.Join(dir, "structure-heatmap.json")
	baseExportPath := filepath.Join(dir, "base-export")
	structureExportPath := filepath.Join(dir, "structure-export")
	exportAllPath := filepath.Join(dir, "json-exports")
	equipmentHistoryManifestPath := filepath.Join(dir, "equipment-history-files.json")
	equipmentHistoryReportPath := filepath.Join(dir, "equipment-history-report.json")
	clusterPath := filepath.Join(dir, "EOS_abc123")
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	profilePath := filepath.Join(dir, "123.arkprofile")
	tribePath := filepath.Join(dir, "777.arktribe")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	dinoID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	stackableID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	equipmentID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeee")
	structureID := uuid.MustParse("22222233-4455-6677-8899-aabbccdddddd")
	pawnID := uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff")
	inventoryID := uuid.MustParse("44444444-4455-6677-8899-aabbccddeeff")
	firstItemID := uuid.MustParse("55555555-4455-6677-8899-aabbccddeeff")
	secondItemID := uuid.MustParse("66666666-4455-6677-8899-aabbccddeeff")
	resourceBlueprint := "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Electronics.PrimalItemResource_Electronics_C'"
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000000: "None",
			0x10000001: "Blueprint'/Game/Test.Test_C'",
			0x10000002: "None",
			0x10000003: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			0x10000004: resourceBlueprint,
			0x10000005: "ItemQuantity",
			0x10000006: "IntProperty",
			0x10000007: "Blueprint'/Game/PrimalEarth/CoreBlueprints/PlayerPawnTest.PlayerPawnTest_C'",
			0x10000008: "LinkedPlayerDataID",
			0x10000009: "MyInventoryComponent",
			0x1000000a: "ObjectProperty",
			0x1000000b: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Inventories/PrimalInventoryTest.PrimalInventoryTest_C'",
			0x1000000c: "InventoryItems",
			0x1000000d: "ArrayProperty",
			0x1000000e: "SavedBaseWorldLocation",
			0x1000000f: "StructProperty",
			0x10000010: "Vector",
			0x10000011: "/Script/CoreUObject",
			0x10000012: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
			0x10000013: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			0x10000014: "StructureID",
			0x10000015: "TargetingTeam",
			0x10000016: "Blueprint'/Game/Test/InventoryItem.InventoryItem_C'",
			0x10000017: "DinoID1",
			0x10000018: "MaxHealth",
			0x10000019: "FloatProperty",
			0x1000001a: "Health",
		}),
		Objects: map[uuid.UUID][]byte{
			objectID:     smokePatchableObjectBytes(0x10000001, 0x10000002, 0x10000017, 0x10000006, 1001),
			dinoID:       testfixtures.GenericObjectBytes(0x10000003, 0x10000002),
			stackableID:  stackableObjectBytes(0x10000004, 0x10000002, 0x10000005, 0x10000006, 250),
			equipmentID:  testfixtures.GenericObjectBytes(0x10000012, 0x10000002),
			structureID:  syntheticSmokeStructureObjectBytes(0x10000013, 0x10000002, 0x10000014, 0x10000006, 0x10000015, 555, 0x10000018, 0x10000019, 0x1000001a),
			pawnID:       playerPawnObjectBytes(0x10000007, 0x10000000, 0x10000008, 0x10000006, 0x10000009, 0x1000000a, 0x1000000e, 0x1000000f, 0x10000010, 0x10000011, inventoryID),
			inventoryID:  inventoryObjectBytes(0x1000000b, 0x10000000, 0x1000000c, 0x1000000d, 0x1000000a, firstItemID, secondItemID),
			firstItemID:  testfixtures.GenericObjectBytes(0x10000016, 0x10000002),
			secondItemID: testfixtures.GenericObjectBytes(0x10000016, 0x10000002),
		},
		Custom: map[string][]byte{
			"ActorTransforms": testfixtures.ActorTransforms(testfixtures.ActorTransform{
				UUID:       structureID,
				X:          11,
				Y:          22,
				Z:          33,
				Quaternion: 1,
			}),
		},
	})
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteTributeFile(t, tributePath, []uint64{11, 22}, []uint64{33})
	testfixtures.WritePlayerArchiveWithOptions(t, profilePath, testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		TribeID:             777,
		NumDeaths:           4,
		ExtraCharacterLevel: 9,
		ExperiencePoints:    123.5,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
		},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, tribePath, testfixtures.TribeArchiveOptions{
		Name:     "Porters",
		TribeID:  777,
		OwnerID:  42,
		NumDinos: 7,
	})

	runExample(t, "map_summary", "map=Valguero_WP", savePath)
	runExample(t, "parse_all", "objects=9 parsed=9 faults=0", savePath)
	runExample(t, "object_classes", "Blueprint'/Game/Test.Test_C'", savePath)
	runExample(t, "class_lookup", "objects=1 classes=1", savePath, "PrimalStructure_Wall_Stone_C")
	runExample(t, "property_filter", "objects=9 classes=8", savePath, "None")
	runExample(t, "property_positions", "properties=3 name_offsets=3 value_offsets=3 encoded=3 positioned=1 offsets_ok=3", savePath, pawnID.String())
	runExample(t, "dino_filter", "dinos=1 tamed=0 wild=1 cryopodded=0 classes=1", savePath)
	runExample(t, "dino_filter", "dinos=1 tamed=0 wild=1 cryopodded=0 classes=1", "--no-cryos", savePath)
	runExample(t, "dino_best_stat", "no_match", savePath)
	runExample(t, "dino_most_mutated", "no_match", savePath)
	runExample(t, "dino_babies", "wild_babies=0 tamed_babies=0", savePath)
	runExample(t, "dino_wild_tamables", "wild_dinos=1 wild_tamables=1", savePath)
	runExample(t, "dino_wild_tamed", "wild_tamed=0 max_level=0", savePath)
	runExample(t, "dino_export_from_save", "dinos=1 rows=1 faults=0 wrote=", savePath, dinoExportPath)
	for _, path := range []string{
		filepath.Join(dinoExportPath, "manifest.json"),
		filepath.Join(dinoExportPath, "dino_"+dinoID.String(), "dino_"+dinoID.String()+".bin"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("dino_export_from_save output %s missing: %v", path, err)
		}
	}
	runExample(t, "dino_heatmap", "cells=0 total=0 max=0 faults=0 wrote=", savePath, dinoHeatmapPath)
	if _, err := os.Stat(dinoHeatmapPath); err != nil {
		t.Fatalf("dino_heatmap output missing: %v", err)
	}
	runExample(t, "dino_heatmap", "cells=0 total=0 max=0 faults=0 wrote=", "--no-cryos", savePath, dinoHeatmapPath)
	runExample(t, "stackable_count", "items=1 total=250", savePath, resourceBlueprint)
	runExample(t, "stackable_owned_by", "tribe_id=555 items=0 total=0", savePath, resourceBlueprint, "555")
	runExample(t, "equipment_summary", "items=1 weapons=1 armor=0 saddles=0 cryopod_saddles=0 shields=0 with_custom_data=0 custom_data_entries=0", savePath)
	runExample(t, "equipment_best", "weapon_damage=0.0 weapon=WeaponBow weapon_crafted=false\narmor=no_match", savePath)
	runExample(t, "equipment_rank", "ranked=0 best_rating=0.0 best_average_stat=0.0 crafted=0 blueprints=0 classes=0", savePath)
	runExample(t, "equipment_ascendant_weapon_bps", "items=0 max_damage=0.0", savePath)
	runExample(t, "equipment_saddles", "item_saddles=0 cryopod_saddles=0 total_saddles=0 max_armor=0.0", savePath)
	runExample(t, "equipment_owned_by", "tribe_id=555 items=0 max_damage=0.0", savePath, "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'", "555")
	runExample(t, "equipment_export_from_save", "items=1 rows=1 faults=0 wrote=", savePath, equipmentExportPath)
	for _, path := range []string{
		filepath.Join(equipmentExportPath, "manifest.json"),
		filepath.Join(equipmentExportPath, "item_"+equipmentID.String()+".bin"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("equipment_export_from_save output %s missing: %v", path, err)
		}
	}
	runExample(t, "export_all_items", "exports=8 wrote=", savePath, exportAllPath)
	if _, err := os.Stat(filepath.Join(exportAllPath, "manifest.json")); err != nil {
		t.Fatalf("export_all_items manifest missing: %v", err)
	}
	if err := os.WriteFile(equipmentHistoryManifestPath, []byte(fmt.Sprintf("[\"%s\",\"%s\"]\n", savePath, savePath)), 0o600); err != nil {
		t.Fatalf("write equipment history manifest: %v", err)
	}
	runExample(t, "equipment_history", "saves=2 initial=1 changes=0 final=1 wrote=", equipmentHistoryManifestPath, equipmentHistoryReportPath)
	if _, err := os.Stat(equipmentHistoryReportPath); err != nil {
		t.Fatalf("equipment_history report missing: %v", err)
	}
	runExample(t, "structure_owner_count", "tribe_id=555 structures=1", savePath, "555")
	runExample(t, "structure_owners", "structures=1 with_tribe_id=1 with_player_id=0 with_tribe_name=0 with_player_name=0 with_original_placer_id=0 unique_tribes=1", savePath)
	runExample(t, "structure_health", "structures=1 with_health=1 damaged=1 repaired=0 without_max_health=0 avg_health=90.0 min_health=90.0 max_health=90.0 faults=0", savePath)
	runExample(t, "structure_owner_locations", "structures=1 owners=1 cells=1 named_cells=1 multi_structure_cells=0 skipped_without_owner=0 skipped_without_location=0 faults=0", savePath, "Valguero", "1")
	runExample(t, "structure_export_from_save", "structures=1 rows=1 faults=0 wrote=", savePath, structureExportPath)
	for _, path := range []string{
		filepath.Join(structureExportPath, "manifest.json"),
		filepath.Join(structureExportPath, "str_"+structureID.String()+".bin"),
		filepath.Join(structureExportPath, "str_"+structureID.String()+"_location.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("structure_export_from_save output %s missing: %v", path, err)
		}
	}
	runExample(t, "base_components", "bases=1 total_structures=1 largest=1 min10=0 faults=0", savePath)
	runExample(t, "base_export_from_save", "bases=1 structures=1 faults=0 wrote=", savePath, baseExportPath)
	for _, path := range []string{
		filepath.Join(baseExportPath, "manifest.json"),
		filepath.Join(baseExportPath, "base_"+structureID.String(), "base.json"),
		filepath.Join(baseExportPath, "base_"+structureID.String(), "str_"+structureID.String()+".bin"),
		filepath.Join(baseExportPath, "base_"+structureID.String(), "str_"+structureID.String()+"_location.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("base_export_from_save output %s missing: %v", path, err)
		}
	}
	runExample(t, "structure_heatmap", "cells=1 total=1 max=1 faults=0 wrote=", savePath, heatmapPath)
	if _, err := os.Stat(heatmapPath); err != nil {
		t.Fatalf("structure_heatmap output missing: %v", err)
	}
	runExample(t, "player_inventory", "location=(11.00,22.00,33.00)", savePath, "42")
	runExample(t, "player_inventories", "players=1 with_inventory=1 without_inventory=0 total_items=2 max_items=2 min_items=2 avg_items=2.00 faults=0", savePath)
	runExample(t, "local_profiles", "average_deaths=4.00 average_level=10.00 average_experience=123.50", dir)
	runExample(t, "local_profiles", "unlocked_engrams=2", dir)
	runExample(t, "player_list", "players=1 with_names=1 highest_level=10", dir)
	runExample(t, "tribe_list", "tribes=1 with_names=1 members=0 dinos=7", dir)
	runExample(t, "player_and_tribe_data", `"players": 1`, dir)
	runExample(t, "player_unlocked_engrams", "unlocked_engrams=2 first=Blueprint'/Game/Engrams/EngramA.EngramA_C' last=Blueprint'/Game/Engrams/EngramB.EngramB_C'", dir)
	runExample(t, "cluster_json", `"id": "EOS_abc123"`, clusterPath)
	runExample(t, "cluster_typed", "cluster=EOS_abc123 items=0 dinos=0 equipment=0 dino_items=0 other_items=0 crafted=0 unsupported_items=0 parsed_dinos=0 unsupported_dinos=0 dino_parse_errors=0 unparsed_dinos=0 dino_ids=0 tamed_dinos=0 female_dinos=0 baby_dinos=0 dead_dinos=0 dinos_with_stats=0 avg_base_level=0.00 max_base_level=0 avg_current_level=0.00 max_current_level=0 embedded_objects=0 parse_errors=0", clusterPath)
	runExample(t, "local_tribute", "player_data_ids=2", tributePath)
	runExample(t, "tribute_json", `"player_data_count": 2`, tributePath)
	runExample(t, "logging_config", "[api] This is an API log.")
	runExample(t, "mutation_copy", "wrote copy:", savePath, copyPath)
	if _, err := os.Stat(copyPath); err != nil {
		t.Fatalf("mutation_copy output missing: %v", err)
	}
	runExample(t, "mutation_copy", "removed class substring:", "remove-class-contains", savePath, classRemoveCopyPath, "/Game/Test.Test_C")
	classRemoveCopy, err := arksave.Open(classRemoveCopyPath)
	if err != nil {
		t.Fatalf("Open(class remove mutation copy) error = %v", err)
	}
	if _, err := classRemoveCopy.ObjectBinary(objectID); err == nil {
		_ = classRemoveCopy.Close()
		t.Fatalf("ObjectBinary(%s) error = nil, want removed object", objectID)
	}
	_ = classRemoveCopy.Close()
	runExample(t, "mutation_copy", "imported base rows: 1", "import-base-binary", savePath, baseImportCopyPath, baseExportPath)
	baseImportCopy, err := arksave.Open(baseImportCopyPath)
	if err != nil {
		t.Fatalf("Open(base import mutation copy) error = %v", err)
	}
	if _, err := baseImportCopy.ObjectBinary(structureID); err != nil {
		_ = baseImportCopy.Close()
		t.Fatalf("ObjectBinary(imported structure %s) error = %v", structureID, err)
	}
	_ = baseImportCopy.Close()
	runExample(t, "mutation_copy", "imported structure rows: 1", "import-structure-binary", savePath, baseImportCopyPath+"-structure", structureExportPath)
	structureImportCopy, err := arksave.Open(baseImportCopyPath + "-structure")
	if err != nil {
		t.Fatalf("Open(structure import mutation copy) error = %v", err)
	}
	if _, err := structureImportCopy.ObjectBinary(structureID); err != nil {
		_ = structureImportCopy.Close()
		t.Fatalf("ObjectBinary(imported structure %s) error = %v", structureID, err)
	}
	_ = structureImportCopy.Close()
	runExample(t, "mutation_copy", "imported dino rows: 1", "import-dino-binary", savePath, baseImportCopyPath+"-dino", dinoExportPath)
	runExample(t, "mutation_copy", "imported equipment rows: 1", "import-equipment-binary", savePath, baseImportCopyPath+"-equipment", equipmentExportPath)
	runExample(t, "mutation_copy", "wrote object bytes:", "put-object-hex", savePath, objectCopyPath, objectID.String(), "090807")
	objectCopy, err := arksave.Open(objectCopyPath)
	if err != nil {
		t.Fatalf("Open(object mutation copy) error = %v", err)
	}
	raw, err := objectCopy.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", objectID, err)
	}
	_ = objectCopy.Close()
	if !bytes.Equal(raw, []byte{9, 8, 7}) {
		t.Fatalf("ObjectBinary(%s) = % x, want 09 08 07", objectID, raw)
	}
	var replacement bytes.Buffer
	testfixtures.WriteIntPropertyID(&replacement, 0x10000017, 0x10000006, 2002)
	runExample(t, "mutation_copy", "replaced object property: DinoID1[0]", "replace-object-property-hex", savePath, propertyCopyPath, objectID.String(), "DinoID1", "0", hex.EncodeToString(replacement.Bytes()))
	propertyCopy, err := arksave.Open(propertyCopyPath)
	if err != nil {
		t.Fatalf("Open(property mutation copy) error = %v", err)
	}
	propertyObject, err := propertyCopy.ParsedObject(objectID)
	if err != nil {
		t.Fatalf("ParsedObject(property mutation copy) error = %v", err)
	}
	_ = propertyCopy.Close()
	gotProperty, ok := propertyObject.Container().Value("DinoID1")
	if !ok || gotProperty != int32(2002) {
		t.Fatalf("DinoID1 = %#v, %v; want int32(2002)", gotProperty, ok)
	}
	runExample(t, "mutation_copy", "wrote custom value:", "put-custom", savePath, customCopyPath, "Extra", "090807")
	customCopy, err := arksave.Open(customCopyPath)
	if err != nil {
		t.Fatalf("Open(custom mutation copy) error = %v", err)
	}
	customValue, err := customCopy.CustomValue("Extra")
	if err != nil {
		t.Fatalf("CustomValue(Extra) error = %v", err)
	}
	_ = customCopy.Close()
	if !bytes.Equal(customValue, []byte{9, 8, 7}) {
		t.Fatalf("CustomValue(Extra) = % x, want 09 08 07", customValue)
	}
}

func TestPlayerInventoriesExampleCountsOnlyInventoryFaults(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "player-inventories.ark")
	inventoryID := uuid.MustParse("44444444-4455-6677-8899-aabbccddeeff")
	firstItemID := uuid.MustParse("55555555-4455-6677-8899-aabbccddeeff")
	secondItemID := uuid.MustParse("66666666-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000000: "None",
			0x10000006: "IntProperty",
			0x10000007: "Blueprint'/Game/PrimalEarth/CoreBlueprints/PlayerPawnTest.PlayerPawnTest_C'",
			0x10000008: "LinkedPlayerDataID",
			0x10000009: "MyInventoryComponent",
			0x1000000a: "ObjectProperty",
			0x1000000b: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Inventories/PrimalInventoryTest.PrimalInventoryTest_C'",
			0x1000000c: "InventoryItems",
			0x1000000d: "ArrayProperty",
			0x1000000e: "SavedBaseWorldLocation",
			0x1000000f: "StructProperty",
			0x10000010: "Vector",
			0x10000011: "/Script/CoreUObject",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("11111111-4455-6677-8899-aabbccddeeff"): testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{PlayerDataID: 42, CharacterName: "Survivor"}),
			uuid.MustParse("22222222-4455-6677-8899-aabbccddeeff"): testfixtures.PlayerGameObjectBytes(testfixtures.PlayerArchiveOptions{PlayerDataID: 99, CharacterName: "Broken"})[:40],
			uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff"): playerPawnObjectBytes(0x10000007, 0x10000000, 0x10000008, 0x10000006, 0x10000009, 0x1000000a, 0x1000000e, 0x1000000f, 0x10000010, 0x10000011, inventoryID),
			inventoryID:  inventoryObjectBytes(0x1000000b, 0x10000000, 0x1000000c, 0x1000000d, 0x1000000a, firstItemID, secondItemID),
			firstItemID:  testfixtures.GenericObjectBytes(0x1000000b, 0x10000000),
			secondItemID: testfixtures.GenericObjectBytes(0x1000000b, 0x10000000),
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "42.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
	})

	runExample(t, "player_inventories", "players=1 with_inventory=1 without_inventory=0 total_items=2 max_items=2 min_items=2 avg_items=2.00 faults=0", savePath)
}

func TestPlayerInventoriesExampleFallsBackToDirectoryPlayers(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "player-inventories-fallback.ark")
	inventoryID := uuid.MustParse("44444444-4455-6677-8899-aabbccddeeff")
	firstItemID := uuid.MustParse("55555555-4455-6677-8899-aabbccddeeff")
	secondItemID := uuid.MustParse("66666666-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, savePath, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000000: "None",
			0x10000006: "IntProperty",
			0x10000007: "Blueprint'/Game/PrimalEarth/CoreBlueprints/PlayerPawnTest.PlayerPawnTest_C'",
			0x10000008: "LinkedPlayerDataID",
			0x10000009: "MyInventoryComponent",
			0x1000000a: "ObjectProperty",
			0x1000000b: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Inventories/PrimalInventoryTest.PrimalInventoryTest_C'",
			0x1000000c: "InventoryItems",
			0x1000000d: "ArrayProperty",
			0x1000000e: "SavedBaseWorldLocation",
			0x1000000f: "StructProperty",
			0x10000010: "Vector",
			0x10000011: "/Script/CoreUObject",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff"): playerPawnObjectBytes(0x10000007, 0x10000000, 0x10000008, 0x10000006, 0x10000009, 0x1000000a, 0x1000000e, 0x1000000f, 0x10000010, 0x10000011, inventoryID),
			inventoryID:  inventoryObjectBytes(0x1000000b, 0x10000000, 0x1000000c, 0x1000000d, 0x1000000a, firstItemID, secondItemID),
			firstItemID:  testfixtures.GenericObjectBytes(0x1000000b, 0x10000000),
			secondItemID: testfixtures.GenericObjectBytes(0x1000000b, 0x10000000),
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "42.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Fallback",
	})

	runExample(t, "player_inventories", "players=1 with_inventory=1 without_inventory=0 total_items=2 max_items=2 min_items=2 avg_items=2.00 faults=0", savePath)
}

func TestExamplesRunAgainstProvidedData(t *testing.T) {
	data := e2etest.DiscoverProvidedData(t)
	if data.SavePath == "" && data.ClusterPath == "" {
		t.Skip("set ARK_E2E_SAVE or ARK_E2E_SAVE_DIR to run provided-data examples E2E")
	}

	if data.SavePath != "" {
		runProvidedExample(t, "map_summary", "map=", data.SavePath)
		runProvidedExample(t, "player_all", "players=", data.SavePath)
		runProvidedExample(t, "equipment_summary", "items=", data.SavePath)
		runProvidedExample(t, "structure_health", "structures=", data.SavePath)
		runProvidedExample(t, "structure_owners", "structures=", data.SavePath)
		runProvidedExample(t, "structure_owner_locations", "structures=", data.SavePath)
	}

	if data.Dir == "" {
		if data.ClusterPath != "" {
			runProvidedExample(t, "cluster_typed", "cluster=", data.ClusterPath)
		}
		return
	}
	if data.ProfileCount > 0 {
		runProvidedExample(t, "player_list", "players=", data.Dir)
		runProvidedExample(t, "local_profiles", "profiles=", data.Dir)
	}
	if data.TribeCount > 0 {
		runProvidedExample(t, "tribe_list", "tribes=", data.Dir)
	}
	if data.TributeCount > 0 {
		runProvidedExample(t, "local_tribute", "tribute_files=", data.Dir)
		runProvidedExample(t, "tribute_json", `"count":`, data.Dir)
		if data.TributePath != "" {
			runProvidedExample(t, "local_tribute", "tribute_file=", data.TributePath)
			runProvidedExample(t, "tribute_json", `"player_data_count":`, data.TributePath)
		}
	}
	if data.ClusterPath != "" {
		runProvidedExample(t, "cluster_typed", "cluster=", data.ClusterPath)
	}
}

func runExample(t *testing.T, name string, want string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"run", "./" + name}, args...)
	cmd := exec.Command("go", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run ./%s %v error = %v\n%s", name, args, err, out)
	}
	if !strings.Contains(string(out), want) {
		t.Fatalf("go run ./%s output %q does not contain %q", name, out, want)
	}
}

func runProvidedExample(t *testing.T, name string, want string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"run", "./" + name}, args...)
	cmd := exec.Command("go", cmdArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go run ./%s failed for provided data: %v\n%s", name, err, e2etest.RedactProvidedPaths(string(out)))
	}
	if !strings.Contains(string(out), want) {
		t.Fatalf("go run ./%s output %q does not contain %q", name, e2etest.RedactProvidedPaths(string(out)), want)
	}
}

func smokePatchableObjectBytes(classNameID uint32, noneNameID uint32, propertyNameID uint32, intPropertyID uint32, value int32) []byte {
	var props bytes.Buffer
	testfixtures.WriteIntPropertyID(&props, propertyNameID, intPropertyID, value)
	return testfixtures.ObjectBytesWithProperties(classNameID, noneNameID, props.Bytes())
}

func playerPawnObjectBytes(classNameID uint32, noneNameID uint32, linkedPlayerDataIDName uint32, intPropertyID uint32, inventoryNameID uint32, objectPropertyID uint32, locationNameID uint32, structPropertyID uint32, vectorNameID uint32, coreObjectNameID uint32, inventoryID uuid.UUID) []byte {
	var props bytes.Buffer
	testfixtures.WriteIntPropertyID(&props, linkedPlayerDataIDName, intPropertyID, 42)
	testfixtures.WriteObjectReferencePropertyID(&props, inventoryNameID, objectPropertyID, inventoryID)
	testfixtures.WriteVectorPropertyID(&props, locationNameID, structPropertyID, vectorNameID, coreObjectNameID, 11, 22, 33)
	return testfixtures.ObjectBytesWithProperties(classNameID, noneNameID, props.Bytes())
}

func inventoryObjectBytes(classNameID uint32, noneNameID uint32, itemsNameID uint32, arrayPropertyID uint32, objectPropertyID uint32, itemIDs ...uuid.UUID) []byte {
	var props bytes.Buffer
	testfixtures.WriteObjectReferenceArrayPropertyID(&props, itemsNameID, arrayPropertyID, objectPropertyID, itemIDs)
	return testfixtures.ObjectBytesWithProperties(classNameID, noneNameID, props.Bytes())
}

func stackableObjectBytes(classNameID uint32, noneNameID uint32, quantityNameID uint32, intPropertyID uint32, quantity int32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, classNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	_ = binary.Write(&buf, binary.LittleEndian, quantityNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, intPropertyID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(4))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, quantity)
	_ = binary.Write(&buf, binary.LittleEndian, noneNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticSmokeStructureObjectBytes(classNameID uint32, noneNameID uint32, structureIDName uint32, intPropertyID uint32, tribeIDName uint32, tribeID int32, maxHealthName uint32, floatPropertyID uint32, currentHealthName uint32) []byte {
	var props bytes.Buffer
	testfixtures.WriteIntPropertyID(&props, structureIDName, intPropertyID, 101)
	testfixtures.WriteFloatPropertyID(&props, maxHealthName, floatPropertyID, 10000)
	testfixtures.WriteFloatPropertyID(&props, currentHealthName, floatPropertyID, 9000)
	testfixtures.WriteIntPropertyID(&props, tribeIDName, intPropertyID, tribeID)
	return testfixtures.ObjectBytesWithProperties(classNameID, noneNameID, props.Bytes())
}
