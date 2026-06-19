package arkproperty

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 4)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeInt32(stream, 250)

	writeName(stream, 3)
	writeName(stream, 4)
	writeInt32(stream, 10)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeArkString(stream, "hello")

	writeName(stream, 5)
	writeName(stream, 6)
	writeInt32(stream, 1)
	writeInt32(stream, 0)
	stream.WriteByte(1)

	writeName(stream, 8)
	writeName(stream, 9)
	writeInt32(stream, 4)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, float32(3.5))

	writeName(stream, 10)
	writeName(stream, 11)
	writeInt32(stream, 8)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, float64(8.25))

	writeName(stream, 12)
	writeName(stream, 13)
	writeInt32(stream, 4)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, uint32(99))

	writeName(stream, 14)
	writeName(stream, 15)
	writeInt32(stream, 8)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, uint64(9876543210))

	writeName(stream, 7)
	writeInt32(stream, 0)

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

func TestParsePropertyReturnsNilForNoneMarkerWithoutTrailingZeros(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{7: "None"})
	stream := bytes.NewBuffer(nil)
	writeName(stream, 7)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 18)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, int16(0))
	stream.Write(ref)
	writeName(stream, 3)

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
	writeArkString(stream, "ItemArchetype")
	writeArkString(stream, "ObjectProperty")
	var body bytes.Buffer
	writeInt32(&body, 1)
	writeArkString(&body, "BlueprintGeneratedClass /Game/Test/Item.Item_C")
	writeInt32(stream, int32(body.Len()))
	writeInt32(stream, 0)
	stream.WriteByte(0)
	stream.Write(body.Bytes())
	writeArkString(stream, "None")

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
	writeArkString(stream, "ItemArchetype")
	writeArkString(stream, "ObjectProperty")
	writeInt32(stream, 4)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeInt32(stream, 1)
	writeArkString(stream, "None")

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 20)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeName(stream, 3)
	writeName(stream, 4)
	writeUInt32(stream, 0)
	writeName(stream, 5)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 8)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeName(stream, 3)
	writeName(stream, 4)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 6)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeInt32(stream, 99)
	stream.Write([]byte{0xaa, 0xbb})

	writeName(stream, 3)
	writeName(stream, 4)
	writeInt32(stream, 7)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	writeArkString(stream, "ok")
	writeName(stream, 5)

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

func TestParseInt64PropertyReadsSignedValue(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "CustomCosmeticModSkinReplacementID",
		2: "Int64Property",
		3: "None",
	})
	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 8)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, int64(-42))
	writeName(stream, 3)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 1)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	stream.WriteByte(0xf9)

	writeName(stream, 3)
	writeName(stream, 4)
	writeInt32(stream, 2)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, int16(-32000))

	writeName(stream, 5)
	writeName(stream, 6)
	writeInt32(stream, 2)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	_ = binary.Write(stream, binary.LittleEndian, uint16(65000))

	writeName(stream, 7)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 0)
	writeInt32(stream, 0)
	stream.WriteByte(0)
	stream.WriteByte(7)
	writeName(stream, 3)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 1)
	writeName(stream, 3)
	writeInt32(stream, 1)
	writeName(stream, 4)
	writeUInt32(stream, 0)
	stream.WriteByte(1)
	writeUInt32(stream, 0)
	writeName(stream, 5)
	writeName(stream, 6)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 3)
	writeUInt32(stream, 3)
	writeInt32(stream, 0)
	writeInt32(stream, 0)
	writeInt32(stream, 12)
	stream.WriteByte(0)
	writeUInt32(stream, 3)
	writeInt32(stream, 5)
	writeInt32(stream, 8)
	writeInt32(stream, 13)
	writeName(stream, 4)

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

