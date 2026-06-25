package arkapi

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestStackableAPIRecognizesApplicableBlueprints(t *testing.T) {
	api := StackableAPI{}
	for _, blueprint := range []string{
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Consumables/PrimalItemConsumable_Berry.PrimalItemConsumable_Berry_C'",
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItemAmmo_ArrowStone.PrimalItemAmmo_ArrowStone_C'",
	} {
		if !api.IsApplicableBlueprint(blueprint) {
			t.Fatalf("IsApplicableBlueprint(%q) = false, want true", blueprint)
		}
	}
	if api.IsApplicableBlueprint("Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'") {
		t.Fatalf("IsApplicableBlueprint(structure) = true, want false")
	}
	if api.IsApplicableBlueprint("/ArkOmega/Buffs/Variants/Other/PrimalItemResource_Crystal_Poop.PrimalItemResource_Crystal_Poop_C") {
		t.Fatalf("IsApplicableBlueprint(modded resource outside Resources directory) = true, want false")
	}
}

func TestStackableAPICountSumsQuantities(t *testing.T) {
	api := StackableAPI{}
	count := api.Count(map[uuid.UUID]arkobject.InventoryItem{
		uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"): {Quantity: 100},
		uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff"): {Quantity: 25},
	})
	if count != 125 {
		t.Fatalf("Count() = %d, want 125", count)
	}
}

func TestStackableAPIAllAndByClassReadLocalSaveItems(t *testing.T) {
	save := openSyntheticStackableSave(t)
	defer save.Close()

	api := NewStackable(save)
	items, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("All() length = %d, want 1", len(items))
	}
	if api.Count(items) != 100 {
		t.Fatalf("Count(All()) = %d, want 100", api.Count(items))
	}

	filtered, err := api.ByClass([]string{"Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'"})
	if err != nil {
		t.Fatalf("ByClass() error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("ByClass() length = %d, want 1", len(filtered))
	}

	resources, err := api.Resources()
	if err != nil {
		t.Fatalf("Resources() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Resources() length = %d, want 1", len(resources))
	}
	ammo, err := api.Ammo()
	if err != nil {
		t.Fatalf("Ammo() error = %v", err)
	}
	if len(ammo) != 0 {
		t.Fatalf("Ammo() length = %d, want 0", len(ammo))
	}
	consumables, err := api.Consumables()
	if err != nil {
		t.Fatalf("Consumables() error = %v", err)
	}
	if len(consumables) != 0 {
		t.Fatalf("Consumables() length = %d, want 0", len(consumables))
	}

	typedAll, err := api.AllStackables()
	if err != nil {
		t.Fatalf("AllStackables() error = %v", err)
	}
	if len(typedAll) != len(items) || api.CountStackables(typedAll) != api.Count(items) {
		t.Fatalf("AllStackables() = len %d count %d, want len %d count %d", len(typedAll), api.CountStackables(typedAll), len(items), api.Count(items))
	}
	for id, item := range typedAll {
		if item.UUID != id || item.Blueprint == "" || item.Quantity != items[id].Quantity {
			t.Fatalf("typed stackable %s = %#v, source %#v", id, item, items[id])
		}
	}

	typedFiltered, err := api.ByClassStackables([]string{"Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'"})
	if err != nil {
		t.Fatalf("ByClassStackables() error = %v", err)
	}
	if len(typedFiltered) != 1 || api.CountStackables(typedFiltered) != 100 {
		t.Fatalf("ByClassStackables() = len %d count %d, want one stack of 100", len(typedFiltered), api.CountStackables(typedFiltered))
	}
}

func TestStackableAPIByClassCanReadExactModdedResourceWithoutBroadeningDefaultScan(t *testing.T) {
	save := openSyntheticMixedStackableSave(t)
	defer save.Close()

	api := NewStackable(save)
	all, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(all) != 1 || api.Count(all) != 100 {
		t.Fatalf("All() = len %d count %d, want only default-applicable vanilla stackable", len(all), api.Count(all))
	}

	moddedBlueprint := "/ArkOmega/Buffs/Variants/Other/PrimalItemResource_Crystal_Poop.PrimalItemResource_Crystal_Poop_C"
	modded, err := api.ByClass([]string{moddedBlueprint})
	if err != nil {
		t.Fatalf("ByClass(modded) error = %v", err)
	}
	if len(modded) != 1 || api.Count(modded) != 100 {
		t.Fatalf("ByClass(modded) = len %d count %d, want exact modded stackable", len(modded), api.Count(modded))
	}
}

func TestStackableAPIAllWithFaultsKeepsValidItemsAndReportsParseFaults(t *testing.T) {
	save := openSyntheticStackableSaveWithFault(t)
	defer save.Close()

	api := NewStackable(save)
	items, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	if len(items) != 1 || api.Count(items) != 100 {
		t.Fatalf("AllWithFaults() items = %#v, want one valid stackable quantity 100", items)
	}
	if len(faults) != 1 || faults[0].ClassName != "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'" || faults[0].Err == nil {
		t.Fatalf("AllWithFaults() faults = %#v, want one stackable parse fault", faults)
	}
	typed, typedFaults, err := api.AllStackablesWithFaults()
	if err != nil {
		t.Fatalf("AllStackablesWithFaults() error = %v", err)
	}
	if len(typed) != len(items) || api.CountStackables(typed) != api.Count(items) || len(typedFaults) != len(faults) {
		t.Fatalf("AllStackablesWithFaults() items=%#v faults=%#v, want parity with InventoryItem path", typed, typedFaults)
	}
}

