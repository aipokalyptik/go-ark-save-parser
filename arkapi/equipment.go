package arkapi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type EquipmentAPI struct {
	save *arksave.Save
}

type EquipmentFilterOptions struct {
	Kinds          []arkobject.EquipmentKind
	Blueprints     []string
	NoBlueprints   bool
	OnlyBlueprints bool
	MinQuality     int32
	MinRating      float64
	MinDurability  float64
	Equipped       *bool
	Crafter        *arkobject.ObjectCrafter
}

const AscendantQualityIndex int32 = 5

var canonicalEquipmentKinds = map[string]arkobject.EquipmentKind{
	"/Game/ScorchedEarth/WeaponChainsaw/PrimalItem_ChainSaw.PrimalItem_ChainSaw_C":                     arkobject.EquipmentWeapon,
	"/Game/ScorchedEarth/WeaponFlamethrower/PrimalItem_WeapFlamethrower.PrimalItem_WeapFlamethrower_C": arkobject.EquipmentWeapon,

	"/Game/ASA/Dinos/YiLing/PrimalItemArmor_YiLingSaddle.PrimalItemArmor_YiLingSaddle_C":                                                         arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Archelon/Dinos/Saddle/PrimalItem_Armor_Archelon_Saddle_ASA.PrimalItem_Armor_Archelon_Saddle_ASA_C":                          arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Ceratosaurus/Dinos/Saddle/PrimalItemArmor_CeratosaurusSaddle_ASA.PrimalItemArmor_CeratosaurusSaddle_ASA_C":                  arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Deinosuchus/Saddle/PrimalItemArmor_Deinosuchus_Saddle_ASA.PrimalItemArmor_Deinosuchus_Saddle_ASA_C":                         arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Fasolasuchus/PrimalItemArmor_FasolaSaddle.PrimalItemArmor_FasolaSaddle_C":                                                   arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Gigantoraptor/PrimalItemArmor_GigantoraptorSaddle.PrimalItemArmor_GigantoraptorSaddle_C":                                    arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Shastasaurus/PrimalItemArmor_ShastaSaddle_Submarine.PrimalItemArmor_ShastaSaddle_Submarine_C":                               arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Xiphactinus/Dinos/Saddle/PrimalItemArmor_XiphSaddle_ASA.PrimalItemArmor_XiphSaddle_ASA_C":                                   arkobject.EquipmentSaddle,
	"/Game/Packs/Steampunk/Dinos/RockDrakeSteampunkSaddle/PrimalItemArmor_RockDrakeSaddle_Steampunk.PrimalItemArmor_RockDrakeSaddle_Steampunk_C": arkobject.EquipmentSaddle,
	"/Game/Aberration/Dinos/Basilisk/PrimalItemArmor_BasiliskSaddle.PrimalItemArmor_BasiliskSaddle_C":                                            arkobject.EquipmentSaddle,
	"/Game/Aberration/Dinos/CaveWolf/PrimalItemArmor_CavewolfSaddle.PrimalItemArmor_CavewolfSaddle_C":                                            arkobject.EquipmentSaddle,
	"/Game/Aberration/Dinos/Crab/PrimalItemArmor_CrabSaddle.PrimalItemArmor_CrabSaddle_C":                                                        arkobject.EquipmentSaddle,
	"/Game/Aberration/Dinos/MoleRat/PrimalItemArmor_MoleRatSaddle.PrimalItemArmor_MoleRatSaddle_C":                                               arkobject.EquipmentSaddle,
	"/Game/Aberration/Dinos/RockDrake/PrimalItemArmor_RockDrakeSaddle.PrimalItemArmor_RockDrakeSaddle_C":                                         arkobject.EquipmentSaddle,
	"/Game/Aberration/Dinos/RockDrake/PrimalItemArmor_RockDrakeSaddle_Tek.PrimalItemArmor_RockDrakeSaddle_Tek_C":                                 arkobject.EquipmentSaddle,
	"/Game/ScorchedEarth/Dinos/Camelsaurus/PrimalItemArmor_CamelsaurusSaddle.PrimalItemArmor_CamelsaurusSaddle_C":                                arkobject.EquipmentSaddle,
	"/Game/ScorchedEarth/Dinos/Mantis/PrimalItemArmor_MantisSaddle.PrimalItemArmor_MantisSaddle_C":                                               arkobject.EquipmentSaddle,
	"/Game/ScorchedEarth/Dinos/Moth/PrimalItemArmor_MothSaddle.PrimalItemArmor_MothSaddle_C":                                                     arkobject.EquipmentSaddle,
	"/Game/ScorchedEarth/Dinos/RockGolem/PrimalItemArmor_RockGolemSaddle.PrimalItemArmor_RockGolemSaddle_C":                                      arkobject.EquipmentSaddle,
	"/Game/ScorchedEarth/Dinos/SpineyLizard/PrimalItemArmor_SpineyLizardSaddle.PrimalItemArmor_SpineyLizardSaddle_C":                             arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Bison/PrimalItemArmor_BisonSaddle.PrimalItemArmor_BisonSaddle_C":                                                            arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Deinotherium/Dinos/Saddle/PrimalItemArmor_DeinotheriumSaddle_ASA.PrimalItemArmor_DeinotheriumSaddle_ASA_C":                  arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Dreadnoughtus/PrimalItemArmor_DreadSaddle_Platform.PrimalItemArmor_DreadSaddle_Platform_C":                                  arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Maelizard/PrimalItemArmor_MaelizardSaddle.PrimalItemArmor_MaelizardSaddle_C":                                                arkobject.EquipmentSaddle,
	"/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C":                                     arkobject.EquipmentSaddle,
	"/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GasBagsSaddle.PrimalItemArmor_GasBagsSaddle_C":                                 arkobject.EquipmentSaddle,
	"/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_OwlSaddle.PrimalItemArmor_OwlSaddle_C":                                         arkobject.EquipmentSaddle,
	"/Game/Aberration/Dinos/CaveWolf/PrimalItemArmor_CavewolfPromoSaddle.PrimalItemArmor_CavewolfPromoSaddle_C":                                  arkobject.EquipmentSaddle,
	"/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_IceJumperSaddle.PrimalItemArmor_IceJumperSaddle_C":                             arkobject.EquipmentSaddle,
	"/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_SpindlesSaddle.PrimalItemArmor_SpindlesSaddle_C":                               arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Helicoprion/Saddle/PrimalItemArmor_Helicoprion.PrimalItemArmor_Helicoprion_C":                                               arkobject.EquipmentSaddle,
	"/Gigantoraptor/Gigantoraptor/PrimalItemArmor_GigantoraptorSaddle.PrimalItemArmor_GigantoraptorSaddle_C":                                     arkobject.EquipmentSaddle,
	"/Game/Fjordur/Dinos/Desmodus/PrimalItemArmor_DesmodusSaddle.PrimalItemArmor_DesmodusSaddle_C":                                               arkobject.EquipmentSaddle,
	"/Game/LostColony/Dinos/AngelFox/PrimalItemArmor_AngelFoxSaddle.PrimalItemArmor_AngelFoxSaddle_C":                                            arkobject.EquipmentSaddle,
	"/Game/LostColony/Dinos/BossBat/Saddle/PrimalItemArmor_BossBatSaddle.PrimalItemArmor_BossBatSaddle_C":                                        arkobject.EquipmentSaddle,
	"/Game/LostColony/Dinos/BossBat/Saddle/PrimalItemArmor_BossBatSaddle_Platform.PrimalItemArmor_BossBatSaddle_Platform_C":                      arkobject.EquipmentSaddle,
	"/Game/LostColony/Dinos/DevilFox/PrimalItemArmor_DevilFoxSaddle.PrimalItemArmor_DevilFoxSaddle_C":                                            arkobject.EquipmentSaddle,
	"/Game/LostColony/Dinos/SnowDragon/PrimalItemArmor_SnowDragonSaddle.PrimalItemArmor_SnowDragonSaddle_C":                                      arkobject.EquipmentSaddle,
	"/Game/LostColony/Dinos/SnowMonster/PrimalItemArmor_SnowMonsterSaddle.PrimalItemArmor_SnowMonsterSaddle_C":                                   arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Megaraptor/PrimalItemArmor_ValMegaraptorSaddle.PrimalItemArmor_ValMegaraptorSaddle_C":                                       arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/FireLion/PrimalItemArmor_Saddle_FireLion.PrimalItemArmor_Saddle_FireLion_C":                                                 arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Cryolophosaurus/Saddle/PrimalItemArmor_CryoSaddle.PrimalItemArmor_CryoSaddle_C":                                             arkobject.EquipmentSaddle,
	"/Game/Valguero/Dinos/Deinonychus/PrimalItemArmor_DeinonychusSaddle.PrimalItemArmor_DeinonychusSaddle_C":                                     arkobject.EquipmentSaddle,
	"/Game/ASA/Dinos/Acrocanthosaurus/Saddle/PrimalItemArmor_AcroSaddle.PrimalItemArmor_AcroSaddle_C":                                            arkobject.EquipmentSaddle,

	"/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothBoots.PrimalItemArmor_DesertClothBoots_C":                 arkobject.EquipmentArmor,
	"/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothGloves.PrimalItemArmor_DesertClothGloves_C":               arkobject.EquipmentArmor,
	"/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothGogglesHelmet.PrimalItemArmor_DesertClothGogglesHelmet_C": arkobject.EquipmentArmor,
	"/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothPants.PrimalItemArmor_DesertClothPants_C":                 arkobject.EquipmentArmor,
	"/Game/ScorchedEarth/Outfits/PrimalItemArmor_DesertClothShirt.PrimalItemArmor_DesertClothShirt_C":                 arkobject.EquipmentArmor,
}

