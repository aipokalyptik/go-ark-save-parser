package arkobject

import (
	"bytes"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

func TestCryopodPayloadsFromObjectExtractsCustomItemByteArrays(t *testing.T) {
	first := []byte{0x78, 0x9c, 0x01, 0x02}
	second := []byte{0x03, 0x04, 0x05}
	object := &GameObject{
		Properties: []arkproperty.Property{
			{
				Name: "CustomItemDatas",
				Type: arkproperty.TypeArray,
				Value: arkproperty.Array{
					ElementType: arkproperty.TypeStruct,
					StructType:  "CustomItemData",
					Values: []any{
						arkproperty.Container{Properties: []arkproperty.Property{
							{
								Name: "CustomDataBytes",
								Type: arkproperty.TypeStruct,
								Value: arkproperty.Container{
									Properties: []arkproperty.Property{{
										Name: "ByteArrays",
										Type: arkproperty.TypeArray,
										Value: arkproperty.Array{
											ElementType: arkproperty.TypeStruct,
											StructType:  "CustomItemByteArray",
											Values: []any{
												arkproperty.Container{Properties: []arkproperty.Property{{
													Name: "Bytes",
													Type: arkproperty.TypeArray,
													Value: arkproperty.Array{
														ElementType: arkproperty.TypeByte,
														Values:      []any{first[0], first[1], first[2], first[3]},
													},
												}}},
												arkproperty.Container{Properties: []arkproperty.Property{{
													Name: "Bytes",
													Type: arkproperty.TypeArray,
													Value: arkproperty.Array{
														ElementType: arkproperty.TypeByte,
														Values:      []any{second[0], second[1], second[2]},
													},
												}}},
											},
										},
									},
									},
								},
							},
						}},
					},
				},
			},
		},
	}

	payloads := CryopodPayloadsFromObject(object)
	if len(payloads) != 2 {
		t.Fatalf("CryopodPayloadsFromObject() length = %d, want 2", len(payloads))
	}
	if !bytes.Equal(payloads[0], first) || !bytes.Equal(payloads[1], second) {
		t.Fatalf("CryopodPayloadsFromObject() = %#v, want %#v and %#v", payloads, first, second)
	}

	first[0] = 0xff
	if payloads[0][0] == first[0] {
		t.Fatalf("CryopodPayloadsFromObject() did not copy byte payload")
	}
}

func TestCryopodPayloadsFromObjectIgnoresMissingOrMalformedCustomData(t *testing.T) {
	for name, object := range map[string]*GameObject{
		"nil": nil,
		"missing": {
			Properties: nil,
		},
		"wrong property type": {
			Properties: []arkproperty.Property{{Name: "CustomItemDatas", Value: "bad"}},
		},
		"empty byte arrays": {
			Properties: []arkproperty.Property{{
				Name: "CustomItemDatas",
				Value: arkproperty.Array{
					ElementType: arkproperty.TypeStruct,
					StructType:  "CustomItemData",
					Values: []any{
						arkproperty.Container{Properties: []arkproperty.Property{{
							Name: "CustomDataBytes",
							Value: arkproperty.Container{
								Properties: []arkproperty.Property{{
									Name: "ByteArrays",
									Value: arkproperty.Array{
										ElementType: arkproperty.TypeStruct,
										StructType:  "CustomItemByteArray",
									},
								}},
							},
						}}},
					},
				},
			}},
		},
		"shallow byte arrays": {
			Properties: []arkproperty.Property{{
				Name: "CustomItemDatas",
				Value: arkproperty.Array{
					ElementType: arkproperty.TypeStruct,
					StructType:  "CustomItemData",
					Values: []any{
						arkproperty.Container{Properties: []arkproperty.Property{{
							Name: "ByteArrays",
							Value: arkproperty.Array{
								ElementType: arkproperty.TypeStruct,
								StructType:  "CustomItemByteArray",
								Values: []any{
									arkproperty.Container{Properties: []arkproperty.Property{{
										Name: "Bytes",
										Value: arkproperty.Array{
											ElementType: arkproperty.TypeByte,
											Values:      []any{byte(1)},
										},
									}}},
								},
							},
						}}},
					},
				},
			}},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if payloads := CryopodPayloadsFromObject(object); len(payloads) != 0 {
				t.Fatalf("CryopodPayloadsFromObject() = %#v, want no payloads", payloads)
			}
		})
	}
}