func TestStackableAPIFilterOwnedByCountsItemsThroughOwnerInventory(t *testing.T) {
	save := openSyntheticStackableOwnedByStructureSave(t)
	defer save.Close()

	api := NewStackable(save)
	items, err := api.ByClass([]string{"Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'"})
	if err != nil {
		t.Fatalf("ByClass() error = %v", err)
	}
	owned, err := api.FilterOwnedBy(items, arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("FilterOwnedBy() error = %v", err)
	}
	if len(owned) != 1 || api.Count(owned) != 100 {
		t.Fatalf("FilterOwnedBy() = len %d count %d, want one owned stack of 100", len(owned), api.Count(owned))
	}

	allOwned, err := api.OwnedBy(arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("OwnedBy() error = %v", err)
	}
	if len(allOwned) != 1 || api.Count(allOwned) != 100 {
		t.Fatalf("OwnedBy() = len %d count %d, want one owned stack of 100", len(allOwned), api.Count(allOwned))
	}

	typedItems, err := api.ByClassStackables([]string{"Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'"})
	if err != nil {
		t.Fatalf("ByClassStackables() error = %v", err)
	}
	typedOwned, err := api.FilterOwnedByStackables(typedItems, arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("FilterOwnedByStackables() error = %v", err)
	}
	if len(typedOwned) != 1 || api.CountStackables(typedOwned) != 100 {
		t.Fatalf("FilterOwnedByStackables() = len %d count %d, want one owned stack of 100", len(typedOwned), api.CountStackables(typedOwned))
	}
	allTypedOwned, err := api.OwnedByStackables(arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("OwnedByStackables() error = %v", err)
	}
	if len(allTypedOwned) != 1 || api.CountStackables(allTypedOwned) != 100 {
		t.Fatalf("OwnedByStackables() = len %d count %d, want one owned stack of 100", len(allTypedOwned), api.CountStackables(allTypedOwned))
	}
}

func TestStackableAPIFilterOwnedByIgnoresMalformedUnrelatedContainers(t *testing.T) {
	save := openSyntheticStackableOwnedByStructureSaveWithFault(t)
	defer save.Close()

	api := NewStackable(save)
	items, err := api.ByClass([]string{"Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'"})
	if err != nil {
		t.Fatalf("ByClass() error = %v", err)
	}
	owned, err := api.FilterOwnedBy(items, arkobject.ObjectOwner{TribeID: 555})
	if err != nil {
		t.Fatalf("FilterOwnedBy() error = %v", err)
	}
	if len(owned) != 1 || api.Count(owned) != 100 {
		t.Fatalf("FilterOwnedBy() = len %d count %d, want one owned stack of 100", len(owned), api.Count(owned))
	}
}

func openSyntheticStackableOwnedByStructureSave(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	ownedItemID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	otherItemID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	otherInventoryID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	return openSyntheticSaveWith(t, "stackables.ark", nil, map[uuid.UUID][]byte{
		structureID: syntheticStructureWithInventoryObjectBytes(inventoryID),
		ownedItemID: syntheticStackableObjectBytesWithOwnerInventory(inventoryID),
		otherItemID: syntheticStackableObjectBytesWithOwnerInventory(otherInventoryID),
	})
}

func openSyntheticStackableOwnedByStructureSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	ownedItemID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	return openSyntheticSaveWith(t, "stackables.ark", nil, map[uuid.UUID][]byte{
		structureID: syntheticStructureWithInventoryObjectBytes(inventoryID),
		faultyID:    truncatedStructureObjectBytes(),
		ownedItemID: syntheticStackableObjectBytesWithOwnerInventory(inventoryID),
	})
}

func openSyntheticMixedStackableSave(t *testing.T) *arksave.Save {
	t.Helper()

	vanillaID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	moddedID := uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111")
	return openSyntheticSaveWith(t, "stackables.ark", nil, map[uuid.UUID][]byte{
		vanillaID: syntheticStackableObjectBytesWithClass(0x1000000b, false),
		moddedID:  syntheticStackableObjectBytesWithClass(0x10000043, false),
	})
}

func openSyntheticStackableSave(t *testing.T) *arksave.Save {
	t.Helper()

	itemID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	blueprintID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	return openSyntheticSaveWith(t, "stackables.ark", nil, map[uuid.UUID][]byte{
		itemID:      syntheticStackableObjectBytes(false),
		blueprintID: syntheticStackableObjectBytes(true),
	})
}

func openSyntheticStackableSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	itemID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "stackables.ark", nil, map[uuid.UUID][]byte{
		itemID:   syntheticStackableObjectBytes(false),
		faultyID: truncatedStackableObjectBytes(),
	})
}

func truncatedStackableObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000b))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticStackableObjectBytes(isBlueprint bool) []byte {
	return syntheticStackableObjectBytesWithClass(0x1000000b, isBlueprint)
}

func syntheticStackableObjectBytesWithClass(className uint32, isBlueprint bool) []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x1000000c, 100)
	if isBlueprint {
		testfixtures.WriteBoolPropertyID(&props, 0x1000000d, 0x1000000e, true)
	}
	return testfixtures.ObjectBytesWithProperties(className, 0x10000004, props.Bytes())
}

func syntheticStackableObjectBytesWithOwnerInventory(ownerInventory uuid.UUID) []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x1000000c, 100)
	testfixtures.WriteObjectReferencePropertyID(&props, 0x10000044, 0x1000001f, ownerInventory)
	return testfixtures.ObjectBytesWithProperties(0x1000000b, 0x10000004, props.Bytes())
}
