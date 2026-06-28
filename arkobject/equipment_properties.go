package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

func customItemDataCount(properties arkproperty.Container) int {
	value, ok := properties.Value("CustomItemDatas")
	if !ok {
		return 0
	}
	switch customData := value.(type) {
	case arkproperty.Array:
		return len(customData.Values)
	case []any:
		return len(customData)
	default:
		return 0
	}
}

func uint16PositionedValue(properties arkproperty.Container, name string, position int32) (uint16, bool) {
	value, ok := properties.PositionedValue(name, position)
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case uint16:
		return v, true
	case uint32:
		return uint16(v), true
	case int32:
		return uint16(v), true
	case int:
		return uint16(v), true
	default:
		return 0, false
	}
}

func numericInt32(properties arkproperty.Container, name string) (int32, bool) {
	value, ok := properties.Value(name)
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case byte:
		return int32(v), true
	case int32:
		return v, true
	case uint32:
		return int32(v), true
	case int:
		return int32(v), true
	default:
		return 0, false
	}
}

func numericFloat64(properties arkproperty.Container, name string) (float64, bool) {
	value, ok := properties.Value(name)
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int32:
		return float64(v), true
	case uint32:
		return float64(v), true
	case byte:
		return float64(v), true
	default:
		return 0, false
	}
}
