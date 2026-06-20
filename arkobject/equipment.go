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
	Rating            float64
	Quality           int32
	CurrentDurability float64
	Stats             EquipmentStats
}

type EquipmentStats struct {
	Internal   map[EquipmentStat]uint16
	Damage     float64
	Durability float64
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
	item.Stats = equipmentStats(properties, kind, object.Blueprint)
	return item
}

func equipmentStats(properties arkproperty.Container, kind EquipmentKind, blueprint string) EquipmentStats {
	stats := EquipmentStats{Internal: map[EquipmentStat]uint16{}}
	for _, stat := range []EquipmentStat{
		EquipmentStatArmor,
		EquipmentStatDurability,
		EquipmentStatDamage,
		EquipmentStatHypothermalResistance,
		EquipmentStatHyperthermalResistance,
	} {
		if value, ok := uint16PositionedValue(properties, "ItemStatValues", int32(stat)); ok {
			stats.Internal[stat] = value
		}
	}
	if value, ok := stats.Internal[EquipmentStatDamage]; ok && kind == EquipmentWeapon {
		stats.Damage = float64(int(value))
		stats.Damage = float64(int((100.0+stats.Damage/100)*10+0.5)) / 10
	}
	if value, ok := stats.Internal[EquipmentStatDurability]; ok {
		stats.Durability = defaultEquipmentDurability(blueprint) * (0.00025*float64(value) + 1)
	}
	return stats
}

func defaultEquipmentDurability(blueprint string) float64 {
	if blueprint == "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'" {
		return 50
	}
	return 1
}

func uint16PositionedValue(properties arkproperty.Container, name string, position int32) (uint16, bool) {
	value, ok := properties.PositionedValue(name, position)
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case uint16:
		return v, true
	case uint32:
		return uint16(v), true
	case int32:
		return uint16(v), true
	case int:
		return uint16(v), true
	default:
		return 0, false
	}
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
