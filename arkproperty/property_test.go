package arkproperty

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
)

func TestParsePropertiesReadsPrimitivePropertiesUntilNone(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1:  "Health",
		2:  "IntProperty",
		3:  "Label",
		4:  "StrProperty",
		5:  "IsActive",
		6:  "BoolProperty",
		7:  "None",
		8:  "Weight",
		9:  "FloatProperty",
		10: "Precise",
		11: "DoubleProperty",
		12: "Count",
		13: "UInt32Property",
		14: "LargeID",
		15: "UInt64Property",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 250)

	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 10)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteArkString(stream, "hello")

	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteNameID(stream, 6)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(1)

	testfixtures.WriteNameID(stream, 8)
	testfixtures.WriteNameID(stream, 9)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, float32(3.5))

	testfixtures.WriteNameID(stream, 10)
	testfixtures.WriteNameID(stream, 11)
	testfixtures.WriteInt32(stream, 8)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, float64(8.25))

	testfixtures.WriteNameID(stream, 12)
	testfixtures.WriteNameID(stream, 13)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, uint32(99))

	testfixtures.WriteNameID(stream, 14)
	testfixtures.WriteNameID(stream, 15)
	testfixtures.WriteInt32(stream, 8)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, uint64(9876543210))

	testfixtures.WriteNameID(stream, 7)
	testfixtures.WriteInt32(stream, 0)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 7 {
		t.Fatalf("ParseAll() length = %d, want 7", len(props))
	}
	if props[0].Name != "Health" || props[0].Type != TypeInt || props[0].Value != int32(250) {
		t.Fatalf("first property = %#v, want Health Int 250", props[0])
	}
	if props[1].Name != "Label" || props[1].Type != TypeString || props[1].Value != "hello" {
		t.Fatalf("second property = %#v, want Label String hello", props[1])
	}
	if props[2].Name != "IsActive" || props[2].Type != TypeBool || props[2].Value != true {
		t.Fatalf("third property = %#v, want IsActive Bool true", props[2])
	}
	if props[3].Name != "Weight" || props[3].Type != TypeFloat || props[3].Value != float32(3.5) {
		t.Fatalf("fourth property = %#v, want Weight Float 3.5", props[3])
	}
	if props[4].Name != "Precise" || props[4].Type != TypeDouble || props[4].Value != float64(8.25) {
		t.Fatalf("fifth property = %#v, want Precise Double 8.25", props[4])
	}
	if props[5].Name != "Count" || props[5].Type != TypeUInt32 || props[5].Value != uint32(99) {
		t.Fatalf("sixth property = %#v, want Count UInt32 99", props[5])
	}
	if props[6].Name != "LargeID" || props[6].Type != TypeUInt64 || props[6].Value != uint64(9876543210) {
		t.Fatalf("seventh property = %#v, want LargeID UInt64 9876543210", props[6])
	}
}

func TestParsePropertiesNormalizesPositionFlagPrimitiveOnlyWhenFlagIsFalse(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Health",
		2: "IntProperty",
		3: "Weight",
		4: "FloatProperty",
		5: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 7)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 250)

	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 8)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, float32(3.5))

	testfixtures.WriteNameID(stream, 5)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() length = %d, want 2", len(props))
	}
	if props[0].Name != "Health" || props[0].Position != 7 {
		t.Fatalf("Health Position = %d, want 7", props[0].Position)
	}
	if props[1].Name != "Weight" || props[1].Position != 0 {
		t.Fatalf("Weight Position = %d, want 0", props[1].Position)
	}
}

func TestParsePropertiesGeneratesUnknownNamesForUnknownPropertyKeys(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		2: "IntProperty",
		7: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 0)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 42)
	testfixtures.WriteNameID(stream, 7)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() returned %d properties, want 1", len(props))
	}
	if props[0].Name != "Unknown_0" || props[0].Value != int32(42) {
		t.Fatalf("property = %#v, want generated unknown key with int32 value 42", props[0])
	}
	if name, ok := ctx.Name(0); ok {
		t.Fatalf("generated property key leaked into context as %q", name)
	}
}

func TestParsePropertiesPreservesUnknownPropertyTypeAndContinues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Mystery",
		2: "MysteryProperty",
		3: "AfterMystery",
		4: "IntProperty",
		5: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	stream.Write([]byte{0xde, 0xad, 0xbe, 0xef})

	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 42)
	testfixtures.WriteNameID(stream, 5)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() returned %d properties, want unknown property plus trailing int: %#v", len(props), props)
	}
	if props[0].Name != "Mystery" || props[0].Type != Type("MysteryProperty") {
		t.Fatalf("unknown property identity = %#v", props[0])
	}
	got, ok := props[0].Value.(UnknownValue)
	if !ok {
		t.Fatalf("unknown property value type = %T, want UnknownValue", props[0].Value)
	}
	if got.TypeName != "MysteryProperty" || !bytes.Equal(got.Raw, []byte{0xde, 0xad, 0xbe, 0xef}) {
		t.Fatalf("UnknownValue = %#v, want preserved raw MysteryProperty payload", got)
	}
	if props[1].Name != "AfterMystery" || props[1].Type != TypeInt || props[1].Value != int32(42) {
		t.Fatalf("trailing property = %#v, want AfterMystery int 42", props[1])
	}
}

func TestParsePropertyReturnsNilForNoneMarkerWithoutTrailingZeros(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{7: "None"})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 7)

	prop, err := ParseOne(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseOne() error = %v", err)
	}
	if prop != nil {
		t.Fatalf("ParseOne() = %#v, want nil", prop)
	}
}

func TestParseObjectPropertyReadsUUIDReference(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Target",
		2: "ObjectProperty",
		3: "None",
	})
	ref := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 18)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, int16(0))
	stream.Write(ref)
	testfixtures.WriteNameID(stream, 3)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	got, ok := props[0].Value.(ObjectReference)
	if !ok {
		t.Fatalf("ObjectProperty value type = %T, want ObjectReference", props[0].Value)
	}
	if got.Type != ObjectReferenceUUID || got.Value != "00112233-4455-6677-8899-aabbccddeeff" {
		t.Fatalf("ObjectReference = %#v, want UUID reference", got)
	}
}

