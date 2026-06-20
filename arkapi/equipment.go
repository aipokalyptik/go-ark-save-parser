package arkapi

import (
	"fmt"
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

func (e *EquipmentAPI) AllWithFaults() (map[uuid.UUID]arkobject.EquipmentItem, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := e.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return e.IsApplicableBlueprint(info.ClassName)
	})
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

func (e *EquipmentAPI) Filtered(opts EquipmentFilterOptions) (map[uuid.UUID]arkobject.EquipmentItem, error) {
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
	return e.filter(func(item arkobject.EquipmentItem) bool {
		if len(allowedKinds) > 0 {
			if _, ok := allowedKinds[item.Kind]; !ok {
				return false
			}
		}
		if len(allowedBlueprints) > 0 {
			if _, ok := allowedBlueprints[item.Blueprint]; !ok {
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
