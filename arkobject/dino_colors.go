package arkobject

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

func colorSetIndices(properties arkproperty.Container) [6]int {
	var values [6]int
	for i := range values {
		value, ok := properties.PositionedValue("ColorSetIndices", int32(i))
		if !ok {
			continue
		}
		switch v := value.(type) {
		case int8:
			values[i] = int(v)
		case int32:
			values[i] = int(v)
		case int:
			values[i] = v
		}
	}
	return values
}

func colorSetNames(properties arkproperty.Container) [6]string {
	var values [6]string
	for i := range values {
		values[i] = "None"
		value, ok := properties.PositionedValue("ColorSetNames", int32(i))
		if !ok {
			continue
		}
		if name, ok := value.(string); ok && name != "" {
			values[i] = name
		}
	}
	return values
}