func TestParseObjectPropertyReadsLocalArchivePathReference(t *testing.T) {
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteArkString(stream, "ItemArchetype")
	testfixtures.WriteArkString(stream, "ObjectProperty")
	var body bytes.Buffer
	testfixtures.WriteInt32(&body, 1)
	testfixtures.WriteArkString(&body, "BlueprintGeneratedClass /Game/Test/Item.Item_C")
	testfixtures.WriteInt32(stream, int32(body.Len()))
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	stream.Write(body.Bytes())
	testfixtures.WriteArkString(stream, "None")

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), nil), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	got, ok := props[0].Value.(ObjectReference)
	if !ok {
		t.Fatalf("ObjectProperty value type = %T, want ObjectReference", props[0].Value)
	}
	if got.Type != ObjectReferencePath || got.Value != "BlueprintGeneratedClass /Game/Test/Item.Item_C" {
		t.Fatalf("ObjectReference = %#v, want local path reference", got)
	}
}

func TestParseObjectPropertyReadsLocalArchiveNonePathReference(t *testing.T) {
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteArkString(stream, "ItemArchetype")
	testfixtures.WriteArkString(stream, "ObjectProperty")
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteArkString(stream, "None")

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), nil), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	got, ok := props[0].Value.(ObjectReference)
	if !ok {
		t.Fatalf("ObjectProperty value type = %T, want ObjectReference", props[0].Value)
	}
	if got.Type != ObjectReferencePath || got.Value != "NONE" {
		t.Fatalf("ObjectReference = %#v, want NONE path reference", got)
	}
}

func TestParseSoftObjectPropertyReadsTerminatedNameList(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "CustomCosmeticBuffToGiveWhenEquipped",
		2: "SoftObjectProperty",
		3: "/Game/PrimalEarth/CoreBlueprints/Items/Armor/Skin/CustomCosmeticBuff",
		4: "CustomCosmeticBuff_C",
		5: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 20)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteNameID(stream, 5)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	if props[0].Type != TypeSoftObject {
		t.Fatalf("property type = %s, want SoftObject", props[0].Type)
	}
	got, ok := props[0].Value.([]string)
	if !ok {
		t.Fatalf("SoftObjectProperty value type = %T, want []string", props[0].Value)
	}
	want := []string{
		"/Game/PrimalEarth/CoreBlueprints/Items/Armor/Skin/CustomCosmeticBuff",
		"CustomCosmeticBuff_C",
	}
	if len(got) != len(want) {
		t.Fatalf("SoftObjectProperty length = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SoftObjectProperty[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestParseNamePropertyReadsNameValue(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "CostumeOverrideRiderSocketName",
		2: "NameProperty",
		3: "RiderSocket",
		4: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 8)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	if props[0].Type != TypeName || props[0].Value != "RiderSocket" {
		t.Fatalf("NameProperty = %#v, want RiderSocket", props[0])
	}
}

func TestParsePrimitivePropertyRealignsToDeclaredDataSize(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Count",
		2: "IntProperty",
		3: "Label",
		4: "StrProperty",
		5: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 6)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 99)
	stream.Write([]byte{0xaa, 0xbb})

	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 7)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteArkString(stream, "ok")
	testfixtures.WriteNameID(stream, 5)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() length = %d, want 2", len(props))
	}
	if props[0].Name != "Count" || props[0].Value != int32(99) {
		t.Fatalf("first property = %#v, want Count 99", props[0])
	}
	if props[1].Name != "Label" || props[1].Value != "ok" {
		t.Fatalf("second property = %#v, want Label ok", props[1])
	}
}

func TestParseStructPropertyReturnsPartialContainerOnDeclaredBodyOverread(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "MyData",
		2: "StructProperty",
		3: "PlayerDataID",
		4: "IntProperty",
		5: "None",
		6: "PlayerData",
	})
	var body bytes.Buffer
	testfixtures.WriteNameID(&body, 3)
	testfixtures.WriteNameID(&body, 4)
	testfixtures.WriteInt32(&body, 4)
	testfixtures.WriteInt32(&body, 0)
	body.WriteByte(0)
	testfixtures.WriteInt32(&body, 42)
	testfixtures.WriteNameID(&body, 5)

	var stream bytes.Buffer
	testfixtures.WriteNameID(&stream, 1)
	testfixtures.WriteNameID(&stream, 2)
	testfixtures.WriteUInt32(&stream, 1)
	testfixtures.WriteNameID(&stream, 6)
	testfixtures.WriteUInt32(&stream, 1)
	testfixtures.WriteNameID(&stream, 6)
	testfixtures.WriteUInt32(&stream, 0)
	testfixtures.WriteUInt32(&stream, uint32(body.Len()-2))
	stream.WriteByte(0)
	stream.Write(body.Bytes())

	prop, err := ParseOne(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err == nil || !isRecoverableCompoundError(err) {
		t.Fatalf("ParseOne() error = %v, want recoverable compound error", err)
	}
	if prop == nil {
		t.Fatalf("ParseOne() prop = nil, want partial property")
	}
	container, ok := prop.Value.(Container)
	if !ok {
		t.Fatalf("MyData value type = %T, want Container", prop.Value)
	}
	if value, ok := container.Value("PlayerDataID"); !ok || value != int32(42) {
		t.Fatalf("PlayerDataID = %#v, %v; want 42, true", value, ok)
	}
}

func TestParsePropertyPreservesEncodedBytes(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Count",
		2: "IntProperty",
		3: "None",
	})
	propertyBytes := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(propertyBytes, 1)
	testfixtures.WriteNameID(propertyBytes, 2)
	testfixtures.WriteInt32(propertyBytes, 4)
	testfixtures.WriteInt32(propertyBytes, 0)
	propertyBytes.WriteByte(0)
	testfixtures.WriteInt32(propertyBytes, 99)

	stream := bytes.NewBuffer(nil)
	stream.Write(propertyBytes.Bytes())
	testfixtures.WriteNameID(stream, 3)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	if !bytes.Equal(props[0].EncodedBytes, propertyBytes.Bytes()) {
		t.Fatalf("EncodedBytes = % x, want % x", props[0].EncodedBytes, propertyBytes.Bytes())
	}
	props[0].EncodedBytes[0] = 0xff
	if propertyBytes.Bytes()[0] == 0xff {
		t.Fatalf("EncodedBytes shares backing storage with source bytes")
	}
}

