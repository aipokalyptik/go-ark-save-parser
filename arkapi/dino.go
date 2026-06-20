package arkapi

import (
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type DinoAPI struct {
	save *arksave.Save
}

const cryopodMaxInflatedBytes int64 = 512 * 1024 * 1024

type DinoFilterOptions struct {
	MinLevel    *int32
	MaxLevel    *int32
	Blueprints  []string
	Tamed       *bool
	Cryopodded  *bool
	StatMinimum int32
	Stats       []arkobject.DinoStat
	GeneTraits  []string
}

type BabyFilterOptions struct {
	IncludeTamed      bool
	IncludeCryopodded bool
	IncludeWild       bool
}

var nonTameableDinoBlueprints = map[string]struct{}{
	"/Game/Aberration/Dinos/Basilisk/MegaBasilisk_Character_BP.MegaBasilisk_Character_BP_C":                                         {},
	"/Game/Aberration/Dinos/ChupaCabra/ChupaCabra_Character_BP_Surface.ChupaCabra_Character_BP_Surface_C":                           {},
	"/Game/Aberration/Dinos/Crab/MegaCrab_Character_BP.MegaCrab_Character_BP_C":                                                     {},
	"/Game/Aberration/Dinos/Lamprey/Lamprey_Character.Lamprey_Character_C":                                                          {},
	"/Game/Aberration/Dinos/Lightbug/Lightbug_Character_BaseBP.Lightbug_Character_BaseBP_C":                                         {},
	"/Game/Aberration/Dinos/Nameless/MegaXenomorph_Character_BP_Male_Surface.MegaXenomorph_Character_BP_Male_Surface_C":             {},
	"/Game/Aberration/Dinos/Nameless/Xenomorph_Character_BP_Male_Surface.Xenomorph_Character_BP_Male_Surface_C":                     {},
	"/Game/ASA/Dinos/Iceworm/Iceworm_Character_Minion_BP_smaller.Iceworm_Character_Minion_BP_smaller_C":                             {},
	"/Game/Extinction/Dinos/Corrupt/Arthropluera/Arthro_Character_BP_Corrupt.Arthro_Character_BP_Corrupt_C":                         {},
	"/Game/Extinction/Dinos/Corrupt/Carno/Carno_Character_BP_Corrupt.Carno_Character_BP_Corrupt_C":                                  {},
	"/Game/Extinction/Dinos/Corrupt/Chalicotherium/Chalico_Character_BP_Corrupt.Chalico_Character_BP_Corrupt_C":                     {},
	"/Game/Extinction/Dinos/Corrupt/Dilo/Dilo_Character_BP_Corrupt.Dilo_Character_BP_Corrupt_C":                                     {},
	"/Game/Extinction/Dinos/Corrupt/Giganotosaurus/Gigant_Character_BP_Corrupt.Gigant_Character_BP_Corrupt_C":                       {},
	"/Game/Extinction/Dinos/Corrupt/Nameless/Xenomorph_Character_BP_Male_Tamed_Corrupt.Xenomorph_Character_BP_Male_Tamed_Corrupt_C": {},
	"/Game/Extinction/Dinos/Corrupt/Paraceratherium/Paracer_Character_BP_Corrupt.Paracer_Character_BP_Corrupt_C":                    {},
	"/Game/Extinction/Dinos/Corrupt/Ptero/Ptero_Character_BP_Corrupt.Ptero_Character_BP_Corrupt_C":                                  {},
	"/Game/Extinction/Dinos/Corrupt/Raptor/Raptor_Character_BP_Corrupt.Raptor_Character_BP_Corrupt_C":                               {},
	"/Game/Extinction/Dinos/Corrupt/Rex/MegaRex_Character_BP_Corrupt.MegaRex_Character_BP_Corrupt_C":                                {},
	"/Game/Extinction/Dinos/Corrupt/Rex/Rex_Character_BP_Corrupt.Rex_Character_BP_Corrupt_C":                                        {},
	"/Game/Extinction/Dinos/Corrupt/RockDrake/RockDrake_Character_BP_Corrupt.RockDrake_Character_BP_Corrupt_C":                      {},
	"/Game/Extinction/Dinos/Corrupt/Spino/Spino_Character_BP_Corrupt.Spino_Character_BP_Corrupt_C":                                  {},
	"/Game/Extinction/Dinos/Corrupt/Stego/Stego_Character_BP_Corrupt.Stego_Character_BP_Corrupt_C":                                  {},
	"/Game/Extinction/Dinos/Corrupt/Trike/Trike_Character_BP_Corrupt.Trike_Character_BP_Corrupt_C":                                  {},
	"/Game/Extinction/Dinos/Corrupt/Wyvern/Wyvern_Character_BP_Fire_Corrupt.Wyvern_Character_BP_Fire_Corrupt_C":                     {},
	"/Game/Extinction/Dinos/Enforcer/Enforcer_Character_BP.Enforcer_Character_BP_C":                                                 {},
	"/Game/Extinction/Dinos/Scout/Scout_Character_BP.Scout_Character_BP_C":                                                          {},
	"/Game/Extinction/Dinos/Tank/Defender_Character_BP.Defender_Character_BP_C":                                                     {},
	"/Game/LostColony/Dinos/Zombie/MegaZombie_Character_BP_Hulking.MegaZombie_Character_BP_Hulking_C":                               {},
	"/Game/PrimalEarth/Dinos/Ant/Ant_Character_BP.Ant_Character_BP_C":                                                               {},
	"/Game/PrimalEarth/Dinos/Ant/FlyingAnt_Character_BP.FlyingAnt_Character_BP_C":                                                   {},
	"/Game/PrimalEarth/Dinos/Bigfoot/Yeti_Character_BP.Yeti_Character_BP_C":                                                         {},
	"/Game/PrimalEarth/Dinos/Carno/MegaCarno_Character_BP.MegaCarno_Character_BP_C":                                                 {},
	"/Game/PrimalEarth/Dinos/Cnidaria/Cnidaria_Character_BP.Cnidaria_Character_BP_C":                                                {},
	"/Game/PrimalEarth/Dinos/Coelacanth/Coel_Character_BP.Coel_Character_BP_C":                                                      {},
	"/Game/PrimalEarth/Dinos/Coelacanth/Coel_Character_BP_Ocean.Coel_Character_BP_Ocean_C":                                          {},
	"/Game/PrimalEarth/Dinos/Direbear/Direbear_Character_Polar.Direbear_Character_Polar_C":                                          {},
	"/Game/PrimalEarth/Dinos/DodoRex/DodoRex_Character_BP.DodoRex_Character_BP_C":                                                   {},
	"/Game/PrimalEarth/Dinos/Dragonfly/Dragonfly_Character_BP.Dragonfly_Character_BP_C":                                             {},
	"/Game/PrimalEarth/Dinos/Giganotosaurus/BionicGigant_Character_BP.BionicGigant_Character_BP_C":                                  {},
	"/Game/PrimalEarth/Dinos/Leech/Leech_Character.Leech_Character_C":                                                               {},
	"/Game/PrimalEarth/Dinos/Leech/Leech_Character_Diseased.Leech_Character_Diseased_C":                                             {},
	"/Game/PrimalEarth/Dinos/Leedsichthys/Alpha_Leedsichthys_Character_BP.Alpha_Leedsichthys_Character_BP_C":                        {},
	"/Game/PrimalEarth/Dinos/Leedsichthys/Leedsichthys_Character_BP.Leedsichthys_Character_BP_C":                                    {},
	"/Game/PrimalEarth/Dinos/Megalodon/MEgaMegalodon_Character_BP.MegaMegalodon_Character_BP_C":                                     {},
	"/Game/PrimalEarth/Dinos/Mosasaurus/Mosa_Character_BP_Mega.Mosa_Character_BP_Mega_C":                                            {},
	"/Game/PrimalEarth/Dinos/Piranha/Piranha_Character_BP.Piranha_Character_BP_C":                                                   {},
	"/Game/PrimalEarth/Dinos/Raptor/MegaRaptor_Character_BP.MegaRaptor_Character_BP_C":                                              {},
	"/Game/PrimalEarth/Dinos/Rex/MegaRex_Character_BP.MegaRex_Character_BP_C":                                                       {},
	"/Game/PrimalEarth/Dinos/Salmon/Salmon_Character_BP.Salmon_Character_BP_C":                                                      {},
	"/Game/PrimalEarth/Dinos/Trilobite/Trilobite_Character.Trilobite_Character_C":                                                   {},
	"/Game/PrimalEarth/Dinos/Tusoteuthis/Mega_Tusoteuthis_Character_BP.Mega_Tusoteuthis_Character_BP_C":                             {},
	"/Game/ScorchedEarth/Dinos/DeathWorm/DeathWorm_Character_BP.Deathworm_Character_BP_C":                                           {},
	"/Game/ScorchedEarth/Dinos/Deathworm/MegaDeathworm_Character_BP.MegaDeathworm_Character_BP_C":                                   {},
	"/Game/ScorchedEarth/Dinos/DodoWyvern/DodoWyvern_Character_BP.DodoWyvern_Character_BP_C":                                        {},
	"/Game/ScorchedEarth/Dinos/Jugbug/Jugbug_Oil_Character_BP.Jugbug_Oil_Character_BP_C":                                            {},
	"/Game/ScorchedEarth/Dinos/Jugbug/Jugbug_Water_Character_BP.Jugbug_Water_Character_BP_C":                                        {},
	"/Game/ScorchedEarth/Dinos/Wyvern/MegaWyvern_Character_BP_Fire.MegaWyvern_Character_BP_Fire_C":                                  {},
	"/PA_Ascension/Dinos/PaleoRaptor/Alpha/Paleo_AlphaRaptor_Character_BP.Paleo_AlphaRaptor_Character_BP_C":                         {},
	"/PA_EVO_Pack_01/Dinos/EVO_Rex/Alpha/EVO_Alpha_Rex_Character_BP.EVO_Alpha_Rex_Character_BP_C":                                   {},
	"/PA_EVO_Pack_02/Dinos/EVO_Mosa/Alpha/EVO_Alpha_Mosa_Character_BP.EVO_Alpha_Mosa_Character_BP_C":                                {},
}

func NewDino(save *arksave.Save) *DinoAPI {
	return &DinoAPI{save: save}
}

func (d *DinoAPI) IsApplicableBlueprint(blueprint string) bool {
	if blueprint == "" {
		return false
	}
	if d.IsCryopodBlueprint(blueprint) {
		return true
	}
	hasDinoPath := strings.Contains(blueprint, "/Creatures/") ||
		strings.Contains(blueprint, "/Dinos/") ||
		strings.Contains(blueprint, "/SDinoVariants/")
	return hasDinoPath && strings.Contains(blueprint, "_Character_")
}

func (d *DinoAPI) IsCryopodBlueprint(blueprint string) bool {
	if blueprint == "" {
		return false
	}
	return strings.Contains(blueprint, "PrimalItem_WeaponEmptyCryopod") ||
		strings.Contains(blueprint, "PrimalItemCryopod") ||
		strings.Contains(blueprint, "SCSCryopod") ||
		strings.Contains(blueprint, "ItemDinoball")
}

func (d *DinoAPI) All() (map[uuid.UUID]arkobject.Dino, error) {
	objects, err := d.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return d.IsApplicableBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for _, info := range objects {
		if d.IsCryopodBlueprint(info.Object.Blueprint) {
			dino, ok, err := arkobject.DinoFromCryopodObject(info.Object, cryopodMaxInflatedBytes)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			out[info.UUID] = dino
			continue
		}
		var location *arkobject.ActorTransform
		if transform, ok := d.save.ActorTransform(info.UUID); ok {
			location = &transform
		}
		dino := arkobject.DinoFromObject(info.Object, location)
		if dino.StatusComponentUUID != nil {
			if statusObject, err := d.save.ParsedObject(*dino.StatusComponentUUID); err == nil {
				dino = arkobject.DinoFromObjectWithStatus(info.Object, statusObject, location)
			}
		}
		out[info.UUID] = dino
	}
	return out, nil
}

func (d *DinoAPI) AllWithFaults() (map[uuid.UUID]arkobject.Dino, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := d.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return d.IsApplicableBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for _, info := range objects {
		if d.IsCryopodBlueprint(info.Object.Blueprint) {
			dino, ok, parseErr := arkobject.DinoFromCryopodObject(info.Object, cryopodMaxInflatedBytes)
			if parseErr != nil {
				faults = append(faults, arksave.FaultyObjectInfo{UUID: info.UUID, ClassName: info.ClassName, Err: parseErr})
				continue
			}
			if ok {
				out[info.UUID] = dino
			}
			continue
		}
		var location *arkobject.ActorTransform
		if transform, ok := d.save.ActorTransform(info.UUID); ok {
			location = &transform
		}
		dino := arkobject.DinoFromObject(info.Object, location)
		if dino.StatusComponentUUID != nil {
			if statusObject, err := d.save.ParsedObject(*dino.StatusComponentUUID); err == nil {
				dino = arkobject.DinoFromObjectWithStatus(info.Object, statusObject, location)
			}
		}
		out[info.UUID] = dino
	}
	return out, faults, nil
}

func (d *DinoAPI) ByClass(blueprints []string) (map[uuid.UUID]arkobject.Dino, error) {
	all, err := d.All()
	if err != nil {
		return nil, err
	}
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if _, ok := allowed[dino.Blueprint]; ok {
			out[id] = dino
		}
	}
	return out, nil
}

