package arkapi

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestStructureAPIGetAllParsesStructureObjects(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	structures, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}

	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	got, ok := structures[id]
	if !ok {
		t.Fatalf("All() missing structure %s: %#v", id, structures)
	}
	if got.ID != 123 || got.Owner.TribeID != 555 || got.Location == nil || got.Location.X != 11 {
		t.Fatalf("Structure = %#v", got)
	}
}

func TestNewStructureFromPathOpensLocalSave(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	api, closeAPI, err := NewStructureFromPath(save.Path())
	if err != nil {
		t.Fatalf("NewStructureFromPath() error = %v", err)
	}
	defer closeAPI()

	structures, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	got, ok := structures[id]
	if !ok {
		t.Fatalf("All() missing structure %s: %#v", id, structures)
	}
	if got.ID != 123 || got.Owner.TribeID != 555 || got.Location == nil {
		t.Fatalf("Structure = %#v, want synthetic structure", got)
	}
}

func TestStructureAPIExportBinaryWritesStructureRowsAndManifest(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	dir := t.TempDir()
	exported, err := NewStructure(save).ExportBinary(dir)
	if err != nil {
		t.Fatalf("ExportBinary() error = %v", err)
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if exported.StructureCount != 1 || exported.RowCount != 1 || exported.FaultCount != 0 || len(exported.Files) != 1 {
		t.Fatalf("ExportBinary() = %#v, want one structure row and no faults", exported)
	}
	if exported.Files[0].UUID != id.String() || exported.Files[0].Kind != "structure_binary" || exported.Files[0].Owner.TribeID != 555 {
		t.Fatalf("ExportBinary() file metadata = %#v", exported.Files[0])
	}
	binPath := filepath.Join(dir, "str_"+id.String()+".bin")
	got, err := os.ReadFile(binPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", binPath, err)
	}
	want, err := save.ObjectBinary(id)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", id, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("exported structure bytes = % x, want save ObjectBinary bytes", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "str_"+id.String()+"_location.json")); err != nil {
		t.Fatalf("stat structure location export: %v", err)
	}

	rawManifest, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		t.Fatalf("ReadFile(manifest.json) error = %v", err)
	}
	var manifest StructureBinaryExport
	if err := json.Unmarshal(rawManifest, &manifest); err != nil {
		t.Fatalf("json.Unmarshal(manifest) error = %v", err)
	}
	if manifest.StructureCount != 1 || manifest.RowCount != 1 || len(manifest.Files) != 1 {
		t.Fatalf("manifest = %#v, want exported structure summary", manifest)
	}
}

func TestExportStructureBinaryFromPathWritesStructureRowsAndManifest(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	dir := t.TempDir()
	exported, err := ExportStructureBinaryFromPath(save.Path(), dir)
	if err != nil {
		t.Fatalf("ExportStructureBinaryFromPath() error = %v", err)
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if exported.StructureCount != 1 || exported.RowCount != 1 || exported.FaultCount != 0 || len(exported.Files) != 1 {
		t.Fatalf("ExportStructureBinaryFromPath() = %#v, want one structure row and no faults", exported)
	}
	got, err := os.ReadFile(filepath.Join(dir, "str_"+id.String()+".bin"))
	if err != nil {
		t.Fatalf("ReadFile(exported structure) error = %v", err)
	}
	want, err := save.ObjectBinary(id)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", id, err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("exported structure bytes differ from save ObjectBinary bytes")
	}
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); err != nil {
		t.Fatalf("manifest missing: %v", err)
	}
}

func TestExportStructureBinaryFromPathReturnsErrorForInvalidSavePath(t *testing.T) {
	_, err := ExportStructureBinaryFromPath(filepath.Join(t.TempDir(), "missing.ark"), filepath.Join(t.TempDir(), "out"))
	if err == nil {
		t.Fatalf("ExportStructureBinaryFromPath() error = nil, want invalid save path error")
	}
}

func TestStructureAPIAllIncludesMissedInventoryContainersAndSkipsEngrams(t *testing.T) {
	save := openSyntheticStructureDiscoverySave(t)
	defer save.Close()

	api := NewStructure(save)
	structures, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}

	normalID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	missedID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	engramID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	if len(structures) != 2 {
		t.Fatalf("All() length = %d, want normal plus missed container: %#v", len(structures), structures)
	}
	if _, ok := structures[normalID]; !ok {
		t.Fatalf("All() missing normal structure %s: %#v", normalID, structures)
	}
	if got, ok := structures[missedID]; !ok || got.InventoryUUID == nil || got.ID != 456 {
		t.Fatalf("All() missed container = %#v, %v; want parsed inventory-bearing structure", got, ok)
	}
	if _, ok := structures[engramID]; ok {
		t.Fatalf("All() included engram structure %s: %#v", engramID, structures)
	}

	summary, _, err := api.OwnerSummaryWithFaults()
	if err != nil {
		t.Fatalf("OwnerSummaryWithFaults() error = %v", err)
	}
	if summary.Structures != 2 {
		t.Fatalf("OwnerSummaryWithFaults() structures = %d, want normal plus missed container only", summary.Structures)
	}
}

