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

	writeName(stream, 7)
	writeInt32(stream, 0)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 6 {
		t.Fatalf("ParseAll() length = %d, want 6", len(props))
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