func (d *DinoAPI) WildByClass(blueprints []string) (map[uuid.UUID]arkobject.Dino, error) {
	byClass, err := d.ByClass(blueprints)
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range byClass {
		if !dino.IsTamed {
			out[id] = dino
		}
	}
	return out, nil
}

func (d *DinoAPI) TamedByClass(blueprints []string, includeCryopodded bool) (map[uuid.UUID]arkobject.Dino, error) {
	byClass, err := d.ByClass(blueprints)
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range byClass {
		if dino.IsTamed && (includeCryopodded || !dino.IsCryopodded) {
			out[id] = dino
		}
	}
	return out, nil
}

func (d *DinoAPI) ByDinoID(dinoID arkobject.DinoID, includeWild bool) (uuid.UUID, arkobject.Dino, bool, error) {
	all, err := d.All()
	if err != nil {
		return uuid.Nil, arkobject.Dino{}, false, err
	}
	for id, dino := range all {
		if !includeWild && !dino.IsTamed {
			continue
		}
		if dino.ID1 == dinoID.ID1 && dino.ID2 == dinoID.ID2 {
			return id, dino, true, nil
		}
	}
	return uuid.Nil, arkobject.Dino{}, false, nil
}

func (d *DinoAPI) Tamed() (map[uuid.UUID]arkobject.Dino, error) {
	all, err := d.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if dino.IsTamed {
			out[id] = dino
		}
	}
	return out, nil
}

