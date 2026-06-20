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

func NewDino(save *arksave.Save) *DinoAPI {
	return &DinoAPI{save: save}
}

func (d *DinoAPI) IsApplicableBlueprint(blueprint string) bool {
	if blueprint == "" {
		return false
	}
	hasDinoPath := strings.Contains(blueprint, "/Creatures/") ||
		strings.Contains(blueprint, "/Dinos/") ||
		strings.Contains(blueprint, "/SDinoVariants/")
	return hasDinoPath && strings.Contains(blueprint, "_Character_")
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

func (d *DinoAPI) Babies() (map[uuid.UUID]arkobject.Dino, error) {
	all, err := d.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.Dino{}
	for id, dino := range all {
		if dino.IsBaby {
			out[id] = dino
		}
	}
	return out, nil
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

func (d *DinoAPI) CountByTamed(dinos map[uuid.UUID]arkobject.Dino) map[bool]int {
	counts := map[bool]int{}
	for _, dino := range dinos {
		counts[dino.IsTamed]++
	}
	return counts
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
