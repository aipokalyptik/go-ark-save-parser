package testfixtures

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestObjectBytesWithPropertiesWrapsObjectHeaderAndNoneMarker(t *testing.T) {
	var props bytes.Buffer
	WriteIntPropertyID(&props, 0x10000002, 0x10000003, 250)

	got := ObjectBytesWithProperties(0x10000001, 0x10000004, props.Bytes())

	var want bytes.Buffer
	_ = binary.Write(&want, binary.LittleEndian, uint32(0x10000001))
	_ = binary.Write(&want, binary.LittleEndian, int32(0))
	_ = binary.Write(&want, binary.LittleEndian, int32(0))
	_ = binary.Write(&want, binary.LittleEndian, int32(0))
	_ = binary.Write(&want, binary.LittleEndian, int32(0))
	_ = binary.Write(&want, binary.LittleEndian, int16(0))
	want.Write(props.Bytes())
	_ = binary.Write(&want, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&want, binary.LittleEndian, int32(0))

	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("ObjectBytesWithProperties() = %x, want %x", got, want.Bytes())
	}
}
