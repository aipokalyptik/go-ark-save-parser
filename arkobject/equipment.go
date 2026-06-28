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

type EquipmentStat int32

const (
	EquipmentStatArmor                  EquipmentStat = 1
	EquipmentStatDurability             EquipmentStat = 2
	EquipmentStatDamage                 EquipmentStat = 3
	EquipmentStatHypothermalResistance  EquipmentStat = 5
	EquipmentStatHyperthermalResistance EquipmentStat = 7
)

type EquipmentItem struct {
	InventoryItem
	Kind              EquipmentKind
	IsEquipped        bool
	IsBlueprint       bool
	CustomDataCount   int
	Rating            float64
	Quality           int32
	CurrentDurability float64
	Stats             EquipmentStats
}

type EquipmentStats struct {
	Internal               map[EquipmentStat]uint16
	Damage                 float64
	Durability             float64
	Armor                  float64
	HypothermalResistance  float64
	HyperthermalResistance float64
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
	item.CustomDataCount = customItemDataCount(properties)
	item.Stats = equipmentStats(properties, kind, object.Blueprint)
	return item
}

func (e EquipmentItem) HasCustomData() bool {
	return e.CustomDataCount > 0
}

func (e EquipmentItem) IsCrafted() bool {
	return e.Crafter != nil && e.Crafter.Valid()
}
