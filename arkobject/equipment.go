package arkobject

import (
	"math"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

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
	item.Stats = equipmentStats(properties, kind, object.Blueprint)
	return item
}

func (e EquipmentItem) IsCrafted() bool {
	return e.Crafter != nil && e.Crafter.Valid()
}

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
		total += float64(e.Stats.Internal[stat])
	}
	return total / float64(len(stats))
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

func defaultEquipmentDurability(blueprint string) float64 {
	blueprint = canonicalEquipmentBlueprintPath(blueprint)
	switch {
	case armorFamily(blueprint, "/Chitin/", "Chitin"):
		return 50
	case armorFamily(blueprint, "/Ghillie/", "Ghillie"),
		armorFamily(blueprint, "/Leather/", "Hide"),
		strings.Contains(blueprint, "PrimalItemArmor_DesertCloth"),
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_ScubaShirt_SuitWithTank.PrimalItemArmor_ScubaShirt_SuitWithTank_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_ScubaBoots_Flippers.PrimalItemArmor_ScubaBoots_Flippers_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_ScubaHelmet_Goggles.PrimalItemArmor_ScubaHelmet_Goggles_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_NightVisionGoggles.PrimalItemArmor_NightVisionGoggles_C":
		return 45
	case armorFamily(blueprint, "/Cloth/", "Cloth"):
		return 25
	case armorFamily(blueprint, "/Riot/", "Riot"),
		armorFamily(blueprint, "/Metal/", "Metal"),
		armorFamily(blueprint, "/TEK/", "Tek"):
		return 120
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_ScubaPants.PrimalItemArmor_ScubaPants_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_GasMask.PrimalItemArmor_GasMask_C":
		return 50
	case strings.Contains(blueprint, "/HazardSuit/"):
		return 85.5
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_MetalShield.PrimalItemArmor_MetalShield_C":
		return 1250
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_TransparentRiotShield.PrimalItemArmor_TransparentRiotShield_C":
		return 2300
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_WoodShield.PrimalItemArmor_WoodShield_C":
		return 350
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Metal/PrimalItemArmor_MinersHelmet.PrimalItemArmor_MinersHelmet_C",
		tekSaddleBlueprint(blueprint),
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponCrossbow.PrimalItem_WeaponCrossbow_C",
		blueprint == "/Game/LostColony/Weapons/FabricatedCrossbow/PrimalItem_WeaponCrossbow_Fab.PrimalItem_WeaponCrossbow_Fab_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponHarpoon.PrimalItem_WeaponHarpoon_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponMachinedShotgun.PrimalItem_WeaponMachinedShotgun_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponRocketLauncher.PrimalItem_WeaponRocketLauncher_C":
		return 120
	case blueprint == "/Game/Aberration/Dinos/MoleRat/PrimalItemArmor_MoleRatSaddle.PrimalItemArmor_MoleRatSaddle_C":
		return 500
	case strings.Contains(blueprint, "PrimalItemArmor_") && (strings.Contains(blueprint, "Saddle") || strings.Contains(blueprint, "/Saddle/")):
		return 100
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponCompoundBow.PrimalItem_WeaponCompoundBow_C":
		return 55
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponSword.PrimalItem_WeaponSword_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponOneShotRifle.PrimalItem_WeaponOneShotRifle_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponMachinedSniper.PrimalItem_WeaponMachinedSniper_C":
		return 70
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponProd.PrimalItem_WeaponProd_C":
		return 10
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C":
		return 50
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponGun.PrimalItem_WeaponGun_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponMachinedPistol.PrimalItem_WeaponMachinedPistol_C":
		return 60
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponShotgun.PrimalItem_WeaponShotgun_C",
		blueprint == "/Game/ScorchedEarth/WeaponChainsaw/PrimalItem_ChainSaw.PrimalItem_ChainSaw_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_TekRifle.PrimalItem_TekRifle_C":
		return 80
	case weaponDurability40Blueprint(blueprint):
		return 40
	case blueprint == "/Game/Aberration/CoreBlueprints/Weapons/PrimalItem_WeaponClimbPick.PrimalItem_WeaponClimbPick_C":
		return 65
	case blueprint == "":
		return 1
	default:
		return 125
	}
}

