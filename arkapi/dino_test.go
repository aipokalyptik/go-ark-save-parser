package arkapi

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func TestDinoAPIRecognizesApplicableBlueprints(t *testing.T) {
	api := DinoAPI{}
	for _, blueprint := range []string{
		"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		"Blueprint'/Game/ASA/Creatures/Foo/Foo_Character_BP.Foo_Character_BP_C'",
		"Blueprint'/Game/Mods/SDinoVariants/Bar/Bar_Character_BP.Bar_Character_BP_C'",
	} {
		if !api.IsApplicableBlueprint(blueprint) {
			t.Fatalf("IsApplicableBlueprint(%q) = false, want true", blueprint)
		}
	}
	if api.IsApplicableBlueprint("Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'") {
		t.Fatalf("IsApplicableBlueprint(weapon) = true, want false")
	}
}

func TestDinoAPIAllAndByClassReadLocalSaveDinos(t *testing.T) {
	save := openSyntheticDinoSave(t)
	defer save.Close()

	api := NewDino(save)
	dinos, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("All() length = %d, want 1", len(dinos))
	}
	for _, dino := range dinos {
		if dino.ID1 != 1001 || !dino.IsTamed || !dino.IsFemale || dino.Location == nil || dino.Location.X != 11 {
			t.Fatalf("Dino = %#v", dino)
		}
	}
	filtered, err := api.ByClass([]string{"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"})
	if err != nil {
		t.Fatalf("ByClass() error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("ByClass() length = %d, want 1", len(filtered))
	}
}

func openSyntheticDinoSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "dinos.ark")
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	otherID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "ActorTransforms", syntheticStructureActorTransforms(dinoID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, dinoID[:], syntheticDinoObjectBytes())
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

func syntheticDinoObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, 1001)
	writeIntProperty(&buf, 0x10000016, 2002)
	writeBoolProperty(&buf, 0x10000017, true)
	writeDoubleProperty(&buf, 0x10000018, 42)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func writeDoubleProperty(buf *bytes.Buffer, name uint32, value float64) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000019))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}
