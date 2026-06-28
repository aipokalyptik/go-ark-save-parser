package arkapi

import (
	"errors"
	"fmt"
	"math"
	"sort"
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

type BabyCounts struct {
	Wild  int
	Tamed int
}

type WildTamableSummary struct {
	WildDinos    int
	WildTamables int
}

type DinoPopulationSummary struct {
	Dinos      int
	Tamed      int
	Wild       int
	Cryopodded int
	Classes    int
}

type DinoBestStatSummary struct {
	UUID      uuid.UUID
	Dino      arkobject.Dino
	Stat      arkobject.DinoStat
	Points    int32
	Found     bool
	Level     int32
	Blueprint string
}

type DinoMostMutatedSummary struct {
	UUID                uuid.UUID
	Dino                arkobject.Dino
	Found               bool
	TotalMutationPoints int32
	MutationPairs       int32
	Level               int32
	Blueprint           string
}

type DinoWildTamedSummary struct {
	Dinos    int
	MaxLevel int32
}

type DinoHeatmapOptions struct {
	MapName           string
	Resolution        int
	Blueprints        []string
	OnlyTamed         bool
	IncludeCryopodded bool
}

type DinoBestStatOptions struct {
	Blueprints      []string
	Stats           []arkobject.DinoStat
	OnlyTamed       bool
	OnlyUntamed     bool
	ExcludeCryopods bool
	BaseStat        bool
	MutatedStat     bool
	LevelUpperBound *int32
}

type DinoPedigreeNode struct {
	UUID            uuid.UUID
	DinoID          arkobject.DinoID
	Blueprint       string
	Name            string
	Generation      int
	IsFemale        bool
	IsBaby          bool
	DescendantCount int
	Children        []DinoPedigreeNode
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

func NewDinoFromPath(savePath string) (*DinoAPI, func() error, error) {
	save, err := arksave.Open(savePath)
	if err != nil {
		return nil, nil, err
	}
	return NewDino(save), save.Close, nil
}

func DinoPopulationSummaryFromPath(savePath string, includeCryopodded bool) (DinoPopulationSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return DinoPopulationSummary{}, nil, err
	}
	defer closeAPI()
	return api.PopulationSummaryWithFaults(includeCryopodded)
}

func DinoWildTamableSummaryFromPath(savePath string) (WildTamableSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return WildTamableSummary{}, nil, err
	}
	defer closeAPI()
	return api.WildTamableSummaryWithFaults()
}

func DinoBabySummaryFromPath(savePath string, opts BabyFilterOptions) (BabyCounts, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return BabyCounts{}, nil, err
	}
	defer closeAPI()
	return api.BabySummaryWithFaults(opts)
}

func DinoBestStatSummaryFromPath(savePath string, opts DinoBestStatOptions) (DinoBestStatSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return DinoBestStatSummary{}, nil, err
	}
	defer closeAPI()

	id, dino, stat, points, found, faults, err := api.BestDinoForStatFilteredWithFaults(opts)
	if err != nil {
		return DinoBestStatSummary{}, nil, err
	}
	return dinoBestStatSummary(id, dino, stat, points, found), faults, nil
}

func DinoMostMutatedSummaryFromPath(savePath string) (DinoMostMutatedSummary, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return DinoMostMutatedSummary{}, err
	}
	defer closeAPI()

	id, dino, total, found, err := api.MostMutatedTamed()
	if err != nil {
		return DinoMostMutatedSummary{}, err
	}
	return dinoMostMutatedSummary(id, dino, total, found), nil
}

func DinoWildTamedSummaryFromPath(savePath string) (DinoWildTamedSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return DinoWildTamedSummary{}, nil, err
	}
	defer closeAPI()

	dinos, faults, err := api.WildTamedWithFaults()
	if err != nil {
		return DinoWildTamedSummary{}, nil, err
	}
	summary := DinoWildTamedSummary{Dinos: len(dinos)}
	if level, ok := api.MaxCurrentLevel(dinos); ok {
		summary.MaxLevel = level
	}
	return summary, faults, nil
}