func NewEquipment(save *arksave.Save) *EquipmentAPI {
	return &EquipmentAPI{save: save}
}

func UpstreamWeaponBlueprints() []string {
	return sortedBlueprintKeys(upstreamWeaponBlueprints)
}

func UpstreamArmorBlueprints() []string {
	return sortedBlueprintKeys(upstreamArmorBlueprints)
}

func sortedBlueprintKeys(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func (e *EquipmentAPI) IsApplicableBlueprint(blueprint string) bool {
	return e.KindForBlueprint(blueprint) != arkobject.EquipmentUnknown
}

func (e *EquipmentAPI) KindForBlueprint(blueprint string) arkobject.EquipmentKind {
	canonical := canonicalBlueprintPath(blueprint)
	switch {
	case strings.Contains(blueprint, "PrimalItemAmmo") || strings.Contains(blueprint, "PrimalItem_WeaponEmptyCryopod"):
		return arkobject.EquipmentUnknown
	case canonicalEquipmentKinds[canonical] != arkobject.EquipmentUnknown:
		return canonicalEquipmentKinds[canonical]
	case strings.Contains(blueprint, "/Weapons/") || strings.Contains(blueprint, "/CursedWeapons/"):
		return arkobject.EquipmentWeapon
	case strings.Contains(blueprint, "/Saddles/") || strings.Contains(blueprint, "/CursedSaddles/"):
		return arkobject.EquipmentSaddle
	case strings.Contains(blueprint, "/Armor/Shields/") || strings.Contains(blueprint, "/CursedArmor/Shields/"):
		return arkobject.EquipmentShield
	case strings.Contains(blueprint, "/Armor/") || strings.Contains(blueprint, "/CursedArmor/"):
		return arkobject.EquipmentArmor
	default:
		return arkobject.EquipmentUnknown
	}
}

func canonicalBlueprintPath(blueprint string) string {
	blueprint = strings.TrimPrefix(blueprint, "Blueprint'")
	blueprint = strings.TrimSuffix(blueprint, "'")
	return blueprint
}

func (e *EquipmentAPI) All() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.AllMatchingBlueprints(nil)
}

