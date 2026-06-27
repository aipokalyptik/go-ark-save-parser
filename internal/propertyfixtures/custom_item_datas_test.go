package propertyfixtures

import (
	"bytes"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

func TestCustomItemDatasPropertyBuildsEntryCountProperty(t *testing.T) {
	prop := CustomItemDatasProperty(2)

	if prop.Name != "CustomItemDatas" || prop.Type != arkproperty.TypeArray {
		t.Fatalf("CustomItemDatasProperty() header = %#v", prop)
	}
	array, ok := prop.Value.(arkproperty.Array)
	if !ok {
		t.Fatalf("CustomItemDatasProperty() value type = %T, want arkproperty.Array", prop.Value)
	}
	if array.ElementType != arkproperty.TypeStruct || array.StructType != "CustomItemData" || len(array.Values) != 2 {
		t.Fatalf("CustomItemDatasProperty() array = %#v, want two CustomItemData entries", array)
	}
}

func TestCryopodCustomItemDatasPropertyBuildsNestedPayloadBytes(t *testing.T) {
	first := []byte{1, 2, 3}
	second := []byte{4, 5}

	prop := CryopodCustomItemDatasProperty(first, second)
	object := &arkobject.GameObject{Properties: []arkproperty.Property{prop}}
	payloads := arkobject.CryopodPayloadsFromObject(object)

	if len(payloads) != 2 || !bytes.Equal(payloads[0], first) || !bytes.Equal(payloads[1], second) {
		t.Fatalf("CryopodPayloadsFromObject() = %#v, want copied payloads", payloads)
	}
	payloads[0][0] = 9
	if first[0] == 9 {
		t.Fatalf("CryopodCustomItemDatasProperty() did not isolate caller payload bytes")
	}
}