func TestParseArrayPropertyReadsUInt64Values(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "ItemIDs",
		2: "ArrayProperty",
		3: "UInt64Property",
		4: "None",
	})
	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 2)
	writeUInt32(stream, 3)
	writeInt32(stream, 0)
	writeInt32(stream, 0)
	writeInt32(stream, 20)
	stream.WriteByte(0)
	writeUInt32(stream, 2)
	_ = binary.Write(stream, binary.LittleEndian, uint64(1001))
	_ = binary.Write(stream, binary.LittleEndian, uint64(1002))
	writeName(stream, 4)

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
	writeName(&element, 5)
	writeName(&element, 6)
	writeInt32(&element, 4)
	writeInt32(&element, 0)
	element.WriteByte(0)
	writeInt32(&element, 88)
	writeName(&element, 7)

	dataSize := uint32(4 + element.Len())
	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, int32(dataSize))
	writeName(stream, 3)
	writeInt32(stream, 1)
	writeName(stream, 4)
	writeUInt32(stream, 1)
	writeName(stream, 4)
	writeUInt32(stream, 0)
	writeUInt32(stream, dataSize)
	stream.WriteByte(0)
	writeUInt32(stream, 1)
	stream.Write(element.Bytes())
	writeName(stream, 7)

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
	writeName(&body, 4)
	writeName(&body, 5)
	writeInt32(&body, 12)
	writeInt32(&body, 0)
	body.WriteByte(0)
	writeArkString(&body, "Porters")
	writeName(&body, 6)
	writeName(&body, 7)
	writeInt32(&body, 4)
	writeInt32(&body, 0)
	body.WriteByte(0)
	writeInt32(&body, 12345)
	writeName(&body, 8)

	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeName(stream, 2)
	writeUInt32(stream, 1)
	writeName(stream, 3)
	writeUInt32(stream, 1)
	writeName(stream, 3)
	writeUInt32(stream, 0)
	writeUInt32(stream, uint32(body.Len()))
	stream.WriteByte(0)
	stream.Write(body.Bytes())
	writeName(stream, 8)

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
	writeName(&body, 4)
	writeName(&body, 5)
	writeInt32(&body, 4)
	writeInt32(&body, 0)
	body.WriteByte(0)
	writeInt32(&body, 42)
	writeName(&body, 6)
	writeName(&body, 7)
	writeInt32(&body, 2)
	writeInt32(&body, 0)
	body.WriteByte(0)
	_ = binary.Write(&body, binary.LittleEndian, int16(5))
	writeName(&body, 8)

	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeName(stream, 2)
	writeUInt32(stream, 1)
	writeName(stream, 3)
	writeUInt32(stream, 1)
	writeName(stream, 3)
	writeUInt32(stream, 0)
	writeUInt32(stream, uint32(body.Len()))
	stream.WriteByte(0)
	stream.Write(body.Bytes())
	writeName(stream, 8)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeUInt32(stream, 1)
	writeName(stream, 3)
	writeUInt32(stream, 1)
	writeName(stream, 3)
	writeUInt32(stream, 0)
	writeUInt32(stream, uint32(len(payload)))
	stream.WriteByte(0)
	stream.Write(payload)
	writeName(stream, 4)

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
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 2)
	writeName(stream, 3)
	writeUInt32(stream, 0)
	writeName(stream, 4)
	writeInt32(stream, 0)
	writeUInt32(stream, 32)
	stream.WriteByte(0)
	writeUInt32(stream, 0)
	writeUInt32(stream, 2)
	writeInt32(stream, 1)
	writeArkString(stream, "one")
	writeInt32(stream, 2)
	writeArkString(stream, "two")
	writeName(stream, 5)

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

func TestParseSetPropertyReadsSimpleValues(t *testing.T) {
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		1: "Unlocked",
		2: "SetProperty",
		3: "IntProperty",
		4: "None",
	})

	stream := bytes.NewBuffer(nil)
	writeName(stream, 1)
	writeName(stream, 2)
	writeInt32(stream, 3)
	writeName(stream, 3)
	writeUInt32(stream, 0)
	writeInt32(stream, 16)
	stream.WriteByte(0)
	writeUInt32(stream, 0)
	writeInt32(stream, 3)
	writeInt32(stream, 10)
	writeInt32(stream, 20)
	writeInt32(stream, 30)
	writeName(stream, 4)

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

func writeName(buf *bytes.Buffer, id uint32) {
	_ = binary.Write(buf, binary.LittleEndian, id)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
}

func writeInt32(buf *bytes.Buffer, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeUInt32(buf *bytes.Buffer, value uint32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}
