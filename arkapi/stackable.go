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
	ids, err := s.save.ObjectIDs()
	if err != nil {
		return nil, err
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for _, id := range ids {
		object, err := s.save.Object(id)
		if err != nil {
			return nil, err
		}
		if !s.IsApplicableBlueprint(object.Blueprint) {
			continue
		}
		if boolProperty(object, "bIsBlueprint") || boolProperty(object, "bIsEngram") {
			continue
		}
		out[id] = arkobject.InventoryItemFromObject(object)
	}
	return out, nil
}

func (s *StackableAPI) ByClass(blueprints []string) (map[uuid.UUID]arkobject.InventoryItem, error) {
	all, err := s.All()
	if err != nil {
		return nil, err
	}
	allowed := map[string]struct{}{}
	for _, blueprint := range blueprints {
		allowed[blueprint] = struct{}{}
	}
	out := map[uuid.UUID]arkobject.InventoryItem{}
	for id, item := range all {
		if _, ok := allowed[item.Blueprint]; ok {
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
