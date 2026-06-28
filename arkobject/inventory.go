package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

type Inventory struct {
	UUID      uuid.UUID
	Object    *GameObject
	ItemUUIDs []uuid.UUID
}

func InventoryFromObject(object *GameObject) Inventory {
	inventory := Inventory{Object: object}
	if object == nil {
		return inventory
	}
	inventory.UUID = object.UUID
	raw, ok := object.Value("InventoryItems")
	if !ok {
		return inventory
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return inventory
	}
	for _, value := range array.Values {
		ref, ok := value.(arkproperty.ObjectReference)
		if !ok {
			continue
		}
		id, ok := uuidFromReferenceValue(ref.Value)
		if ok {
			inventory.ItemUUIDs = append(inventory.ItemUUIDs, id)
		}
	}
	return inventory
}

func (i Inventory) NumberOfItems() int {
	seen := make(map[uuid.UUID]struct{}, len(i.ItemUUIDs))
	for _, id := range i.ItemUUIDs {
		seen[id] = struct{}{}
	}
	return len(seen)
}
