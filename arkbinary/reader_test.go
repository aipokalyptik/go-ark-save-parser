package arkbinary

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	"github.com/google/uuid"
)

func TestReaderReadsPrimitiveValuesAndTracksPosition(t *testing.T) {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int32(-42))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0xdecafbad))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(0xbeef))
	_ = binary.Write(&buf, binary.LittleEndian, uint64(0x1122334455667788))
	_ = binary.Write(&buf, binary.LittleEndian, int64(-123456789))
	_ = binary.Write(&buf, binary.LittleEndian, float32(3.5))
	_ = binary.Write(&buf, binary.LittleEndian, float64(-7.25))
	buf.WriteByte(1)
	buf.WriteByte(0x7f)

	r := NewReader(buf.Bytes(), nil)
	if got := r.Position(); got != 0 {
		t.Fatalf("initial position = %d, want 0", got)
	}
	if got, err := r.ReadInt32(); err != nil || got != -42 {
		t.Fatalf("ReadInt32() = %d, %v; want -42, nil", got, err)
	}
	if got, err := r.PeekUInt32(); err != nil || got != 0xdecafbad {
		t.Fatalf("PeekUInt32() = %#x, %v; want 0xdecafbad, nil", got, err)
	}
	if got := r.Position(); got != 4 {
		t.Fatalf("position after peek = %d, want 4", got)
	}
	if got, err := r.ReadUInt32(); err != nil || got != 0xdecafbad {
		t.Fatalf("ReadUInt32() = %#x, %v; want 0xdecafbad, nil", got, err)
	}
	if got, err := r.ReadUInt16(); err != nil || got != 0xbeef {
		t.Fatalf("ReadUInt16() = %#x, %v; want 0xbeef, nil", got, err)
	}
	if got, err := r.ReadUInt64(); err != nil || got != 0x1122334455667788 {
		t.Fatalf("ReadUInt64() = %#x, %v; want 0x1122334455667788, nil", got, err)
	}
	if got, err := r.ReadInt64(); err != nil || got != -123456789 {
		t.Fatalf("ReadInt64() = %d, %v; want -123456789, nil", got, err)
	}
	if got, err := r.ReadFloat32(); err != nil || math.Abs(float64(got-3.5)) > 0.00001 {
		t.Fatalf("ReadFloat32() = %f, %v; want 3.5, nil", got, err)
	}
	if got, err := r.ReadFloat64(); err != nil || got != -7.25 {
		t.Fatalf("ReadFloat64() = %f, %v; want -7.25, nil", got, err)
	}
	if got, err := r.ReadBool(); err != nil || !got {
		t.Fatalf("ReadBool() = %v, %v; want true, nil", got, err)
	}
	if got, err := r.ReadByte(); err != nil || got != 0x7f {
		t.Fatalf("ReadByte() = %#x, %v; want 0x7f, nil", got, err)
	}
	if r.HasMore() {
		t.Fatalf("HasMore() = true, want false")
	}
}

func TestReaderRejectsOutOfBoundsReads(t *testing.T) {
	r := NewReader([]byte{1, 2, 3}, nil)
	if _, err := r.ReadUInt32(); err == nil {
		t.Fatalf("ReadUInt32() error = nil, want underflow")
	}
	if got := r.Position(); got != 0 {
		t.Fatalf("position after failed read = %d, want 0", got)
	}
	if err := r.SetPosition(4); err == nil {
		t.Fatalf("SetPosition beyond end error = nil, want error")
	}
}

func TestReaderSliceReturnsCopiedBoundsCheckedBytes(t *testing.T) {
	raw := []byte{0, 1, 2, 3, 4}
	r := NewReader(raw, nil)
	got, err := r.Slice(1, 4)
	if err != nil {
		t.Fatalf("Slice() error = %v", err)
	}
	want := []byte{1, 2, 3}
	if !bytes.Equal(got, want) {
		t.Fatalf("Slice() = %v, want %v", got, want)
	}
	got[0] = 99
	if raw[1] != 1 {
		t.Fatalf("Slice() returned backing storage; raw[1] = %d, want 1", raw[1])
	}
	if r.Position() != 0 {
		t.Fatalf("Slice() moved position to %d, want 0", r.Position())
	}
	if _, err := r.Slice(3, 6); err == nil {
		t.Fatalf("Slice() out of bounds error = nil, want error")
	}
}