func (d *DinoAPI) Wild() (map[uuid.UUID]arkobject.Dino, error) {
	all, err := d.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if !dino.IsTamed {
			out[id] = dino
		}
	}
	return out, nil
}

func (d *DinoAPI) WildTamable() (map[uuid.UUID]arkobject.Dino, error) {
	wild, err := d.Wild()
	if err != nil {
		return nil, err
	}
	return d.FilterWildTamable(wild), nil
}

func (d *DinoAPI) FilterWildTamable(dinos map[uuid.UUID]arkobject.Dino) map[uuid.UUID]arkobject.Dino {
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range dinos {
		if dino.IsTamed {
			continue
		}
		if d.IsNonTameableDino(dino.Blueprint) {
			continue
		}
		out[id] = dino
	}
	return out
}

func (d *DinoAPI) IsNonTameableDino(blueprint string) bool {
	_, ok := nonTameableDinoBlueprints[canonicalBlueprintPath(blueprint)]
	return ok
}

func (d *DinoAPI) Babies() (map[uuid.UUID]arkobject.Dino, error) {
	return d.BabiesFiltered(BabyFilterOptions{
		IncludeTamed:      true,
		IncludeCryopodded: true,
		IncludeWild:       true,
	})
}

func (d *DinoAPI) BabiesFiltered(opts BabyFilterOptions) (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		if !dino.IsBaby {
			return false
		}
		if dino.IsCryopodded && !opts.IncludeCryopodded {
			return false
		}
		if dino.IsTamed {
			return opts.IncludeTamed
		}
		return opts.IncludeWild
	})
}

