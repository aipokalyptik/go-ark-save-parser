package arkobject

import "github.com/google/uuid"

type StackableItem struct {
	UUID           uuid.UUID
	Blueprint      string
	Object         *GameObject
	Quantity       int32
	OwnerInventory *uuid.UUID
	Crafter        *ObjectCrafter
}

func StackableItemFromInventoryItem(item InventoryItem) StackableItem {
	return StackableItem{
		UUID:           item.UUID,
		Blueprint:      item.Blueprint,
		Object:         item.Object,
		Quantity:       item.Quantity,
		OwnerInventory: item.OwnerInventory,
		Crafter:        item.Crafter,
	}
}

func StackableItemFromObject(object *GameObject) StackableItem {
	return StackableItemFromInventoryItem(InventoryItemFromObject(object))
}
