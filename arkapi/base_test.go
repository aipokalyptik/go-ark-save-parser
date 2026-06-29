package arkapi

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestBaseAPIAtGroupsNearbyOwnedStructures(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewBase(save, "Valguero")
	base, err := api.At(arkobject.MapCoords{Lat: 50, Long: 50}, 0.3, &arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("At() error = %v", err)
	}
	if base == nil {
		t.Fatalf("At() = nil, want base")
	}
	if base.StructureCount != 2 || base.Owner.TribeID != 555 {
		t.Fatalf("Base = %#v", base)
	}
	if base.KeystoneUUID != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") {
		t.Fatalf("KeystoneUUID = %s", base.KeystoneUUID)
	}
}

func TestBaseAPIAtExpandsNearbyStructureToConnectedBase(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	structures := NewStructure(save)
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	first, ok, err := structures.ByID(firstID)
	if err != nil {
		t.Fatalf("ByID() error = %v", err)
	}
	if !ok {
		t.Fatalf("ByID() ok = false, want true")
	}

	api := NewBase(save, "Valguero")
	base, err := api.At(first.Location.AsMapCoords("Valguero"), 0.000001, &arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("At() error = %v", err)
	}
	if base == nil {
		t.Fatalf("At() = nil, want connected base")
	}
	if base.StructureCount != 2 {
		t.Fatalf("At() StructureCount = %d, want connected structure count 2", base.StructureCount)
	}
	if _, ok := base.Structures[uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")]; !ok {
		t.Fatalf("At() did not include linked structure outside radius: %#v", base.Structures)
	}
}

func TestBaseAPIAtReturnsNilWhenNoStructuresMatch(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewBase(save, "Valguero")
	base, err := api.At(arkobject.MapCoords{Lat: 10, Long: 10}, 0.1, nil)
	if err != nil {
		t.Fatalf("At() error = %v", err)
	}
	if base != nil {
		t.Fatalf("At() = %#v, want nil", base)
	}
}

func TestBaseAPIAllGroupsLinkedStructures(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewBase(save, "Valguero")
	bases, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(bases) != 1 {
		t.Fatalf("All() length = %d, want 1: %#v", len(bases), bases)
	}
	base := bases[0]
	if base.StructureCount != 2 || base.Owner.TribeID != 555 || base.AverageLocation == nil {
		t.Fatalf("Base = %#v", base)
	}
	if base.KeystoneUUID != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") {
		t.Fatalf("KeystoneUUID = %s", base.KeystoneUUID)
	}
	if base.AverageLocation.X != 500 || base.AverageLocation.Y != 500 {
		t.Fatalf("AverageLocation = %#v", base.AverageLocation)
	}
	minTwo, err := api.AllWithMinStructures(2)
	if err != nil {
		t.Fatalf("AllWithMinStructures(2) error = %v", err)
	}
	if len(minTwo) != 1 {
		t.Fatalf("AllWithMinStructures(2) length = %d, want 1", len(minTwo))
	}
	minThree, err := api.AllWithMinStructures(3)
	if err != nil {
		t.Fatalf("AllWithMinStructures(3) error = %v", err)
	}
	if len(minThree) != 0 {
		t.Fatalf("AllWithMinStructures(3) length = %d, want 0", len(minThree))
	}
}