func dinoBestStatSummary(id uuid.UUID, dino arkobject.Dino, stat arkobject.DinoStat, points int32, found bool) DinoBestStatSummary {
	summary := DinoBestStatSummary{
		UUID:      id,
		Dino:      dino,
		Stat:      stat,
		Points:    points,
		Found:     found,
		Blueprint: dino.ShortName(),
	}
	if dino.Stats != nil {
		summary.Level = dino.Stats.CurrentLevel
	}
	return summary
}

func dinoMostMutatedSummary(id uuid.UUID, dino arkobject.Dino, total int32, found bool) DinoMostMutatedSummary {
	summary := DinoMostMutatedSummary{
		UUID:                id,
		Dino:                dino,
		Found:               found,
		TotalMutationPoints: total,
		MutationPairs:       total / 2,
		Blueprint:           dino.ShortName(),
	}
	if dino.Stats != nil {
		summary.Level = dino.Stats.CurrentLevel
	}
	return summary
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
			if statusObject, err := d.statusObject(*dino.StatusComponentUUID); err == nil {
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
			if statusObject, err := d.statusObject(*dino.StatusComponentUUID); err == nil {
				dino = arkobject.DinoFromObjectWithStatus(info.Object, statusObject, location)
			}
		}
		out[info.UUID] = dino
	}
	return out, faults, nil
}

func (d *DinoAPI) statusObject(id uuid.UUID) (*arkobject.GameObject, error) {
	statusObject, err := d.save.ParsedObject(id)
	if err == nil {
		return statusObject, nil
	}
	partial, partialErr := d.save.ParsedObjectPartial(id)
	if partial != nil {
		return partial, nil
	}
	return nil, partialErr
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

func (d *DinoAPI) WildWithFaults() (map[uuid.UUID]arkobject.Dino, []arksave.FaultyObjectInfo, error) {
	all, faults, err := d.AllWithFaults()
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if !dino.IsTamed {
			out[id] = dino
		}
	}
	return out, faults, nil
}

func (d *DinoAPI) WildTamable() (map[uuid.UUID]arkobject.Dino, error) {
	wild, err := d.Wild()
	if err != nil {
		return nil, err
	}
	return d.FilterWildTamable(wild), nil
}

func (d *DinoAPI) WildTamableWithFaults() (map[uuid.UUID]arkobject.Dino, []arksave.FaultyObjectInfo, error) {
	wild, faults, err := d.WildWithFaults()
	if err != nil {
		return nil, nil, err
	}
	return d.FilterWildTamable(wild), faults, nil
}

func (d *DinoAPI) WildTamableSummaryForDinos(dinos map[uuid.UUID]arkobject.Dino) WildTamableSummary {
	wild := 0
	for _, dino := range dinos {
		if !dino.IsTamed {
			wild++
		}
	}
	return WildTamableSummary{WildDinos: wild, WildTamables: len(d.FilterWildTamable(dinos))}
}

func (d *DinoAPI) WildTamableSummaryWithFaults() (WildTamableSummary, []arksave.FaultyObjectInfo, error) {
	wild, faults, err := d.WildWithFaults()
	if err != nil {
		return WildTamableSummary{}, nil, err
	}
	return d.WildTamableSummaryForDinos(wild), faults, nil
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
		return babyFilterMatches(dino, opts)
	})
}

func (d *DinoAPI) BabiesFilteredWithFaults(opts BabyFilterOptions) (map[uuid.UUID]arkobject.Dino, []arksave.FaultyObjectInfo, error) {
	all, faults, err := d.AllWithFaults()
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if babyFilterMatches(dino, opts) {
			out[id] = dino
		}
	}
	return out, faults, nil
}

func babyFilterMatches(dino arkobject.Dino, opts BabyFilterOptions) bool {
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
}

func (d *DinoAPI) BabiesWithFaults() (map[uuid.UUID]arkobject.Dino, []arksave.FaultyObjectInfo, error) {
	return d.BabiesFilteredWithFaults(BabyFilterOptions{
		IncludeTamed:      true,
		IncludeCryopodded: true,
		IncludeWild:       true,
	})
}