func (d *DinoAPI) InCryopods() (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return dino.IsCryopodded
	})
}

func (d *DinoAPI) NotInCryopods() (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return !dino.IsCryopodded
	})
}

func (d *DinoAPI) Females() (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return dino.IsFemale
	})
}

func (d *DinoAPI) Males() (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return !dino.IsFemale
	})
}

func (d *DinoAPI) Dead() (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return dino.IsDead
	})
}

func (d *DinoAPI) Alive() (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return !dino.IsDead
	})
}

func (d *DinoAPI) LevelAtLeast(level int32) (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return dino.Stats != nil && dino.Stats.CurrentLevel >= level
	})
}

func (d *DinoAPI) WildLevelAtLeast(level int32) (map[uuid.UUID]arkobject.Dino, error) {
	byLevel, err := d.LevelAtLeast(level)
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range byLevel {
		if !dino.IsTamed {
			out[id] = dino
		}
	}
	return out, nil
}

func (d *DinoAPI) TamedLevelAtLeast(level int32) (map[uuid.UUID]arkobject.Dino, error) {
	byLevel, err := d.LevelAtLeast(level)
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range byLevel {
		if dino.IsTamed {
			out[id] = dino
		}
	}
	return out, nil
}

func (d *DinoAPI) WithStatAtLeast(value int32, stats ...arkobject.DinoStat) (map[uuid.UUID]arkobject.Dino, error) {
	return d.withStatAtLeast(value, arkobject.StatScopeCombined, stats...)
}

