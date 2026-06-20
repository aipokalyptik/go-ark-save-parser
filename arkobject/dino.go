package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

type BabyStage string

const (
	BabyStageUnknown    BabyStage = ""
	BabyStageBaby       BabyStage = "Baby"
	BabyStageJuvenile   BabyStage = "Juvenile"
	BabyStageAdolescent BabyStage = "Adolescent"
)

type Dino struct {
	UUID                uuid.UUID
	Blueprint           string
	Object              *GameObject
	ID1                 uint32
	ID2                 uint32
	IsFemale            bool
	IsTamed             bool
	IsBaby              bool
	IsDead              bool
	IsCryopodded        bool
	MaturationPercent   float64
	BabyStage           BabyStage
	StatusComponentUUID *uuid.UUID
	InventoryUUID       *uuid.UUID
	TamedName           string
	IsNeutered          bool
	Owner               DinoOwner
	GeneTraits          []string
	Location            *ActorTransform
}

func DinoFromObject(object *GameObject, location *ActorTransform) Dino {
	dino := Dino{Object: object, Location: location}
	if object == nil {
		return dino
	}
	properties := arkproperty.Container{Properties: object.Properties}
	dino.UUID = object.UUID
	dino.Blueprint = object.Blueprint
	dino.ID1 = uint32Value(properties, "DinoID1")
	dino.ID2 = uint32Value(properties, "DinoID2")
	dino.IsFemale = boolValue(properties, "bIsFemale")
	dino.IsDead = boolValue(properties, "bIsDead")
	dino.IsBaby = boolValue(properties, "bIsBaby")
	if dino.IsBaby {
		dino.MaturationPercent = float64Value(properties, "BabyAge") * 100
		dino.BabyStage = babyStageForPercent(dino.MaturationPercent)
	}
	_, dino.IsTamed = properties.Value("TamedTimeStamp")
	dino.StatusComponentUUID = objectReferenceUUID(properties, "MyCharacterStatusComponent")
	dino.InventoryUUID = objectReferenceUUID(properties, "MyInventoryComponent")
	dino.TamedName = stringValue(properties, "TamedName")
	dino.IsNeutered = boolValue(properties, "bNeutered")
	dino.Owner = DinoOwnerFromContainer(properties)
	dino.GeneTraits = stringArrayValue(properties, "GeneTraits")
	if dino.Location == nil {
		dino.Location = actorTransformValue(properties, "SavedBaseWorldLocation")
	}
	return dino
}

func babyStageForPercent(percent float64) BabyStage {
	if percent < 10 {
		return BabyStageBaby
	}
	if percent < 50 {
		return BabyStageJuvenile
	}
	return BabyStageAdolescent
}

func uint32Value(properties arkproperty.Container, name string) uint32 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case uint32:
		return v
	case int32:
		return uint32(v)
	case int:
		return uint32(v)
	default:
		return 0
	}
}

func objectReferenceUUID(properties arkproperty.Container, name string) *uuid.UUID {
	value, ok := properties.Value(name)
	if !ok {
		return nil
	}
	ref, ok := value.(arkproperty.ObjectReference)
	if !ok || ref.Type != arkproperty.ObjectReferenceUUID {
		return nil
	}
	id, ok := uuidFromReferenceValue(ref.Value)
	if !ok {
		return nil
	}
	return &id
}

func actorTransformValue(properties arkproperty.Container, name string) *ActorTransform {
	value, ok := properties.Value(name)
	if !ok {
		return nil
	}
	switch v := value.(type) {
	case ActorTransform:
		return &v
	case *ActorTransform:
		return v
	default:
		return nil
	}
}