func TestParseInt64PropertyReadsSignedValue(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "CustomCosmeticModSkinReplacementID",
		2: "Int64Property",
		3: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 8)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, int64(-42))
	testfixtures.WriteNameID(stream, 3)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	if props[0].Type != TypeInt64 || props[0].Value != int64(-42) {
		t.Fatalf("Int64Property = %#v, want -42", props[0])
	}
}

func TestParseSmallIntegerProperties(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Int8Value",
		2: "Int8Property",
		3: "Int16Value",
		4: "Int16Property",
		5: "UInt16Value",
		6: "UInt16Property",
		7: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	stream.WriteByte(0xf9)

	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, int16(-32000))

	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteNameID(stream, 6)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, uint16(65000))

	testfixtures.WriteNameID(stream, 7)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 3 {
		t.Fatalf("ParseAll() length = %d, want 3", len(props))
	}
	if props[0].Type != TypeInt8 || props[0].Value != int8(-7) {
		t.Fatalf("Int8Property = %#v", props[0])
	}
	if props[1].Type != TypeInt16 || props[1].Value != int16(-32000) {
		t.Fatalf("Int16Property = %#v", props[1])
	}
	if props[2].Type != TypeUInt16 || props[2].Value != uint16(65000) {
		t.Fatalf("UInt16Property = %#v", props[2])
	}
}

func TestParseBytePropertyReadsRawByteValue(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "PaintRegion",
		2: "ByteProperty",
		3: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	stream.WriteByte(7)
	testfixtures.WriteNameID(stream, 3)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	if props[0].Type != TypeByte || props[0].Value != byte(7) {
		t.Fatalf("ByteProperty = %#v, want raw byte 7", props[0])
	}
}

func TestParseBytePropertyReadsEnumValue(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "QualityTier",
		2: "ByteProperty",
		3: "EPrimalItemQuality",
		4: "/Script/ShooterGame.EPrimalItemQuality",
		5: "EPrimalItemQuality::Journeyman",
		6: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteUInt32(stream, 0)
	stream.WriteByte(1)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteNameID(stream, 6)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	got, ok := props[0].Value.(EnumValue)
	if !ok {
		t.Fatalf("ByteProperty enum value type = %T, want EnumValue", props[0].Value)
	}
	if props[0].Type != TypeEnum || got.Name != "EPrimalItemQuality::Journeyman" {
		t.Fatalf("Enum ByteProperty = %#v, want Journeyman enum", props[0])
	}
}

func TestParseArrayPropertyReadsIntValues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Levels",
		2: "ArrayProperty",
		3: "IntProperty",
		4: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 3)
	testfixtures.WriteUInt32(stream, 3)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 12)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 3)
	testfixtures.WriteInt32(stream, 5)
	testfixtures.WriteInt32(stream, 8)
	testfixtures.WriteInt32(stream, 13)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	if props[0].Type != TypeArray {
		t.Fatalf("property type = %s, want Array", props[0].Type)
	}
	got, ok := props[0].Value.(Array)
	if !ok {
		t.Fatalf("ArrayProperty value type = %T, want Array", props[0].Value)
	}
	want := []any{int32(5), int32(8), int32(13)}
	if got.ElementType != TypeInt || len(got.Values) != len(want) {
		t.Fatalf("Array = %#v, want Int array length %d", got, len(want))
	}
	for i := range want {
		if got.Values[i] != want[i] {
			t.Fatalf("Array value %d = %#v, want %#v", i, got.Values[i], want[i])
		}
	}
}

func TestParseCompactPositionedNumericProperties(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "NumberOfLevelUpPointsApplied",
		2: "IntProperty",
		3: "CurrentStatusValues",
		4: "FloatProperty",
		5: "None",
	})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 8)
	testfixtures.WriteInt32(stream, 508)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 7)
	testfixtures.WriteFloat32(stream, 321.25)
	testfixtures.WriteNameID(stream, 5)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() length = %d, want 2: %#v", len(props), props)
	}
	if props[0].Position != 8 || props[0].Value != int32(508) {
		t.Fatalf("compact int property = %#v, want position 8 value 508", props[0])
	}
	if props[1].Position != 7 || props[1].Value != float32(321.25) {
		t.Fatalf("compact float property = %#v, want position 7 value 321.25", props[1])
	}
}

func TestParseArrayPropertyReadsUInt64Values(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "ItemIDs",
		2: "ArrayProperty",
		3: "UInt64Property",
		4: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteUInt32(stream, 3)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 20)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 2)
	_ = binary.Write(stream, binary.LittleEndian, uint64(1001))
	_ = binary.Write(stream, binary.LittleEndian, uint64(1002))
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeArray {
		t.Fatalf("ParseAll() = %#v, want one array property", props)
	}
	got, ok := props[0].Value.(Array)
	if !ok {
		t.Fatalf("ArrayProperty value type = %T, want Array", props[0].Value)
	}
	if got.ElementType != TypeUInt64 || len(got.Values) != 2 || got.Values[0] != uint64(1001) || got.Values[1] != uint64(1002) {
		t.Fatalf("Array = %#v, want UInt64 values 1001 and 1002", got)
	}
}

func TestParseArrayPropertyRealignsToDeclaredElementDataSize(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Levels",
		2: "ArrayProperty",
		3: "IntProperty",
		4: "Label",
		5: "StrProperty",
		6: "None",
	})
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteUInt32(stream, 3)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 6)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteInt32(stream, 5)
	stream.Write([]byte{0xaa, 0xbb})

	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteInt32(stream, 7)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteArkString(stream, "ok")
	testfixtures.WriteNameID(stream, 6)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() length = %d, want 2", len(props))
	}
	if props[0].Name != "Levels" || props[1].Name != "Label" || props[1].Value != "ok" {
		t.Fatalf("ParseAll() = %#v, want array followed by Label ok", props)
	}
}

