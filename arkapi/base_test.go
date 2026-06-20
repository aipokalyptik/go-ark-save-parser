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
		otherID:  syntheticObjectBytes(0x10000001),
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
		faultyID: truncatedStructureObjectBytes(),
	})
}

func syntheticBaseStructureObjectBytes(structureID int32, linked ...uuid.UUID) []byte {
	var props bytes.Buffer
	writeIntProperty(&props, 0x10000006, structureID)
	writeFloatProperty(&props, 0x10000007, 10000)
	writeFloatProperty(&props, 0x10000008, 9000)
	writeIntProperty(&props, 0x10000009, 555)
	if len(linked) > 0 {
		writeObjectReferenceArrayProperty(&props, 0x1000001d, linked)
	}
	return testfixtures.ObjectBytesWithProperties(0x10000005, 0x10000004, props.Bytes())
}

func syntheticBaseActorTransforms(first uuid.UUID, second uuid.UUID) []byte {
	var buf bytes.Buffer
	buf.Write(first[:])
	for _, value := range []float64{0, 0, 0, 0, 0, 0, 1} {
		_ = binary.Write(&buf, binary.LittleEndian, value)
	}
	buf.Write(second[:])
	for _, value := range []float64{1000, 1000, 0, 0, 0, 0, 1} {
		_ = binary.Write(&buf, binary.LittleEndian, value)
	}
	buf.Write(uuid.Nil[:])
	return buf.Bytes()
}

func writeObjectReferenceArrayProperty(buf *bytes.Buffer, name uint32, values []uuid.UUID) {
	dataSize := int32(len(values) * 18)
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001e))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(len(values)))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001f))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(dataSize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, id := range values {
		_ = binary.Write(buf, binary.LittleEndian, int16(0))
		buf.Write(id[:])
	}
}
