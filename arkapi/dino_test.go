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
		if dino.ColorSetIndices != [6]int{11, 0, 0, 44, 0, 0} {
			t.Fatalf("ColorSetIndices = %#v", dino.ColorSetIndices)
		}
		if dino.ColorSetNames != [6]string{"None", "Blue", "None", "None", "Black", "None"} {
			t.Fatalf("ColorSetNames = %#v", dino.ColorSetNames)
		}
		if dino.UploadedFromServerName != "TheIsland" {
			t.Fatalf("UploadedFromServerName = %q", dino.UploadedFromServerName)
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
		if len(dino.ParsedGeneTraits) != 2 || dino.ParsedGeneTraits[0].Name != "MutableMelee" || dino.ParsedGeneTraits[0].Level != 2 {
			t.Fatalf("dino parsed gene traits = %#v", dino.ParsedGeneTraits)
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

func TestDinoAPIReadsLinkedStatusComponentStats(t *testing.T) {
	save := openSyntheticDinoStatsSave(t)
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
		if dino.Stats == nil {
			t.Fatalf("Stats = nil, want parsed status component")
		}
		if dino.Stats.BaseLevel != 12 || dino.Stats.CurrentLevel != 12 {
			t.Fatalf("Dino stats levels = %#v", dino.Stats)
		}
		if dino.Stats.BaseStatPoints.Health != 5 || dino.Stats.AddedStatPoints.MeleeDamage != 2 {
			t.Fatalf("Dino stat points = %#v", dino.Stats)
		}
		if dino.Stats.ImprintingPercent != 87.5 {
			t.Fatalf("Dino imprinting = %f", dino.Stats.ImprintingPercent)
		}
	}
}

func TestDinoAPIFiltersByLevelAndStats(t *testing.T) {
	save := openSyntheticDinoStatsSave(t)
	defer save.Close()

	api := NewDino(save)
	highLevel, err := api.LevelAtLeast(12)
	if err != nil {
		t.Fatalf("LevelAtLeast() error = %v", err)
	}
	if len(highLevel) != 1 {
		t.Fatalf("LevelAtLeast(12) length = %d, want 1", len(highLevel))
	}
	tooHigh, err := api.LevelAtLeast(13)
	if err != nil {
		t.Fatalf("LevelAtLeast(13) error = %v", err)
	}
	if len(tooHigh) != 0 {
		t.Fatalf("LevelAtLeast(13) length = %d, want 0", len(tooHigh))
	}
	health, err := api.WithStatAtLeast(6, arkobject.DinoStatHealth)
	if err != nil {
		t.Fatalf("WithStatAtLeast(health) error = %v", err)
	}
	if len(health) != 1 {
		t.Fatalf("WithStatAtLeast(health) length = %d, want 1", len(health))
	}
	weight, err := api.WithStatAtLeast(8, arkobject.DinoStatWeight)
	if err != nil {
		t.Fatalf("WithStatAtLeast(weight) error = %v", err)
	}
	if len(weight) != 0 {
		t.Fatalf("WithStatAtLeast(weight) length = %d, want 0", len(weight))
	}
	baseHealth, err := api.WithBaseStatAtLeast(5, arkobject.DinoStatHealth)
	if err != nil {
		t.Fatalf("WithBaseStatAtLeast() error = %v", err)
	}
	if len(baseHealth) != 1 {
		t.Fatalf("WithBaseStatAtLeast() length = %d, want 1", len(baseHealth))
	}
	mutatedHealth, err := api.WithMutatedStatAtLeast(1, arkobject.DinoStatHealth)
	if err != nil {
		t.Fatalf("WithMutatedStatAtLeast() error = %v", err)
	}
	if len(mutatedHealth) != 1 {
		t.Fatalf("WithMutatedStatAtLeast() length = %d, want 1", len(mutatedHealth))
	}
	minLevel := int32(12)
	maxLevel := int32(12)
	tamed := false
	filtered, err := api.Filtered(DinoFilterOptions{
		MinLevel:    &minLevel,
		MaxLevel:    &maxLevel,
		Tamed:       &tamed,
		StatMinimum: 6,
		Stats:       []arkobject.DinoStat{arkobject.DinoStatHealth},
		Blueprints:  []string{"Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"},
	})
	if err != nil {
		t.Fatalf("Filtered() error = %v", err)
	}
	if len(filtered) != 1 {
		t.Fatalf("Filtered() length = %d, want 1", len(filtered))
	}
	maxLevel = 11
	filtered, err = api.Filtered(DinoFilterOptions{MaxLevel: &maxLevel})
	if err != nil {
		t.Fatalf("Filtered(max) error = %v", err)
	}
	if len(filtered) != 0 {
		t.Fatalf("Filtered(max) length = %d, want 0", len(filtered))
	}
}

func TestDinoAPIFiltersByGeneTrait(t *testing.T) {
	save := openSyntheticDinoDetailSave(t)
	defer save.Close()

	api := NewDino(save)
	byName, err := api.WithGeneTrait("MutableMelee")
	if err != nil {
		t.Fatalf("WithGeneTrait(name) error = %v", err)
	}
	if len(byName) != 1 {
		t.Fatalf("WithGeneTrait(name) length = %d, want 1", len(byName))
	}
	byLevel, err := api.WithGeneTrait("MutableMelee", 2)
	if err != nil {
		t.Fatalf("WithGeneTrait(name, level) error = %v", err)
	}
	if len(byLevel) != 1 {
		t.Fatalf("WithGeneTrait(name, level) length = %d, want 1", len(byLevel))
	}
	wrongLevel, err := api.WithGeneTrait("MutableMelee", 3)
	if err != nil {
		t.Fatalf("WithGeneTrait(wrong level) error = %v", err)
	}
	if len(wrongLevel) != 0 {
		t.Fatalf("WithGeneTrait(wrong level) length = %d, want 0", len(wrongLevel))
	}
	fallback, err := api.WithGeneTrait("Robust")
	if err != nil {
		t.Fatalf("WithGeneTrait(fallback) error = %v", err)
	}
	if len(fallback) != 1 {
		t.Fatalf("WithGeneTrait(fallback) length = %d, want 1", len(fallback))
	}
}

func TestDinoAPICountsByLevelClassAndTamedState(t *testing.T) {
	api := NewDino(nil)
	dinos := map[uuid.UUID]arkobject.Dino{
		uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			IsTamed:  true,
			Stats:    &arkobject.DinoStats{CurrentLevel: 12},
		},
		uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			IsTamed:  false,
			Stats:    &arkobject.DinoStats{CurrentLevel: 12},
		},
		uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000"): {
			Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Dodo/Dodo_Character_BP.Dodo_Character_BP_C'",
			IsTamed:  true,
			Stats:    &arkobject.DinoStats{CurrentLevel: 8},
		},
	}

	byLevel := api.CountByLevel(dinos)
	if byLevel[12] != 2 || byLevel[8] != 1 {
		t.Fatalf("CountByLevel() = %#v", byLevel)
	}
	byClass := api.CountByClass(dinos)
	if byClass["Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'"] != 2 ||
		byClass["Blueprint'/Game/PrimalEarth/Dinos/Dodo/Dodo_Character_BP.Dodo_Character_BP_C'"] != 1 {
		t.Fatalf("CountByClass() = %#v", byClass)
	}
	byTamed := api.CountByTamed(dinos)
	if byTamed[true] != 2 || byTamed[false] != 1 {
		t.Fatalf("CountByTamed() = %#v", byTamed)
	}
}

func openSyntheticDinoStatsSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "dinos.ark")
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	statusID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, dinoID[:], syntheticDinoStatsObjectBytes(statusID))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, statusID[:], syntheticDinoStatusObjectBytes())
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func syntheticDinoStatsObjectBytes(statusID uuid.UUID) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000015, 1001)
	writeIntProperty(&buf, 0x10000016, 2002)
	writeObjectReferenceProperty(&buf, 0x10000035, statusID)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticDinoStatusObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000036))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x10000037, 12)
	writePositionedIntProperty(&buf, 0x10000038, 0, 5)
	writePositionedIntProperty(&buf, 0x10000038, 7, 3)
	writePositionedIntProperty(&buf, 0x10000039, 8, 2)
	writePositionedIntProperty(&buf, 0x1000003a, 0, 1)
	writePositionedFloatProperty(&buf, 0x1000003b, 0, 1234.5)
	writePositionedFloatProperty(&buf, 0x1000003b, 7, 321.25)
	writeFloatProperty(&buf, 0x1000003c, 0.875)
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
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
	writePositionedInt8Property(&buf, 0x1000002e, 0, 11)
	writePositionedInt8Property(&buf, 0x1000002e, 3, 44)
	writePositionedNameProperty(&buf, 0x1000002f, 1, 0x10000034)
	writePositionedNameProperty(&buf, 0x1000002f, 4, 0x10000031)
	writeStringProperty(&buf, 0x10000030, "\nTheIsland")
	writeNameArrayProperty(&buf, 0x1000003d, []uint32{0x1000003e, 0x1000003f})
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

func writePositionedInt8Property(buf *bytes.Buffer, name uint32, position int32, value int8) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000032))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	_ = binary.Write(buf, binary.LittleEndian, int32(position))
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(byte(value))
}

func writePositionedNameProperty(buf *bytes.Buffer, name uint32, position int32, valueName uint32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000033))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(position))
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, valueName)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
}

func writePositionedIntProperty(buf *bytes.Buffer, name uint32, position int32, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000003))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writePositionedFloatProperty(buf *bytes.Buffer, name uint32, position int32, value float32) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000000a))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeNameArrayProperty(buf *bytes.Buffer, name uint32, values []uint32) {
	dataSize := uint32(4 + len(values)*8)
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001e))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(dataSize))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000033))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, dataSize)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		_ = binary.Write(buf, binary.LittleEndian, value)
		_ = binary.Write(buf, binary.LittleEndian, int32(0))
	}
}
