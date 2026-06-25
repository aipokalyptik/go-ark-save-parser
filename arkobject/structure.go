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

func (s Structure) IsOwnedBy(owner ObjectOwner) bool {
	if s.Owner.PlayerID != 0 && s.Owner.PlayerID == owner.PlayerID {
		return true
	}
	if s.Owner.PlayerName != "" && s.Owner.PlayerName == owner.PlayerName {
		return true
	}
	if s.Owner.TribeName != "" && s.Owner.TribeName == owner.TribeName {
		return true
	}
	if s.Owner.TribeID != 0 && s.Owner.TribeID == owner.TribeID {
		return true
	}
	if s.Owner.OriginalPlacerID != 0 && s.Owner.OriginalPlacerID == owner.OriginalPlacerID {
		return true
	}
	return false
}

func objectReferenceUUIDArray(properties arkproperty.Container, name string) []uuid.UUID {
	raw, ok := properties.Value(name)
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]uuid.UUID, 0, len(array.Values))
	for _, value := range array.Values {
		ref, ok := value.(arkproperty.ObjectReference)
		if !ok || ref.Type != arkproperty.ObjectReferenceUUID {
			continue
		}
		id, ok := uuidFromReferenceValue(ref.Value)
		if ok {
			out = append(out, id)
		}
	}
	return out
}

func float64Value(properties arkproperty.Container, name string) float64 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int32:
		return float64(v)
	case uint32:
		return float64(v)
	default:
		return 0
	}
}

func boolValue(properties arkproperty.Container, name string) bool {
	value, ok := properties.Value(name)
	if !ok {
		return false
	}
	out, _ := value.(bool)
	return out
}
