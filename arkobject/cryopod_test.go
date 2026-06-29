package arkobject

import (
	"bytes"
	"encoding/binary"
	"errors"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/propertyfixtures"
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
	payload := testfixtures.CryopodDinoPayload(t, dinoID, statusID, testfixtures.CryopodDinoPayloadOptions{})
	cryopod := &GameObject{
		UUID:      uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122"),
		Blueprint: "Blueprint'/Game/Extinction/CoreBlueprints/Weapons/PrimalItem_WeaponEmptyCryopod.PrimalItem_WeaponEmptyCryopod_C'",
		Properties: []arkproperty.Property{
			propertyfixtures.CryopodCustomItemDatasProperty(payload),
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
	payload := testfixtures.CryopodDinoPayload(t, dinoID, statusID, testfixtures.CryopodDinoPayloadOptions{Reversed: true})
	cryopod := &GameObject{
		UUID:       uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122"),
		Properties: []arkproperty.Property{propertyfixtures.CryopodCustomItemDatasProperty(payload)},
	}

	dino, ok, err := DinoFromCryopodObject(cryopod, 1<<20)
	if err != nil {
		t.Fatalf("DinoFromCryopodObject() error = %v", err)
	}
	if !ok || dino.UUID != dinoID || dino.Stats == nil || dino.Stats.BaseLevel != 12 {
		t.Fatalf("DinoFromCryopodObject() = %#v, %v; want reversed embedded dino/status parsed", dino, ok)
	}
}

func TestDinoFromCryopodObjectClassifiesMalformedPayload(t *testing.T) {
	var payload bytes.Buffer
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0x0406))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0))
	_ = binary.Write(&payload, binary.LittleEndian, uint32(0))
	cryopod := &GameObject{
		UUID:       uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122"),
		Properties: []arkproperty.Property{propertyfixtures.CryopodCustomItemDatasProperty(payload.Bytes())},
	}

	_, _, err := DinoFromCryopodObject(cryopod, 1<<20)
	if !errors.Is(err, arkbinary.ErrUnsupportedEmbeddedDataVersion) {
		t.Fatalf("DinoFromCryopodObject() error = %v, want ErrUnsupportedEmbeddedDataVersion", err)
	}
	var cryopodErr *CryopodPayloadError
	if !errors.As(err, &cryopodErr) || cryopodErr.Kind != CryopodPayloadDinoParse {
		t.Fatalf("DinoFromCryopodObject() typed error = %#v, want dino parse kind", cryopodErr)
	}
}

func TestSaddleFromCryopodObjectParsesModernEmbeddedSaddle(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	cryopodID := uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122")
	dinoPayload := testfixtures.CryopodDinoPayload(t, dinoID, statusID, testfixtures.CryopodDinoPayloadOptions{})
	saddlePayload := testfixtures.CryopodSaddlePayload()
	cryopod := &GameObject{
		UUID:      cryopodID,
		Blueprint: "Blueprint'/Game/Extinction/CoreBlueprints/Weapons/PrimalItem_WeaponEmptyCryopod.PrimalItem_WeaponEmptyCryopod_C'",
		Properties: []arkproperty.Property{
			propertyfixtures.CryopodCustomItemDatasProperty(dinoPayload, saddlePayload),
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

func TestSaddleFromCryopodObjectClassifiesUnsupportedVersion(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")
	var unsupportedSaddle bytes.Buffer
	testfixtures.WriteUInt32(&unsupportedSaddle, 8)
	testfixtures.WriteUInt32(&unsupportedSaddle, 0)
	testfixtures.WriteUInt32(&unsupportedSaddle, 0)
	testfixtures.WriteUInt32(&unsupportedSaddle, 0)
	cryopod := &GameObject{
		UUID: uuid.MustParse("21222324-2526-2728-292a-2b2c2d2e2122"),
		Properties: []arkproperty.Property{
			propertyfixtures.CryopodCustomItemDatasProperty(
				testfixtures.CryopodDinoPayload(t, dinoID, statusID, testfixtures.CryopodDinoPayloadOptions{}),
				unsupportedSaddle.Bytes(),
			),
		},
	}

	_, ok, err := SaddleFromCryopodObject(cryopod)
	if ok || !errors.Is(err, ErrUnsupportedCryopodSaddleVersion) {
		t.Fatalf("SaddleFromCryopodObject() ok/error = %v/%v, want unsupported saddle version", ok, err)
	}
	var cryopodErr *CryopodPayloadError
	if !errors.As(err, &cryopodErr) || cryopodErr.Kind != CryopodPayloadUnsupportedSaddleVersion || cryopodErr.Version != 0 {
		t.Fatalf("SaddleFromCryopodObject() typed error = %#v, want unsupported version 0", cryopodErr)
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
