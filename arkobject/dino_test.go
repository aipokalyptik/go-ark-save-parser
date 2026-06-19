package arkobject

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

func TestDinoFromObjectReadsCoreFields(t *testing.T) {
	statusID := "00112233-4455-6677-8899-aabbccddeeff"
	object := &GameObject{
		UUID:      uuid.MustParse("11112222-3333-4444-5555-666677778888"),
		Blueprint: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		Properties: []arkproperty.Property{
			{Name: "DinoID1", Type: arkproperty.TypeUInt32, Value: uint32(1001)},
			{Name: "DinoID2", Type: arkproperty.TypeUInt32, Value: uint32(2002)},
			{Name: "bIsFemale", Type: arkproperty.TypeBool, Value: true},
			{Name: "bIsDead", Type: arkproperty.TypeBool, Value: false},
			{Name: "bIsBaby", Type: arkproperty.TypeBool, Value: true},
			{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)},
			{Name: "MyCharacterStatusComponent", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{
				Type:  arkproperty.ObjectReferenceUUID,
				Value: statusID,
			}},
			{Name: "GeneTraits", Type: arkproperty.TypeArray, Value: arkproperty.Array{
				ElementType: arkproperty.TypeString,
				Values:      []any{"MutableMelee[2]", "Robust"},
			}},
			{Name: "SavedBaseWorldLocation", Type: arkproperty.TypeStruct, Value: ActorTransform{X: 1, Y: 2, Z: 3}},
		},
	}

	dino := DinoFromObject(object, nil)

	if dino.UUID != object.UUID || dino.Blueprint != object.Blueprint {
		t.Fatalf("Dino identity = %#v", dino)
	}
	if dino.ID1 != 1001 || dino.ID2 != 2002 || !dino.IsFemale || dino.IsDead || !dino.IsBaby || !dino.IsTamed {
		t.Fatalf("Dino flags/id = %#v", dino)
	}
	if dino.StatusComponentUUID == nil || dino.StatusComponentUUID.String() != statusID {
		t.Fatalf("StatusComponentUUID = %v, want %s", dino.StatusComponentUUID, statusID)
	}
	if len(dino.GeneTraits) != 2 || dino.GeneTraits[0] != "MutableMelee[2]" {
		t.Fatalf("GeneTraits = %#v", dino.GeneTraits)
	}
	if dino.Location == nil || dino.Location.X != 1 || dino.Location.Z != 3 {
		t.Fatalf("Location = %#v", dino.Location)
	}
}

func TestDinoFromObjectPrefersSaveContextLocation(t *testing.T) {
	object := &GameObject{UUID: uuid.MustParse("11112222-3333-4444-5555-666677778888")}
	location := ActorTransform{X: 10, Y: 20, Z: 30}

	dino := DinoFromObject(object, &location)

	if dino.Location == nil || dino.Location.X != 10 || dino.Location.Y != 20 {
		t.Fatalf("Location = %#v, want save context location", dino.Location)
	}
}
