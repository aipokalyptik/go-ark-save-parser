package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

type DinoStatPoints struct {
	Health        int32
	Stamina       int32
	Torpidity     int32
	Oxygen        int32
	Food          int32
	Water         int32
	Temperature   int32
	Weight        int32
	MeleeDamage   int32
	MovementSpeed int32
	Fortitude     int32
	CraftingSpeed int32
}

type DinoStatValues struct {
	Health        float64
	Stamina       float64
	Torpidity     float64
	Oxygen        float64
	Food          float64
	Water         float64
	Temperature   float64
	Weight        float64
	MeleeDamage   float64
	MovementSpeed float64
	Fortitude     float64
	CraftingSpeed float64
}

type DinoStats struct {
	BaseLevel         int32
	CurrentLevel      int32
	BaseStatPoints    DinoStatPoints
	AddedStatPoints   DinoStatPoints
	MutatedStatPoints DinoStatPoints
	StatValues        DinoStatValues
	ImprintingPercent float64
}

func DinoStatsFromObject(object *GameObject) DinoStats {
	if object == nil {
		return DinoStats{}
	}
	properties := arkproperty.Container{Properties: object.Properties}
	stats := DinoStats{
		BaseLevel:         int32Value(properties, "BaseCharacterLevel"),
		BaseStatPoints:    statPoints(properties, "NumberOfLevelUpPointsApplied"),
		AddedStatPoints:   statPoints(properties, "NumberOfLevelUpPointsAppliedTamed"),
		MutatedStatPoints: statPoints(properties, "NumberOfMutationsAppliedTamed"),
		StatValues:        statValues(properties, "CurrentStatusValues"),
		ImprintingPercent: float64Value(properties, "DinoImprintingQuality") * 100,
	}
	stats.CurrentLevel = stats.BaseStatPoints.level(true) +
		stats.AddedStatPoints.level(false) +
		stats.MutatedStatPoints.level(false)
	return stats
}

func statPoints(properties arkproperty.Container, name string) DinoStatPoints {
	return DinoStatPoints{
		Health:        int32PositionedValue(properties, name, 0),
		Stamina:       int32PositionedValue(properties, name, 1),
		Torpidity:     int32PositionedValue(properties, name, 2),
		Oxygen:        int32PositionedValue(properties, name, 3),
		Food:          int32PositionedValue(properties, name, 4),
		Water:         int32PositionedValue(properties, name, 5),
		Temperature:   int32PositionedValue(properties, name, 6),
		Weight:        int32PositionedValue(properties, name, 7),
		MeleeDamage:   int32PositionedValue(properties, name, 8),
		MovementSpeed: int32PositionedValue(properties, name, 9),
		Fortitude:     int32PositionedValue(properties, name, 10),
		CraftingSpeed: int32PositionedValue(properties, name, 11),
	}
}

func statValues(properties arkproperty.Container, name string) DinoStatValues {
	return DinoStatValues{
		Health:        float64PositionedValue(properties, name, 0),
		Stamina:       float64PositionedValue(properties, name, 1),
		Torpidity:     float64PositionedValue(properties, name, 2),
		Oxygen:        float64PositionedValue(properties, name, 3),
		Food:          float64PositionedValue(properties, name, 4),
		Water:         float64PositionedValue(properties, name, 5),
		Temperature:   float64PositionedValue(properties, name, 6),
		Weight:        float64PositionedValue(properties, name, 7),
		MeleeDamage:   float64PositionedValue(properties, name, 8),
		MovementSpeed: float64PositionedValue(properties, name, 9),
		Fortitude:     float64PositionedValue(properties, name, 10),
		CraftingSpeed: float64PositionedValue(properties, name, 11),
	}
}

func (p DinoStatPoints) level(includeInitial bool) int32 {
	level := p.Health + p.Stamina + p.Torpidity + p.Oxygen + p.Food + p.Water +
		p.Temperature + p.Weight + p.MeleeDamage + p.MovementSpeed + p.Fortitude +
		p.CraftingSpeed
	if includeInitial {
		level++
	}
	return level
}

func int32PositionedValue(properties arkproperty.Container, name string, position int32) int32 {
	value, ok := properties.PositionedValue(name, position)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int32:
		return v
	case int8:
		return int32(v)
	case uint32:
		return int32(v)
	case int:
		return int32(v)
	default:
		return 0
	}
}

func float64PositionedValue(properties arkproperty.Container, name string, position int32) float64 {
	value, ok := properties.PositionedValue(name, position)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int32:
		return float64(v)
	case uint32:
		return float64(v)
	default:
		return 0
	}
}
