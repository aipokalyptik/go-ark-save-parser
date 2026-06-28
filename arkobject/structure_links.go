package arkobject

import (
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/google/uuid"
)

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
