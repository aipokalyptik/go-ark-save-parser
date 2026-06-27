package testfixtures

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
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

func TestTruncatedObjectWithPropertiesBytesTrimsWrappedObject(t *testing.T) {
	var props bytes.Buffer
	WriteIntPropertyID(&props, 0x10000002, 0x10000003, 250)
	full := ObjectBytesWithProperties(0x10000001, 0x10000004, props.Bytes())

	got := TruncatedObjectWithPropertiesBytes(0x10000001, 0x10000004, props.Bytes(), 10)

	if !bytes.Equal(got, full[:len(full)-10]) {
		t.Fatalf("TruncatedObjectWithPropertiesBytes() = %x, want %x", got, full[:len(full)-10])
	}

	untruncated := TruncatedObjectWithPropertiesBytes(0x10000001, 0x10000004, props.Bytes(), 0)
	if !bytes.Equal(untruncated, full) {
		t.Fatalf("TruncatedObjectWithPropertiesBytes(zero) = %x, want %x", untruncated, full)
	}

	empty := TruncatedObjectWithPropertiesBytes(0x10000001, 0x10000004, props.Bytes(), len(full))
	if empty == nil || len(empty) != 0 {
		t.Fatalf("TruncatedObjectWithPropertiesBytes(full) length = %d nil=%v, want non-nil empty slice", len(empty), empty == nil)
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

func TestPlayerPawnGameObjectBytesWritesParseablePawnObject(t *testing.T) {
	inventoryID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	pawnID := uuid.MustParse("11112222-3333-4444-5555-666677778888")

	object, err := arkobject.ParseGameObject(pawnID, PlayerPawnGameObjectBytes(42, inventoryID), nil, nil)
	if err != nil {
		t.Fatalf("ParseGameObject(player pawn) error = %v", err)
	}
	if object.Blueprint != "Blueprint'/Game/PrimalEarth/CoreBlueprints/PlayerPawnTest.PlayerPawnTest_C'" {
		t.Fatalf("Blueprint = %q, want player pawn test blueprint", object.Blueprint)
	}
	if got, ok := object.Value("LinkedPlayerDataID"); !ok || got != int32(42) {
		t.Fatalf("LinkedPlayerDataID = %#v, %v; want 42, true", got, ok)
	}
	rawInventory, ok := object.Value("MyInventoryComponent")
	if !ok {
		t.Fatalf("MyInventoryComponent missing")
	}
	inventoryRef, ok := rawInventory.(arkproperty.ObjectReference)
	if !ok {
		t.Fatalf("MyInventoryComponent type = %T, want ObjectReference", rawInventory)
	}
	if inventoryRef.Type != arkproperty.ObjectReferencePath || inventoryRef.Value != inventoryID.String() {
		t.Fatalf("MyInventoryComponent = %#v, want object-path reference %s", inventoryRef, inventoryID)
	}
	if _, ok := object.Value("SavedBaseWorldLocation"); !ok {
		t.Fatalf("SavedBaseWorldLocation missing")
	}
}

func TestInventoryGameObjectBytesWritesParseableInventoryObject(t *testing.T) {
	inventoryID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	firstItemID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	secondItemID := uuid.MustParse("99992222-3333-4444-5555-666677778888")

	object, err := arkobject.ParseGameObject(inventoryID, InventoryGameObjectBytes(inventoryID, firstItemID, secondItemID), nil, nil)
	if err != nil {
		t.Fatalf("ParseGameObject(inventory) error = %v", err)
	}
	inventory := arkobject.InventoryFromObject(object)
	if inventory.UUID != inventoryID {
		t.Fatalf("Inventory UUID = %s, want %s", inventory.UUID, inventoryID)
	}
	if inventory.NumberOfItems() != 2 {
		t.Fatalf("NumberOfItems() = %d, want 2; item UUIDs %#v", inventory.NumberOfItems(), inventory.ItemUUIDs)
	}
	for _, want := range []uuid.UUID{firstItemID, secondItemID} {
		if !inventoryHasItem(inventory, want) {
			t.Fatalf("inventory item UUIDs = %#v, missing %s", inventory.ItemUUIDs, want)
		}
	}
}

func inventoryHasItem(inventory arkobject.Inventory, want uuid.UUID) bool {
	for _, got := range inventory.ItemUUIDs {
		if got == want {
			return true
		}
	}
	return false
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