func TestParseArrayPropertyReadsGenericStructValues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Items",
		2: "ArrayProperty",
		3: "StructProperty",
		4: "ItemStruct",
		5: "Durability",
		6: "IntProperty",
		7: "None",
	})

	var element bytes.Buffer
	testfixtures.WriteNameID(&element, 5)
	testfixtures.WriteNameID(&element, 6)
	testfixtures.WriteInt32(&element, 4)
	testfixtures.WriteInt32(&element, 0)
	element.WriteByte(0)
	testfixtures.WriteInt32(&element, 88)
	testfixtures.WriteNameID(&element, 7)

	dataSize := uint32(4 + element.Len())
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, int32(dataSize))
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, dataSize)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 1)
	stream.Write(element.Bytes())
	testfixtures.WriteNameID(stream, 7)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeArray {
		t.Fatalf("ParseAll() = %#v, want one array property", props)
	}
	got, ok := props[0].Value.(Array)
	if !ok {
		t.Fatalf("ArrayProperty value type = %T, want Array", props[0].Value)
	}
	if got.ElementType != TypeStruct || got.StructType != "ItemStruct" || len(got.Values) != 1 {
		t.Fatalf("Array = %#v, want one ItemStruct value", got)
	}
	container, ok := got.Values[0].(Container)
	if !ok {
		t.Fatalf("struct array element type = %T, want Container", got.Values[0])
	}
	value, ok := container.Value("Durability")
	if !ok || value != int32(88) {
		t.Fatalf("Durability = %#v, %v; want 88, true", value, ok)
	}
}

func TestParseArrayPropertyReadsCustomItemDataByteArrays(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1:  "CustomItemDatas",
		2:  "ArrayProperty",
		3:  "StructProperty",
		4:  "CustomItemData",
		5:  "CustomDataBytes",
		6:  "CustomItemByteArrays",
		7:  "ByteArrays",
		8:  "CustomItemByteArray",
		9:  "Bytes",
		10: "ByteProperty",
		11: "None",
	})

	payload := []byte{0x78, 0x9c, 0x01, 0x02}
	byteArrayElement := bytes.NewBuffer(nil)
	testfixtures.WriteByteArrayPropertyID(byteArrayElement, 9, 2, 10, payload)
	testfixtures.WriteNameID(byteArrayElement, 11)

	customDataBytes := bytes.NewBuffer(nil)
	testfixtures.WriteStructArrayPropertyID(customDataBytes, 7, 2, 3, 8, [][]byte{byteArrayElement.Bytes()})
	testfixtures.WriteNameID(customDataBytes, 11)

	customItemData := bytes.NewBuffer(nil)
	testfixtures.WriteStructPropertyID(customItemData, 5, 3, 6, customDataBytes.Bytes())
	testfixtures.WriteNameID(customItemData, 11)

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteStructArrayPropertyID(stream, 1, 2, 3, 4, [][]byte{customItemData.Bytes()})
	testfixtures.WriteNameID(stream, 11)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeArray {
		t.Fatalf("ParseAll() = %#v, want CustomItemDatas array", props)
	}
	customDatas, ok := props[0].Value.(Array)
	if !ok || customDatas.StructType != "CustomItemData" || len(customDatas.Values) != 1 {
		t.Fatalf("CustomItemDatas = %#v, want one CustomItemData", props[0].Value)
	}
	customData, ok := customDatas.Values[0].(Container)
	if !ok {
		t.Fatalf("CustomItemData value type = %T, want Container", customDatas.Values[0])
	}
	customDataBytesValue, ok := customData.Value("CustomDataBytes")
	if !ok {
		t.Fatalf("CustomDataBytes missing from %#v", customData)
	}
	customDataBytesContainer, ok := customDataBytesValue.(Container)
	if !ok {
		t.Fatalf("CustomDataBytes value type = %T, want Container", customDataBytesValue)
	}
	byteArraysValue, ok := customDataBytesContainer.Value("ByteArrays")
	if !ok {
		t.Fatalf("ByteArrays missing from %#v", customDataBytesContainer)
	}
	byteArrays, ok := byteArraysValue.(Array)
	if !ok || byteArrays.StructType != "CustomItemByteArray" || len(byteArrays.Values) != 1 {
		t.Fatalf("ByteArrays = %#v, want one CustomItemByteArray", byteArraysValue)
	}
	byteArray, ok := byteArrays.Values[0].(Container)
	if !ok {
		t.Fatalf("CustomItemByteArray value type = %T, want Container", byteArrays.Values[0])
	}
	bytesValue, ok := byteArray.Value("Bytes")
	if !ok {
		t.Fatalf("Bytes missing from %#v", byteArray)
	}
	bytesArray, ok := bytesValue.(Array)
	if !ok || bytesArray.ElementType != TypeByte || len(bytesArray.Values) != len(payload) {
		t.Fatalf("Bytes = %#v, want byte array length %d", bytesValue, len(payload))
	}
	for i, want := range payload {
		if bytesArray.Values[i] != want {
			t.Fatalf("Bytes[%d] = %#v, want %#v", i, bytesArray.Values[i], want)
		}
	}
}

func TestParseArrayPropertyReadsCustomItemDataAfterTrailingPadding(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1:  "CustomItemDatas",
		2:  "ArrayProperty",
		3:  "StructProperty",
		4:  "CustomItemData",
		5:  "CustomDataBytes",
		6:  "CustomItemByteArrays",
		7:  "ByteArrays",
		8:  "CustomItemByteArray",
		9:  "Bytes",
		10: "ByteProperty",
		11: "None",
	})

	first := customItemDataElementWithPayload([]byte{0x01})
	first.Write(make([]byte, 4))
	testfixtures.WriteNameID(first, 11)
	second := customItemDataElementWithPayload([]byte{0x02, 0x03})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteStructArrayPropertyID(stream, 1, 2, 3, 4, [][]byte{first.Bytes(), second.Bytes()})
	testfixtures.WriteNameID(stream, 11)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	customDatas := props[0].Value.(Array)
	if len(customDatas.Values) != 2 {
		t.Fatalf("CustomItemDatas length = %d, want 2", len(customDatas.Values))
	}
	secondContainer := customDatas.Values[1].(Container)
	secondPayload := customItemDataPayload(t, secondContainer)
	if !bytes.Equal(secondPayload, []byte{0x02, 0x03}) {
		t.Fatalf("second CustomItemData payload = % x, want 02 03", secondPayload)
	}
}

