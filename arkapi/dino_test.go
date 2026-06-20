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

func TestDinoAPIFiltersBySexDeathAndBabyState(t *testing.T) {
	save := openSyntheticDinoFilterSave(t)
	defer save.Close()

	api := NewDino(save)
	females, err := api.Females()
	if err != nil {
		t.Fatalf("Females() error = %v", err)
	}
	if len(females) != 1 {
		t.Fatalf("Females() length = %d, want 1", len(females))
	}
	males, err := api.Males()
	if err != nil {
		t.Fatalf("Males() error = %v", err)
	}
	if len(males) != 1 {
		t.Fatalf("Males() length = %d, want 1", len(males))
	}
	dead, err := api.Dead()
	if err != nil {
		t.Fatalf("Dead() error = %v", err)
	}
	if len(dead) != 1 {
		t.Fatalf("Dead() length = %d, want 1", len(dead))
	}
	alive, err := api.Alive()
	if err != nil {
		t.Fatalf("Alive() error = %v", err)
	}
	if len(alive) != 1 {
		t.Fatalf("Alive() length = %d, want 1", len(alive))
	}
	babies, err := api.Babies()
	if err != nil {
		t.Fatalf("Babies() error = %v", err)
	}
	if len(babies) != 1 {
		t.Fatalf("Babies() length = %d, want 1", len(babies))
	}
}

func TestDinoAPIReadsTamedDetailsAndOwner(t *testing.T) {
	save := openSyntheticDinoDetailSave(t)
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
		if dino.TamedName != "Blue" || !dino.IsNeutered {
			t.Fatalf("dino tamed details = %#v", dino)
		}
		if dino.InventoryUUID == nil || dino.InventoryUUID.String() != "99999999-aaaa-bbbb-cccc-ddddeeeeffff" {
			t.Fatalf("InventoryUUID = %v", dino.InventoryUUID)
		}
		if dino.Owner.TribeName != "Porters" || dino.Owner.TamerTribeID != 555 || dino.Owner.TargetTeam != 555 {
			t.Fatalf("dino owner tribe fields = %#v", dino.Owner)
		}
		if dino.Owner.PlayerName != "Survivor" || dino.Owner.PlayerID != 42 || dino.Owner.ImprinterUniqueID != "eos-survivor" {
			t.Fatalf("dino owner player fields = %#v", dino.Owner)
		}
	}
}

func TestDinoAPIReadsBabyMaturationStage(t *testing.T) {
	save := openSyntheticDinoBabyStageSave(t)
	defer save.Close()

	api := NewDino(save)
	babies, err := api.Babies()
	if err != nil {
		t.Fatalf("Babies() error = %v", err)
	}
	if len(babies) != 1 {
		t.Fatalf("Babies() length = %d, want 1", len(babies))
	}
	for _, dino := range babies {
		if dino.MaturationPercent != 75 || dino.BabyStage != arkobject.BabyStageAdolescent {
			t.Fatalf("baby maturation = %#v", dino)
		}
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

func openSyntheticDinoFilterSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "dinos.ark")
	femaleAliveID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	maleDeadID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, femaleAliveID[:], syntheticDinoObjectBytesWithFlags(1001, 2002, true, false, false, true))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, maleDeadID[:], syntheticDinoObjectBytesWithFlags(3003, 4004, false, true, true, false))
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func openSyntheticDinoDetailSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "dinos.ark")
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, dinoID[:], syntheticDinoDetailObjectBytes())
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func openSyntheticDinoBabyStageSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "dinos.ark")
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, dinoID[:], syntheticDinoBabyObjectBytes(0.75))
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
	return syntheticDinoObjectBytesWithFlags(1001, 2002, true, false, false, true)
}

func syntheticDinoDetailObjectBytes() []byte {
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
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
	writeObjectReferenceProperty(&buf, 0x10000023, inventoryID)
	writeStringProperty(&buf, 0x10000024, "Blue")
	writeBoolProperty(&buf, 0x10000025, true)
	writeStringProperty(&buf, 0x10000026, "Porters")
	writeIntProperty(&buf, 0x10000027, 555)
	writeStringProperty(&buf, 0x10000028, "Porters")
	writeStringProperty(&buf, 0x10000029, "Survivor")
	writeStringProperty(&buf, 0x1000002a, "Survivor")
	writeStringProperty(&buf, 0x1000002b, "eos-survivor")
	writeIntProperty(&buf, 0x1000002c, 42)
	writeIntProperty(&buf, 0x10000009, 555)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticDinoBabyObjectBytes(age float32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, 1001)
	writeIntProperty(&buf, 0x10000016, 2002)
	writeBoolProperty(&buf, 0x10000021, true)
	writeFloatProperty(&buf, 0x1000002d, age)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticDinoObjectBytesWithFlags(id1 int32, id2 int32, isFemale bool, isDead bool, isBaby bool, isTamed bool) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, id1)
	writeIntProperty(&buf, 0x10000016, id2)
	writeBoolProperty(&buf, 0x10000017, isFemale)
	writeBoolProperty(&buf, 0x10000020, isDead)
	writeBoolProperty(&buf, 0x10000021, isBaby)
	if isTamed {
		writeDoubleProperty(&buf, 0x10000018, 42)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func writeObjectReferenceProperty(buf *bytes.Buffer, name uint32, id uuid.UUID) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001f))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(18))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, int16(0))
	buf.Write(id[:])
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