func (d *DinoAPI) BabySummaryWithFaults(opts BabyFilterOptions) (BabyCounts, []arksave.FaultyObjectInfo, error) {
	babies, faults, err := d.BabiesFilteredWithFaults(opts)
	if err != nil {
		return BabyCounts{}, nil, err
	}
	return d.CountBabiesByTamed(babies), faults, nil
}

func (d *DinoAPI) InCryopods() (map[uuid.UUID]arkobject.Dino, error) {
	return d.filter(func(dino arkobject.Dino) bool {
		return dino.IsCryopodded
	})
}

func (d *DinoAPI) SaddlesFromCryopods() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	objects, err := d.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return d.IsCryopodBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for _, info := range objects {
		saddle, ok, err := arkobject.SaddleFromCryopodObject(info.Object)
		if err != nil {
			continue
		}
		if ok {
			out[info.UUID] = saddle
		}
	}
	return out, nil
}

func (d *DinoAPI) SaddlesFromCryopodsWithFaults() (map[uuid.UUID]arkobject.EquipmentItem, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := d.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return d.IsCryopodBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for _, info := range objects {
		saddle, ok, parseErr := arkobject.SaddleFromCryopodObject(info.Object)
		if parseErr != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: info.UUID, ClassName: info.ClassName, Err: parseErr})
			continue
		}
		if ok {
			out[info.UUID] = saddle
		}
	}
	return out, faults, nil
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

func (d *DinoAPI) WildTamedWithFaults() (map[uuid.UUID]arkobject.Dino, []arksave.FaultyObjectInfo, error) {
	all, faults, err := d.AllWithFaults()
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if dino.IsWildTamed() {
			out[id] = dino
		}
	}
	return out, faults, nil
}

func (d *DinoAPI) Heatmap(mapName string, resolution int, dinos map[uuid.UUID]arkobject.Dino, blueprints []string, onlyTamed bool) ([][]int, error) {
	if resolution <= 0 {
		return nil, fmt.Errorf("resolution must be positive")
	}
	if mapName == "" && d.save != nil && d.save.Context != nil {
		mapName = d.save.Context.MapName
	}
	var err error
	if dinos == nil {
		dinos, err = d.All()
		if err != nil {
			return nil, err
		}
	}
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	heatmap := make([][]int, resolution)
	for i := range heatmap {
		heatmap[i] = make([]int, resolution)
	}
	for _, dino := range dinos {
		if dino.Location == nil {
			continue
		}
		if onlyTamed && !dino.IsTamed {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[dino.Blueprint]; !ok {
				continue
			}
		}
		coords := dino.Location.AsMapCoords(mapName)
		x := int(math.Floor(coords.Lat))
		y := int(math.Floor(coords.Long))
		if x < 0 || x >= resolution || y < 0 || y >= resolution {
			continue
		}
		heatmap[x][y]++
	}
	return heatmap, nil
}