func (e *EquipmentAPI) AllMatchingBlueprints(match func(string) bool) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	objects, err := e.save.ParsedObjects(e.equipmentBlueprintFilter(match))
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for _, info := range objects {
		kind := e.KindForBlueprint(info.Object.Blueprint)
		if kind == arkobject.EquipmentUnknown {
			continue
		}
		if boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.EquipmentItemFromObject(info.Object, kind)
	}
	return out, nil
}

func (e *EquipmentAPI) AllWithFaults() (map[uuid.UUID]arkobject.EquipmentItem, []arksave.FaultyObjectInfo, error) {
	return e.AllMatchingBlueprintsWithFaults(nil)
}

func (e *EquipmentAPI) AllMatchingBlueprintsWithFaults(match func(string) bool) (map[uuid.UUID]arkobject.EquipmentItem, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := e.save.ParsedObjectsWithFaults(e.equipmentBlueprintFilter(match))
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for _, info := range objects {
		kind := e.KindForBlueprint(info.Object.Blueprint)
		if kind == arkobject.EquipmentUnknown {
			continue
		}
		if boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.EquipmentItemFromObject(info.Object, kind)
	}
	return out, faults, nil
}

func (e *EquipmentAPI) equipmentBlueprintFilter(match func(string) bool) func(arksave.ObjectClassInfo) bool {
	return func(info arksave.ObjectClassInfo) bool {
		if match != nil && !match(info.ClassName) {
			return false
		}
		return e.IsApplicableBlueprint(info.ClassName)
	}
}

