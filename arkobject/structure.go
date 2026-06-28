package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

type Structure struct {
	UUID                          uuid.UUID
	Blueprint                     string
	Object                        *GameObject
	Owner                         ObjectOwner
	ID                            int32
	MaxHealth                     float64
	CurrentHealth                 float64
	Location                      *ActorTransform
	InventoryUUID                 *uuid.UUID
	ItemCount                     int32
	MaxItemCount                  int32
	LinkedStructureUUIDs          []uuid.UUID
	OriginalCreationTime          float64
	LastEnterStasisTime           float64
	HasResetDecayTime             bool
	SavedWhenStasised             bool
	WasPlacementSnapped           bool
	LastInAllyRangeTimeSerialized float64
}

func StructureFromObject(object *GameObject, location *ActorTransform) Structure {
	structure := Structure{Object: object, Location: location}
	if object == nil {
		return structure
	}
	properties := arkproperty.Container{Properties: object.Properties}
	structure.UUID = object.UUID
	structure.Blueprint = object.Blueprint
	structure.Owner = ObjectOwnerFromContainer(properties)
	structure.ID = int32Value(properties, "StructureID")
	structure.MaxHealth = float64Value(properties, "MaxHealth")
	structure.CurrentHealth = float64Value(properties, "Health")
	if structure.CurrentHealth == 0 {
		structure.CurrentHealth = structure.MaxHealth
	}
	structure.InventoryUUID = objectReferenceUUID(properties, "MyInventoryComponent")
	structure.ItemCount = int32Value(properties, "CurrentItemCount")
	structure.MaxItemCount = int32Value(properties, "MaxItemCount")
	structure.LinkedStructureUUIDs = objectReferenceUUIDArray(properties, "LinkedStructures")
	structure.OriginalCreationTime = float64Value(properties, "OriginalCreationTime")
	structure.LastEnterStasisTime = float64Value(properties, "LastEnterStasisTime")
	structure.HasResetDecayTime = boolValue(properties, "bHasResetDecayTime")
	structure.SavedWhenStasised = boolValue(properties, "bSavedWhenStasised")
	structure.WasPlacementSnapped = boolValue(properties, "bWasPlacementSnapped")
	structure.LastInAllyRangeTimeSerialized = float64Value(properties, "LastInAllyRangeTimeSerialized")
	return structure
}

func (s Structure) OpenSlots() int32 {
	return s.MaxItemCount - s.ItemCount
}

func (s Structure) IsEmpty() bool {
	return s.ItemCount == 0
}