func defaultEquipmentArmor(blueprint string) float64 {
	blueprint = canonicalEquipmentBlueprintPath(blueprint)
	switch {
	case armorFamily(blueprint, "/Chitin/", "Chitin"):
		return 50
	case armorFamily(blueprint, "/Ghillie/", "Ghillie"):
		return 32
	case armorFamily(blueprint, "/Leather/", "Hide"):
		return 20
	case strings.Contains(blueprint, "PrimalItemArmor_DesertCloth"),
		armorFamily(blueprint, "/Fur/", "Fur"),
		strings.Contains(blueprint, "PrimalItemArmor_Arctic"):
		return 40
	case armorFamily(blueprint, "/Cloth/", "Cloth"):
		return 10
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Metal/PrimalItemArmor_MinersHelmet.PrimalItemArmor_MinersHelmet_C":
		return 120
	case armorFamily(blueprint, "/Riot/", "Riot"):
		return 115
	case armorFamily(blueprint, "/Metal/", "Metal"):
		return 100
	case armorFamily(blueprint, "/TEK/", "Tek"):
		return 180
	case strings.Contains(blueprint, "/SCUBA/"):
		return 1
	case strings.Contains(blueprint, "/HazardSuit/"):
		return 65
	case tekSaddleBlueprint(blueprint):
		return 45
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_Paracer_Saddle.PrimalItemArmor_Paracer_Saddle_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_DiplodocusSaddle.PrimalItemArmor_DiplodocusSaddle_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_SauroSaddle.PrimalItemArmor_SauroSaddle_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_ParacerSaddle_Platform.PrimalItemArmor_ParacerSaddle_Platform_C",
		blueprint == "/Game/ASA/Dinos/Archelon/Dinos/Saddle/PrimalItem_Armor_Archelon_Saddle_ASA.PrimalItem_Armor_Archelon_Saddle_ASA_C",
		blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_TurtleSaddle.PrimalItemArmor_TurtleSaddle_C":
		return 20
	case blueprint == "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_TitanSaddle_Platform.PrimalItemArmor_TitanSaddle_Platform_C":
		return 1
	case strings.Contains(blueprint, "PrimalItemArmor_") && (strings.Contains(blueprint, "Saddle") || strings.Contains(blueprint, "/Saddle/")):
		return 25
	default:
		return 1
	}
}

func defaultEquipmentHypothermal(blueprint string) float64 {
	blueprint = canonicalEquipmentBlueprintPath(blueprint)
	if value, ok := armorThermalDefaultForBlueprint(blueprint); ok {
		return value.hypothermal
	}
	if value, ok := armorThermalDefaults[blueprint]; ok {
		return value.hypothermal
	}
	if strings.Contains(blueprint, "PrimalItemArmor_ClothShirt") {
		return 8
	}
	return 0
}

func defaultEquipmentHyperthermal(blueprint string) float64 {
	blueprint = canonicalEquipmentBlueprintPath(blueprint)
	if value, ok := armorThermalDefaultForBlueprint(blueprint); ok {
		return value.hyperthermal
	}
	if value, ok := armorThermalDefaults[blueprint]; ok {
		return value.hyperthermal
	}
	if strings.Contains(blueprint, "PrimalItemArmor_ClothShirt") {
		return 15
	}
	return 0
}

func round1(value float64) float64 {
	return math.Round(value*10) / 10
}

func canonicalEquipmentBlueprintPath(blueprint string) string {
	blueprint = strings.TrimPrefix(blueprint, "Blueprint'")
	return strings.TrimSuffix(blueprint, "'")
}

func armorFamily(blueprint string, pathPart string, namePart string) bool {
	return strings.Contains(blueprint, pathPart) || strings.Contains(blueprint, "_Cursed") && strings.Contains(blueprint, namePart)
}

func tekSaddleBlueprint(blueprint string) bool {
	switch blueprint {
	case "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_Tapejara_Tek.PrimalItemArmor_Tapejara_Tek_C",
		"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_RexSaddle_Tek.PrimalItemArmor_RexSaddle_Tek_C",
		"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_MosaSaddle_Tek.PrimalItemArmor_MosaSaddle_Tek_C",
		"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_MegalodonSaddle_Tek.PrimalItemArmor_MegalodonSaddle_Tek_C",
		"/Game/Aberration/Dinos/RockDrake/PrimalItemArmor_RockDrakeSaddle_Tek.PrimalItemArmor_RockDrakeSaddle_Tek_C":
		return true
	default:
		return false
	}
}