func TestStructureAPIGetOwnedByFiltersByOwner(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	structures, err := api.OwnedBy(arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("OwnedBy() error = %v", err)
	}
	if len(structures) != 1 {
		t.Fatalf("OwnedBy() length = %d, want 1", len(structures))
	}
}

func TestStructureAPICountOwnedByTribe(t *testing.T) {
	save := openSyntheticMixedOwnedStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	count, err := api.CountOwnedByTribe(555)
	if err != nil {
		t.Fatalf("CountOwnedByTribe() error = %v", err)
	}
	if count != 2 {
		t.Fatalf("CountOwnedByTribe(555) = %d, want 2", count)
	}

	count, err = api.CountOwnedByTribe(777)
	if err != nil {
		t.Fatalf("CountOwnedByTribe(777) error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountOwnedByTribe(777) = %d, want 1", count)
	}
}

func TestStructureAPIOwnerSummaryCountsOwnerFields(t *testing.T) {
	save := openSyntheticMixedOwnedStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	summary, faults, err := api.OwnerSummaryWithFaults()
	if err != nil {
		t.Fatalf("OwnerSummaryWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("OwnerSummaryWithFaults() faults = %#v, want none", faults)
	}
	if summary.Structures != 3 || summary.WithTribeID != 3 || summary.UniqueTribes != 2 {
		t.Fatalf("OwnerSummaryWithFaults() = %#v, want 3 structures, 3 tribe IDs, 2 unique tribes", summary)
	}
	if summary.WithPlayerID != 0 || summary.WithTribeName != 0 || summary.WithPlayerName != 0 || summary.WithOriginalPlacerID != 0 {
		t.Fatalf("OwnerSummaryWithFaults() owner optional counts = %#v, want only tribe IDs", summary)
	}
}

func TestStructureAPIHealthSummaryCountsDamagedStructures(t *testing.T) {
	save := openSyntheticStructureHealthSave(t)
	defer save.Close()

	api := NewStructure(save)
	summary, faults, err := api.HealthSummaryWithFaults()
	if err != nil {
		t.Fatalf("HealthSummaryWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("HealthSummaryWithFaults() faults = %#v, want none", faults)
	}
	want := StructureHealthSummary{
		Structures:           3,
		WithHealth:           2,
		Damaged:              1,
		TotalMaxHealth:       20000,
		TotalCurrentHealth:   19000,
		AverageHealthPercent: 95,
		MinimumHealthPercent: 90,
		MaximumHealthPercent: 100,
		FullyRepaired:        1,
		WithoutMaxHealth:     1,
	}
	if summary != want {
		t.Fatalf("HealthSummaryWithFaults() = %#v, want %#v", summary, want)
	}
}

func TestStructureSummaryFromPathHelpersReturnTypedSummariesAndFaults(t *testing.T) {
	healthSave := openSyntheticStructureHealthSave(t)
	defer healthSave.Close()

	health, faults, err := StructureHealthSummaryFromPath(healthSave.Path())
	if err != nil {
		t.Fatalf("StructureHealthSummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("StructureHealthSummaryFromPath() faults = %#v, want none", faults)
	}
	if health.Structures != 3 || health.WithHealth != 2 || health.Damaged != 1 {
		t.Fatalf("StructureHealthSummaryFromPath() = %#v, want damaged structure summary", health)
	}

	ownedSave := openSyntheticMixedOwnedStructureSave(t)
	defer ownedSave.Close()

	owners, faults, err := StructureOwnerSummaryFromPath(ownedSave.Path())
	if err != nil {
		t.Fatalf("StructureOwnerSummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("StructureOwnerSummaryFromPath() faults = %#v, want none", faults)
	}
	if owners.Structures != 3 || owners.WithTribeID != 3 || owners.UniqueTribes != 2 {
		t.Fatalf("StructureOwnerSummaryFromPath() = %#v, want 3 structures and 2 unique tribes", owners)
	}

	tribe, faults, err := StructureTribeOwnershipSummaryFromPath(ownedSave.Path(), 555)
	if err != nil {
		t.Fatalf("StructureTribeOwnershipSummaryFromPath() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("StructureTribeOwnershipSummaryFromPath() faults = %#v, want none", faults)
	}
	wantTribe := StructureTribeOwnershipSummary{TribeID: 555, Structures: 2}
	if tribe != wantTribe {
		t.Fatalf("StructureTribeOwnershipSummaryFromPath() = %#v, want %#v", tribe, wantTribe)
	}
}

func TestStructureAPIOwnerLocationsGroupsOwnedStructuresByRoundedMapCell(t *testing.T) {
	save := openSyntheticStructureOwnerLocationSave(t)
	defer save.Close()

	export, faults, err := NewStructure(save).OwnerLocationsWithFaults("Valguero", 1, nil)
	if err != nil {
		t.Fatalf("OwnerLocationsWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("OwnerLocationsWithFaults() faults = %#v, want none", faults)
	}
	if export.Structures != 3 || export.Owners != 2 || export.Cells != 2 || export.NamedCells != 1 || export.MultiStructureCells != 1 {
		t.Fatalf("OwnerLocationsWithFaults() = %#v, want 3 structures, 2 owners, 2 cells, 1 named, 1 multi", export)
	}
	if len(export.OwnersByLocation) != 2 {
		t.Fatalf("OwnersByLocation length = %d, want 2", len(export.OwnersByLocation))
	}
	if export.OwnersByLocation[0].Owner != "555" || len(export.OwnersByLocation[0].Cells) != 1 || export.OwnersByLocation[0].Cells[0].Count != 2 || export.OwnersByLocation[0].Cells[0].Name != "" {
		t.Fatalf("first owner bucket = %#v, want owner 555 with one multi-structure cell", export.OwnersByLocation[0])
	}
	if export.OwnersByLocation[1].Owner != "777" || len(export.OwnersByLocation[1].Cells) != 1 || export.OwnersByLocation[1].Cells[0].Count != 0 || export.OwnersByLocation[1].Cells[0].Name == "" {
		t.Fatalf("second owner bucket = %#v, want owner 777 with one named singleton cell", export.OwnersByLocation[1])
	}
}

func TestStructureOwnerLocationsFromPathWithFaultsUsesSharedSaveAPIs(t *testing.T) {
	save := openSyntheticStructureOwnerLocationSave(t)
	defer save.Close()

	export, faults, err := StructureOwnerLocationsFromPathWithFaults(save.Path(), "Valguero", 1)
	if err != nil {
		t.Fatalf("StructureOwnerLocationsFromPathWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("StructureOwnerLocationsFromPathWithFaults() faults = %#v, want none", faults)
	}
	if export.Structures != 3 || export.Owners != 2 || export.Cells != 2 || export.NamedCells != 1 || export.MultiStructureCells != 1 {
		t.Fatalf("StructureOwnerLocationsFromPathWithFaults() = %#v, want 3 structures, 2 owners, 2 cells, 1 named, 1 multi", export)
	}
	if len(export.OwnersByLocation) != 2 {
		t.Fatalf("OwnersByLocation length = %d, want 2", len(export.OwnersByLocation))
	}
}

func TestStructureAPIOwnerLocationsFullUsesParsedStructures(t *testing.T) {
	save := openSyntheticStructureOwnerLocationSave(t)
	defer save.Close()

	export, faults, err := NewStructure(save).OwnerLocationsFullWithFaults("Valguero", 1, nil)
	if err != nil {
		t.Fatalf("OwnerLocationsFullWithFaults() error = %v", err)
	}
	if len(faults) != 0 || export.FaultCount != 0 {
		t.Fatalf("OwnerLocationsFullWithFaults() faults = %d / %#v, want none", len(faults), export)
	}
	if export.Structures != 3 || export.Owners != 2 || export.Cells != 2 || export.NamedCells != 1 || export.MultiStructureCells != 1 {
		t.Fatalf("OwnerLocationsFullWithFaults() = %#v, want 3 structures, 2 owners, 2 cells, 1 named, 1 multi", export)
	}
	if len(export.OwnersByLocation) != 2 {
		t.Fatalf("OwnersByLocation length = %d, want 2", len(export.OwnersByLocation))
	}
	if export.OwnersByLocation[0].Owner != "555" || len(export.OwnersByLocation[0].Cells) != 1 || export.OwnersByLocation[0].Cells[0].Count != 2 || export.OwnersByLocation[0].Cells[0].Name != "" {
		t.Fatalf("first owner bucket = %#v, want owner 555 with one multi-structure cell", export.OwnersByLocation[0])
	}
	if export.OwnersByLocation[1].Owner != "777" || len(export.OwnersByLocation[1].Cells) != 1 || export.OwnersByLocation[1].Cells[0].Count != 0 || export.OwnersByLocation[1].Cells[0].Name == "" {
		t.Fatalf("second owner bucket = %#v, want owner 777 with one named singleton cell", export.OwnersByLocation[1])
	}
}

func TestStructureAPIOwnerLocationsReportsSkippedCandidates(t *testing.T) {
	save := openSyntheticStructureOwnerLocationSkippedSave(t)
	defer save.Close()

	export, faults, err := NewStructure(save).OwnerLocationsWithFaults("Valguero", 1, nil)
	if err != nil {
		t.Fatalf("OwnerLocationsWithFaults() error = %v", err)
	}
	if len(faults) != 0 || export.FaultCount != 0 {
		t.Fatalf("OwnerLocationsWithFaults() faults = %d / %#v, want no selected-property parse faults", len(faults), export)
	}
	if export.Structures != 3 || export.Owners != 1 || export.Cells != 1 || export.NamedCells != 1 || export.MultiStructureCells != 0 {
		t.Fatalf("OwnerLocationsWithFaults() = %#v, want one valid owner/location cell plus skipped candidates", export)
	}
	if export.SkippedWithoutOwner != 1 || export.SkippedWithoutLocation != 1 {
		t.Fatalf("OwnerLocationsWithFaults() skips = owner:%d location:%d, want 1/1", export.SkippedWithoutOwner, export.SkippedWithoutLocation)
	}
	if len(export.OwnersByLocation) != 1 || export.OwnersByLocation[0].Owner != "555" {
		t.Fatalf("OwnersByLocation = %#v, want one owner 555 bucket", export.OwnersByLocation)
	}
}

func TestStructureAPIFilterByOwnerFiltersProvidedStructures(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewStructure(save)
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	otherID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	first, ok, err := api.ByID(firstID)
	if err != nil {
		t.Fatalf("ByID() error = %v", err)
	}
	if !ok {
		t.Fatalf("ByID() ok = false, want true")
	}
	structures := map[uuid.UUID]arkobject.Structure{
		firstID: first,
		otherID: {
			ID:    999,
			Owner: arkobject.ObjectOwner{TribeID: 777},
		},
	}
	owner := arkobject.ObjectOwner{TribeID: 555}

	filtered, err := api.FilterByOwner(structures, &owner, 0, false)
	if err != nil {
		t.Fatalf("FilterByOwner() error = %v", err)
	}
	if len(filtered) != 1 || filtered[firstID].ID != 101 {
		t.Fatalf("FilterByOwner() = %#v, want first structure only", filtered)
	}

	byTribeID, err := api.FilterByOwner(structures, nil, 777, false)
	if err != nil {
		t.Fatalf("FilterByOwner(tribeID) error = %v", err)
	}
	if len(byTribeID) != 1 || byTribeID[otherID].ID != 999 {
		t.Fatalf("FilterByOwner(tribeID) = %#v, want other structure only", byTribeID)
	}
}

func TestStructureAPIFilterByOwnerPreservesUpstreamInvertAndValidation(t *testing.T) {
	api := StructureAPI{}
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	structures := map[uuid.UUID]arkobject.Structure{
		firstID:  {ID: 101, Owner: arkobject.ObjectOwner{TribeID: 555}},
		secondID: {ID: 102, Owner: arkobject.ObjectOwner{TribeID: 777}},
	}
	owner := arkobject.ObjectOwner{TribeID: 555}

	inverted, err := api.FilterByOwner(structures, &owner, 0, true)
	if err != nil {
		t.Fatalf("FilterByOwner(invert) error = %v", err)
	}
	if len(inverted) != 2 {
		t.Fatalf("FilterByOwner(invert) length = %d, want upstream-compatible all structures", len(inverted))
	}

	if _, err := api.FilterByOwner(structures, nil, 0, false); err == nil {
		t.Fatalf("FilterByOwner(no filter) error = nil, want error")
	}
}

func TestStructureAPIByClassOwnedByFiltersClassSubsetByOwner(t *testing.T) {
	save := openSyntheticMixedOwnedStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	vaultBlueprint := "Blueprint'/Game/Structures/Storage/PrimalStructureItemContainer_StorageBox_Huge.PrimalStructureItemContainer_StorageBox_Huge_C'"
	vaults, err := api.ByClassOwnedBy([]string{vaultBlueprint}, arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("ByClassOwnedBy() error = %v", err)
	}
	if len(vaults) != 1 {
		t.Fatalf("ByClassOwnedBy(vault, tribe 555) length = %d, want 1", len(vaults))
	}
	for _, structure := range vaults {
		if structure.Blueprint != vaultBlueprint || structure.Owner.TribeID != 555 {
			t.Fatalf("ByClassOwnedBy() structure = %#v", structure)
		}
	}
}

func TestStructureAPIGetByClassFiltersBlueprints(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	structures, err := api.ByClass([]string{"Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'"})
	if err != nil {
		t.Fatalf("ByClass() error = %v", err)
	}
	if len(structures) != 1 {
		t.Fatalf("ByClass() length = %d, want 1", len(structures))
	}
}

func TestStructureAPIGetByIDReturnsSingleStructure(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	structure, ok, err := api.ByID(id)
	if err != nil {
		t.Fatalf("ByID() error = %v", err)
	}
	if !ok {
		t.Fatalf("ByID() ok = false, want true")
	}
	if structure.ID != 123 || structure.Owner.TribeID != 555 || structure.Location == nil || structure.Location.X != 11 {
		t.Fatalf("ByID() structure = %#v", structure)
	}

	_, ok, err = api.ByID(uuid.MustParse("11111111-2222-3333-4444-555555555555"))
	if err != nil {
		t.Fatalf("ByID(missing) error = %v", err)
	}
	if ok {
		t.Fatalf("ByID(missing) ok = true, want false")
	}
}

func TestStructureAPIConnectedStructuresFollowsLinkedStructureUUIDs(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewStructure(save)
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	first, ok, err := api.ByID(firstID)
	if err != nil {
		t.Fatalf("ByID() error = %v", err)
	}
	if !ok {
		t.Fatalf("ByID() ok = false, want true")
	}

	connected, err := api.ConnectedStructures(map[uuid.UUID]arkobject.Structure{
		firstID: first,
	})
	if err != nil {
		t.Fatalf("ConnectedStructures() error = %v", err)
	}
	if len(connected) != 2 {
		t.Fatalf("ConnectedStructures() length = %d, want 2: %#v", len(connected), connected)
	}
	if _, ok := connected[firstID]; !ok {
		t.Fatalf("ConnectedStructures() missing seed structure %s", firstID)
	}
	if got, ok := connected[secondID]; !ok || got.ID != 102 {
		t.Fatalf("ConnectedStructures() linked structure = %#v, %v; want ID 102", got, ok)
	}
}

func TestStructureAPIGetAtLocationFiltersByMapCoordsAndClass(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	all, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	structure := all[id]
	coords := structure.Location.AsMapCoords("Valguero")

	nearby, err := api.AtLocation("Valguero", coords, 0.01, nil)
	if err != nil {
		t.Fatalf("AtLocation() error = %v", err)
	}
	if len(nearby) != 1 {
		t.Fatalf("AtLocation() length = %d, want 1", len(nearby))
	}

	filtered, err := api.AtLocation("Valguero", coords, 0.01, []string{structure.Blueprint})
	if err != nil {
		t.Fatalf("AtLocation(class) error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("AtLocation(class) length = %d, want 1", len(filtered))
	}

	missed, err := api.AtLocation("Valguero", arkobject.MapCoords{Lat: 1, Long: 1}, 0.01, nil)
	if err != nil {
		t.Fatalf("AtLocation(miss) error = %v", err)
	}
	if len(missed) != 0 {
		t.Fatalf("AtLocation(miss) length = %d, want 0", len(missed))
	}
}

func TestStructureAPIGetAtLocationWithFaultsKeepsValidStructures(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()

	api := NewStructure(save)
	all, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	if len(faults) != 1 {
		t.Fatalf("AllWithFaults() faults = %d, want 1", len(faults))
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	structure := all[id]
	coords := structure.Location.AsMapCoords("Valguero")

	nearby, faults, err := api.AtLocationWithFaults("Valguero", coords, 0.01, nil)
	if err != nil {
		t.Fatalf("AtLocationWithFaults() error = %v", err)
	}
	if len(faults) != 1 {
		t.Fatalf("AtLocationWithFaults() faults = %d, want 1", len(faults))
	}
	if len(nearby) != 1 {
		t.Fatalf("AtLocationWithFaults() length = %d, want 1", len(nearby))
	}
	if _, ok := nearby[id]; !ok {
		t.Fatalf("AtLocationWithFaults() missing valid structure %s", id)
	}
}

func TestStructureAPIAtLocationSummaryWithFaultsCountsNearbyAndConnected(t *testing.T) {
	save := openSyntheticBaseSave(t)
	defer save.Close()

	api := NewStructure(save)
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	first, ok, err := api.ByID(firstID)
	if err != nil {
		t.Fatalf("ByID() error = %v", err)
	}
	if !ok {
		t.Fatalf("ByID() ok = false, want true")
	}
	coords := first.Location.AsMapCoords("Valguero")

	summary, faults, err := api.AtLocationSummaryWithFaults("Valguero", coords, 0.01, nil)
	if err != nil {
		t.Fatalf("AtLocationSummaryWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("AtLocationSummaryWithFaults() faults = %#v, want none", faults)
	}
	if summary.Structures != 1 || summary.Connected != 2 {
		t.Fatalf("AtLocationSummaryWithFaults() = %#v, want 1 structure and 2 connected", summary)
	}
}

func TestStructureAPIFilterByLocationFiltersProvidedStructures(t *testing.T) {
	save := openSyntheticStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	all, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	structure := all[id]
	coords := structure.Location.AsMapCoords("Valguero")

	filtered := api.FilterByLocation("Valguero", coords, 0.01, map[uuid.UUID]arkobject.Structure{
		id: structure,
		uuid.MustParse("11111111-2222-3333-4444-555555555555"): {
			ID:       999,
			Location: &arkobject.ActorTransform{X: 999999, Y: 999999},
		},
	})
	if len(filtered) != 1 {
		t.Fatalf("FilterByLocation() length = %d, want 1: %#v", len(filtered), filtered)
	}
	if got := filtered[id]; got.ID != 123 {
		t.Fatalf("FilterByLocation() structure = %#v, want ID 123", got)
	}
}

func TestStructureAPIHeatmapCountsStructureMapCells(t *testing.T) {
	api := StructureAPI{}
	first := arkobject.MapCoords{Lat: 12.4, Long: 34.6}.AsActorTransform("Valguero")
	second := arkobject.MapCoords{Lat: 12.8, Long: 34.1}.AsActorTransform("Valguero")
	third := arkobject.MapCoords{Lat: 70.2, Long: 10.9}.AsActorTransform("Valguero")
	owner := arkobject.ObjectOwner{TribeID: 555}
	structures := map[uuid.UUID]arkobject.Structure{
		uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): {
			ID:        101,
			Blueprint: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			Owner:     owner,
			Location:  &first,
		},
		uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): {
			ID:        102,
			Blueprint: "Blueprint'/Game/Structures/Stone/PrimalStructure_Door_Stone.PrimalStructure_Door_Stone_C'",
			Owner:     owner,
			Location:  &second,
		},
		uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000"): {
			ID:        103,
			Blueprint: "Blueprint'/Game/Structures/Wood/PrimalStructure_Wall_Wood.PrimalStructure_Wall_Wood_C'",
			Owner:     arkobject.ObjectOwner{TribeID: 777},
			Location:  &third,
		},
	}

	heatmap, err := api.Heatmap("Valguero", 100, structures, nil, &owner, 2)
	if err != nil {
		t.Fatalf("Heatmap() error = %v", err)
	}
	if len(heatmap) != 100 || len(heatmap[0]) != 100 {
		t.Fatalf("Heatmap() dimensions = %dx%d, want 100x100", len(heatmap), len(heatmap[0]))
	}
	if heatmap[12][34] != 2 {
		t.Fatalf("Heatmap()[12][34] = %d, want 2", heatmap[12][34])
	}
	if heatmap[70][10] != 0 {
		t.Fatalf("Heatmap()[70][10] = %d, want owner-filtered and thresholded 0", heatmap[70][10])
	}
}

func TestStructureAPIHeatmapFiltersProvidedStructuresByClass(t *testing.T) {
	api := StructureAPI{}
	first := arkobject.MapCoords{Lat: 12.4, Long: 34.6}.AsActorTransform("Valguero")
	second := arkobject.MapCoords{Lat: 12.8, Long: 34.1}.AsActorTransform("Valguero")
	structures := map[uuid.UUID]arkobject.Structure{
		uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): {
			ID:        101,
			Blueprint: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			Location:  &first,
		},
		uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): {
			ID:        102,
			Blueprint: "Blueprint'/Game/Structures/Stone/PrimalStructure_Door_Stone.PrimalStructure_Door_Stone_C'",
			Location:  &second,
		},
	}

	heatmap, err := api.Heatmap("Valguero", 100, structures, []string{"Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'"}, nil, 1)
	if err != nil {
		t.Fatalf("Heatmap(class) error = %v", err)
	}
	if heatmap[12][34] != 1 {
		t.Fatalf("Heatmap(class)[12][34] = %d, want only one matching blueprint", heatmap[12][34])
	}

	if _, err := api.Heatmap("Valguero", 0, structures, nil, nil, 1); err == nil {
		t.Fatalf("Heatmap(resolution 0) error = nil, want error")
	}
}

func TestStructureAPIHeatmapSummaryWithFaultsCountsSaveStructures(t *testing.T) {
	save := openSyntheticStructureOwnerLocationSave(t)
	defer save.Close()

	summary, faults, err := NewStructure(save).HeatmapSummaryWithFaults(StructureHeatmapOptions{
		MapName:      "Valguero",
		Resolution:   100,
		MinInSection: 2,
	})
	if err != nil {
		t.Fatalf("HeatmapSummaryWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("HeatmapSummaryWithFaults() faults = %#v, want none", faults)
	}
	if summary.NonzeroCells != 1 || summary.Total != 2 || summary.Max != 2 || summary.Faults != 0 {
		t.Fatalf("HeatmapSummaryWithFaults() = %#v, want one populated cell with two structures", summary)
	}
}

func TestStructureAPISelectedHeatmapSummaryWithFaultsCountsSaveStructures(t *testing.T) {
	save := openSyntheticStructureOwnerLocationSave(t)
	defer save.Close()

	summary, faults, err := NewStructure(save).SelectedHeatmapSummaryWithFaults(StructureHeatmapOptions{
		MapName:      "Valguero",
		Resolution:   100,
		MinInSection: 2,
	})
	if err != nil {
		t.Fatalf("SelectedHeatmapSummaryWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("SelectedHeatmapSummaryWithFaults() faults = %#v, want none", faults)
	}
	if summary.NonzeroCells != 1 || summary.Total != 2 || summary.Max != 2 || summary.Faults != 0 {
		t.Fatalf("SelectedHeatmapSummaryWithFaults() = %#v, want one populated cell with two structures", summary)
	}
}

func TestStructureAPIContainerOfInventoryFindsInventoryBearingStructure(t *testing.T) {
	save := openSyntheticStructureWithInventorySave(t)
	defer save.Close()

	api := NewStructure(save)
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	id, structure, ok, err := api.ContainerOfInventory(inventoryID)
	if err != nil {
		t.Fatalf("ContainerOfInventory() error = %v", err)
	}
	if !ok {
		t.Fatalf("ContainerOfInventory() ok = false, want true")
	}
	if id != uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff") {
		t.Fatalf("ContainerOfInventory() id = %s", id)
	}
	if structure.InventoryUUID == nil || *structure.InventoryUUID != inventoryID || structure.Owner.TribeID != 555 {
		t.Fatalf("ContainerOfInventory() structure = %#v", structure)
	}

	_, _, ok, err = api.ContainerOfInventory(uuid.MustParse("11111111-2222-3333-4444-555555555555"))
	if err != nil {
		t.Fatalf("ContainerOfInventory(missing) error = %v", err)
	}
	if ok {
		t.Fatalf("ContainerOfInventory(missing) ok = true, want false")
	}
}

func TestStructureAPIAllWithFaultsKeepsValidStructuresAndReportsParseFaults(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()

	api := NewStructure(save)
	structures, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if len(structures) != 1 {
		t.Fatalf("AllWithFaults() structures length = %d, want 1", len(structures))
	}
	if _, ok := structures[id]; !ok {
		t.Fatalf("AllWithFaults() missing valid structure %s: %#v", id, structures)
	}
	if len(faults) != 1 || faults[0].ClassName != "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'" || faults[0].Err == nil {
		t.Fatalf("AllWithFaults() faults = %#v, want one structure parse fault", faults)
	}
}

func TestStructureAPIOwnedByWithFaultsKeepsOwnedStructuresAndReportsFaults(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()

	api := NewStructure(save)
	structures, faults, err := api.OwnedByWithFaults(arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("OwnedByWithFaults() error = %v", err)
	}
	id := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if len(structures) != 1 {
		t.Fatalf("OwnedByWithFaults() structures length = %d, want 1", len(structures))
	}
	if _, ok := structures[id]; !ok {
		t.Fatalf("OwnedByWithFaults() missing valid owned structure %s: %#v", id, structures)
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("OwnedByWithFaults() faults = %#v, want one structure parse fault", faults)
	}
}

func TestStructureAPICountOwnedByTribeWithFaultsUsesSelectedProperties(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()

	api := NewStructure(save)
	count, faults, err := api.CountOwnedByTribeWithFaults(555)
	if err != nil {
		t.Fatalf("CountOwnedByTribeWithFaults() error = %v", err)
	}
	if count != 1 {
		t.Fatalf("CountOwnedByTribeWithFaults() count = %d, want 1", count)
	}
	if len(faults) != 0 {
		t.Fatalf("CountOwnedByTribeWithFaults() faults = %#v, want no selected-property parse faults", faults)
	}
}

func TestStructureAPITribeOwnershipSummaryWithFaultsUsesSelectedProperties(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()

	api := NewStructure(save)
	summary, faults, err := api.TribeOwnershipSummaryWithFaults(555)
	if err != nil {
		t.Fatalf("TribeOwnershipSummaryWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("TribeOwnershipSummaryWithFaults() faults = %#v, want no selected-property parse faults", faults)
	}
	want := StructureTribeOwnershipSummary{TribeID: 555, Structures: 1}
	if summary != want {
		t.Fatalf("TribeOwnershipSummaryWithFaults() = %#v, want %#v", summary, want)
	}
}

func TestStructureAPISelectedPropertyPathsKeepFalseEngramFlag(t *testing.T) {
	save := openSyntheticFalseEngramStructureSave(t)
	defer save.Close()

	api := NewStructure(save)
	count, faults, err := api.CountOwnedByTribeWithFaults(555)
	if err != nil {
		t.Fatalf("CountOwnedByTribeWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("CountOwnedByTribeWithFaults() faults = %#v, want none", faults)
	}
	if count != 1 {
		t.Fatalf("CountOwnedByTribeWithFaults() count = %d, want explicit false bIsEngram structure only", count)
	}

	summary, faults, err := api.OwnerSummaryWithFaults()
	if err != nil {
		t.Fatalf("OwnerSummaryWithFaults() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("OwnerSummaryWithFaults() faults = %#v, want none", faults)
	}
	if summary.Structures != 1 || summary.WithTribeID != 1 || summary.UniqueTribes != 1 {
		t.Fatalf("OwnerSummaryWithFaults() = %#v, want explicit false bIsEngram structure only", summary)
	}
}

func openSyntheticStructureSave(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	otherID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	return openSyntheticSaveWith(t, "structures.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransforms(structureID),
	}, map[uuid.UUID][]byte{
		structureID: syntheticStructureObjectBytes(),
		otherID:     syntheticObjectBytes(0x10000001),
	})
}

func openSyntheticStructureDiscoverySave(t *testing.T) *arksave.Save {
	t.Helper()

	normalID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	missedID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	engramID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", nil, map[uuid.UUID][]byte{
		normalID: syntheticStructureObjectBytes(),
		missedID: syntheticStructureContainerObjectBytes(456, uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")),
		engramID: syntheticStructureEngramObjectBytes(),
	})
}

func openSyntheticMixedOwnedStructureSave(t *testing.T) *arksave.Save {
	t.Helper()

	firstVaultID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondVaultID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	wallID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", nil, map[uuid.UUID][]byte{
		firstVaultID:  syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 101, 555),
		secondVaultID: syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 102, 777),
		wallID:        syntheticStructureObjectBytesWithClassAndOwner(0x10000005, 103, 555),
	})
}

func openSyntheticStructureOwnerLocationSave(t *testing.T) *arksave.Save {
	t.Helper()

	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	thirdID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransformsFor(map[uuid.UUID][3]float64{
			firstID:  {100000, 100000, 33},
			secondID: {100000, 100000, 33},
			thirdID:  {-100000, -100000, 66},
		}),
	}, map[uuid.UUID][]byte{
		firstID:  syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 101, 555),
		secondID: syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 102, 555),
		thirdID:  syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 103, 777),
	})
}

