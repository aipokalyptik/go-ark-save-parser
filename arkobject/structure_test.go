package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestStructureFromObjectReadsCoreFields(t *testing.T) {
	linkedID := "00112233-4455-6677-8899-aabbccddeeff"
	inventoryID := "99999999-aaaa-bbbb-cccc-ddddeeeeffff"
	object := &GameObject{
		UUID:      uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Blueprint: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
		Properties: []arkproperty.Property{
			{Name: "StructureID", Type: arkproperty.TypeInt, Value: int32(123)},
			{Name: "MaxHealth", Type: arkproperty.TypeFloat, Value: float32(10000)},
			{Name: "Health", Type: arkproperty.TypeFloat, Value: float32(7500)},
			{Name: "TargetingTeam", Type: arkproperty.TypeInt, Value: int32(555)},
			{Name: "OwnerName", Type: arkproperty.TypeString, Value: "Builders"},
			{Name: "LinkedStructures", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeObject,
				Values: []any{
					arkproperty.ObjectReference{Type: arkproperty.ObjectReferenceUUID, Value: linkedID},
				},
			}},
			{Name: "MyInventoryComponent", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{
				Type:  arkproperty.ObjectReferenceUUID,
				Value: inventoryID,
			}},
			{Name: "CurrentItemCount", Type: arkproperty.TypeInt, Value: int32(12)},
			{Name: "MaxItemCount", Type: arkproperty.TypeInt, Value: int32(300)},
			{Name: "bSavedWhenStasised", Type: arkproperty.TypeBool, Value: true},
		},
	}
	location := ActorTransform{X: 10, Y: 20, Z: 30}

	structure := StructureFromObject(object, &location)

	if structure.UUID != object.UUID || structure.Blueprint != object.Blueprint {
		t.Fatalf("Structure identity = %#v", structure)
	}
	if structure.ID != 123 || structure.MaxHealth != 10000 || structure.CurrentHealth != 7500 {
		t.Fatalf("Structure health/id = %#v", structure)
	}
	if structure.Owner.TribeID != 555 || structure.Owner.TribeName != "Builders" {
		t.Fatalf("Structure owner = %#v", structure.Owner)
	}
	if structure.Location == nil || structure.Location.X != 10 {
		t.Fatalf("Structure location = %#v", structure.Location)
	}
	if len(structure.LinkedStructureUUIDs) != 1 || structure.LinkedStructureUUIDs[0].String() != linkedID {
		t.Fatalf("LinkedStructureUUIDs = %#v", structure.LinkedStructureUUIDs)
	}
	if structure.InventoryUUID == nil || structure.InventoryUUID.String() != inventoryID {
		t.Fatalf("InventoryUUID = %v, want %s", structure.InventoryUUID, inventoryID)
	}
	if structure.ItemCount != 12 || structure.MaxItemCount != 300 || structure.OpenSlots() != 288 || structure.IsEmpty() {
		t.Fatalf("inventory counts = current %d max %d open %d empty %v", structure.ItemCount, structure.MaxItemCount, structure.OpenSlots(), structure.IsEmpty())
	}
	if !structure.SavedWhenStasised {
		t.Fatalf("SavedWhenStasised = false, want true")
	}
}

func TestStructureIsOwnedByMatchesAnyKnownOwnerField(t *testing.T) {
	structure := Structure{Owner: ObjectOwner{TribeID: 555, TribeName: "Builders"}}
	if !structure.IsOwnedBy(ObjectOwner{TribeID: 555}) {
		t.Fatalf("IsOwnedBy tribe id = false, want true")
	}
	if !structure.IsOwnedBy(ObjectOwner{TribeName: "Builders"}) {
		t.Fatalf("IsOwnedBy tribe name = false, want true")
	}
	if structure.IsOwnedBy(ObjectOwner{TribeID: 777, TribeName: "Other"}) {
		t.Fatalf("IsOwnedBy unrelated owner = true, want false")
	}
}