func (d *DinoAPI) HeatmapSummaryWithFaults(opts DinoHeatmapOptions) (HeatmapSummary, []arksave.FaultyObjectInfo, error) {
	dinos, faults, err := d.AllWithFaults()
	if err != nil {
		return HeatmapSummary{}, nil, err
	}
	if !opts.IncludeCryopodded {
		for id, dino := range dinos {
			if dino.IsCryopodded {
				delete(dinos, id)
			}
		}
		faults = nil
	}
	heatmap, err := d.Heatmap(opts.MapName, opts.Resolution, dinos, opts.Blueprints, opts.OnlyTamed)
	if err != nil {
		return HeatmapSummary{}, nil, err
	}
	return SummarizeHeatmap(heatmap, len(faults)), faults, nil
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

func (d *DinoAPI) ChildrenByAncestor(dinos map[uuid.UUID]arkobject.Dino) map[arkobject.DinoID][]uuid.UUID {
	children := map[arkobject.DinoID][]uuid.UUID{}
	for id, dino := range dinos {
		if !dino.IsTamed {
			continue
		}
		for _, ancestorID := range dino.AncestorIDs {
			if ancestorID.IsZero() {
				continue
			}
			children[ancestorID] = append(children[ancestorID], id)
		}
	}
	for ancestorID := range children {
		sort.Slice(children[ancestorID], func(i, j int) bool {
			return children[ancestorID][i].String() < children[ancestorID][j].String()
		})
	}
	return children
}

func (d *DinoAPI) DescendantsOf(dinos map[uuid.UUID]arkobject.Dino, root arkobject.DinoID) map[uuid.UUID]arkobject.Dino {
	children := d.ChildrenByAncestor(dinos)
	descendants := map[uuid.UUID]arkobject.Dino{}
	queue := append([]uuid.UUID(nil), children[root]...)
	seen := map[uuid.UUID]struct{}{}
	for len(queue) > 0 {
		childID := queue[0]
		queue = queue[1:]
		if _, ok := seen[childID]; ok {
			continue
		}
		seen[childID] = struct{}{}
		child, ok := dinos[childID]
		if !ok {
			continue
		}
		arkDinoID := arkobject.DinoID{ID1: child.ID1, ID2: child.ID2}
		if arkDinoID == root {
			continue
		}
		descendants[childID] = child
		if !arkDinoID.IsZero() {
			queue = append(queue, children[arkDinoID]...)
		}
	}
	return descendants
}

func (d *DinoAPI) PedigreeTree(dinos map[uuid.UUID]arkobject.Dino, rootID uuid.UUID) (DinoPedigreeNode, bool) {
	root, ok := dinos[rootID]
	if !ok || !root.IsTamed {
		return DinoPedigreeNode{}, false
	}
	children := d.ChildrenByAncestor(dinos)
	seen := map[uuid.UUID]struct{}{}
	tree := d.pedigreeTreeFromChildren(dinos, children, rootID, seen)
	return tree, true
}

func (d *DinoAPI) PedigreeTrees(dinos map[uuid.UUID]arkobject.Dino) []DinoPedigreeNode {
	out := make([]DinoPedigreeNode, 0, len(dinos))
	for _, id := range sortedUUIDKeys(dinos) {
		tree, ok := d.PedigreeTree(dinos, id)
		if ok {
			out = append(out, tree)
		}
	}
	return out
}

func (d *DinoAPI) pedigreeTreeFromChildren(dinos map[uuid.UUID]arkobject.Dino, children map[arkobject.DinoID][]uuid.UUID, id uuid.UUID, seen map[uuid.UUID]struct{}) DinoPedigreeNode {
	if _, ok := seen[id]; ok {
		return DinoPedigreeNode{}
	}
	seen[id] = struct{}{}
	dino := dinos[id]
	dinoID := arkobject.DinoID{ID1: dino.ID1, ID2: dino.ID2}
	node := DinoPedigreeNode{
		UUID:       id,
		DinoID:     dinoID,
		Blueprint:  dino.Blueprint,
		Name:       dino.TamedName,
		Generation: dino.Generation,
		IsFemale:   dino.IsFemale,
		IsBaby:     dino.IsBaby,
	}
	for _, childID := range children[dinoID] {
		if childID == id {
			continue
		}
		child, ok := dinos[childID]
		if !ok || !child.IsTamed {
			continue
		}
		childNode := d.pedigreeTreeFromChildren(dinos, children, childID, copySeenUUIDs(seen))
		if childNode.UUID == uuid.Nil {
			continue
		}
		node.DescendantCount += 1 + childNode.DescendantCount
		node.Children = append(node.Children, childNode)
	}
	return node
}

func copySeenUUIDs(seen map[uuid.UUID]struct{}) map[uuid.UUID]struct{} {
	out := make(map[uuid.UUID]struct{}, len(seen))
	for id := range seen {
		out[id] = struct{}{}
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

func (d *DinoAPI) MaxCurrentLevel(dinos map[uuid.UUID]arkobject.Dino) (int32, bool) {
	var max int32
	ok := false
	for _, dino := range dinos {
		if dino.Stats == nil {
			continue
		}
		if !ok || dino.Stats.CurrentLevel > max {
			max = dino.Stats.CurrentLevel
			ok = true
		}
	}
	return max, ok
}

func (d *DinoAPI) PopulationSummaryForDinos(dinos map[uuid.UUID]arkobject.Dino) DinoPopulationSummary {
	byTamed := d.CountByTamed(dinos)
	byCryopodded := d.CountByCryopodded(dinos)
	return DinoPopulationSummary{
		Dinos:      len(dinos),
		Tamed:      byTamed[true],
		Wild:       byTamed[false],
		Cryopodded: byCryopodded[true],
		Classes:    len(d.CountByClass(dinos)),
	}
}

func (d *DinoAPI) PopulationSummaryWithFaults(includeCryopodded bool) (DinoPopulationSummary, []arksave.FaultyObjectInfo, error) {
	dinos, faults, err := d.AllWithFaults()
	if err != nil {
		return DinoPopulationSummary{}, nil, err
	}
	if !includeCryopodded {
		for id, dino := range dinos {
			if dino.IsCryopodded {
				delete(dinos, id)
			}
		}
	}
	return d.PopulationSummaryForDinos(dinos), faults, nil
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

func (d *DinoAPI) CountBabiesByTamed(dinos map[uuid.UUID]arkobject.Dino) BabyCounts {
	var counts BabyCounts
	for _, dino := range dinos {
		if !dino.IsBaby {
			continue
		}
		if dino.IsTamed {
			counts.Tamed++
		} else {
			counts.Wild++
		}
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
	id, dino, stat, points, ok := bestDinoForStat(all, nil, scopes...)
	return id, dino, stat, points, ok, nil
}

func (d *DinoAPI) BestDinoForStatFiltered(opts DinoBestStatOptions) (uuid.UUID, arkobject.Dino, arkobject.DinoStat, int32, bool, error) {
	filtered, err := d.filteredForBestStat(opts)
	if err != nil {
		return uuid.Nil, arkobject.Dino{}, 0, 0, false, err
	}
	id, dino, stat, points, ok := bestDinoForStat(filtered, opts.Stats, bestStatScopes(opts)...)
	return id, dino, stat, points, ok, nil
}

func (d *DinoAPI) BestDinoForStatFilteredWithFaults(opts DinoBestStatOptions) (uuid.UUID, arkobject.Dino, arkobject.DinoStat, int32, bool, []arksave.FaultyObjectInfo, error) {
	filtered, faults, err := d.filteredForBestStatWithFaults(opts)
	if err != nil {
		return uuid.Nil, arkobject.Dino{}, 0, 0, false, nil, err
	}
	id, dino, stat, points, ok := bestDinoForStat(filtered, opts.Stats, bestStatScopes(opts)...)
	return id, dino, stat, points, ok, faults, nil
}

func (d *DinoAPI) filteredForBestStat(opts DinoBestStatOptions) (map[uuid.UUID]arkobject.Dino, error) {
	if opts.OnlyTamed && opts.OnlyUntamed {
		return nil, errors.New("cannot specify both OnlyTamed and OnlyUntamed")
	}
	if opts.BaseStat && opts.MutatedStat {
		return nil, errors.New("cannot specify both BaseStat and MutatedStat")
	}
	tamed := (*bool)(nil)
	switch {
	case opts.OnlyTamed:
		value := true
		tamed = &value
	case opts.OnlyUntamed:
		value := false
		tamed = &value
	}
	cryopodded := (*bool)(nil)
	if opts.ExcludeCryopods {
		value := false
		cryopodded = &value
	}
	filtered, err := d.Filtered(DinoFilterOptions{
		MaxLevel:   opts.LevelUpperBound,
		Blueprints: opts.Blueprints,
		Tamed:      tamed,
		Cryopodded: cryopodded,
	})
	if err != nil {
		return nil, err
	}
	return filtered, nil
}

func (d *DinoAPI) filteredForBestStatWithFaults(opts DinoBestStatOptions) (map[uuid.UUID]arkobject.Dino, []arksave.FaultyObjectInfo, error) {
	if opts.OnlyTamed && opts.OnlyUntamed {
		return nil, nil, errors.New("cannot specify both OnlyTamed and OnlyUntamed")
	}
	if opts.BaseStat && opts.MutatedStat {
		return nil, nil, errors.New("cannot specify both BaseStat and MutatedStat")
	}
	tamed := (*bool)(nil)
	switch {
	case opts.OnlyTamed:
		value := true
		tamed = &value
	case opts.OnlyUntamed:
		value := false
		tamed = &value
	}
	cryopodded := (*bool)(nil)
	if opts.ExcludeCryopods {
		value := false
		cryopodded = &value
	}
	all, faults, err := d.AllWithFaults()
	if err != nil {
		return nil, nil, err
	}
	filtered := filterDinos(all, DinoFilterOptions{
		MaxLevel:   opts.LevelUpperBound,
		Blueprints: opts.Blueprints,
		Tamed:      tamed,
		Cryopodded: cryopodded,
	})
	return filtered, faults, nil
}

func bestStatScopes(opts DinoBestStatOptions) []arkobject.StatScope {
	if opts.BaseStat {
		return []arkobject.StatScope{arkobject.StatScopeBase}
	}
	if opts.MutatedStat {
		return []arkobject.StatScope{arkobject.StatScopeMutated}
	}
	return []arkobject.StatScope{arkobject.StatScopeCombined}
}

func bestDinoForStat(dinos map[uuid.UUID]arkobject.Dino, stats []arkobject.DinoStat, scopes ...arkobject.StatScope) (uuid.UUID, arkobject.Dino, arkobject.DinoStat, int32, bool) {
	allowedStats := map[arkobject.DinoStat]struct{}{}
	for _, stat := range stats {
		allowedStats[stat] = struct{}{}
	}
	var bestID uuid.UUID
	var bestDino arkobject.Dino
	var bestStat arkobject.DinoStat
	var bestPoints int32
	found := false
	for id, dino := range dinos {
		if dino.Stats == nil {
			continue
		}
		if len(allowedStats) > 0 {
			for _, stat := range stats {
				points := dino.Stats.Points(stat, scopes...)
				if !found || points > bestPoints || (points == bestPoints && id.String() < bestID.String()) {
					bestID = id
					bestDino = dino
					bestStat = stat
					bestPoints = points
					found = true
				}
			}
		} else {
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
	}
	return bestID, bestDino, bestStat, bestPoints, found
}

func (d *DinoAPI) MostMutatedTamed() (uuid.UUID, arkobject.Dino, int32, bool, error) {
	all, _, err := d.AllWithFaults()
	if err != nil {
		return uuid.Nil, arkobject.Dino{}, 0, false, err
	}
	var bestID uuid.UUID
	var bestDino arkobject.Dino
	var bestTotal int32
	found := false
	for id, dino := range all {
		if !dino.IsTamed || dino.Stats == nil {
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
	all, err := d.All()
	if err != nil {
		return nil, err
	}
	return filterDinos(all, opts), nil
}

func filterDinos(dinos map[uuid.UUID]arkobject.Dino, opts DinoFilterOptions) map[uuid.UUID]arkobject.Dino {
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
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range dinos {
		if len(allowedBlueprints) > 0 {
			if _, ok := allowedBlueprints[dino.Blueprint]; !ok {
				continue
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
				continue
			}
		}
		if opts.Tamed != nil && dino.IsTamed != *opts.Tamed {
			continue
		}
		if opts.Cryopodded != nil && dino.IsCryopodded != *opts.Cryopodded {
			continue
		}
		if opts.MinLevel != nil || opts.MaxLevel != nil || opts.StatMinimum != 0 {
			if dino.Stats == nil {
				continue
			}
		}
		if opts.MinLevel != nil && dino.Stats.CurrentLevel < *opts.MinLevel {
			continue
		}
		if opts.MaxLevel != nil && dino.Stats.CurrentLevel > *opts.MaxLevel {
			continue
		}
		if opts.StatMinimum != 0 {
			statsAbove := dino.Stats.StatsAtLeast(opts.StatMinimum, arkobject.StatScopeCombined)
			if len(statsAbove) == 0 {
				continue
			}
			if len(allowedStats) > 0 {
				matched := false
				for _, stat := range statsAbove {
					if _, ok := allowedStats[stat]; ok {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}
			}
		}
		out[id] = dino
	}
	return out
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