func TestParseStructPropertyReadsPackedVector(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "SavedBaseWorldLocation",
		2: "StructProperty",
		3: "Vector",
		4: "/Script/CoreUObject",
		5: "None",
	})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, 24)
	stream.WriteByte(8)
	testfixtures.WriteFloat64(stream, 11)
	testfixtures.WriteFloat64(stream, 22)
	testfixtures.WriteFloat64(stream, 33)
	testfixtures.WriteNameID(stream, 5)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeStruct {
		t.Fatalf("ParseAll() = %#v, want one struct property", props)
	}
	got, ok := props[0].Value.(Vector)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want Vector", props[0].Value)
	}
	if got.X != 11 || got.Y != 22 || got.Z != 33 {
		t.Fatalf("Vector = %#v, want 11/22/33", got)
	}
}

func TestParseStructPropertyReadsPackedRotator(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Rotation",
		2: "StructProperty",
		3: "Rotator",
		4: "None",
	})
	body := bytes.NewBuffer(nil)
	testfixtures.WriteFloat64(body, 10.5)
	testfixtures.WriteFloat64(body, 20.5)
	testfixtures.WriteFloat64(body, 30.5)

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteStructPropertyID(stream, 1, 2, 3, body.Bytes())
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	got, ok := props[0].Value.(Rotator)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want Rotator", props[0].Value)
	}
	if got.Pitch != 10.5 || got.Roll != 20.5 || got.Yaw != 30.5 {
		t.Fatalf("Rotator = %#v, want pitch/roll/yaw 10.5/20.5/30.5", got)
	}
}

func TestParseStructPropertyReadsPackedQuat(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Orientation",
		2: "StructProperty",
		3: "Quat",
		4: "None",
	})
	body := bytes.NewBuffer(nil)
	testfixtures.WriteFloat64(body, 1.25)
	testfixtures.WriteFloat64(body, 2.25)
	testfixtures.WriteFloat64(body, 3.25)
	testfixtures.WriteFloat64(body, 4.25)

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteStructPropertyID(stream, 1, 2, 3, body.Bytes())
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	got, ok := props[0].Value.(Quat)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want Quat", props[0].Value)
	}
	if got.X != 1.25 || got.Y != 2.25 || got.Z != 3.25 || got.W != 4.25 {
		t.Fatalf("Quat = %#v, want 1.25/2.25/3.25/4.25", got)
	}
}

func TestParseStructPropertyReadsPackedColor(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Tint",
		2: "StructProperty",
		3: "Color",
		4: "None",
	})
	body := []byte{10, 20, 30, 40}

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteStructPropertyID(stream, 1, 2, 3, body)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	got, ok := props[0].Value.(Color)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want Color", props[0].Value)
	}
	if got != (Color{R: 10, G: 20, B: 30, A: 40}) {
		t.Fatalf("Color = %#v, want 10/20/30/40", got)
	}
}

func TestParseStructPropertyReadsPackedLinearColor(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "LinearTint",
		2: "StructProperty",
		3: "LinearColor",
		4: "None",
	})
	body := bytes.NewBuffer(nil)
	testfixtures.WriteFloat32(body, 0.1)
	testfixtures.WriteFloat32(body, 0.2)
	testfixtures.WriteFloat32(body, 0.3)
	testfixtures.WriteFloat32(body, 0.4)

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteStructPropertyID(stream, 1, 2, 3, body.Bytes())
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	got, ok := props[0].Value.(LinearColor)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want LinearColor", props[0].Value)
	}
	if got.R != float32(0.1) || got.G != float32(0.2) || got.B != float32(0.3) || got.A != float32(0.4) {
		t.Fatalf("LinearColor = %#v, want 0.1/0.2/0.3/0.4", got)
	}
}

func TestParseStructPropertyReadsUniqueNetIDAndContinues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "UniqueID",
		2: "StructProperty",
		3: "UniqueNetIdRepl",
		4: "PlayerName",
		5: "StrProperty",
		6: "None",
	})
	var payload bytes.Buffer
	payload.WriteByte(0xf9)
	testfixtures.WriteArkString(&payload, "RedpointEOS")
	payload.WriteByte(16)
	payload.Write([]byte{0x00, 0x02, 0x38, 0xca, 0xa8, 0xe4, 0x4e, 0xf7, 0x9b, 0x67, 0x9a, 0x07, 0x24, 0x32, 0x08, 0x28})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(payload.Len()))
	stream.WriteByte(8)
	stream.Write(payload.Bytes())
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteNameID(stream, 5)
	var stringPayload bytes.Buffer
	testfixtures.WriteArkString(&stringPayload, "PlatformName")
	testfixtures.WriteInt32(stream, int32(stringPayload.Len()))
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	stream.Write(stringPayload.Bytes())
	testfixtures.WriteNameID(stream, 6)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() length = %d, want 2: %#v", len(props), props)
	}
	got, ok := props[0].Value.(UniqueNetID)
	if !ok {
		t.Fatalf("UniqueID value type = %T, want UniqueNetID", props[0].Value)
	}
	if got.Unknown != 0xf9 || got.ValueType != "RedpointEOS" || got.Value != "000238caa8e44ef79b679a0724320828" {
		t.Fatalf("UniqueNetID = %#v", got)
	}
	if props[1].Name != "PlayerName" || props[1].Value != "PlatformName" {
		t.Fatalf("following property = %#v", props[1])
	}
}

func TestParseStructPropertyReadsNestedPropertyContainer(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "TribeData",
		2: "StructProperty",
		3: "TribeDataStruct",
		4: "TribeName",
		5: "StrProperty",
		6: "TribeID",
		7: "IntProperty",
		8: "None",
	})

	var body bytes.Buffer
	testfixtures.WriteNameID(&body, 4)
	testfixtures.WriteNameID(&body, 5)
	testfixtures.WriteInt32(&body, 12)
	testfixtures.WriteInt32(&body, 0)
	body.WriteByte(0)
	testfixtures.WriteArkString(&body, "Porters")
	testfixtures.WriteNameID(&body, 6)
	testfixtures.WriteNameID(&body, 7)
	testfixtures.WriteInt32(&body, 4)
	testfixtures.WriteInt32(&body, 0)
	body.WriteByte(0)
	testfixtures.WriteInt32(&body, 12345)
	testfixtures.WriteNameID(&body, 8)

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(body.Len()))
	stream.WriteByte(0)
	stream.Write(body.Bytes())
	testfixtures.WriteNameID(stream, 8)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 {
		t.Fatalf("ParseAll() length = %d, want 1", len(props))
	}
	if props[0].Type != TypeStruct {
		t.Fatalf("property type = %s, want Struct", props[0].Type)
	}
	container, ok := props[0].Value.(Container)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want Container", props[0].Value)
	}
	name, ok := container.Value("TribeName")
	if !ok || name != "Porters" {
		t.Fatalf("TribeName = %#v, %v; want Porters, true", name, ok)
	}
	tribeID, ok := container.Value("TribeID")
	if !ok || tribeID != int32(12345) {
		t.Fatalf("TribeID = %#v, %v; want 12345, true", tribeID, ok)
	}
}

