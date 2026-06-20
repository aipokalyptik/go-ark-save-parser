package arkapi

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
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
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000b))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x1000000c, 100)
	if isBlueprint {
		writeBoolProperty(&buf, 0x1000000d, true)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func writeBoolProperty(buf *bytes.Buffer, name uint32, value bool) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000000e))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	if value {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
}
