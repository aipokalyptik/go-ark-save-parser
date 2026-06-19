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

func TestEquipmentAPIClassifiesBlueprints(t *testing.T) {
	api := EquipmentAPI{}
	for blueprint, want := range map[string]arkobject.EquipmentKind{
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'":                         arkobject.EquipmentWeapon,
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_Saddle.PrimalItemArmor_Saddle_C'":         arkobject.EquipmentSaddle,
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'":   arkobject.EquipmentArmor,
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_WoodShield.PrimalItemArmor_WoodShield_C'": arkobject.EquipmentShield,
	} {
		got := api.KindForBlueprint(blueprint)
		if got != want {
			t.Fatalf("KindForBlueprint(%q) = %s, want %s", blueprint, got, want)
		}
	}
	if api.IsApplicableBlueprint("Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItemAmmo_ArrowStone.PrimalItemAmmo_ArrowStone_C'") {
		t.Fatalf("IsApplicableBlueprint(ammo) = true, want false")
	}
}

func TestEquipmentAPIAllAndByKindReadLocalSaveItems(t *testing.T) {
	save := openSyntheticEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	items, err := api.All()
	if err != nil {
		t.Fatalf("All() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("All() length = %d, want 1", len(items))
	}
	for _, item := range items {
		if item.Kind != arkobject.EquipmentWeapon || item.Rating != 7.5 || item.Quality != 3 {
			t.Fatalf("Equipment item = %#v", item)
		}
	}
	weapons, err := api.ByKind(arkobject.EquipmentWeapon)
	if err != nil {
		t.Fatalf("ByKind(weapon) error = %v", err)
	}
	if len(weapons) != 1 {
		t.Fatalf("ByKind(weapon) length = %d, want 1", len(weapons))
	}
}

func TestEquipmentAPIByCrafterFiltersLocalSaveItems(t *testing.T) {
	save := openSyntheticEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	items, err := api.ByCrafter(arkobject.ObjectCrafter{CharacterName: "Survivor", TribeName: "Porters"})
	if err != nil {
		t.Fatalf("ByCrafter() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ByCrafter() length = %d, want 1", len(items))
	}
	items, err = api.ByCrafter(arkobject.ObjectCrafter{CharacterName: "Other", TribeName: "Porters"})
	if err != nil {
		t.Fatalf("ByCrafter(other) error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("ByCrafter(other) length = %d, want 0", len(items))
	}
}

func openSyntheticEquipmentSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "equipment.ark")
	weaponID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	engramID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, weaponID[:], syntheticEquipmentObjectBytes(false))
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, engramID[:], syntheticEquipmentObjectBytes(true))
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func syntheticEquipmentObjectBytes(isEngram bool) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000f))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x1000000c, 1)
	writeFloatProperty(&buf, 0x10000010, 7.5)
	writeIntProperty(&buf, 0x10000011, 3)
	writeFloatProperty(&buf, 0x10000012, 0.75)
	writeStringProperty(&buf, 0x1000001b, "Survivor")
	writeStringProperty(&buf, 0x1000001c, "Porters")
	if isEngram {
		writeBoolProperty(&buf, 0x10000013, true)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func writeStringProperty(buf *bytes.Buffer, name uint32, value string) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x1000001a))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+5))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	writeArkString(buf, value)
}