func TestParseStructPropertyRejectsDeclaredSizeOverread(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "TribeData",
		2: "StructProperty",
		3: "TribeDataStruct",
		4: "Count",
		5: "IntProperty",
		6: "None",
	})

	var body bytes.Buffer
	testfixtures.WriteNameID(&body, 4)
	testfixtures.WriteNameID(&body, 5)
	testfixtures.WriteInt32(&body, 4)
	testfixtures.WriteInt32(&body, 0)
	body.WriteByte(0)
	testfixtures.WriteInt32(&body, 42)
	testfixtures.WriteNameID(&body, 6)

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(body.Len()-1))
	stream.WriteByte(0)
	stream.Write(body.Bytes())

	_, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err == nil {
		t.Fatalf("ParseAll() error = nil, want declared size overread error")
	}
}

func TestParseStructPropertyKeepsPartialContainerBeforeNestedError(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "MyData",
		2: "StructProperty",
		3: "PlayerDataStruct",
		4: "PlayerDataID",
		5: "IntProperty",
		6: "Owner",
		7: "ObjectProperty",
		8: "None",
	})

	var body bytes.Buffer
	testfixtures.WriteNameID(&body, 4)
	testfixtures.WriteNameID(&body, 5)
	testfixtures.WriteInt32(&body, 4)
	testfixtures.WriteInt32(&body, 0)
	body.WriteByte(0)
	testfixtures.WriteInt32(&body, 42)
	testfixtures.WriteNameID(&body, 6)
	testfixtures.WriteNameID(&body, 7)
	testfixtures.WriteInt32(&body, 2)
	testfixtures.WriteInt32(&body, 0)
	body.WriteByte(0)
	_ = binary.Write(&body, binary.LittleEndian, int16(5))
	testfixtures.WriteNameID(&body, 8)

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(body.Len()))
	stream.WriteByte(0)
	stream.Write(body.Bytes())
	testfixtures.WriteNameID(stream, 8)

	props, err := ParseAllPartial(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err == nil {
		t.Fatalf("ParseAllPartial() error = nil, want nested parse error")
	}
	container, ok := props[0].Value.(Container)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want Container", props[0].Value)
	}
	value, ok := container.Value("PlayerDataID")
	if !ok || value != int32(42) {
		t.Fatalf("PlayerDataID = %#v, %v; want 42, true", value, ok)
	}
}

func TestParseStructPropertyFallsBackToRawUnknownStruct(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "CustomPacked",
		2: "StructProperty",
		3: "MysteryStruct",
		4: "None",
	})
	payload := []byte{0xde, 0xad, 0xbe, 0xef, 0x01, 0x02}

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(len(payload)))
	stream.WriteByte(0)
	stream.Write(payload)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeStruct {
		t.Fatalf("ParseAll() = %#v, want one struct property", props)
	}
	got, ok := props[0].Value.(UnknownStruct)
	if !ok {
		t.Fatalf("StructProperty value type = %T, want UnknownStruct", props[0].Value)
	}
	if got.TypeName != "MysteryStruct" {
		t.Fatalf("UnknownStruct.TypeName = %q, want MysteryStruct", got.TypeName)
	}
	if !bytes.Equal(got.Raw, payload) {
		t.Fatalf("UnknownStruct.Raw = % x, want % x", got.Raw, payload)
	}
}

func TestParseMapPropertyReadsSimpleKeyValueEntries(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "LabelsByLevel",
		2: "MapProperty",
		3: "IntProperty",
		4: "StrProperty",
		5: "None",
	})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteUInt32(stream, 32)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, 2)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteArkString(stream, "one")
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteArkString(stream, "two")
	testfixtures.WriteNameID(stream, 5)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeMap {
		t.Fatalf("ParseAll() = %#v, want one map property", props)
	}
	got, ok := props[0].Value.(Map)
	if !ok {
		t.Fatalf("MapProperty value type = %T, want Map", props[0].Value)
	}
	if got.KeyType != TypeInt || got.ValueType != TypeString || len(got.Entries) != 2 {
		t.Fatalf("Map = %#v, want Int->String length 2", got)
	}
	if got.Entries[0].Key != int32(1) || got.Entries[0].Value != "one" {
		t.Fatalf("first map entry = %#v", got.Entries[0])
	}
	if got.Entries[1].Key != int32(2) || got.Entries[1].Value != "two" {
		t.Fatalf("second map entry = %#v", got.Entries[1])
	}
}

func TestParseMapPropertyReadsStructValues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "EntriesByLevel",
		2: "MapProperty",
		3: "IntProperty",
		4: "StructProperty",
		5: "EntryStruct",
		6: "Label",
		7: "StrProperty",
		8: "None",
	})

	var entry bytes.Buffer
	testfixtures.WriteNameID(&entry, 6)
	testfixtures.WriteNameID(&entry, 7)
	testfixtures.WriteInt32(&entry, 10)
	testfixtures.WriteInt32(&entry, 0)
	entry.WriteByte(0)
	testfixtures.WriteArkString(&entry, "alpha")
	testfixtures.WriteNameID(&entry, 8)

	bodySize := 8 + 4 + entry.Len()
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(bodySize))
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteInt32(stream, 7)
	stream.Write(entry.Bytes())
	testfixtures.WriteNameID(stream, 8)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeMap {
		t.Fatalf("ParseAll() = %#v, want one map property", props)
	}
	got, ok := props[0].Value.(Map)
	if !ok {
		t.Fatalf("MapProperty value type = %T, want Map", props[0].Value)
	}
	if got.KeyType != TypeInt || got.ValueType != TypeStruct || len(got.Entries) != 1 {
		t.Fatalf("Map = %#v, want Int->Struct length 1", got)
	}
	if got.Entries[0].Key != int32(7) {
		t.Fatalf("map key = %#v, want 7", got.Entries[0].Key)
	}
	container, ok := got.Entries[0].Value.(Container)
	if !ok {
		t.Fatalf("map value type = %T, want Container", got.Entries[0].Value)
	}
	value, ok := container.Value("Label")
	if !ok || value != "alpha" {
		t.Fatalf("Label = %#v, %v; want alpha, true", value, ok)
	}
}

