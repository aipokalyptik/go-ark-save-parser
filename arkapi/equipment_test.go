package arkapi

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
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
		if item.Stats.Internal[arkobject.EquipmentStatDamage] != 1234 || item.Stats.Damage != 112.3 {
			t.Fatalf("Equipment stats = %#v", item.Stats)
		}
	}
	weapons, err := api.ByKind(arkobject.EquipmentWeapon)
	if err != nil {
		t.Fatalf("ByKind(weapon) error = %v", err)
	}
	if len(weapons) != 1 {
		t.Fatalf("ByKind(weapon) length = %d, want 1", len(weapons))
	}
	weapons, err = api.Weapons()
	if err != nil {
		t.Fatalf("Weapons() error = %v", err)
	}
	if len(weapons) != 1 {
		t.Fatalf("Weapons() length = %d, want 1", len(weapons))
	}
}

func TestEquipmentAPIAllWithFaultsKeepsValidItemsAndReportsParseFaults(t *testing.T) {
	save := openSyntheticEquipmentSaveWithFault(t)
	defer save.Close()

	api := NewEquipment(save)
	items, faults, err := api.AllWithFaults()
	if err != nil {
		t.Fatalf("AllWithFaults() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("AllWithFaults() items length = %d, want 1", len(items))
	}
	for _, item := range items {
		if item.Kind != arkobject.EquipmentWeapon || item.Rating != 7.5 || item.Quality != 3 {
			t.Fatalf("Equipment item = %#v", item)
		}
	}
	if len(faults) != 1 || faults[0].ClassName != "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'" || faults[0].Err == nil {
		t.Fatalf("AllWithFaults() faults = %#v, want one equipment parse fault", faults)
	}
}

func TestEquipmentAPIReadsArmorStatValues(t *testing.T) {
	save := openSyntheticArmorEquipmentSave(t)
	defer save.Close()

	api := NewEquipment(save)
	armor, err := api.ByKind(arkobject.EquipmentArmor)
	if err != nil {
		t.Fatalf("ByKind(armor) error = %v", err)
	}
	if len(armor) != 1 {
		t.Fatalf("ByKind(armor) length = %d, want 1", len(armor))
	}
	armor, err = api.Armor()
	if err != nil {
		t.Fatalf("Armor() error = %v", err)
	}
	if len(armor) != 1 {
		t.Fatalf("Armor() length = %d, want 1", len(armor))
	}
	saddles, err := api.Saddles()
	if err != nil {
		t.Fatalf("Saddles() error = %v", err)
	}
	if len(saddles) != 0 {
		t.Fatalf("Saddles() length = %d, want 0", len(saddles))
	}
	shields, err := api.Shields()
	if err != nil {
		t.Fatalf("Shields() error = %v", err)
	}
	if len(shields) != 0 {
		t.Fatalf("Shields() length = %d, want 0", len(shields))
	}
	for _, item := range armor {
		if item.Stats.Armor != 12 || item.Stats.HypothermalResistance != 8.8 || item.Stats.HyperthermalResistance != 15.6 {
			t.Fatalf("Armor equipment stats = %#v", item.Stats)
		}
	}
}

func TestEquipmentAPICountsQuantities(t *testing.T) {
	api := NewEquipment(nil)
	items := map[uuid.UUID]arkobject.EquipmentItem{
		uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): {
			InventoryItem: arkobject.InventoryItem{Quantity: 2},
		},
		uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): {
			InventoryItem: arkobject.InventoryItem{Quantity: 3},
		},
	}
	if got := api.Count(items); got != 5 {
		t.Fatalf("Count() = %d, want 5", got)
	}
}

