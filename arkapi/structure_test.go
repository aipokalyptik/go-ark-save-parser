package arkapi

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
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

func openSyntheticStructureSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "structures.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransforms(structureID),
	}, map[uuid.UUID][]byte{
		structureID: syntheticStructureObjectBytes(),
		faultyID:    truncatedStructureObjectBytes(),
	})
}

func syntheticStructureObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000005))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))

	writeIntProperty(&buf, 0x10000006, 123)
	writeFloatProperty(&buf, 0x10000007, 10000)
	writeFloatProperty(&buf, 0x10000008, 9000)
	writeIntProperty(&buf, 0x10000009, 555)

	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func truncatedStructureObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000005))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticStructureActorTransforms(id uuid.UUID) []byte {
	var buf bytes.Buffer
	buf.Write(id[:])
	for _, value := range []float64{11, 22, 33, 0, 0, 0, 1} {
		_ = binary.Write(&buf, binary.LittleEndian, value)
	}
	buf.Write(uuid.Nil[:])
	return buf.Bytes()
}

func writeIntProperty(buf *bytes.Buffer, name uint32, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000003))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeFloatProperty(buf *bytes.Buffer, name uint32, value float32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000000a))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}
