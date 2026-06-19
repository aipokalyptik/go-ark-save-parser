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
		1: "Health",
		2: "IntProperty",
		3: "Label",
		4: "StrProperty",
		5: "IsActive",
		6: "BoolProperty",
		7: "None",
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

	writeName(stream, 7)
	writeInt32(stream, 0)

	props, err := ParseAll(arkbinary.NewReader(stream.Bytes(), ctx), -1)
	if err != nil {
		t.Fatalf("ParseAll() error = %v", err)
	}
	if len(props) != 3 {
		t.Fatalf("ParseAll() length = %d, want 3", len(props))
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

func writeName(buf *bytes.Buffer, id uint32) {
	_ = binary.Write(buf, binary.LittleEndian, id)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
}

func writeInt32(buf *bytes.Buffer, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}