func TestEquipmentAPIFiltersByParsedEquipmentStats(t *testing.T) {
	weaponSave := openSyntheticEquipmentSave(t)
	defer weaponSave.Close()

	weaponAPI := NewEquipment(weaponSave)
	highDamage, err := weaponAPI.WithMinDamage(112)
	if err != nil {
		t.Fatalf("WithMinDamage() error = %v", err)
	}
	if len(highDamage) != 1 {
		t.Fatalf("WithMinDamage(112) length = %d, want 1", len(highDamage))
	}
	tooMuchDamage, err := weaponAPI.WithMinDamage(113)
	if err != nil {
		t.Fatalf("WithMinDamage(113) error = %v", err)
	}
	if len(tooMuchDamage) != 0 {
		t.Fatalf("WithMinDamage(113) length = %d, want 0", len(tooMuchDamage))
	}
	actualDurability, err := weaponAPI.WithMinActualDurability(62.5)
	if err != nil {
		t.Fatalf("WithMinActualDurability() error = %v", err)
	}
	if len(actualDurability) != 1 {
		t.Fatalf("WithMinActualDurability(62.5) length = %d, want 1", len(actualDurability))
	}
	tooMuchActualDurability, err := weaponAPI.WithMinActualDurability(62.6)
	if err != nil {
		t.Fatalf("WithMinActualDurability(62.6) error = %v", err)
	}
	if len(tooMuchActualDurability) != 0 {
		t.Fatalf("WithMinActualDurability(62.6) length = %d, want 0", len(tooMuchActualDurability))
	}

	armorSave := openSyntheticArmorEquipmentSave(t)
	defer armorSave.Close()

	armorAPI := NewEquipment(armorSave)
	armored, err := armorAPI.WithMinArmor(12)
	if err != nil {
		t.Fatalf("WithMinArmor() error = %v", err)
	}
	if len(armored) != 1 {
		t.Fatalf("WithMinArmor(12) length = %d, want 1", len(armored))
	}
	tooMuchArmor, err := armorAPI.WithMinArmor(13)
	if err != nil {
		t.Fatalf("WithMinArmor(13) error = %v", err)
	}
	if len(tooMuchArmor) != 0 {
		t.Fatalf("WithMinArmor(13) length = %d, want 0", len(tooMuchArmor))
	}
	cold, err := armorAPI.WithMinHypothermalResistance(8.8)
	if err != nil {
		t.Fatalf("WithMinHypothermalResistance() error = %v", err)
	}
	if len(cold) != 1 {
		t.Fatalf("WithMinHypothermalResistance() length = %d, want 1", len(cold))
	}
	heat, err := armorAPI.WithMinHyperthermalResistance(15.6)
	if err != nil {
		t.Fatalf("WithMinHyperthermalResistance() error = %v", err)
	}
	if len(heat) != 1 {
		t.Fatalf("WithMinHyperthermalResistance() length = %d, want 1", len(heat))
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

func TestEquipmentAPIFiltersByStateAndStats(t *testing.T) {
	save := openSyntheticEquipmentFilterSave(t)
	defer save.Close()

	api := NewEquipment(save)
	equipped, err := api.Equipped()
	if err != nil {
		t.Fatalf("Equipped() error = %v", err)
	}
	if len(equipped) != 1 {
		t.Fatalf("Equipped() length = %d, want 1", len(equipped))
	}
	blueprints, err := api.Blueprints()
	if err != nil {
		t.Fatalf("Blueprints() error = %v", err)
	}
	if len(blueprints) != 1 {
		t.Fatalf("Blueprints() length = %d, want 1", len(blueprints))
	}
	byQuality, err := api.ByQuality(5)
	if err != nil {
		t.Fatalf("ByQuality() error = %v", err)
	}
	if len(byQuality) != 1 {
		t.Fatalf("ByQuality() length = %d, want 1", len(byQuality))
	}
	rated, err := api.WithMinRating(7)
	if err != nil {
		t.Fatalf("WithMinRating() error = %v", err)
	}
	if len(rated) != 1 {
		t.Fatalf("WithMinRating() length = %d, want 1", len(rated))
	}
	durable, err := api.WithMinDurability(0.5)
	if err != nil {
		t.Fatalf("WithMinDurability() error = %v", err)
	}
	if len(durable) != 1 {
		t.Fatalf("WithMinDurability() length = %d, want 1", len(durable))
	}
}

func openSyntheticEquipmentSave(t *testing.T) *arksave.Save {
	t.Helper()

	weaponID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	engramID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		weaponID: syntheticEquipmentObjectBytes(false),
		engramID: syntheticEquipmentObjectBytes(true),
	})
}

func openSyntheticEquipmentSaveWithFault(t *testing.T) *arksave.Save {
	t.Helper()

	weaponID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	faultyID := uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		weaponID: syntheticEquipmentObjectBytes(false),
		faultyID: truncatedEquipmentObjectBytes(),
	})
}

func openSyntheticEquipmentFilterSave(t *testing.T) *arksave.Save {
	t.Helper()

	equippedID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	blueprintID := uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		equippedID:  syntheticEquipmentObjectBytesWithFlags(false, true, false, 7.5, 5, 0.75),
		blueprintID: syntheticEquipmentObjectBytesWithFlags(false, false, true, 2.5, 1, 0.25),
	})
}

func openSyntheticArmorEquipmentSave(t *testing.T) *arksave.Save {
	t.Helper()

	armorID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	return openSyntheticSaveWith(t, "equipment.ark", nil, map[uuid.UUID][]byte{
		armorID: syntheticArmorEquipmentObjectBytes(),
	})
}

func syntheticEquipmentObjectBytes(isEngram bool) []byte {
	return syntheticEquipmentObjectBytesWithFlags(isEngram, false, false, 7.5, 3, 0.75)
}

func syntheticEquipmentObjectBytesWithFlags(isEngram bool, isEquipped bool, isBlueprint bool, rating float32, quality int32, durability float32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000f))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x1000000c, 1)
	writeFloatProperty(&buf, 0x10000010, rating)
	writeIntProperty(&buf, 0x10000011, quality)
	writeFloatProperty(&buf, 0x10000012, durability)
	writePositionedUInt16Property(&buf, 0x10000040, 2, 1000)
	writePositionedUInt16Property(&buf, 0x10000040, 3, 1234)
	writeStringProperty(&buf, 0x1000001b, "Survivor")
	writeStringProperty(&buf, 0x1000001c, "Porters")
	if isEngram {
		writeBoolProperty(&buf, 0x10000013, true)
	}
	if isEquipped {
		writeBoolProperty(&buf, 0x10000022, true)
	}
	if isBlueprint {
		writeBoolProperty(&buf, 0x1000000d, true)
	}
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func truncatedEquipmentObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000f))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticArmorEquipmentObjectBytes() []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000042))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	writeIntProperty(&buf, 0x1000000c, 1)
	writePositionedUInt16Property(&buf, 0x10000040, 1, 1000)
	writePositionedUInt16Property(&buf, 0x10000040, 5, 500)
	writePositionedUInt16Property(&buf, 0x10000040, 7, 200)
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

func writePositionedUInt16Property(buf *bytes.Buffer, name uint32, position int32, value uint16) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0x10000041))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(2))
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(1)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, value)
}