func TestReaderReadsArkStrings(t *testing.T) {
	t.Run("ascii null terminated", func(t *testing.T) {
		var data bytes.Buffer
		_ = binary.Write(&data, binary.LittleEndian, int32(6))
		data.WriteString("hello")
		data.WriteByte(0)

		r := NewReader(data.Bytes(), nil)
		got, err := r.ReadString()
		if err != nil {
			t.Fatalf("ReadString() error = %v", err)
		}
		if got == nil || *got != "hello" {
			t.Fatalf("ReadString() = %v, want hello", got)
		}
	})

	t.Run("zero length returns nil", func(t *testing.T) {
		r := NewReader([]byte{0, 0, 0, 0}, nil)
		got, err := r.ReadString()
		if err != nil {
			t.Fatalf("ReadString() error = %v", err)
		}
		if got != nil {
			t.Fatalf("ReadString() = %q, want nil", *got)
		}
	})

	t.Run("utf16 negative length", func(t *testing.T) {
		var data bytes.Buffer
		_ = binary.Write(&data, binary.LittleEndian, int32(-3))
		data.Write([]byte{'h', 0, 'i', 0, 0, 0})

		r := NewReader(data.Bytes(), nil)
		got, err := r.ReadString()
		if err != nil {
			t.Fatalf("ReadString() error = %v", err)
		}
		if got == nil || *got != "hi" {
			t.Fatalf("ReadString() = %v, want hi", got)
		}
	})
}

func TestReaderReadsUUIDAsBigEndianBytes(t *testing.T) {
	want := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	r := NewReader(want[:], nil)
	got, err := r.ReadUUID()
	if err != nil {
		t.Fatalf("ReadUUID() error = %v", err)
	}
	if got != want {
		t.Fatalf("ReadUUID() = %s, want %s", got, want)
	}
}

func TestReaderReadsNamesFromContext(t *testing.T) {
	ctx := NewContext()
	ctx.SetNames(map[uint32]string{
		0x10000000: "ObjectName",
		7:          "NPCZoneVolume",
	})

	var data bytes.Buffer
	_ = binary.Write(&data, binary.LittleEndian, uint32(0x10000000))
	_ = binary.Write(&data, binary.LittleEndian, int32(0))
	_ = binary.Write(&data, binary.LittleEndian, uint32(7))
	_ = binary.Write(&data, binary.LittleEndian, int32(123))

	r := NewReader(data.Bytes(), ctx)
	got, err := r.ReadName("")
	if err != nil || got != "ObjectName" {
		t.Fatalf("ReadName() = %q, %v; want ObjectName, nil", got, err)
	}
	got, err = r.ReadName("")
	if err != nil || got != "NPCZoneVolume_0x7b" {
		t.Fatalf("ReadName() = %q, %v; want NPCZoneVolume_0x7b, nil", got, err)
	}
}

func TestReaderDoesNotSuffixBlueprintPathsContainingNPCZoneVolume(t *testing.T) {
	ctx := NewContext()
	ctx.SetNames(map[uint32]string{
		7: "Blueprint'/Game/Path/NPCZoneVolume.NPCZoneVolume_C'",
	})
	var data bytes.Buffer
	_ = binary.Write(&data, binary.LittleEndian, uint32(7))
	_ = binary.Write(&data, binary.LittleEndian, int32(0))

	got, err := NewReader(data.Bytes(), ctx).ReadName("")
	if err != nil {
		t.Fatalf("ReadName() error = %v", err)
	}
	if got != "Blueprint'/Game/Path/NPCZoneVolume.NPCZoneVolume_C'" {
		t.Fatalf("ReadName() = %q, want unsuffixed blueprint path", got)
	}
}

func TestInflateZlibData(t *testing.T) {
	got, err := InflateZlib([]byte{0x78, 0x9c, 0xcb, 0x48, 0xcd, 0xc9, 0xc9, 0x07, 0x00, 0x06, 0x2c, 0x02, 0x15})
	if err != nil {
		t.Fatalf("InflateZlib() error = %v", err)
	}
	if string(got) != "hello" {
		t.Fatalf("InflateZlib() = %q, want hello", got)
	}
}

func TestWildcardDecompressMatchesUpstreamRules(t *testing.T) {
	got, err := WildcardDecompress([]byte{
		0x01,
		0xf0, 0x02,
		0xf1, 0x34,
		0xf3,
		0xff, 0xaa, 0xbb,
	})
	if err != nil {
		t.Fatalf("WildcardDecompress() error = %v", err)
	}
	want := []byte{
		0x01,
		0x02,
		0xf3, 0xf4,
		0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0xaa, 0x00, 0x00, 0x00, 0xbb, 0x00, 0x00, 0x00,
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("WildcardDecompress() = % x, want % x", got, want)
	}
}
