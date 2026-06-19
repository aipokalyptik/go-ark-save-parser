package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

type EquipmentKind string

const (
	EquipmentUnknown EquipmentKind = ""
	EquipmentWeapon  EquipmentKind = "weapon"
	EquipmentSaddle  EquipmentKind = "saddle"
	EquipmentArmor   EquipmentKind = "armor"
	EquipmentShield  EquipmentKind = "shield"
)

type EquipmentItem struct {
	InventoryItem
	Kind              EquipmentKind
	IsEquipped        bool
	IsBlueprint       bool
	Rating            float64
	Quality           int32
	CurrentDurability float64
}

func EquipmentItemFromObject(object *GameObject, kind EquipmentKind) EquipmentItem {
	item := EquipmentItem{
		InventoryItem:     InventoryItemFromObject(object),
		Kind:              kind,
		Rating:            1,
		Quality:           0,
		CurrentDurability: 1,
	}
	if object == nil {
		return item
	}
	properties := arkproperty.Container{Properties: object.Properties}
	item.IsEquipped = boolValue(properties, "bEquippedItem")
	item.IsBlueprint = boolValue(properties, "bIsBlueprint")
	if value, ok := numericFloat64(properties, "ItemRating"); ok {
		item.Rating = value
	}
	if value, ok := numericInt32(properties, "ItemQualityIndex"); ok {
		item.Quality = value
	}
	if value, ok := numericFloat64(properties, "SavedDurability"); ok {
		item.CurrentDurability = value
	}
	return item
}

func numericInt32(properties arkproperty.Container, name string) (int32, bool) {
	value, ok := properties.Value(name)
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case byte:
		return int32(v), true
	case int32:
		return v, true
	case uint32:
		return int32(v), true
	case int:
		return int32(v), true
	default:
		return 0, false
	}
}

func numericFloat64(properties arkproperty.Container, name string) (float64, bool) {
	value, ok := properties.Value(name)
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int32:
		return float64(v), true
	case uint32:
		return float64(v), true
	case byte:
		return float64(v), true
	default:
		return 0, false
	}
}
