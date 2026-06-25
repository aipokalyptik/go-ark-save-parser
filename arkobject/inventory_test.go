package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestInventoryItemFromObjectReadsQuantityAndOwner(t *testing.T) {
	ownerID := "00112233-4455-6677-8899-aabbccddeeff"
	item := InventoryItemFromObject(&GameObject{
		UUID:      uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
		Properties: []arkproperty.Property{
			{Name: "ItemQuantity", Type: arkproperty.TypeInt, Value: int32(37)},
			{Name: "OwnerInventory", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{
				Type:  arkproperty.ObjectReferenceUUID,
				Value: ownerID,
			}},
		},
	})

	if item.UUID.String() != "11112222-3333-4444-5555-666677778888" {
		t.Fatalf("InventoryItem UUID = %s", item.UUID)
	}
	if item.Quantity != 37 {
		t.Fatalf("InventoryItem.Quantity = %d, want 37", item.Quantity)
	}
	if item.OwnerInventory == nil || item.OwnerInventory.String() != ownerID {
		t.Fatalf("InventoryItem.OwnerInventory = %v, want %s", item.OwnerInventory, ownerID)
	}
}

func TestInventoryItemFromObjectReadsCrafterMetadata(t *testing.T) {
	item := InventoryItemFromObject(&GameObject{
		UUID:      uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Blueprint: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
		Properties: []arkproperty.Property{
			{Name: "CrafterCharacterName", Type: arkproperty.TypeString, Value: "Survivor"},
			{Name: "CrafterTribeName", Type: arkproperty.TypeString, Value: "Porters"},
		},
	})

	if item.Crafter == nil {
		t.Fatalf("InventoryItem.Crafter = nil, want crafter metadata")
	}
	if item.Crafter.CharacterName != "Survivor" || item.Crafter.TribeName != "Porters" {
		t.Fatalf("InventoryItem.Crafter = %#v, want Survivor/Porters", item.Crafter)
	}
	if !item.Crafter.Valid() {
		t.Fatalf("InventoryItem.Crafter.Valid() = false, want true")
	}
}

func TestInventoryFromObjectReadsReferencedItems(t *testing.T) {
	first := "00112233-4455-6677-8899-aabbccddeeff"
	second := "11112222-3333-4444-5555-666677778888"
	inventory := InventoryFromObject(&GameObject{
		UUID: uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"),
		Properties: []arkproperty.Property{
			{Name: "InventoryItems", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeObject,
				Values: []any{
					arkproperty.ObjectReference{Type: arkproperty.ObjectReferenceUUID, Value: first},
					arkproperty.ObjectReference{Type: arkproperty.ObjectReferencePath, Value: second},
				},
			}},
		},
	})

	if inventory.NumberOfItems() != 2 {
		t.Fatalf("NumberOfItems() = %d, want 2", inventory.NumberOfItems())
	}
	if inventory.ItemUUIDs[0].String() != first || inventory.ItemUUIDs[1].String() != second {
		t.Fatalf("Inventory.ItemUUIDs = %#v", inventory.ItemUUIDs)
	}
}

func TestInventoryNumberOfItemsCollapsesDuplicateReferences(t *testing.T) {
	first := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	second := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	inventory := Inventory{ItemUUIDs: []uuid.UUID{first, first, second}}

	if inventory.NumberOfItems() != 2 {
		t.Fatalf("NumberOfItems() = %d, want unique item count 2", inventory.NumberOfItems())
	}
}
