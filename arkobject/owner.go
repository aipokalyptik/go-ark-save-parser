package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

type ObjectOwner struct {
	OriginalPlacerID int32
	TribeName        string
	PlayerName       string
	PlayerID         int32
	TribeID          int32
}

func ObjectOwnerFromContainer(properties arkproperty.Container) ObjectOwner {
	return ObjectOwner{
		OriginalPlacerID: int32Value(properties, "OriginalPlacerPlayerID"),
		TribeName:        stringValue(properties, "OwnerName"),
		PlayerName:       stringValue(properties, "OwningPlayerName"),
		PlayerID:         int32Value(properties, "OwningPlayerID"),
		TribeID:          int32Value(properties, "TargetingTeam"),
	}
}

func ObjectOwnerFromObject(object *GameObject) ObjectOwner {
	if object == nil {
		return ObjectOwner{}
	}
	return ObjectOwnerFromContainer(arkproperty.Container{Properties: object.Properties})
}

func (o ObjectOwner) Equal(other ObjectOwner) bool {
	if o.TribeID != 0 && other.TribeID != 0 && o.TribeID != other.TribeID {
		return false
	}
	if o.PlayerID != 0 && other.PlayerID != 0 && o.PlayerID != other.PlayerID {
		return false
	}
	return true
}

func int32Value(properties arkproperty.Container, name string) int32 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case int32:
		return v
	case uint32:
		return int32(v)
	default:
		return 0
	}
}

func stringValue(properties arkproperty.Container, name string) string {
	value, ok := properties.Value(name)
	if !ok {
		return ""
	}
	text, _ := value.(string)
	return text
}
