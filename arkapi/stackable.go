package arkapi

import (
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

type StackableAPI struct {
	save *arksave.Save
}

func NewStackable(save *arksave.Save) *StackableAPI {
	return &StackableAPI{save: save}
}

func (s *StackableAPI) IsApplicableBlueprint(blueprint string) bool {
	return strings.Contains(blueprint, "Resources/PrimalItemResource") ||
		strings.Contains(blueprint, "/PrimalItemConsumable") ||
		strings.Contains(blueprint, "PrimalItemAmmo")
}

func (s *StackableAPI) Count(items map[uuid.UUID]arkobject.InventoryItem) int32 {
	var count int32
	for _, item := range items {
		count += item.Quantity
	}
	return count
}

func (s *StackableAPI) All() (map[uuid.UUID]arkobject.InventoryItem, error) {
	objects, err := s.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return s.IsApplicableBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for _, info := range objects {
		if boolProperty(info.Object, "bIsBlueprint") || boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.InventoryItemFromObject(info.Object)
	}
	return out, nil
}

func (s *StackableAPI) AllWithFaults() (map[uuid.UUID]arkobject.InventoryItem, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := s.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return s.IsApplicableBlueprint(info.ClassName)
	})
	if err != nil {
		return nil, nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for _, info := range objects {
		if boolProperty(info.Object, "bIsBlueprint") || boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.InventoryItemFromObject(info.Object)
	}
	return out, faults, nil
}

func (s *StackableAPI) ByClass(blueprints []string) (map[uuid.UUID]arkobject.InventoryItem, error) {
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	objects, err := s.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		_, ok := allowed[info.ClassName]
		return ok
	})
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for _, info := range objects {
		if boolProperty(info.Object, "bIsBlueprint") || boolProperty(info.Object, "bIsEngram") {
			continue
		}
		out[info.UUID] = arkobject.InventoryItemFromObject(info.Object)
	}
	return out, nil
}

func (s *StackableAPI) Resources() (map[uuid.UUID]arkobject.InventoryItem, error) {
	return s.byBlueprintSubstring("Resources/PrimalItemResource")
}

func (s *StackableAPI) Ammo() (map[uuid.UUID]arkobject.InventoryItem, error) {
	return s.byBlueprintSubstring("PrimalItemAmmo")
}

func (s *StackableAPI) Consumables() (map[uuid.UUID]arkobject.InventoryItem, error) {
	return s.byBlueprintSubstring("/PrimalItemConsumable")
}

func (s *StackableAPI) OwnedBy(owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.InventoryItem, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	return s.FilterOwnedBy(all, owner)
}

func (s *StackableAPI) FilterOwnedBy(items map[uuid.UUID]arkobject.InventoryItem, owner arkobject.ObjectOwner) (map[uuid.UUID]arkobject.InventoryItem, error) {
	structures := NewStructure(s.save)
	containers := map[uuid.UUID]*arkobject.Structure{}
	out := map[uuid.UUID]arkobject.InventoryItem{}
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

func (s *StackableAPI) byBlueprintSubstring(substring string) (map[uuid.UUID]arkobject.InventoryItem, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for id, item := range all {
		if strings.Contains(item.Blueprint, substring) {
			out[id] = item
		}
	}
	return out, nil
}

func boolProperty(object *arkobject.GameObject, name string) bool {
	value, ok := object.Value(name)
	if !ok {
		return false
	}
	out, _ := value.(bool)
	return out
}
