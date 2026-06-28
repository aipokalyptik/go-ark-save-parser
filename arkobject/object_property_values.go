package arkobject

import (
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

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

func arrayLength(properties arkproperty.Container, name string) int {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	array, ok := value.(arkproperty.Array)
	if !ok {
		return 0
	}
	return len(array.Values)
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

func uploadedFromServerName(properties arkproperty.Container) string {
	return strings.TrimPrefix(stringValue(properties, "UploadedFromServerName"), "\n")
}