func weaponDurability40Blueprint(blueprint string) bool {
	switch blueprint {
	case "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponSickle.PrimalItem_WeaponSickle_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponMetalHatchet.PrimalItem_WeaponMetalHatchet_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponMetalPick.PrimalItem_WeaponMetalPick_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponStoneHatchet.PrimalItem_WeaponStoneHatchet_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponStonePick.PrimalItem_WeaponStonePick_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponFishingRod.PrimalItem_WeaponFishingRod_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponRifle.PrimalItem_WeaponRifle_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponSlingshot.PrimalItem_WeaponSlingshot_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponStoneClub.PrimalItem_WeaponStoneClub_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponPike.PrimalItem_WeaponPike_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponLance.PrimalItem_WeaponLance_C",
		"/Game/ScorchedEarth/WeaponFlamethrower/PrimalItem_WeapFlamethrower.PrimalItem_WeapFlamethrower_C",
		"/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponTorch.PrimalItem_WeaponTorch_C":
		return true
	default:
		return strings.Contains(blueprint, "/Game/Packs/Wasteland/Dinos/Doggo/PrimalItemArmor_DinoCompanionSaddle_Doggo")
	}
}

type armorThermalDefault struct {
	hypothermal  float64
	hyperthermal float64
}

func armorThermalDefaultForBlueprint(blueprint string) (armorThermalDefault, bool) {
	slot := armorSlot(blueprint)
	if slot == "" {
		return armorThermalDefault{}, false
	}
	switch {
	case armorFamily(blueprint, "/Cloth/", "Cloth"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 8, hyperthermal: 15}, true
		case "helmet":
			return armorThermalDefault{hypothermal: 4, hyperthermal: 15}, true
		case "boots", "gloves":
			return armorThermalDefault{hypothermal: 5, hyperthermal: 15}, true
		}
	case armorFamily(blueprint, "/Chitin/", "Chitin"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 10, hyperthermal: -5}, true
		case "helmet":
			return armorThermalDefault{hypothermal: 6, hyperthermal: -5}, true
		case "boots", "gloves":
			return armorThermalDefault{hypothermal: 7, hyperthermal: -5}, true
		}
	case strings.Contains(blueprint, "PrimalItemArmor_DesertCloth"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 8, hyperthermal: 25}, true
		case "helmet":
			return armorThermalDefault{hypothermal: 5, hyperthermal: 30}, true
		case "boots", "gloves":
			return armorThermalDefault{hypothermal: 5, hyperthermal: 25}, true
		}
	case armorFamily(blueprint, "/Fur/", "Fur"),
		armorFamily(blueprint, "/Armor/", "Arctic") && strings.Contains(blueprint, "Arctic"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 65, hyperthermal: -30}, true
		case "helmet":
			return armorThermalDefault{hypothermal: 52, hyperthermal: -25}, true
		case "boots", "gloves":
			return armorThermalDefault{hypothermal: 34, hyperthermal: -10}, true
		}
	case armorFamily(blueprint, "/Leather/", "Hide"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 20, hyperthermal: -5}, true
		case "helmet", "boots", "gloves":
			return armorThermalDefault{hypothermal: 15, hyperthermal: -5}, true
		}
	case armorFamily(blueprint, "/Ghillie/", "Ghillie"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 8, hyperthermal: 30}, true
		case "helmet":
			return armorThermalDefault{hypothermal: 4, hyperthermal: 35}, true
		case "boots", "gloves":
			return armorThermalDefault{hypothermal: 2, hyperthermal: 30}, true
		}
	case armorFamily(blueprint, "/Riot/", "Riot"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 15, hyperthermal: -10}, true
		case "helmet", "boots", "gloves":
			return armorThermalDefault{hypothermal: 10, hyperthermal: -10}, true
		}
	case armorFamily(blueprint, "/Metal/", "Metal"):
		switch slot {
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 15, hyperthermal: -7}, true
		case "helmet":
			return armorThermalDefault{hypothermal: 10, hyperthermal: -3}, true
		case "boots", "gloves":
			return armorThermalDefault{hypothermal: 10, hyperthermal: -4}, true
		}
	case armorFamily(blueprint, "/TEK/", "Tek"):
		switch slot {
		case "helmet":
			return armorThermalDefault{hypothermal: 5, hyperthermal: 30}, true
		case "shirt", "pants":
			return armorThermalDefault{hypothermal: 15, hyperthermal: -7}, true
		case "boots", "gloves":
			return armorThermalDefault{hypothermal: 10, hyperthermal: -4}, true
		}
	case strings.Contains(blueprint, "/SCUBA/"):
		switch slot {
		case "helmet", "boots":
			return armorThermalDefault{hypothermal: 15, hyperthermal: -5}, true
		case "shirt":
			return armorThermalDefault{hypothermal: 40, hyperthermal: -5}, true
		case "pants":
			return armorThermalDefault{hypothermal: 200, hyperthermal: 0}, true
		}
	case strings.Contains(blueprint, "/HazardSuit/"):
		return armorThermalDefault{hypothermal: 10, hyperthermal: 60}, true
	}
	return armorThermalDefault{}, false
}

