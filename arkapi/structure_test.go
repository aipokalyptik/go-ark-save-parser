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

func openSyntheticStructureSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "structures.ark")
	structureID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	otherID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "ActorTransforms", syntheticStructureActorTransforms(structureID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, structureID[:], syntheticStructureObjectBytes())
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
