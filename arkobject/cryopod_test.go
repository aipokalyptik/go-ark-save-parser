package arkobject

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
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

func TestDinoFromCryopodObjectParsesEmbeddedDinoAndStatus(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	payload := syntheticCryopodDinoPayload(t, dinoID, statusID)
	cryopod := &GameObject{
		UUID:      uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122"),
		Blueprint: "Blueprint'/Game/Extinction/CoreBlueprints/Weapons/PrimalItem_WeaponEmptyCryopod.PrimalItem_WeaponEmptyCryopod_C'",
		Properties: []arkproperty.Property{
			customItemDatasProperty(payload),
		},
	}

	dino, ok, err := DinoFromCryopodObject(cryopod, 1<<20)
	if err != nil {
		t.Fatalf("DinoFromCryopodObject() error = %v", err)
	}
	if !ok {
		t.Fatalf("DinoFromCryopodObject() ok = false, want true")
	}
	if dino.UUID != dinoID || dino.ID1 != 1001 || dino.ID2 != 2002 {
		t.Fatalf("dino identity = %#v", dino)
	}
	if !dino.IsTamed || !dino.IsCryopodded {
		t.Fatalf("dino flags = %#v, want tamed cryopodded", dino)
	}
	if dino.Location == nil || !dino.Location.InCryopod {
		t.Fatalf("dino location = %#v, want in cryopod", dino.Location)
	}
	if dino.Stats == nil || dino.Stats.BaseLevel != 12 {
		t.Fatalf("dino stats = %#v, want base level 12", dino.Stats)
	}
}

func TestDinoFromCryopodObjectFindsReversedEmbeddedDinoAndStatus(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	payload := syntheticCryopodDinoPayloadWithOrder(t, dinoID, statusID, true)
	cryopod := &GameObject{
		UUID:       uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122"),
		Properties: []arkproperty.Property{customItemDatasProperty(payload)},
	}

	dino, ok, err := DinoFromCryopodObject(cryopod, 1<<20)
	if err != nil {
		t.Fatalf("DinoFromCryopodObject() error = %v", err)
	}
	if !ok || dino.UUID != dinoID || dino.Stats == nil || dino.Stats.BaseLevel != 12 {
		t.Fatalf("DinoFromCryopodObject() = %#v, %v; want reversed embedded dino/status parsed", dino, ok)
	}
}

func TestSaddleFromCryopodObjectParsesModernEmbeddedSaddle(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	cryopodID := uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122")
	dinoPayload := syntheticCryopodDinoPayload(t, dinoID, statusID)
	saddlePayload := syntheticCryopodSaddlePayload()
	cryopod := &GameObject{
		UUID:      cryopodID,
		Blueprint: "Blueprint'/Game/Extinction/CoreBlueprints/Weapons/PrimalItem_WeaponEmptyCryopod.PrimalItem_WeaponEmptyCryopod_C'",
		Properties: []arkproperty.Property{
			customItemDatasProperty(dinoPayload, saddlePayload),
		},
	}

	saddle, ok, err := SaddleFromCryopodObject(cryopod)
	if err != nil {
		t.Fatalf("SaddleFromCryopodObject() error = %v", err)
	}
	if !ok {
		t.Fatalf("SaddleFromCryopodObject() ok = false, want true")
	}
	if saddle.UUID != cryopodID {
		t.Fatalf("saddle UUID = %s, want containing cryopod UUID %s", saddle.UUID, cryopodID)
	}
	if saddle.Kind != EquipmentSaddle {
		t.Fatalf("saddle kind = %q, want saddle", saddle.Kind)
	}
	wantBlueprint := "/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C"
	if saddle.Blueprint != wantBlueprint {
		t.Fatalf("saddle blueprint = %q, want %q", saddle.Blueprint, wantBlueprint)
	}
	if saddle.Quantity != 1 {
		t.Fatalf("saddle quantity = %d, want default 1", saddle.Quantity)
	}
}

func TestDinoFromCryopodObjectIgnoresEmptyCryopod(t *testing.T) {
	dino, ok, err := DinoFromCryopodObject(&GameObject{}, 1<<20)
	if err != nil {
		t.Fatalf("DinoFromCryopodObject() error = %v", err)
	}
	if ok || dino.UUID != uuid.Nil {
		t.Fatalf("DinoFromCryopodObject() = %#v, %v; want no dino", dino, ok)
	}
}