func armorSlot(blueprint string) string {
	switch {
	case strings.Contains(blueprint, "Boots") || strings.Contains(blueprint, "Flippers"):
		return "boots"
	case strings.Contains(blueprint, "Gloves"):
		return "gloves"
	case strings.Contains(blueprint, "Helmet") || strings.Contains(blueprint, "Goggles"):
		return "helmet"
	case strings.Contains(blueprint, "Pants"):
		return "pants"
	case strings.Contains(blueprint, "Shirt"):
		return "shirt"
	default:
		return ""
	}
}

var armorThermalDefaults = map[string]armorThermalDefault{
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Chitin/PrimalItemArmor_ChitinShirt.PrimalItemArmor_ChitinShirt_C":              {hypothermal: 10, hyperthermal: -5},
	"/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothGogglesHelmet.PrimalItemArmor_DesertClothGogglesHelmet_C":            {hypothermal: 5, hyperthermal: 30},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C":                 {hypothermal: 8, hyperthermal: 15},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothPants.PrimalItemArmor_ClothPants_C":                 {hypothermal: 8, hyperthermal: 15},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothHelmet.PrimalItemArmor_ClothHelmet_C":               {hypothermal: 4, hyperthermal: 15},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothBoots.PrimalItemArmor_ClothBoots_C":                 {hypothermal: 5, hyperthermal: 15},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothGloves.PrimalItemArmor_ClothGloves_C":               {hypothermal: 5, hyperthermal: 15},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Metal/PrimalItemArmor_MetalShirt.PrimalItemArmor_MetalShirt_C":                 {hypothermal: 15, hyperthermal: -7},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/TEK/PrimalItemArmor_TekHelmet.PrimalItemArmor_TekHelmet_C":                     {hypothermal: 5, hyperthermal: 30},
	"/Game/Aberration/CoreBlueprints/Items/Armor/HazardSuit/PrimalItemArmor_HazardSuitShirt.PrimalItemArmor_HazardSuitShirt_C":   {hypothermal: 10, hyperthermal: 60},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_ScubaPants.PrimalItemArmor_ScubaPants_C":                 {hypothermal: 200, hyperthermal: 0},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Leather/PrimalItemArmor_HideShirt.PrimalItemArmor_HideShirt_C":                 {hypothermal: 20, hyperthermal: -5},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Ghillie/PrimalItemArmor_GhillieHelmet.PrimalItemArmor_GhillieHelmet_C":         {hypothermal: 4, hyperthermal: 35},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Fur/PrimalItemArmor_FurShirt.PrimalItemArmor_FurShirt_C":                       {hypothermal: 65, hyperthermal: -30},
	"/Game/LostColony/CoreBlueprints/Items/CursedArmor/PrimalItemArmor_MetalShirt_Cursed.PrimalItemArmor_MetalShirt_Cursed_C":    {hypothermal: 15, hyperthermal: -7},
	"/Game/LostColony/CoreBlueprints/Items/CursedArmor/PrimalItemArmor_RiotShirt_Cursed.PrimalItemArmor_RiotShirt_Cursed_C":      {hypothermal: 15, hyperthermal: -10},
	"/Game/LostColony/CoreBlueprints/Items/CursedArmor/PrimalItemArmor_TekHelmet_Cursed.PrimalItemArmor_TekHelmet_Cursed_C":      {hypothermal: 5, hyperthermal: 30},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_GasMask.PrimalItemArmor_GasMask_C":                       {hypothermal: 10, hyperthermal: -2},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Metal/PrimalItemArmor_MinersHelmet.PrimalItemArmor_MinersHelmet_C":             {hypothermal: 10, hyperthermal: -3},
	"/Game/PrimalEarth/CoreBlueprints/Items/Armor/SCUBA/PrimalItemArmor_NightVisionGoggles.PrimalItemArmor_NightVisionGoggles_C": {hypothermal: 15, hyperthermal: -5},
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