func openSyntheticStructureOwnerLocationSkippedSave(t *testing.T) *arksave.Save {
	t.Helper()

	validID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	noOwnerID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	noLocationID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransformsFor(map[uuid.UUID][3]float64{
			validID:   {100000, 100000, 33},
			noOwnerID: {200000, 200000, 33},
		}),
	}, map[uuid.UUID][]byte{
		validID:      syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 101, 555),
		noOwnerID:    syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 102, 0),
		noLocationID: syntheticStructureObjectBytesWithClassAndOwner(0x10000051, 103, 777),
	})
}

func openSyntheticStructureHealthSave(t *testing.T) *arksave.Save {
	t.Helper()

	fullID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	damagedID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	noHealthID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", nil, map[uuid.UUID][]byte{
		fullID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			StructureID:   101,
			TribeID:       555,
			MaxHealth:     10000,
			CurrentHealth: 10000,
		}),
		damagedID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			StructureID:   102,
			TribeID:       555,
			MaxHealth:     10000,
			CurrentHealth: 9000,
		}),
		noHealthID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			StructureID: 103,
			TribeID:     555,
		}),
	})
}

func openSyntheticStructureSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransforms(structureID),
	}, map[uuid.UUID][]byte{
		structureID: syntheticStructureObjectBytes(),
		faultyID:    testfixtures.TruncatedObjectBytes(0x10000005),
	})
}

