package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

func containerValue(properties arkproperty.Container, name string) (arkproperty.Container, bool) {
	value, ok := properties.Value(name)
	if !ok {
		return arkproperty.Container{}, false
	}
	container, ok := value.(arkproperty.Container)
	return container, ok
}

func uint64Value(properties arkproperty.Container, name string) uint64 {
	value, ok := properties.Value(name)
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case uint64:
		return v
	case uint32:
		return uint64(v)
	case int32:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int:
		if v < 0 {
			return 0
		}
		return uint64(v)
	default:
		return 0
	}
}

func uniqueIDValue(properties arkproperty.Container) string {
	value, ok := properties.Value("UniqueID")
	if !ok {
		return ""
	}
	switch v := value.(type) {
	case arkproperty.ObjectReference:
		text, _ := v.Value.(string)
		return text
	case string:
		return v
	default:
		return ""
	}
}

func stringArrayValue(properties arkproperty.Container, name string) []string {
	raw, ok := properties.Value(name)
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(array.Values))
	for _, value := range array.Values {
		if text, ok := value.(string); ok {
			out = append(out, text)
		}
	}
	return out
}

func objectReferenceStringArrayValue(properties arkproperty.Container, name string) []string {
	raw, ok := properties.Value(name)
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(array.Values))
	for _, value := range array.Values {
		switch v := value.(type) {
		case arkproperty.ObjectReference:
			if text, ok := v.Value.(string); ok && text != "" {
				out = append(out, text)
			}
		case string:
			if v != "" {
				out = append(out, v)
			}
		}
	}
	return out
}

func int32ArrayValue(properties arkproperty.Container, name string) []int32 {
	raw, ok := properties.Value(name)
	if !ok {
		return nil
	}
	array, ok := raw.(arkproperty.Array)
	if !ok {
		return nil
	}
	out := make([]int32, 0, len(array.Values))
	for _, value := range array.Values {
		switch v := value.(type) {
		case int32:
			out = append(out, v)
		case uint32:
			out = append(out, int32(v))
		case int:
			out = append(out, int32(v))
		}
	}
	return out
}