func (d *DinoAPI) WithBaseStatAtLeast(value int32, stats ...arkobject.DinoStat) (map[uuid.UUID]arkobject.Dino, error) {
	return d.withStatAtLeast(value, arkobject.StatScopeBase, stats...)
}

func (d *DinoAPI) WithMutatedStatAtLeast(value int32, stats ...arkobject.DinoStat) (map[uuid.UUID]arkobject.Dino, error) {
	return d.withStatAtLeast(value, arkobject.StatScopeMutated, stats...)
}

func (d *DinoAPI) WithGeneTrait(name string, levels ...int) (map[uuid.UUID]arkobject.Dino, error) {
	allowedLevels := map[int]struct{}{}
	for _, level := range levels {
		allowedLevels[level] = struct{}{}
	}
	return d.filter(func(dino arkobject.Dino) bool {
		for _, trait := range dino.ParsedGeneTraits {
			if trait.Name != name {
				continue
			}
			if len(allowedLevels) == 0 {
				return true
			}
			if _, ok := allowedLevels[trait.Level]; ok {
				return true
			}
		}
		return false
	})
}

func (d *DinoAPI) OwnedByTribe(tribeID int32, includeCryopodded bool) (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return dino.IsTamed && (includeCryopodded || !dino.IsCryopodded) && dino.Owner.TargetTeam == tribeID
	})
}

func (d *DinoAPI) WildTamed() (map[uuid.UUID]arkobject.Dino, error) {
	tamed, err := d.Tamed()
	if err != nil {
		return nil, err
	}
	return d.FilterWildTamed(tamed), nil
}

func (d *DinoAPI) FilterWildTamed(dinos map[uuid.UUID]arkobject.Dino) map[uuid.UUID]arkobject.Dino {
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range dinos {
		if dino.IsWildTamed() {
			out[id] = dino
		}
	}
	return out
}

