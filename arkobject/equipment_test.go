package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestEquipmentItemFromObjectReadsBaseEquipmentFields(t *testing.T) {
	object := &GameObject{
		UUID:      uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"),
		Blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
		Properties: []arkproperty.Property{
			{Name: "ItemQuantity", Type: arkproperty.TypeInt, Value: int32(1)},
			{Name: "bIsBlueprint", Type: arkproperty.TypeBool, Value: true},
			{Name: "bEquippedItem", Type: arkproperty.TypeBool, Value: false},
			{Name: "ItemRating", Type: arkproperty.TypeFloat, Value: float32(7.5)},
			{Name: "ItemQualityIndex", Type: arkproperty.TypeByte, Value: byte(3)},
			{Name: "SavedDurability", Type: arkproperty.TypeFloat, Value: float32(0.75)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatDurability), Value: uint16(1000)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatDamage), Value: uint16(1234)},
			{Name: "CrafterCharacterName", Type: arkproperty.TypeString, Value: "Survivor"},
			{Name: "CrafterTribeName", Type: arkproperty.TypeString, Value: "Porters"},
		},
	}

	item := EquipmentItemFromObject(object, EquipmentWeapon)

	if item.UUID != object.UUID || item.Blueprint != object.Blueprint || item.Kind != EquipmentWeapon {
		t.Fatalf("EquipmentItem identity = %#v", item)
	}
	if item.Quantity != 1 || !item.IsBlueprint || item.IsEquipped {
		t.Fatalf("EquipmentItem inventory flags = %#v", item)
	}
	if item.Rating != 7.5 || item.Quality != 3 || item.CurrentDurability != 0.75 {
		t.Fatalf("EquipmentItem equipment fields = %#v", item)
	}
	if item.Stats.Internal[EquipmentStatDamage] != 1234 || item.Stats.Internal[EquipmentStatDurability] != 1000 {
		t.Fatalf("EquipmentItem internal stats = %#v", item.Stats.Internal)
	}
	if item.Stats.Damage != 112.3 || item.Stats.Durability != 62.5 {
		t.Fatalf("EquipmentItem actual stats = %#v", item.Stats)
	}
	if item.Crafter == nil || item.Crafter.CharacterName != "Survivor" || item.Crafter.TribeName != "Porters" {
		t.Fatalf("EquipmentItem.Crafter = %#v", item.Crafter)
	}
}

func TestEquipmentItemFromObjectReadsArmorStats(t *testing.T) {
	object := &GameObject{
		UUID:      uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"),
		Blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'",
		Properties: []arkproperty.Property{
			{Name: "ItemQuantity", Type: arkproperty.TypeInt, Value: int32(1)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatArmor), Value: uint16(1000)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatHypothermalResistance), Value: uint16(500)},
			{Name: "ItemStatValues", Type: arkproperty.TypeUInt16, Position: int32(EquipmentStatHyperthermalResistance), Value: uint16(200)},
		},
	}

	item := EquipmentItemFromObject(object, EquipmentArmor)

	if item.Stats.Internal[EquipmentStatArmor] != 1000 || item.Stats.Internal[EquipmentStatHypothermalResistance] != 500 {
		t.Fatalf("EquipmentItem internal armor stats = %#v", item.Stats.Internal)
	}
	if item.Stats.Armor != 12 || item.Stats.HypothermalResistance != 8.8 || item.Stats.HyperthermalResistance != 15.6 {
		t.Fatalf("EquipmentItem armor stats = %#v", item.Stats)
	}
}