func TestBaseAPIExportBinaryWritesMetadataAndStructureRows(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	outDir := filepath.Join(t.TempDir(), "base-export")
	exported, err := NewBase(save, "Valguero").ExportBinary(outDir)
	if err != nil {
		t.Fatalf("ExportBinary() error = %v", err)
	}
	if exported.BaseCount != 1 || exported.StructureCount != 2 || exported.FaultCount != 0 {
		t.Fatalf("ExportBinary() = %#v, want one two-structure base without faults", exported)
	}
	baseDir := filepath.Join(outDir, "base_"+firstID.String())
	for _, path := range []string{
		filepath.Join(outDir, "manifest.json"),
		filepath.Join(baseDir, "base.json"),
		filepath.Join(baseDir, "str_"+firstID.String()+"_location.json"),
		filepath.Join(baseDir, "str_"+secondID.String()+"_location.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("exported path %s missing: %v", path, err)
		}
	}
	for _, id := range []uuid.UUID{firstID, secondID} {
		got, err := os.ReadFile(filepath.Join(baseDir, "str_"+id.String()+".bin"))
		if err != nil {
			t.Fatalf("read exported structure %s: %v", id, err)
		}
		want, err := save.ObjectBinary(id)
		if err != nil {
			t.Fatalf("ObjectBinary(%s) error = %v", id, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("exported structure %s bytes differ from save row", id)
		}
	}
}

func TestExportBaseBinaryFromPathWritesMetadataAndStructureRows(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	outDir := filepath.Join(t.TempDir(), "base-export")
	exported, err := ExportBaseBinaryFromPath(save.Path(), "Valguero", outDir)
	if err != nil {
		t.Fatalf("ExportBaseBinaryFromPath() error = %v", err)
	}
	if exported.BaseCount != 1 || exported.StructureCount != 2 || exported.FaultCount != 0 {
		t.Fatalf("ExportBaseBinaryFromPath() = %#v, want one two-structure base without faults", exported)
	}
	baseDir := filepath.Join(outDir, "base_"+firstID.String())
	for _, path := range []string{
		filepath.Join(outDir, "manifest.json"),
		filepath.Join(baseDir, "base.json"),
		filepath.Join(baseDir, "str_"+firstID.String()+".bin"),
		filepath.Join(baseDir, "str_"+secondID.String()+".bin"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("exported path %s missing: %v", path, err)
		}
	}
	for _, id := range []uuid.UUID{firstID, secondID} {
		got, err := os.ReadFile(filepath.Join(baseDir, "str_"+id.String()+".bin"))
		if err != nil {
			t.Fatalf("read exported structure %s: %v", id, err)
		}
		want, err := save.ObjectBinary(id)
		if err != nil {
			t.Fatalf("ObjectBinary(%s) error = %v", id, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("exported structure %s differs from save row", id)
		}
	}
}

func TestExportBaseBinaryFromPathReturnsErrorForInvalidSavePath(t *testing.T) {
	_, err := ExportBaseBinaryFromPath(filepath.Join(t.TempDir(), "missing.ark"), "Valguero", filepath.Join(t.TempDir(), "out"))
	if err == nil {
		t.Fatalf("ExportBaseBinaryFromPath() error = nil, want invalid save path error")
	}
}

func TestBaseAPIAllBasesAppliesUpstreamStyleOptions(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewBase(save, "Valguero")
	defaults, err := api.AllBases(BaseQueryOptions{})
	if err != nil {
		t.Fatalf("AllBases(defaults) error = %v", err)
	}
	if len(defaults) != 0 {
		t.Fatalf("AllBases(defaults) length = %d, want upstream default min_structures 10 filter", len(defaults))
	}

	minTwo, err := api.AllBases(BaseQueryOptions{MinStructures: 2})
	if err != nil {
		t.Fatalf("AllBases(min two) error = %v", err)
	}
	if len(minTwo) != 1 || minTwo[0].StructureCount != 2 {
		t.Fatalf("AllBases(min two) = %#v, want one two-structure base", minTwo)
	}
}

func TestBaseAPIAllBasesSupportsConnectedOnlyMode(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewBase(save, "Valguero")
	bases, err := api.AllBases(BaseQueryOptions{OnlyConnected: true, MinStructures: 2})
	if err != nil {
		t.Fatalf("AllBases(connected) error = %v", err)
	}
	if len(bases) != 1 || bases[0].KeystoneUUID != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") {
		t.Fatalf("AllBases(connected) = %#v, want connected component base", bases)
	}
}

func TestBaseAPIAllWithFaultsKeepsValidBasesAndReportsStructureParseFaults(t *testing.T) {
	save := openSyntheticBaseSaveWithFault(t)
	defer save.Close()

	api := NewBase(save, "Valguero")
	bases, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	if len(bases) != 1 {
		t.Fatalf("AllWithFaults() bases length = %d, want 1: %#v", len(bases), bases)
	}
	if bases[0].StructureCount != 2 || bases[0].Owner.TribeID != 555 {
		t.Fatalf("Base = %#v", bases[0])
	}
	if len(faults) != 1 || faults[0].ClassName != "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'" || faults[0].Err == nil {
		t.Fatalf("AllWithFaults() faults = %#v, want one structure parse fault", faults)
	}
}

func TestBaseAPISummaryWithFaultsKeepsValidBasesAndCountsLargestBase(t *testing.T) {
	save := openSyntheticBaseSaveWithFault(t)
	defer save.Close()

	api := NewBase(save, "Valguero")
	summary, faults, err := api.SummaryWithFaults()
	if err != nil {
		t.Fatalf("SummaryWithFaults() error = %v", err)
	}

	want := BaseSummary{
		Bases:             1,
		Structures:        2,
		LargestBase:       2,
		BasesWithLocation: 1,
		BasesWithTribeID:  1,
		UniqueTribes:      1,
		Faults:            1,
	}
	if summary != want {
		t.Fatalf("SummaryWithFaults() = %#v, want %#v", summary, want)
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("SummaryWithFaults() faults = %#v, want one parse fault", faults)
	}
}

func TestBaseSummaryFromPathReturnsSummaryAndFaults(t *testing.T) {
	save := openSyntheticBaseSaveWithFault(t)
	defer save.Close()

	summary, faults, err := BaseSummaryFromPath(save.Path())
	if err != nil {
		t.Fatalf("BaseSummaryFromPath() error = %v", err)
	}

	want := BaseSummary{
		Bases:             1,
		Structures:        2,
		LargestBase:       2,
		BasesWithLocation: 1,
		BasesWithTribeID:  1,
		UniqueTribes:      1,
		Faults:            1,
	}
	if summary != want {
		t.Fatalf("BaseSummaryFromPath() = %#v, want %#v", summary, want)
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("BaseSummaryFromPath() faults = %#v, want one parse fault", faults)
	}
}

func TestBaseAPIComponentStatsUsesLinkedStructures(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	stats, err := NewBase(save, "Valguero").ComponentStats()
	if err != nil {
		t.Fatalf("ComponentStats() error = %v", err)
	}
	if stats.Components != 1 || stats.TotalStructures != 2 || stats.LargestComponent != 2 || stats.ComponentsAtLeast10 != 0 {
		t.Fatalf("ComponentStats() = %#v, want one two-structure component", stats)
	}
}

func TestBaseComponentStatsFromPathReturnsTypedStats(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	stats, err := BaseComponentStatsFromPath(save.Path(), "Valguero")
	if err != nil {
		t.Fatalf("BaseComponentStatsFromPath() error = %v", err)
	}
	if stats.Components != 1 || stats.TotalStructures != 2 || stats.LargestComponent != 2 || stats.ComponentsAtLeast10 != 0 {
		t.Fatalf("BaseComponentStatsFromPath() = %#v, want one two-structure component", stats)
	}
}

func TestNewBaseFromPathOpensLocalSave(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api, closeAPI, err := NewBaseFromPath(save.Path(), "")
	if err != nil {
		t.Fatalf("NewBaseFromPath() error = %v", err)
	}
	defer closeAPI()

	stats, err := api.ComponentStats()
	if err != nil {
		t.Fatalf("ComponentStats() error = %v", err)
	}
	if stats.Components != 1 || stats.TotalStructures != 2 || stats.LargestComponent != 2 {
		t.Fatalf("ComponentStats() = %#v, want one two-structure component", stats)
	}
}

func TestBaseAPIComponentStatsSkipsUpstreamUnparsedBunkerBase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "base-bunker.ark")
	normalID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	bunkerID := uuid.MustParse("99999999-bbbb-cccc-dddd-eeeeffffffff")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000000: "None",
			0x10000001: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			0x10000002: "/Game/LostColony/Structures/TekBunker/Structures/BP_Bunker_Base.BP_Bunker_Base_C",
			0x10000003: "IntProperty",
			0x10000004: "StructureID",
			0x10000005: "None",
			0x10000006: "TargetingTeam",
		}),
		Objects: map[uuid.UUID][]byte{
			normalID: syntheticStructureObjectBytesForClass(0x10000001, 0x10000000, 0x10000004, 0x10000006, 555),
			bunkerID: syntheticStructureObjectBytesForClass(0x10000002, 0x10000000, 0x10000004, 0x10000006, 555),
		},
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer save.Close()

	stats, err := NewBase(save, "Valguero").ComponentStats()
	if err != nil {
		t.Fatalf("ComponentStats() error = %v", err)
	}
	if stats.TotalStructures != 1 || stats.Components != 1 {
		t.Fatalf("ComponentStats() = %#v, want bunker base excluded for upstream parity", stats)
	}
}

