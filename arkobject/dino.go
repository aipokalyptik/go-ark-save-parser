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
	UUID                          uuid.UUID
	Blueprint                     string
	Object                        *GameObject
	ID1                           uint32
	ID2                           uint32
	IsFemale                      bool
	IsTamed                       bool
	IsBaby                        bool
	IsDead                        bool
	IsCryopodded                  bool
	TamedTimeStamp                float64
	LastInAllyRangeTimeSerialized float64
	Generation                    int
	AncestorIDs                   []DinoID
	MaturationPercent             float64
	BabyStage                     BabyStage
	StatusComponentUUID           *uuid.UUID
	InventoryUUID                 *uuid.UUID
	TamedName                     string
	IsNeutered                    bool
	ColorSetIndices               [6]int
	ColorSetNames                 [6]string
	UploadedFromServerName        string
	Stats                         *DinoStats
	Owner                         DinoOwner
	GeneTraits                    []string
	ParsedGeneTraits              []GeneTrait
	Location                      *ActorTransform
}

type GeneTrait struct {
	Raw   string
	Name  string
	Level int
}

type DinoID struct {
	ID1 uint32
	ID2 uint32
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
	dino.TamedTimeStamp = float64Value(properties, "TamedTimeStamp")
	dino.LastInAllyRangeTimeSerialized = float64Value(properties, "LastInAllyRangeTimeSerialized")
	if dino.IsTamed {
		dino.Generation = max(arrayLength(properties, "DinoAncestors"), arrayLength(properties, "DinoAncestorsMale")) + 1
	}
	dino.AncestorIDs = ancestorIDs(properties)
	dino.StatusComponentUUID = objectReferenceUUID(properties, "MyCharacterStatusComponent")
	dino.InventoryUUID = objectReferenceUUID(properties, "MyInventoryComponent")
	dino.TamedName = stringValue(properties, "TamedName")
	dino.IsNeutered = boolValue(properties, "bNeutered")
	dino.ColorSetIndices = colorSetIndices(properties)
	dino.ColorSetNames = colorSetNames(properties)
	dino.UploadedFromServerName = uploadedFromServerName(properties)
	dino.Owner = DinoOwnerFromContainer(properties)
	dino.GeneTraits = stringArrayValue(properties, "GeneTraits")
	dino.ParsedGeneTraits = parseGeneTraits(dino.GeneTraits)
	if dino.Location == nil {
		dino.Location = actorTransformValue(properties, "SavedBaseWorldLocation")
	}
	return dino
}

func DinoFromObjectWithStatus(object *GameObject, statusObject *GameObject, location *ActorTransform) Dino {
	dino := DinoFromObject(object, location)
	if statusObject != nil {
		stats := DinoStatsFromObject(statusObject)
		dino.Stats = &stats
	}
	return dino
}

func (d Dino) ShortName() string {
	if d.Object != nil {
		return d.Object.ShortName()
	}
	return ShortNameFromBlueprint(d.Blueprint)
}

func (d Dino) IsWildTamed() bool {
	if !d.IsTamed {
		return false
	}
	if d.Object != nil {
		_, hasFemaleAncestors := (arkproperty.Container{Properties: d.Object.Properties}).Value("DinoAncestors")
		return !hasFemaleAncestors
	}
	return len(d.AncestorIDs) == 0
}