func TestParseMapPropertyReadsEnumKeyedStructValues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1:  "EntriesByKind",
		2:  "MapProperty",
		3:  "ByteProperty",
		4:  "EEntryKind",
		5:  "/Script/ShooterGame.EEntryKind",
		6:  "StructProperty",
		7:  "EntryStruct",
		8:  "EEntryKind::Alpha",
		9:  "Label",
		10: "StrProperty",
		11: "None",
	})

	var entry bytes.Buffer
	testfixtures.WriteNameID(&entry, 9)
	testfixtures.WriteNameID(&entry, 10)
	testfixtures.WriteInt32(&entry, 10)
	testfixtures.WriteInt32(&entry, 0)
	entry.WriteByte(0)
	testfixtures.WriteArkString(&entry, "alpha")
	testfixtures.WriteNameID(&entry, 11)

	bodySize := 4 + 4 + 8 + entry.Len()
	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteNameID(stream, 6)
	testfixtures.WriteInt32(stream, 1)
	testfixtures.WriteNameID(stream, 7)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 7)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(bodySize))
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, 1)
	testfixtures.WriteNameID(stream, 8)
	stream.Write(entry.Bytes())
	testfixtures.WriteNameID(stream, 11)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeMap {
		t.Fatalf("ParseAll() = %#v, want one map property", props)
	}
	got, ok := props[0].Value.(Map)
	if !ok {
		t.Fatalf("MapProperty value type = %T, want Map", props[0].Value)
	}
	if got.KeyType != TypeEnum || got.ValueType != TypeStruct || len(got.Entries) != 1 {
		t.Fatalf("Map = %#v, want Enum->Struct length 1", got)
	}
	key, ok := got.Entries[0].Key.(EnumValue)
	if !ok || key.Name != "EEntryKind::Alpha" {
		t.Fatalf("map key = %#v, want enum Alpha", got.Entries[0].Key)
	}
	container, ok := got.Entries[0].Value.(Container)
	if !ok {
		t.Fatalf("map value type = %T, want Container", got.Entries[0].Value)
	}
	value, ok := container.Value("Label")
	if !ok || value != "alpha" {
		t.Fatalf("Label = %#v, %v; want alpha, true", value, ok)
	}
}

func TestParseMapPropertySkipsStructKeyedMapAndContinues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "EntriesByStruct",
		2: "MapProperty",
		3: "StructProperty",
		4: "IntProperty",
		5: "AfterMap",
		6: "None",
	})
	body := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteUInt32(stream, uint32(len(body)))
	stream.WriteByte(0)
	stream.Write(body)
	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 42)
	testfixtures.WriteNameID(stream, 6)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() returned %d properties, want skipped map placeholder plus trailing int: %#v", len(props), props)
	}
	got, ok := props[0].Value.(Map)
	if !ok || got.KeyType != TypeStruct || got.ValueType != TypeInt || len(got.Entries) != 0 {
		t.Fatalf("struct-keyed map = %#v, want empty Struct->Int placeholder", props[0].Value)
	}
	if props[1].Name != "AfterMap" || props[1].Type != TypeInt || props[1].Value != int32(42) {
		t.Fatalf("trailing property = %#v, want AfterMap int 42", props[1])
	}
}

func TestParseMapPropertySkipsUnsupportedValueTypeAndContinues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "EntriesByLevel",
		2: "MapProperty",
		3: "IntProperty",
		4: "MysteryProperty",
		5: "AfterMap",
		6: "None",
	})
	body := []byte{0, 0, 0, 0, 1, 0, 0, 0, 7, 0, 0, 0, 0xdd, 0xcc, 0xbb, 0xaa}

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 2)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteUInt32(stream, uint32(len(body)))
	stream.WriteByte(0)
	stream.Write(body)

	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 42)
	testfixtures.WriteNameID(stream, 6)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() returned %d properties, want skipped map placeholder plus trailing int: %#v", len(props), props)
	}
	got, ok := props[0].Value.(Map)
	if !ok || got.KeyType != TypeInt || got.ValueType != Type("MysteryProperty") || len(got.Entries) != 0 {
		t.Fatalf("unsupported-value map = %#v, want empty Int->MysteryProperty placeholder", props[0].Value)
	}
	if props[1].Name != "AfterMap" || props[1].Type != TypeInt || props[1].Value != int32(42) {
		t.Fatalf("trailing property = %#v, want AfterMap int 42", props[1])
	}
}

func TestParseSetPropertyReadsSimpleValues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Unlocked",
		2: "SetProperty",
		3: "IntProperty",
		4: "None",
	})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 3)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 20)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 3)
	testfixtures.WriteInt32(stream, 10)
	testfixtures.WriteInt32(stream, 20)
	testfixtures.WriteInt32(stream, 30)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeSet {
		t.Fatalf("ParseAll() = %#v, want one set property", props)
	}
	got, ok := props[0].Value.(Set)
	if !ok {
		t.Fatalf("SetProperty value type = %T, want Set", props[0].Value)
	}
	if got.ElementType != TypeInt || len(got.Values) != 3 {
		t.Fatalf("Set = %#v, want Int set length 3", got)
	}
	for i, want := range []any{int32(10), int32(20), int32(30)} {
		if got.Values[i] != want {
			t.Fatalf("set value %d = %#v, want %#v", i, got.Values[i], want)
		}
	}
}

