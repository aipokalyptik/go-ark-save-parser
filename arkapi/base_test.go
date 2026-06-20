package arkapi

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
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

	path := filepath.Join(t.TempDir(), "base.ark")
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	otherID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "ActorTransforms", syntheticBaseActorTransforms(firstID, secondID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, firstID[:], syntheticBaseStructureObjectBytes(101, secondID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, secondID[:], syntheticBaseStructureObjectBytes(102, firstID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, otherID[:], syntheticObjectBytes(0x10000001))
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func openSyntheticBaseSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "base.ark")
	firstID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	secondID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "ActorTransforms", syntheticBaseActorTransforms(firstID, secondID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, firstID[:], syntheticBaseStructureObjectBytes(101, secondID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, secondID[:], syntheticBaseStructureObjectBytes(102, firstID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, faultyID[:], truncatedStructureObjectBytes())
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func syntheticBaseStructureObjectBytes(structureID int32, linked ...uuid.UUID) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000005))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000006, structureID)
	writeFloatProperty(&buf, 0x10000007, 10000)
	writeFloatProperty(&buf, 0x10000008, 9000)
	writeIntProperty(&buf, 0x10000009, 555)
	if len(linked) > 0 {
		writeObjectReferenceArrayProperty(&buf, 0x1000001d, linked)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
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
