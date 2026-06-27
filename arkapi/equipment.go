package arkapi

import (
	"encoding/json"
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

type EquipmentRankOptions struct {
	MinRating        float64
	ExcludeCrafted   bool
	IgnoredNameParts []string
}

type EquipmentRankStats struct {
	Ranked          int
	BestRating      float64
	BestAverageStat float64
	Crafted         int
	Blueprints      int
	Classes         int
}

type EquipmentSummary struct {
	Items                int
	TotalQuantity        int32
	ByKind               map[arkobject.EquipmentKind]int
	CryopodSaddles       int
	Blueprints           int
	Crafted              int
	Equipped             int
	Classes              int
	MaxQuality           int32
	MaxRating            float64
	MaxDamage            float64
	MaxArmor             float64
	MaxStatDurability    float64
	MaxCurrentDurability float64
}

type EquipmentSaddleSummary struct {
	ItemSaddles    int
	CryopodSaddles int
	TotalSaddles   int
	MaxArmor       float64
}

type EquipmentOwnedSummary struct {
	Items     int
	MaxDamage float64
}

type equipmentHistoryIdentity struct {
	Blueprint   string  `json:"blueprint"`
	Kind        string  `json:"kind"`
	IsBlueprint bool    `json:"is_blueprint"`
	Rating      float64 `json:"rating"`
	Quality     int32   `json:"quality"`
	Damage      float64 `json:"damage,omitempty"`
	Armor       float64 `json:"armor,omitempty"`
	Durability  float64 `json:"durability,omitempty"`
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

func UpstreamSaddleBlueprints() []string {
	return sortedBlueprintKeys(upstreamSaddleBlueprints)
}

func UpstreamShieldBlueprints() []string {
	return sortedBlueprintKeys(upstreamShieldBlueprints)
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
	containers, err := selectedInventoryContainerOwners(e.save)
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range items {
		if item.OwnerInventory == nil {
			continue
		}
		containerOwner, ok := containers[*item.OwnerInventory]
		if ok && ownerMatches(containerOwner, owner) {
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

func (e *EquipmentAPI) BestArmor(items map[uuid.UUID]arkobject.EquipmentItem) (uuid.UUID, arkobject.EquipmentItem, bool) {
	return e.best(items, func(item arkobject.EquipmentItem) (float64, bool) {
		switch item.Kind {
		case arkobject.EquipmentArmor, arkobject.EquipmentSaddle, arkobject.EquipmentShield:
			return item.Stats.Armor, true
		default:
			return 0, false
		}
	})
}

func (e *EquipmentAPI) BestActualDurability(items map[uuid.UUID]arkobject.EquipmentItem) (uuid.UUID, arkobject.EquipmentItem, bool) {
	return e.best(items, func(item arkobject.EquipmentItem) (float64, bool) {
		return item.Stats.Durability, true
	})
}

func (e *EquipmentAPI) BestWeaponDamageWithFaults(opts EquipmentFilterOptions) (uuid.UUID, arkobject.EquipmentItem, bool, []arksave.FaultyObjectInfo, error) {
	items, faults, err := e.FilteredWithFaults(opts)
	if err != nil {
		return uuid.Nil, arkobject.EquipmentItem{}, false, nil, err
	}
	id, item, ok := e.BestWeaponDamage(items)
	return id, item, ok, faults, nil
}

func (e *EquipmentAPI) BestActualDurabilityWithFaults(opts EquipmentFilterOptions) (uuid.UUID, arkobject.EquipmentItem, bool, []arksave.FaultyObjectInfo, error) {
	items, faults, err := e.FilteredWithFaults(opts)
	if err != nil {
		return uuid.Nil, arkobject.EquipmentItem{}, false, nil, err
	}
	id, item, ok := e.BestActualDurability(items)
	return id, item, ok, faults, nil
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

func (e *EquipmentAPI) RankedCandidatesWithFaults() (map[uuid.UUID]arkobject.EquipmentItem, []arksave.FaultyObjectInfo, error) {
	allowedKinds := map[string]arkobject.EquipmentKind{}
	for _, group := range []struct {
		kind       arkobject.EquipmentKind
		blueprints []string
	}{
		{kind: arkobject.EquipmentWeapon, blueprints: UpstreamWeaponBlueprints()},
		{kind: arkobject.EquipmentArmor, blueprints: UpstreamArmorBlueprints()},
		{kind: arkobject.EquipmentShield, blueprints: UpstreamShieldBlueprints()},
		{kind: arkobject.EquipmentSaddle, blueprints: UpstreamSaddleBlueprints()},
	} {
		for _, blueprint := range group.blueprints {
			allowedKinds[blueprint] = group.kind
		}
	}
	objects, faults, err := e.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		_, ok := allowedKinds[canonicalBlueprintPath(info.ClassName)]
		return ok
	})
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for _, info := range objects {
		kind, ok := allowedKinds[canonicalBlueprintPath(info.Object.Blueprint)]
		if !ok {
			continue
		}
		if boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.EquipmentItemFromObject(info.Object, kind)
	}
	return out, faults, nil
}

func (e *EquipmentAPI) RankStats(items map[uuid.UUID]arkobject.EquipmentItem, opts EquipmentRankOptions) EquipmentRankStats {
	stats := EquipmentRankStats{}
	classes := map[string]struct{}{}
	for _, item := range items {
		if opts.MinRating != 0 && item.Rating < opts.MinRating {
			continue
		}
		if opts.ExcludeCrafted && item.IsCrafted() {
			continue
		}
		if ignoredEquipmentNamePart(item, opts.IgnoredNameParts) {
			continue
		}
		stats.Ranked++
		if item.Rating > stats.BestRating {
			stats.BestRating = item.Rating
		}
		if average := item.AverageStat(); average > stats.BestAverageStat {
			stats.BestAverageStat = average
		}
		if item.IsCrafted() {
			stats.Crafted++
		}
		if item.IsBlueprint {
			stats.Blueprints++
		}
		classes[item.Blueprint] = struct{}{}
	}
	stats.Classes = len(classes)
	return stats
}

func (e *EquipmentAPI) Summary(items map[uuid.UUID]arkobject.EquipmentItem) EquipmentSummary {
	summary := EquipmentSummary{ByKind: map[arkobject.EquipmentKind]int{}}
	classes := map[string]struct{}{}
	for _, item := range items {
		summary.Items++
		summary.TotalQuantity += item.Quantity
		summary.ByKind[item.Kind]++
		if item.IsBlueprint {
			summary.Blueprints++
		}
		if item.IsCrafted() {
			summary.Crafted++
		}
		if item.IsEquipped {
			summary.Equipped++
		}
		if item.Quality > summary.MaxQuality {
			summary.MaxQuality = item.Quality
		}
		if item.Rating > summary.MaxRating {
			summary.MaxRating = item.Rating
		}
		if item.Stats.Damage > summary.MaxDamage {
			summary.MaxDamage = item.Stats.Damage
		}
		if item.Stats.Armor > summary.MaxArmor {
			summary.MaxArmor = item.Stats.Armor
		}
		if item.Stats.Durability > summary.MaxStatDurability {
			summary.MaxStatDurability = item.Stats.Durability
		}
		if item.CurrentDurability > summary.MaxCurrentDurability {
			summary.MaxCurrentDurability = item.CurrentDurability
		}
		if item.Blueprint != "" {
			classes[canonicalBlueprintPath(item.Blueprint)] = struct{}{}
		}
	}
	summary.Classes = len(classes)
	return summary
}

func (e *EquipmentAPI) SummaryWithFaults(opts EquipmentFilterOptions) (EquipmentSummary, []arksave.FaultyObjectInfo, error) {
	items, faults, err := e.FilteredWithFaults(opts)
	if err != nil {
		return EquipmentSummary{}, nil, err
	}
	return e.Summary(items), faults, nil
}

func (e *EquipmentAPI) SummaryIncludingCryopodSaddlesWithFaults(opts EquipmentFilterOptions) (EquipmentSummary, []arksave.FaultyObjectInfo, error) {
	match, err := equipmentFilterPredicate(opts)
	if err != nil {
		return EquipmentSummary{}, nil, err
	}
	directItems, faults, err := e.FilteredWithFaults(opts)
	if err != nil {
		return EquipmentSummary{}, nil, err
	}
	summary := e.Summary(directItems)
	cryopodSaddles, cryopodFaults, err := NewDino(e.save).SaddlesFromCryopodsWithFaults()
	if err != nil {
		return EquipmentSummary{}, nil, err
	}
	filteredCryopodSaddles := map[uuid.UUID]arkobject.EquipmentItem{}
	for id, item := range cryopodSaddles {
		if match(item) {
			filteredCryopodSaddles[id] = item
		}
	}
	cryopodSummary := e.Summary(filteredCryopodSaddles)
	summary.Items += cryopodSummary.Items
	summary.TotalQuantity += cryopodSummary.TotalQuantity
	for kind, count := range cryopodSummary.ByKind {
		summary.ByKind[kind] += count
	}
	summary.CryopodSaddles = cryopodSummary.Items
	summary.Blueprints += cryopodSummary.Blueprints
	summary.Crafted += cryopodSummary.Crafted
	summary.Equipped += cryopodSummary.Equipped
	if cryopodSummary.MaxQuality > summary.MaxQuality {
		summary.MaxQuality = cryopodSummary.MaxQuality
	}
	if cryopodSummary.MaxRating > summary.MaxRating {
		summary.MaxRating = cryopodSummary.MaxRating
	}
	if cryopodSummary.MaxDamage > summary.MaxDamage {
		summary.MaxDamage = cryopodSummary.MaxDamage
	}
	if cryopodSummary.MaxArmor > summary.MaxArmor {
		summary.MaxArmor = cryopodSummary.MaxArmor
	}
	if cryopodSummary.MaxStatDurability > summary.MaxStatDurability {
		summary.MaxStatDurability = cryopodSummary.MaxStatDurability
	}
	if cryopodSummary.MaxCurrentDurability > summary.MaxCurrentDurability {
		summary.MaxCurrentDurability = cryopodSummary.MaxCurrentDurability
	}
	directClasses := map[string]struct{}{}
	for _, item := range directItems {
		if item.Blueprint != "" {
			directClasses[canonicalBlueprintPath(item.Blueprint)] = struct{}{}
		}
	}
	for _, item := range filteredCryopodSaddles {
		if item.Blueprint != "" {
			directClasses[canonicalBlueprintPath(item.Blueprint)] = struct{}{}
		}
	}
	summary.Classes = len(directClasses)
	faults = append(faults, cryopodFaults...)
	return summary, faults, nil
}

func EquipmentHistorySnapshotFromPath(path string) (map[string]struct{}, error) {
	save, err := arksave.Open(path)
	if err != nil {
		return nil, err
	}
	defer save.Close()

	exported, err := NewJSON(save).ExportDomain("equipment")
	if err != nil {
		return nil, err
	}
	items, ok := exported.Items.([]EquipmentInfo)
	if !ok {
		return nil, fmt.Errorf("equipment export item type %T", exported.Items)
	}
	out := map[string]struct{}{}
	for _, item := range items {
		identity := equipmentHistoryIdentity{
			Blueprint:   item.Blueprint,
			Kind:        item.Kind,
			IsBlueprint: item.IsBlueprint,
			Rating:      item.Rating,
			Quality:     item.Quality,
		}
		if item.Stats != nil {
			identity.Damage = item.Stats.Damage
			identity.Armor = item.Stats.Armor
			identity.Durability = item.Stats.Durability
		}
		data, err := json.Marshal(identity)
		if err != nil {
			return nil, err
		}
		out[string(data)] = struct{}{}
	}
	return out, nil
}

func DiffEquipmentHistorySnapshots(previous map[string]struct{}, current map[string]struct{}) (int, int) {
	added := 0
	for key := range current {
		if _, ok := previous[key]; !ok {
			added++
		}
	}
	removed := 0
	for key := range previous {
		if _, ok := current[key]; !ok {
			removed++
		}
	}
	return added, removed
}

func (e *EquipmentAPI) SaddleSummaryWithFaults() (EquipmentSaddleSummary, []arksave.FaultyObjectInfo, error) {
	itemSaddles, faults, err := e.FilteredWithFaults(EquipmentFilterOptions{
		Kinds:      []arkobject.EquipmentKind{arkobject.EquipmentSaddle},
		Blueprints: UpstreamSaddleBlueprints(),
	})
	if err != nil {
		return EquipmentSaddleSummary{}, nil, err
	}
	cryopodSaddles, cryopodFaults, err := NewDino(e.save).SaddlesFromCryopodsWithFaults()
	if err != nil {
		return EquipmentSaddleSummary{}, nil, err
	}
	summary := EquipmentSaddleSummary{
		ItemSaddles:    len(itemSaddles),
		CryopodSaddles: len(cryopodSaddles),
		TotalSaddles:   len(itemSaddles) + len(cryopodSaddles),
	}
	if _, saddle, ok := e.BestArmor(itemSaddles); ok {
		summary.MaxArmor = saddle.Stats.Armor
	}
	if _, saddle, ok := e.BestArmor(cryopodSaddles); ok && saddle.Stats.Armor > summary.MaxArmor {
		summary.MaxArmor = saddle.Stats.Armor
	}
	faults = append(faults, cryopodFaults...)
	return summary, faults, nil
}

func (e *EquipmentAPI) OwnedSummaryWithFaults(opts EquipmentFilterOptions, owner arkobject.ObjectOwner) (EquipmentOwnedSummary, []arksave.FaultyObjectInfo, error) {
	items, faults, err := e.FilteredWithFaults(opts)
	if err != nil {
		return EquipmentOwnedSummary{}, nil, err
	}
	owned, err := e.FilterOwnedBy(items, owner)
	if err != nil {
		return EquipmentOwnedSummary{}, faults, err
	}
	summary := EquipmentOwnedSummary{Items: len(owned)}
	if _, item, ok := e.BestWeaponDamage(owned); ok {
		summary.MaxDamage = item.Stats.Damage
	}
	return summary, faults, nil
}

func ignoredEquipmentNamePart(item arkobject.EquipmentItem, parts []string) bool {
	shortName := arkobject.ShortNameFromBlueprint(item.Blueprint)
	for _, part := range parts {
		if strings.Contains(shortName, part) {
			return true
		}
	}
	return false
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

func (e *EquipmentAPI) CanonicalCountWithFaults(kind arkobject.EquipmentKind, blueprints []string) (int, []arksave.FaultyObjectInfo, error) {
	items, faults, err := e.FilteredWithFaults(EquipmentFilterOptions{
		Kinds:      []arkobject.EquipmentKind{kind},
		Blueprints: blueprints,
	})
	if err != nil {
		return 0, nil, err
	}
	return len(items), faults, nil
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
