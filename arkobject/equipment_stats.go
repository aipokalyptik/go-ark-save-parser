package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

func (e EquipmentItem) ImplementedStats() []EquipmentStat {
	switch e.Kind {
	case EquipmentWeapon:
		return []EquipmentStat{EquipmentStatDurability, EquipmentStatDamage}
	case EquipmentArmor:
		return []EquipmentStat{
			EquipmentStatDurability,
			EquipmentStatArmor,
			EquipmentStatHypothermalResistance,
			EquipmentStatHyperthermalResistance,
		}
	case EquipmentSaddle:
		return []EquipmentStat{EquipmentStatDurability, EquipmentStatArmor}
	case EquipmentShield:
		return []EquipmentStat{EquipmentStatDurability}
	default:
		return nil
	}
}

func (e EquipmentItem) AverageStat() float64 {
	stats := e.ImplementedStats()
	if len(stats) == 0 {
		return 0
	}
	var total float64
	for _, stat := range stats {
		total += e.effectiveInternalStat(stat)
	}
	return total / float64(len(stats))
}

func (e EquipmentItem) effectiveInternalStat(stat EquipmentStat) float64 {
	value, ok := e.Stats.Internal[stat]
	raw := float64(value)
	switch stat {
	case EquipmentStatDamage:
		if ok && raw >= 100 {
			return raw
		}
		return 100
	case EquipmentStatDurability:
		return defaultedInternalStat(raw, ok, defaultEquipmentDurability(e.Blueprint))
	case EquipmentStatArmor:
		return defaultedInternalStat(raw, ok, defaultEquipmentArmor(e.Blueprint))
	case EquipmentStatHypothermalResistance:
		return defaultedThermalInternalStat(raw, ok, defaultEquipmentHypothermal(e.Blueprint))
	case EquipmentStatHyperthermalResistance:
		return defaultedThermalInternalStat(raw, ok, defaultEquipmentHyperthermal(e.Blueprint))
	default:
		if ok {
			return raw
		}
		return 0
	}
}

func defaultedInternalStat(raw float64, ok bool, fallback float64) float64 {
	if ok && raw >= fallback {
		return raw
	}
	return fallback
}

func defaultedThermalInternalStat(raw float64, ok bool, fallback float64) float64 {
	if !ok || raw == 0 || fallback == 0 {
		return 0
	}
	if raw >= fallback {
		return raw
	}
	return fallback
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
	if value, ok := stats.Internal[EquipmentStatArmor]; ok && (kind == EquipmentArmor || kind == EquipmentSaddle || kind == EquipmentShield) {
		stats.Armor = round1(defaultEquipmentArmor(blueprint) * (0.0002*float64(value) + 1))
	}
	if value, ok := stats.Internal[EquipmentStatHypothermalResistance]; ok && kind == EquipmentArmor {
		stats.HypothermalResistance = round1(defaultEquipmentHypothermal(blueprint) * (0.0002*float64(value) + 1))
	}
	if value, ok := stats.Internal[EquipmentStatHyperthermalResistance]; ok && kind == EquipmentArmor {
		stats.HyperthermalResistance = round1(defaultEquipmentHyperthermal(blueprint) * (0.0002*float64(value) + 1))
	}
	return stats
}