func (d *DinoAPI) ContainerOfInventory(inventoryID uuid.UUID, includeCryopodded bool) (uuid.UUID, arkobject.Dino, bool, error) {
	tamed, err := d.Tamed()
	if err != nil {
		return uuid.Nil, arkobject.Dino{}, false, err
	}
	for id, dino := range tamed {
		if !includeCryopodded && dino.IsCryopodded {
			continue
		}
		if dino.InventoryUUID != nil && *dino.InventoryUUID == inventoryID {
			return id, dino, true, nil
		}
	}
	return uuid.Nil, arkobject.Dino{}, false, nil
}

func (d *DinoAPI) ChildlessTamed() (map[uuid.UUID]arkobject.Dino, error) {
	tamed, err := d.Tamed()
	if err != nil {
		return nil, err
	}
	return d.FilterChildlessTamed(tamed), nil
}

func (d *DinoAPI) FilterChildlessTamed(dinos map[uuid.UUID]arkobject.Dino) map[uuid.UUID]arkobject.Dino {
	ancestorIDs := map[arkobject.DinoID]struct{}{}
	for _, dino := range dinos {
		if !dino.IsTamed || dino.IsBaby {
			continue
		}
		for _, ancestorID := range dino.AncestorIDs {
			if !ancestorID.IsZero() {
				ancestorIDs[ancestorID] = struct{}{}
			}
		}
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range dinos {
		if !dino.IsTamed {
			continue
		}
		dinoID := arkobject.DinoID{ID1: dino.ID1, ID2: dino.ID2}
		if _, ok := ancestorIDs[dinoID]; !ok {
			out[id] = dino
		}
	}
	return out
}

func (d *DinoAPI) CountByLevel(dinos map[uuid.UUID]arkobject.Dino) map[int32]int {
	counts := map[int32]int{}
	for _, dino := range dinos {
		if dino.Stats == nil {
			continue
		}
		counts[dino.Stats.CurrentLevel]++
	}
	return counts
}

func (d *DinoAPI) CountByClass(dinos map[uuid.UUID]arkobject.Dino) map[string]int {
	counts := map[string]int{}
	for _, dino := range dinos {
		counts[dino.Blueprint]++
	}
	return counts
}

func (d *DinoAPI) CountByShortName(dinos map[uuid.UUID]arkobject.Dino) map[string]int {
	counts := map[string]int{}
	for _, dino := range dinos {
		counts[dino.ShortName()]++
	}
	return counts
}

func (d *DinoAPI) CountByTamed(dinos map[uuid.UUID]arkobject.Dino) map[bool]int {
	counts := map[bool]int{}
	for _, dino := range dinos {
		counts[dino.IsTamed]++
	}
	return counts
}

func (d *DinoAPI) CountByCryopodded(dinos map[uuid.UUID]arkobject.Dino) map[bool]int {
	counts := map[bool]int{}
	for _, dino := range dinos {
		counts[dino.IsCryopodded]++
	}
	return counts
}

func (d *DinoAPI) CountCryopoddedByClass(dinos map[uuid.UUID]arkobject.Dino) map[string]int {
	counts := map[string]int{"all": 0}
	for _, dino := range dinos {
		if !dino.IsCryopodded {
			continue
		}
		counts["all"]++
		counts[dino.Blueprint]++
	}
	return counts
}

func (d *DinoAPI) CountCryopoddedByShortName(dinos map[uuid.UUID]arkobject.Dino) map[string]int {
	counts := map[string]int{"all": 0}
	for _, dino := range dinos {
		if !dino.IsCryopodded {
			continue
		}
		counts["all"]++
		counts[dino.ShortName()]++
	}
	return counts
}

func (d *DinoAPI) BestDinoForStat(scopes ...arkobject.StatScope) (uuid.UUID, arkobject.Dino, arkobject.DinoStat, int32, bool, error) {
	all, err := d.All()
	if err != nil {
		return uuid.Nil, arkobject.Dino{}, 0, 0, false, err
	}
	var bestID uuid.UUID
	var bestDino arkobject.Dino
	var bestStat arkobject.DinoStat
	var bestPoints int32
	found := false
	for id, dino := range all {
		if dino.Stats == nil {
			continue
		}
		stat, points, ok := dino.Stats.BestStat(scopes...)
		if !ok {
			continue
		}
		if !found || points > bestPoints || (points == bestPoints && id.String() < bestID.String()) {
			bestID = id
			bestDino = dino
			bestStat = stat
			bestPoints = points
			found = true
		}
	}
	return bestID, bestDino, bestStat, bestPoints, found, nil
}

func (d *DinoAPI) MostMutatedTamed() (uuid.UUID, arkobject.Dino, int32, bool, error) {
	tamed, err := d.Tamed()
	if err != nil {
		return uuid.Nil, arkobject.Dino{}, 0, false, err
	}
	var bestID uuid.UUID
	var bestDino arkobject.Dino
	var bestTotal int32
	found := false
	for id, dino := range tamed {
		if dino.Stats == nil {
			continue
		}
		total := dino.Stats.TotalMutations()
		if !found || total > bestTotal || (total == bestTotal && id.String() < bestID.String()) {
			bestID = id
			bestDino = dino
			bestTotal = total
			found = true
		}
	}
	return bestID, bestDino, bestTotal, found, nil
}

func (d *DinoAPI) Filtered(opts DinoFilterOptions) (map[uuid.UUID]arkobject.Dino, error) {
	allowedBlueprints := map[string]struct{}{}
	for _, blueprint := range opts.Blueprints {
		allowedBlueprints[blueprint] = struct{}{}
	}
	allowedStats := map[arkobject.DinoStat]struct{}{}
	for _, stat := range opts.Stats {
		allowedStats[stat] = struct{}{}
	}
	allowedTraits := map[string]struct{}{}
	for _, trait := range opts.GeneTraits {
		allowedTraits[trait] = struct{}{}
	}
	return d.filter(func(dino arkobject.Dino) bool {
		if len(allowedBlueprints) > 0 {
			if _, ok := allowedBlueprints[dino.Blueprint]; !ok {
				return false
			}
		}
		if len(allowedTraits) > 0 {
			matched := false
			for _, trait := range dino.ParsedGeneTraits {
				if _, ok := allowedTraits[trait.Name]; ok {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
		if opts.Tamed != nil && dino.IsTamed != *opts.Tamed {
			return false
		}
		if opts.Cryopodded != nil && dino.IsCryopodded != *opts.Cryopodded {
			return false
		}
		if opts.MinLevel != nil || opts.MaxLevel != nil || opts.StatMinimum != 0 {
			if dino.Stats == nil {
				return false
			}
		}
		if opts.MinLevel != nil && dino.Stats.CurrentLevel < *opts.MinLevel {
			return false
		}
		if opts.MaxLevel != nil && dino.Stats.CurrentLevel > *opts.MaxLevel {
			return false
		}
		if opts.StatMinimum != 0 {
			statsAbove := dino.Stats.StatsAtLeast(opts.StatMinimum, arkobject.StatScopeCombined)
			if len(statsAbove) == 0 {
				return false
			}
			if len(allowedStats) > 0 {
				for _, stat := range statsAbove {
					if _, ok := allowedStats[stat]; ok {
						return true
					}
				}
				return false
			}
		}
		return true
	})
}

func (d *DinoAPI) withStatAtLeast(value int32, scope arkobject.StatScope, stats ...arkobject.DinoStat) (map[uuid.UUID]arkobject.Dino, error) {
	allowed := map[arkobject.DinoStat]struct{}{}
	for _, stat := range stats {
		allowed[stat] = struct{}{}
	}
	return d.filter(func(dino arkobject.Dino) bool {
		if dino.Stats == nil {
			return false
		}
		for _, stat := range dino.Stats.StatsAtLeast(value, scope) {
			if len(allowed) == 0 {
				return true
			}
			if _, ok := allowed[stat]; ok {
				return true
			}
		}
		return false
	})
}

func (d *DinoAPI) filter(match func(arkobject.Dino) bool) (map[uuid.UUID]arkobject.Dino, error) {
	all, err := d.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if match(dino) {
			out[id] = dino
		}
	}
	return out, nil
}