func customItemDatasProperty(payloads ...[]byte) arkproperty.Property {
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

func syntheticCryopodSaddlePayload() []byte {
	var buf bytes.Buffer
	testfixtures.WriteUInt32(&buf, 8)
	testfixtures.WriteUInt32(&buf, 7)
	testfixtures.WriteUInt32(&buf, 0)
	testfixtures.WriteUInt32(&buf, 0)
	writeCryopodSaddlePathObjectProperty(&buf, "ItemArchetype", "BlueprintGeneratedClass /Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C")
	testfixtures.WriteArkString(&buf, "None")
	return buf.Bytes()
}

func writeCryopodSaddlePathObjectProperty(buf *bytes.Buffer, name string, path string) {
	testfixtures.WriteArkString(buf, name)
	testfixtures.WriteArkString(buf, "ObjectProperty")
	var body bytes.Buffer
	testfixtures.WriteInt32(&body, 1)
	testfixtures.WriteArkString(&body, path)
	testfixtures.WriteInt32(buf, int32(body.Len()))
	testfixtures.WriteInt32(buf, 0)
	buf.WriteByte(0)
	buf.Write(body.Bytes())
}

func syntheticCryopodDinoPayload(t *testing.T, dinoID uuid.UUID, statusID uuid.UUID) []byte {
	return syntheticCryopodDinoPayloadWithOrder(t, dinoID, statusID, false)
}

func syntheticCryopodDinoPayloadWithOrder(t *testing.T, dinoID uuid.UUID, statusID uuid.UUID, reversed bool) []byte {
	t.Helper()

	var decoded bytes.Buffer
	testfixtures.WriteInt32(&decoded, 0)
	testfixtures.WriteInt32(&decoded, 0)
	testfixtures.WriteUInt32(&decoded, 2)
	var dinoOffsetPos int
	var statusOffsetPos int
	if reversed {
		statusOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, statusID, "Status", []string{"S0"})
		dinoOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, dinoID, "Dino", []string{"D0"})
	} else {
		dinoOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, dinoID, "Dino", []string{"D0"})
		statusOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, statusID, "Status", []string{"S0"})
	}

	dinoPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000001, 1001)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000002, 2002)
	writeCryopodEmbeddedNameDoubleProperty(&decoded, 0x10000003, 42)
	writeCryopodEmbeddedNone(&decoded)
	statusPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000005, 12)
	writeCryopodEmbeddedNone(&decoded)

	binary.LittleEndian.PutUint32(decoded.Bytes()[dinoOffsetPos:dinoOffsetPos+4], uint32(dinoPropsOffset))
	binary.LittleEndian.PutUint32(decoded.Bytes()[statusOffsetPos:statusOffsetPos+4], uint32(statusPropsOffset))

	namesOffset := decoded.Len()
	testfixtures.WriteUInt32(&decoded, 7)
	testfixtures.WriteArkString(&decoded, "None")
	testfixtures.WriteArkString(&decoded, "DinoID1")
	testfixtures.WriteArkString(&decoded, "DinoID2")
	testfixtures.WriteArkString(&decoded, "TamedTimeStamp")
	testfixtures.WriteArkString(&decoded, "IntProperty")
	testfixtures.WriteArkString(&decoded, "BaseCharacterLevel")
	testfixtures.WriteArkString(&decoded, "DoubleProperty")

	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(decoded.Bytes()); err != nil {
		t.Fatalf("zlib write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("zlib close: %v", err)
	}

	var payload bytes.Buffer
	testfixtures.WriteUInt32(&payload, 0x0407)
	testfixtures.WriteUInt32(&payload, uint32(decoded.Len()))
	testfixtures.WriteUInt32(&payload, uint32(namesOffset))
	payload.Write(compressed.Bytes())
	return payload.Bytes()
}

func writeCryopodEmbeddedObjectHeader(buf *bytes.Buffer, id uuid.UUID, className string, names []string) int {
	buf.Write(id[:])
	testfixtures.WriteArkString(buf, className)
	testfixtures.WriteUInt32(buf, 0)
	testfixtures.WriteStringArray(buf, names)
	testfixtures.WriteUInt32(buf, 0)
	testfixtures.WriteInt32(buf, 0)
	testfixtures.WriteUInt32(buf, 0)
	offsetPos := buf.Len()
	testfixtures.WriteInt32(buf, 0)
	testfixtures.WriteUInt32(buf, 0)
	return offsetPos
}

func writeCryopodEmbeddedNameIntProperty(buf *bytes.Buffer, nameID uint32, value int32) {
	writeCryopodEmbeddedName(buf, nameID)
	writeCryopodEmbeddedName(buf, 0x10000004)
	testfixtures.WriteInt32(buf, 4)
	testfixtures.WriteInt32(buf, 0)
	buf.WriteByte(0)
	testfixtures.WriteInt32(buf, value)
}

func writeCryopodEmbeddedNameDoubleProperty(buf *bytes.Buffer, nameID uint32, value float64) {
	writeCryopodEmbeddedName(buf, nameID)
	writeCryopodEmbeddedName(buf, 0x10000006)
	testfixtures.WriteInt32(buf, 8)
	testfixtures.WriteInt32(buf, 0)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeCryopodEmbeddedNone(buf *bytes.Buffer) {
	writeCryopodEmbeddedName(buf, 0x10000000)
}

func writeCryopodEmbeddedName(buf *bytes.Buffer, nameID uint32) {
	testfixtures.WriteUInt32(buf, nameID)
	testfixtures.WriteInt32(buf, 0)
}