func syntheticStructureObjectBytesForClass(classID uint32, noneID uint32, structureIDName uint32, tribeIDName uint32, tribeID int32) []byte {
	return testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
		ClassID:           classID,
		NoneID:            noneID,
		StructureIDNameID: structureIDName,
		TribeIDNameID:     tribeIDName,
		StructureID:       101,
		TribeID:           tribeID,
	})
}

func openSyntheticBaseSave(t *testing.T) *arksave.Save {
	t.Helper()

	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	otherID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	return openSyntheticSaveWith(t, "base.ark", map[string][]byte{
		"ActorTransforms": syntheticBaseActorTransforms(firstID, secondID),
	}, map[uuid.UUID][]byte{
		firstID:  syntheticBaseStructureObjectBytes(101, secondID),
		secondID: syntheticBaseStructureObjectBytes(102, firstID),
		otherID:  testfixtures.ObjectBytesWithIntProperty(0x10000001, 0x10000004, 0x10000002, 0x10000003, 250),
	})
}

func openSyntheticBaseSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "base.ark", map[string][]byte{
		"ActorTransforms": syntheticBaseActorTransforms(firstID, secondID),
	}, map[uuid.UUID][]byte{
		firstID:  syntheticBaseStructureObjectBytes(101, secondID),
		secondID: syntheticBaseStructureObjectBytes(102, firstID),
		faultyID: testfixtures.TruncatedObjectBytes(0x10000005),
	})
}

func syntheticBaseStructureObjectBytes(structureID int32, linked ...uuid.UUID) []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x10000006, structureID)
	writeFloatProperty(&props, 0x10000007, 10000)
	writeFloatProperty(&props, 0x10000008, 9000)
	writeIntProperty(&props, 0x10000009, 555)
	if len(linked) > 0 {
		testfixtures.WriteObjectReferenceArrayPropertyID(&props, 0x1000001d, 0x1000001e, 0x1000001f, linked)
	}
	return testfixtures.ObjectBytesWithProperties(0x10000005, 0x10000004, props.Bytes())
}

func syntheticBaseActorTransforms(first uuid.UUID, second uuid.UUID) []byte {
	return testfixtures.ActorTransforms(
		testfixtures.ActorTransform{UUID: first, Quaternion: 1},
		testfixtures.ActorTransform{UUID: second, X: 1000, Y: 1000, Quaternion: 1},
	)
}