func openSyntheticFalseEngramStructureSave(t *testing.T) *arksave.Save {
	t.Helper()

	falseEngramID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	trueEngramID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", nil, map[uuid.UUID][]byte{
		falseEngramID: syntheticStructureObjectBytesWithEngramFlag(false),
		trueEngramID:  syntheticStructureObjectBytesWithEngramFlag(true),
	})
}

func openSyntheticStructureWithInventorySave(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	return openSyntheticSaveWith(t, "structures.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransforms(structureID),
	}, map[uuid.UUID][]byte{
		structureID: syntheticStructureWithInventoryObjectBytes(uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")),
	})
}

func syntheticStructureObjectBytes() []byte {
	return testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
		StructureID:   123,
		TribeID:       555,
		MaxHealth:     10000,
		CurrentHealth: 9000,
	})
}

func syntheticStructureObjectBytesWithEngramFlag(isEngram bool) []byte {
	return testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
		StructureID:   123,
		TribeID:       555,
		MaxHealth:     10000,
		CurrentHealth: 9000,
		IsEngram:      &isEngram,
	})
}

func syntheticStructureObjectBytesWithClassAndOwner(classID uint32, structureID int32, tribeID int32) []byte {
	return testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
		ClassID:       classID,
		StructureID:   structureID,
		TribeID:       tribeID,
		MaxHealth:     10000,
		CurrentHealth: 9000,
	})
}

