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
		out[info.UUID] = arkobject.DinoFromObject(info.Object, location)
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
