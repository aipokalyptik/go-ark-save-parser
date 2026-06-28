package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

type InventoryItem struct {
	UUID           uuid.UUID
	Blueprint      string
	Object         *GameObject
	Quantity       int32
	OwnerInventory *uuid.UUID
	Crafter        *ObjectCrafter
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
