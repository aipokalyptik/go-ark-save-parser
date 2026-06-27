package propertyfixtures

import "github.com/aipokalyptik/go-ark-save-parser/arkproperty"

func CustomItemDatasProperty(entries int) arkproperty.Property {
	values := make([]any, 0, entries)
	for i := 0; i < entries; i++ {
		values = append(values, arkproperty.Container{})
	}
	return arkproperty.Property{
		Name: "CustomItemDatas",
		Type: arkproperty.TypeArray,
		Value: arkproperty.Array{
			ElementType: arkproperty.TypeStruct,
			StructType:  "CustomItemData",
			Values:      values,
		},
	}
}

func CryopodCustomItemDatasProperty(payloads ...[]byte) arkproperty.Property {
	byteArrayValues := make([]any, 0, len(payloads))
	for _, payload := range payloads {
		values := make([]any, 0, len(payload))
		for _, value := range payload {
			values = append(values, value)
		}
		byteArrayValues = append(byteArrayValues, arkproperty.Container{Properties: []arkproperty.Property{{
			Name: "Bytes",
			Type: arkproperty.TypeArray,
			Value: arkproperty.Array{
				ElementType: arkproperty.TypeByte,
				Values:      values,
			},
		}}})
	}
	return arkproperty.Property{
		Name: "CustomItemDatas",
		Type: arkproperty.TypeArray,
		Value: arkproperty.Array{
			ElementType: arkproperty.TypeStruct,
			StructType:  "CustomItemData",
			Values: []any{
				arkproperty.Container{Properties: []arkproperty.Property{{
					Name: "CustomDataBytes",
					Type: arkproperty.TypeStruct,
					Value: arkproperty.Container{Properties: []arkproperty.Property{{
						Name: "ByteArrays",
						Type: arkproperty.TypeArray,
						Value: arkproperty.Array{
							ElementType: arkproperty.TypeStruct,
							StructType:  "CustomItemByteArray",
							Values:      byteArrayValues,
						},
					}}},
				}}},
			},
		},
	}
}
