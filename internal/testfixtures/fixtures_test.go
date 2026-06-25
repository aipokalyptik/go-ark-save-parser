package testfixtures

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"testing"

	"github.com/google/uuid"
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

func TestActorTransformsWritesEntriesAndNilTerminator(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	got := ActorTransforms(ActorTransform{
		UUID:       id,
		X:          1,
		Y:          2,
		Z:          3,
		Pitch:      4,
		Roll:       5,
		Yaw:        6,
		Quaternion: 7,
	})

	var want bytes.Buffer
	want.Write(id[:])
	for _, value := range []float64{1, 2, 3, 4, 5, 6, 7} {
		_ = binary.Write(&want, binary.LittleEndian, value)
	}
	want.Write(uuid.Nil[:])

	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("ActorTransforms() = %x, want %x", got, want.Bytes())
	}
}

func TestCryopodDinoPayloadWritesSupportedCompressedArchive(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")

	payload := CryopodDinoPayload(t, dinoID, statusID, CryopodDinoPayloadOptions{Health: 6})

	if len(payload) <= 12 {
		t.Fatalf("CryopodDinoPayload() length = %d, want compressed body", len(payload))
	}
	if got := binary.LittleEndian.Uint32(payload[0:4]); got != 0x0407 {
		t.Fatalf("payload version = %#x, want 0x0407", got)
	}
	decodedSize := int(binary.LittleEndian.Uint32(payload[4:8]))
	namesOffset := int(binary.LittleEndian.Uint32(payload[8:12]))
	if decodedSize <= 0 || namesOffset <= 0 || namesOffset >= decodedSize {
		t.Fatalf("decodedSize=%d namesOffset=%d, want valid embedded archive offsets", decodedSize, namesOffset)
	}

	reader, err := zlib.NewReader(bytes.NewReader(payload[12:]))
	if err != nil {
		t.Fatalf("zlib reader: %v", err)
	}
	decoded, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatalf("zlib read: %v", err)
	}
	if len(decoded) != decodedSize {
		t.Fatalf("decoded length = %d, want %d", len(decoded), decodedSize)
	}
}

func TestMinimalEmbeddedCryopodPayloadWritesMinimalNameTable(t *testing.T) {
	dinoID := uuid.MustParse("01020304-0506-0708-090a-0b0c0d0e0102")
	statusID := uuid.MustParse("11121314-1516-1718-191a-1b1c1d1e1112")

	payload := MinimalEmbeddedCryopodPayload(t, dinoID, statusID)

	if got := binary.LittleEndian.Uint32(payload[0:4]); got != 0x0407 {
		t.Fatalf("payload version = %#x, want 0x0407", got)
	}
	decodedSize := int(binary.LittleEndian.Uint32(payload[4:8]))
	namesOffset := int(binary.LittleEndian.Uint32(payload[8:12]))
	reader, err := zlib.NewReader(bytes.NewReader(payload[12:]))
	if err != nil {
		t.Fatalf("zlib reader: %v", err)
	}
	decoded, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatalf("zlib read: %v", err)
	}
	if len(decoded) != decodedSize {
		t.Fatalf("decoded length = %d, want %d", len(decoded), decodedSize)
	}
	if got := binary.LittleEndian.Uint32(decoded[8:12]); got != 2 {
		t.Fatalf("embedded object count = %d, want 2", got)
	}

	names := decoded[namesOffset:]
	if got := binary.LittleEndian.Uint32(names[0:4]); got != 4 {
		t.Fatalf("name count = %d, want 4", got)
	}
	names = names[4:]
	for _, want := range []string{"None", "DinoID1", "IntProperty", "BaseCharacterLevel"} {
		got, rest := readFixtureArkString(t, names)
		if got != want {
			t.Fatalf("name = %q, want %q", got, want)
		}
		names = rest
	}
}

func TestCryopodSaddlePayloadWritesSupportedNoHeaderPayload(t *testing.T) {
	payload := CryopodSaddlePayload()

	if len(payload) <= 16 {
		t.Fatalf("CryopodSaddlePayload() length = %d, want properties", len(payload))
	}
	if got := binary.LittleEndian.Uint32(payload[0:4]); got != 8 {
		t.Fatalf("payload prefix[0] = %d, want 8", got)
	}
	if got := binary.LittleEndian.Uint32(payload[4:8]); got != 7 {
		t.Fatalf("payload prefix[1] = %d, want 7", got)
	}
	if !bytes.Contains(payload, []byte("ItemArchetype")) {
		t.Fatalf("CryopodSaddlePayload() missing ItemArchetype property")
	}
}

func readFixtureArkString(t *testing.T, data []byte) (string, []byte) {
	t.Helper()
	if len(data) < 4 {
		t.Fatalf("ark string length missing")
	}
	size := int(int32(binary.LittleEndian.Uint32(data[0:4])))
	if size <= 0 || len(data) < 4+size {
		t.Fatalf("ark string size = %d for %d bytes", size, len(data))
	}
	raw := data[4 : 4+size]
	if raw[len(raw)-1] != 0 {
		t.Fatalf("ark string %q missing terminator", raw)
	}
	return string(raw[:len(raw)-1]), data[4+size:]
}