func syntheticStructureWithInventoryObjectBytes(inventoryID uuid.UUID) []byte {
	return testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
		StructureID:   123,
		TribeID:       555,
		MaxHealth:     10000,
		CurrentHealth: 9000,
		InventoryID:   inventoryID,
		ItemCount:     12,
		MaxItemCount:  300,
	})
}

func syntheticStructureContainerObjectBytes(structureID int32, inventoryID uuid.UUID) []byte {
	return testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
		ClassID:     0x10000043,
		StructureID: structureID,
		TribeID:     555,
		MaxHealth:   10000,
		InventoryID: inventoryID,
	})
}

func syntheticStructureEngramObjectBytes() []byte {
	isEngram := true
	return testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
		StructureID: 789,
		TribeID:     555,
		IsEngram:    &isEngram,
	})
}

func syntheticStructureActorTransforms(id uuid.UUID) []byte {
	return syntheticStructureActorTransformsFor(map[uuid.UUID][3]float64{
		id: {11, 22, 33},
	})
}

func syntheticStructureActorTransformsFor(locations map[uuid.UUID][3]float64) []byte {
	transforms := make([]testfixtures.ActorTransform, 0, len(locations))
	for _, id := range sortedUUIDKeys(locations) {
		location := locations[id]
		transforms = append(transforms, testfixtures.ActorTransform{
			UUID:       id,
			X:          location[0],
			Y:          location[1],
			Z:          location[2],
			Quaternion: 1,
		})
	}
	return testfixtures.ActorTransforms(transforms...)
}
