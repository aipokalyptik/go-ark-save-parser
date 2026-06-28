package arkobject

import "github.com/google/uuid"

func uuidFromReferenceValue(value any) (uuid.UUID, bool) {
	switch v := value.(type) {
	case string:
		id, err := uuid.Parse(v)
		return id, err == nil
	case uuid.UUID:
		return v, true
	default:
		return uuid.Nil, false
	}
}