func TestParseSetPropertyUsesSecondSerializedCount(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Unlocked",
		2: "SetProperty",
		3: "IntProperty",
		4: "None",
	})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 20)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 99)
	testfixtures.WriteInt32(stream, 3)
	testfixtures.WriteInt32(stream, 10)
	testfixtures.WriteInt32(stream, 20)
	testfixtures.WriteInt32(stream, 30)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 1 || props[0].Type != TypeSet {
		t.Fatalf("ParseAll() = %#v, want one set property", props)
	}
	got, ok := props[0].Value.(Set)
	if !ok {
		t.Fatalf("SetProperty value type = %T, want Set", props[0].Value)
	}
	if got.ElementType != TypeInt || len(got.Values) != 3 {
		t.Fatalf("Set = %#v, want Int set length 3", got)
	}
	for i, want := range []any{int32(10), int32(20), int32(30)} {
		if got.Values[i] != want {
			t.Fatalf("set value %d = %#v, want %#v", i, got.Values[i], want)
		}
	}
}

func TestParseSetPropertyReadsObjectValuesAsUUIDs(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "LinkedObjects",
		2: "SetProperty",
		3: "ObjectProperty",
		4: "None",
	})
	first := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	second := []byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 40)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 0)
	testfixtures.WriteInt32(stream, 2)
	stream.Write(first)
	stream.Write(second)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	got, ok := props[0].Value.(Set)
	if !ok || len(got.Values) != 2 {
		t.Fatalf("Set value = %#v, want two object references", props[0].Value)
	}
	if got.Values[0] != "00112233-4455-6677-8899-aabbccddeeff" || got.Values[1] != "ffeeddcc-bbaa-9988-7766-554433221100" {
		t.Fatalf("Set object values = %#v", got.Values)
	}
}

func TestParseSetPropertyReadsCompactObjectValues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "LinkedObjects",
		2: "SetProperty",
		3: "ObjectProperty",
		4: "None",
	})
	first := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0xff}
	second := []byte{0xff, 0xee, 0xdd, 0xcc, 0xbb, 0xaa, 0x99, 0x88, 0x77, 0x66, 0x55, 0x44, 0x33, 0x22, 0x11, 0x00}

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 5)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 2)
	stream.Write(first)
	stream.Write(second)
	testfixtures.WriteNameID(stream, 4)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	got, ok := props[0].Value.(Set)
	if !ok || len(got.Values) != 2 {
		t.Fatalf("Set value = %#v, want two object references", props[0].Value)
	}
	if got.Values[0] != "00112233-4455-6677-8899-aabbccddeeff" || got.Values[1] != "ffeeddcc-bbaa-9988-7766-554433221100" {
		t.Fatalf("Set object values = %#v", got.Values)
	}
}

func TestParseSetPropertySkipsUnsupportedElementTypeAndContinues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "UnsupportedSet",
		2: "SetProperty",
		3: "MysteryProperty",
		4: "AfterSet",
		5: "IntProperty",
		6: "None",
	})

	stream := bytes.NewBuffer(nil)
	testfixtures.WriteNameID(stream, 1)
	testfixtures.WriteNameID(stream, 2)
	testfixtures.WriteInt32(stream, 3)
	testfixtures.WriteNameID(stream, 3)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteInt32(stream, 8)
	stream.WriteByte(0)
	testfixtures.WriteUInt32(stream, 0)
	testfixtures.WriteUInt32(stream, 0xaabbccdd)
	testfixtures.WriteUInt32(stream, 0x11223344)

	testfixtures.WriteNameID(stream, 4)
	testfixtures.WriteNameID(stream, 5)
	testfixtures.WriteInt32(stream, 4)
	testfixtures.WriteInt32(stream, 0)
	stream.WriteByte(0)
	testfixtures.WriteInt32(stream, 42)
	testfixtures.WriteNameID(stream, 6)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 2 {
		t.Fatalf("ParseAll() returned %d properties, want skipped set placeholder plus trailing int: %#v", len(props), props)
	}
	got, ok := props[0].Value.(Set)
	if !ok || got.ElementType != Type("MysteryProperty") || len(got.Values) != 0 {
		t.Fatalf("unsupported set = %#v, want empty MysteryProperty placeholder", props[0].Value)
	}
	if props[1].Name != "AfterSet" || props[1].Type != TypeInt || props[1].Value != int32(42) {
		t.Fatalf("trailing property = %#v, want AfterSet int 42", props[1])
	}
}

func customItemDataElementWithPayload(payload []byte) *bytes.Buffer {
	byteArrayElement := bytes.NewBuffer(nil)
	testfixtures.WriteByteArrayPropertyID(byteArrayElement, 9, 2, 10, payload)
	testfixtures.WriteNameID(byteArrayElement, 11)

	customDataBytes := bytes.NewBuffer(nil)
	testfixtures.WriteStructArrayPropertyID(customDataBytes, 7, 2, 3, 8, [][]byte{byteArrayElement.Bytes()})
	testfixtures.WriteNameID(customDataBytes, 11)

	customItemData := bytes.NewBuffer(nil)
	testfixtures.WriteStructPropertyID(customItemData, 5, 3, 6, customDataBytes.Bytes())
	testfixtures.WriteNameID(customItemData, 11)
	return customItemData
}

func customItemDataPayload(t *testing.T, customData Container) []byte {
	t.Helper()
	customDataBytesValue, ok := customData.Value("CustomDataBytes")
	if !ok {
		t.Fatalf("CustomDataBytes missing from %#v", customData)
	}
	customDataBytes, ok := customDataBytesValue.(Container)
	if !ok {
		t.Fatalf("CustomDataBytes type = %T, want Container", customDataBytesValue)
	}
	byteArraysValue, ok := customDataBytes.Value("ByteArrays")
	if !ok {
		t.Fatalf("ByteArrays missing from %#v", customDataBytes)
	}
	byteArrays, ok := byteArraysValue.(Array)
	if !ok || len(byteArrays.Values) != 1 {
		t.Fatalf("ByteArrays = %#v, want one byte array", byteArraysValue)
	}
	byteArray := byteArrays.Values[0].(Container)
	bytesValue, ok := byteArray.Value("Bytes")
	if !ok {
		t.Fatalf("Bytes missing from %#v", byteArray)
	}
	bytesArray := bytesValue.(Array)
	out := make([]byte, 0, len(bytesArray.Values))
	for _, value := range bytesArray.Values {
		out = append(out, value.(byte))
	}
	return out
}