func (e *EquipmentAPI) ByKind(kind arkobject.EquipmentKind) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	all, err := e.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range all {
		if item.Kind == kind {
			out[id] = item
		}
	}
	return out, nil
}

func (e *EquipmentAPI) Weapons() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.ByKind(arkobject.EquipmentWeapon)
}

func (e *EquipmentAPI) Saddles() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.ByKind(arkobject.EquipmentSaddle)
}

func (e *EquipmentAPI) Armor() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.ByKind(arkobject.EquipmentArmor)
}

func (e *EquipmentAPI) Shields() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.ByKind(arkobject.EquipmentShield)
}

func (e *EquipmentAPI) Count(items map[uuid.UUID]arkobject.EquipmentItem) int32 {
	var count int32
	for _, item := range items {
		count += item.Quantity
	}
	return count
}

func (e *EquipmentAPI) ByClass(blueprints []string) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	all, err := e.All()
	if err != nil {
		return nil, err
	}
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range all {
		if _, ok := allowed[item.Blueprint]; ok {
			out[id] = item
		}
	}
	return out, nil
}

func (e *EquipmentAPI) ByKindClass(kind arkobject.EquipmentKind, blueprints []string) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	all, err := e.All()
	if err != nil {
		return nil, err
	}
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range all {
		if item.Kind != kind {
			continue
		}
		if _, ok := allowed[item.Blueprint]; ok {
			out[id] = item
		}
	}
	return out, nil
}

func (e *EquipmentAPI) ByCrafter(crafter arkobject.ObjectCrafter) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	all, err := e.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range all {
		if item.Crafter == nil {
			continue
		}
		if *item.Crafter == crafter {
			out[id] = item
		}
	}
	return out, nil
}

func (e *EquipmentAPI) OwnedBy(owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	all, err := e.All()
	if err != nil {
		return nil, err
	}
	return e.FilterOwnedBy(all, owner)
}

func (e *EquipmentAPI) FilterOwnedBy(items map[uuid.UUID]arkobject.EquipmentItem, owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	structures := NewStructure(e.save)
	containers := map[uuid.UUID]*arkobject.Structure{}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range items {
		if item.OwnerInventory == nil {
			continue
		}
		container, cached := containers[*item.OwnerInventory]
		if !cached {
			_, structure, ok, err := structures.ContainerOfInventory(*item.OwnerInventory)
			if err != nil {
				return nil, err
			}
			if ok {
				container = &structure
			}
			containers[*item.OwnerInventory] = container
		}
		if container != nil && container.IsOwnedBy(owner) {
			out[id] = item
		}
	}
	return out, nil
}

func (e *EquipmentAPI) Equipped() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.IsEquipped
	})
}

func (e *EquipmentAPI) Blueprints() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.IsBlueprint
	})
}

func (e *EquipmentAPI) ByQuality(quality int32) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.Quality == quality
	})
}

func (e *EquipmentAPI) WithMinRating(min float64) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.Rating >= min
	})
}

func (e *EquipmentAPI) WithMinDurability(min float64) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.CurrentDurability >= min
	})
}

func (e *EquipmentAPI) WithMinActualDurability(min float64) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.Stats.Durability >= min
	})
}

func (e *EquipmentAPI) WithMinDamage(min float64) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.Kind == arkobject.EquipmentWeapon && item.Stats.Damage >= min
	})
}

func (e *EquipmentAPI) WithMinArmor(min float64) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return (item.Kind == arkobject.EquipmentArmor || item.Kind == arkobject.EquipmentSaddle || item.Kind == arkobject.EquipmentShield) &&
			item.Stats.Armor >= min
	})
}

func (e *EquipmentAPI) WithMinHypothermalResistance(min float64) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.Kind == arkobject.EquipmentArmor && item.Stats.HypothermalResistance >= min
	})
}

func (e *EquipmentAPI) WithMinHyperthermalResistance(min float64) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	return e.filter(func(item arkobject.EquipmentItem) bool {
		return item.Kind == arkobject.EquipmentArmor && item.Stats.HyperthermalResistance >= min
	})
}

