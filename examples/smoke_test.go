package examples_test

import (
	"bytes"
	"encoding/binary"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestExamplesRunAgainstLocalSyntheticFixtures(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	copyPath := filepath.Join(dir, "copy.ark")
	objectCopyPath := filepath.Join(dir, "object-copy.ark")
	customCopyPath := filepath.Join(dir, "custom-copy.ark")
	clusterPath := filepath.Join(dir, "EOS_abc123")
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	profilePath := filepath.Join(dir, "123.arkprofile")
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
		}),
		Objects: map[uuid.UUID][]byte{
			objectID:    testfixtures.GenericObjectBytes(0x10000001, 0x10000002),
			dinoID:      testfixtures.GenericObjectBytes(0x10000003, 0x10000002),
			stackableID: stackableObjectBytes(0x10000004, 0x10000002, 0x10000005, 0x10000006, 250),
			equipmentID: testfixtures.GenericObjectBytes(0x10000012, 0x10000002),
			structureID: syntheticSmokeStructureObjectBytes(0x10000013, 0x10000002, 0x10000014, 0x10000006, 0x10000015, 555),
			pawnID:      playerPawnObjectBytes(0x10000007, 0x10000000, 0x10000008, 0x10000006, 0x10000009, 0x1000000a, 0x1000000e, 0x1000000f, 0x10000010, 0x10000011, inventoryID),
			inventoryID: inventoryObjectBytes(0x1000000b, 0x10000000, 0x1000000c, 0x1000000d, 0x1000000a, firstItemID, secondItemID),
		},
	})
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteTributeFile(t, tributePath, []uint64{11, 22}, []uint64{33})
	testfixtures.WritePlayerArchiveWithOptions(t, profilePath, testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
		TribeID:       777,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
		},
	})

	runExample(t, "map_summary", "map=Valguero_WP", savePath)
	runExample(t, "object_classes", "Blueprint'/Game/Test.Test_C'", savePath)
	runExample(t, "class_lookup", "objects=1 classes=1", savePath, "PrimalStructure_Wall_Stone_C")
	runExample(t, "property_filter", "objects=7 classes=7", savePath, "None")
	runExample(t, "dino_filter", "dinos=1 tamed=0 wild=1 cryopodded=0 classes=1", savePath)
	runExample(t, "dino_filter", "dinos=1 tamed=0 wild=1 cryopodded=0 classes=1", "--no-cryos", savePath)
	runExample(t, "dino_best_stat", "no_match", savePath)
	runExample(t, "dino_most_mutated", "no_match", savePath)
	runExample(t, "dino_babies", "wild_babies=0 tamed_babies=0", savePath)
	runExample(t, "dino_wild_tamables", "wild_dinos=1 wild_tamables=1", savePath)
	runExample(t, "dino_wild_tamed", "wild_tamed=0 max_level=0", savePath)
	runExample(t, "stackable_count", "items=1 total=250", savePath, resourceBlueprint)
	runExample(t, "stackable_owned_by", "tribe_id=555 items=0 total=0", savePath, resourceBlueprint, "555")
	runExample(t, "equipment_summary", "items=1 weapons=1 armor=0 saddles=0 cryopod_saddles=0 shields=0", savePath)
	runExample(t, "equipment_best", "weapon_damage=0.0 weapon=WeaponBow weapon_crafted=false\narmor=no_match", savePath)
	runExample(t, "equipment_ascendant_weapon_bps", "items=0 max_damage=0.0", savePath)
	runExample(t, "equipment_saddles", "item_saddles=0 cryopod_saddles=0 total_saddles=0 max_armor=0.0", savePath)
	runExample(t, "equipment_owned_by", "tribe_id=555 items=0 max_damage=0.0", savePath, "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'", "555")
	runExample(t, "structure_owner_count", "tribe_id=555 structures=1", savePath, "555")
	runExample(t, "base_components", "bases=1 total_structures=1 largest=1 min10=0 faults=0", savePath)
	runExample(t, "player_inventory", "location=(11.00,22.00,33.00)", savePath, "42")
	runExample(t, "local_profiles", "unlocked_engrams=2", dir)
	runExample(t, "cluster_json", `"id": "EOS_abc123"`, clusterPath)
	runExample(t, "local_tribute", "player_data_ids=2", tributePath)
	runExample(t, "tribute_json", `"player_data_count": 2`, tributePath)
	runExample(t, "mutation_copy", "wrote copy:", savePath, copyPath)
	if _, err := os.Stat(copyPath); err != nil {
		t.Fatalf("mutation_copy output missing: %v", err)
	}
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

func playerPawnObjectBytes(classNameID uint32, noneNameID uint32, linkedPlayerDataIDName uint32, intPropertyID uint32, inventoryNameID uint32, objectPropertyID uint32, locationNameID uint32, structPropertyID uint32, vectorNameID uint32, coreObjectNameID uint32, inventoryID uuid.UUID) []byte {
	var props bytes.Buffer
	testfixtures.WriteIntPropertyID(&props, linkedPlayerDataIDName, intPropertyID, 42)
	testfixtures.WriteObjectReferencePropertyID(&props, inventoryNameID, objectPropertyID, inventoryID)
	writeVectorPropertyID(&props, locationNameID, structPropertyID, vectorNameID, coreObjectNameID, 11, 22, 33)
	return testfixtures.ObjectBytesWithProperties(classNameID, noneNameID, props.Bytes())
}

func writeVectorPropertyID(buf *bytes.Buffer, name uint32, structProperty uint32, vectorName uint32, coreObjectName uint32, x float64, y float64, z float64) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, structProperty)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, vectorName)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, coreObjectName)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(24))
	buf.WriteByte(8)
	_ = binary.Write(buf, binary.LittleEndian, x)
	_ = binary.Write(buf, binary.LittleEndian, y)
	_ = binary.Write(buf, binary.LittleEndian, z)
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

func syntheticSmokeStructureObjectBytes(classNameID uint32, noneNameID uint32, structureIDName uint32, intPropertyID uint32, tribeIDName uint32, tribeID int32) []byte {
	var props bytes.Buffer
	testfixtures.WriteIntPropertyID(&props, structureIDName, intPropertyID, 101)
	testfixtures.WriteIntPropertyID(&props, tribeIDName, intPropertyID, tribeID)
	return testfixtures.ObjectBytesWithProperties(classNameID, noneNameID, props.Bytes())
}
