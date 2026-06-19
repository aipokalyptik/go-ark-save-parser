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
		if !ok || ref.Type != arkproperty.ObjectReferenceUUID {
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
	return len(i.ItemUUIDs)
}

type InventoryItem struct {
	UUID           uuid.UUID
	Blueprint      string
	Object         *GameObject
	Quantity       int32
	OwnerInventory *uuid.UUID
	Crafter        *ObjectCrafter
}

type ObjectCrafter struct {
	CharacterName string
	TribeName     string
}

func (c ObjectCrafter) Valid() bool {
	return c.CharacterName != "" || c.TribeName != ""
}

func InventoryItemFromObject(object *GameObject) InventoryItem {
	item := InventoryItem{Object: object, Quantity: 1}
	if object == nil {
		return item
	}
	item.UUID = object.UUID
	item.Blueprint = object.Blueprint
	if value, ok := object.Value("ItemQuantity"); ok {
		if quantity, ok := value.(int32); ok {
			item.Quantity = quantity
		}
	}
	if value, ok := object.Value("OwnerInventory"); ok {
		if ref, ok := value.(arkproperty.ObjectReference); ok {
			if id, ok := uuidFromReferenceValue(ref.Value); ok {
				item.OwnerInventory = &id
			}
		}
	}
	crafter := ObjectCrafter{}
	if value, ok := object.Value("CrafterCharacterName"); ok {
		if name, ok := value.(string); ok {
			crafter.CharacterName = name
		}
	}
	if value, ok := object.Value("CrafterTribeName"); ok {
		if name, ok := value.(string); ok {
			crafter.TribeName = name
		}
	}
	if crafter.Valid() {
		item.Crafter = &crafter
	}
	return item
}

func uuidFromReferenceValue(value any) (uuid.UUID, bool) {
	switch v := value.(type) {
	case string:
		id, err := uuid.Parse(v)
		return id, err == nil
	case uuid.UUID:
		return v, true
	default:
		return uuid.Nil, false
	}
}