func (e *EquipmentAPI) BestWeaponDamage(items map[uuid.UUID]arkobject.EquipmentItem) (uuid.UUID, arkobject.EquipmentItem, bool) {
	return e.best(items, func(item arkobject.EquipmentItem) (float64, bool) {
		return item.Stats.Damage, item.Kind == arkobject.EquipmentWeapon
	})
}

func (e *EquipmentAPI) BestActualDurability(items map[uuid.UUID]arkobject.EquipmentItem) (uuid.UUID, arkobject.EquipmentItem, bool) {
	return e.best(items, func(item arkobject.EquipmentItem) (float64, bool) {
		return item.Stats.Durability, true
	})
}

func (e *EquipmentAPI) FilterAscendantWeaponBlueprints(items map[uuid.UUID]arkobject.EquipmentItem) map[uuid.UUID]arkobject.EquipmentItem {
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range items {
		if item.Kind == arkobject.EquipmentWeapon && item.IsBlueprint && item.Quality >= AscendantQualityIndex {
			out[id] = item
		}
	}
	return out
}

func (e *EquipmentAPI) best(items map[uuid.UUID]arkobject.EquipmentItem, value func(arkobject.EquipmentItem) (float64, bool)) (uuid.UUID, arkobject.EquipmentItem, bool) {
	var bestID uuid.UUID
	var bestItem arkobject.EquipmentItem
	var bestValue float64
	found := false
	for id, item := range items {
		current, ok := value(item)
		if !ok {
			continue
		}
		if !found || current > bestValue || (current == bestValue && id.String() < bestID.String()) {
			bestID = id
			bestItem = item
			bestValue = current
			found = true
		}
	}
	return bestID, bestItem, found
}

func (e *EquipmentAPI) Filtered(opts EquipmentFilterOptions) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	match, err := equipmentFilterPredicate(opts)
	if err != nil {
		return nil, err
	}
	return e.filter(match)
}

func (e *EquipmentAPI) FilteredWithFaults(opts EquipmentFilterOptions) (map[uuid.UUID]arkobject.EquipmentItem, []arksave.FaultyObjectInfo, error) {
	match, err := equipmentFilterPredicate(opts)
	if err != nil {
		return nil, nil, err
	}
	all, faults, err := e.AllWithFaults()
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range all {
		if match(item) {
			out[id] = item
		}
	}
	return out, faults, nil
}

func equipmentFilterPredicate(opts EquipmentFilterOptions) (func(arkobject.EquipmentItem) bool, error) {
	if opts.NoBlueprints && opts.OnlyBlueprints {
		return nil, fmt.Errorf("cannot filter by both no blueprints and only blueprints")
	}
	allowedKinds := map[arkobject.EquipmentKind]struct{}{}
	for _, kind := range opts.Kinds {
		allowedKinds[kind] = struct{}{}
	}
	allowedBlueprints := map[string]struct{}{}
	for _, blueprint := range opts.Blueprints {
		allowedBlueprints[blueprint] = struct{}{}
	}
	return func(item arkobject.EquipmentItem) bool {
		if len(allowedKinds) > 0 {
			if _, ok := allowedKinds[item.Kind]; !ok {
				return false
			}
		}
		if len(allowedBlueprints) > 0 {
			_, exact := allowedBlueprints[item.Blueprint]
			_, canonical := allowedBlueprints[canonicalBlueprintPath(item.Blueprint)]
			if !exact && !canonical {
				return false
			}
		}
		if opts.NoBlueprints && item.IsBlueprint {
			return false
		}
		if opts.OnlyBlueprints && !item.IsBlueprint {
			return false
		}
		if item.Quality < opts.MinQuality {
			return false
		}
		if opts.MinRating != 0 && item.Rating < opts.MinRating {
			return false
		}
		if opts.MinDurability != 0 && item.CurrentDurability < opts.MinDurability {
			return false
		}
		if opts.Equipped != nil && item.IsEquipped != *opts.Equipped {
			return false
		}
		if opts.Crafter != nil {
			if item.Crafter == nil || *item.Crafter != *opts.Crafter {
				return false
			}
		}
		return true
	}, nil
}

func (e *EquipmentAPI) filter(match func(arkobject.EquipmentItem) bool) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	all, err := e.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range all {
		if match(item) {
			out[id] = item
		}
	}
	return out, nil
}
