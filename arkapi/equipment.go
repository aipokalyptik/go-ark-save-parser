package arkapi

import (
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type EquipmentAPI struct {
	save *arksave.Save
}

func NewEquipment(save *arksave.Save) *EquipmentAPI {
	return &EquipmentAPI{save: save}
}

func (e *EquipmentAPI) IsApplicableBlueprint(blueprint string) bool {
	return e.KindForBlueprint(blueprint) != arkobject.EquipmentUnknown
}

func (e *EquipmentAPI) KindForBlueprint(blueprint string) arkobject.EquipmentKind {
	switch {
	case strings.Contains(blueprint, "PrimalItemAmmo") || strings.Contains(blueprint, "PrimalItem_WeaponEmptyCryopod"):
		return arkobject.EquipmentUnknown
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

func (e *EquipmentAPI) All() (map[uuid.UUID]arkobject.EquipmentItem, error) {
	objects, err := e.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return e.IsApplicableBlueprint(info.ClassName)
	})
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
