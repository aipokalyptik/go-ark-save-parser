package testfixtures

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"io"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkbinary"
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

func TestObjectBytesWithIntPropertyWrapsSingleIntProperty(t *testing.T) {
	var props bytes.Buffer
	WriteIntPropertyID(&props, 0x10000002, 0x10000003, 250)
	want := ObjectBytesWithProperties(0x10000001, 0x10000004, props.Bytes())

	got := ObjectBytesWithIntProperty(0x10000001, 0x10000004, 0x10000002, 0x10000003, 250)
	if !bytes.Equal(got, want) {
		t.Fatalf("ObjectBytesWithIntProperty() = %x, want %x", got, want)
	}
}

func TestObjectBytesWithNamePayloadWrapsCustomObjectNames(t *testing.T) {
	var names bytes.Buffer
	WriteNameID(&names, 0x10000002)
	WriteInt32(&names, 0)
	var props bytes.Buffer
	WriteIntPropertyID(&props, 0x10000003, 0x10000004, 250)

	got := ObjectBytesWithNamePayload(0x10000001, names.Bytes(), 9, props.Bytes(), 0x10000005)

	var want bytes.Buffer
	_ = binary.Write(&want, binary.LittleEndian, uint32(0x10000001))
	_ = binary.Write(&want, binary.LittleEndian, int32(0))
	_ = binary.Write(&want, binary.LittleEndian, uint32(0))
	_ = binary.Write(&want, binary.LittleEndian, int32(1))
	want.Write(names.Bytes())
	_ = binary.Write(&want, binary.LittleEndian, int32(0))
	_ = binary.Write(&want, binary.LittleEndian, int16(9))
	want.Write(props.Bytes())
	_ = binary.Write(&want, binary.LittleEndian, uint32(0x10000005))
	_ = binary.Write(&want, binary.LittleEndian, int32(0))

	if !bytes.Equal(got, want.Bytes()) {
		t.Fatalf("ObjectBytesWithNamePayload() = %x, want %x", got, want.Bytes())
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

func TestWriteCustomItemDatasPropertyIDWritesParseablePayloadBytes(t *testing.T) {
	payload := []byte{7, 8, 9}
	var props bytes.Buffer
	WriteCustomItemDatasPropertyID(&props, payload)

	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		0x10000004: "None",
		0x1000001e: "ArrayProperty",
		0x10000047: "Blueprint'/Game/Extinction/CoreBlueprints/Weapons/PrimalItem_WeaponEmptyCryopod.PrimalItem_WeaponEmptyCryopod_C'",
		0x10000048: "CustomItemDatas",
		0x10000049: "StructProperty",
		0x1000004a: "CustomItemData",
		0x1000004b: "CustomDataBytes",
		0x1000004c: "CustomItemByteArrays",
		0x1000004d: "ByteArrays",
		0x1000004e: "CustomItemByteArray",
		0x1000004f: "Bytes",
		0x10000050: "ByteProperty",
	})
	object, err := arkobject.ParseGameObject(uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"), ObjectBytesWithProperties(0x10000047, 0x10000004, props.Bytes()), ctx, nil)
	if err != nil {
		t.Fatalf("ParseGameObject() error = %v", err)
	}
	payloads := arkobject.CryopodPayloadsFromObject(object)
	if len(payloads) != 1 || !bytes.Equal(payloads[0], payload) {
		t.Fatalf("CryopodPayloadsFromObject() = %#v, want written payload", payloads)
	}
}

func TestWriteVectorPropertyIDWritesParseableVectorStruct(t *testing.T) {
	var props bytes.Buffer
	WriteVectorPropertyID(&props, 0x10000048, 0x10000049, 0x1000004a, 0x1000004b, 11, 22, 33)

	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		0x10000004: "None",
		0x10000047: "Blueprint'/Game/PrimalEarth/CoreBlueprints/PlayerPawnTest.PlayerPawnTest_C'",
		0x10000048: "SavedBaseWorldLocation",
		0x10000049: "StructProperty",
		0x1000004a: "Vector",
		0x1000004b: "/Script/CoreUObject",
	})
	object, err := arkobject.ParseGameObject(uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"), ObjectBytesWithProperties(0x10000047, 0x10000004, props.Bytes()), ctx, nil)
	if err != nil {
		t.Fatalf("ParseGameObject() error = %v", err)
	}
	raw, ok := object.Value("SavedBaseWorldLocation")
	if !ok {
		t.Fatalf("SavedBaseWorldLocation missing")
	}
	got, ok := raw.(arkproperty.Vector)
	if !ok {
		t.Fatalf("SavedBaseWorldLocation type = %T, want Vector", raw)
	}
	if got.X != 11 || got.Y != 22 || got.Z != 33 {
		t.Fatalf("SavedBaseWorldLocation = %#v, want 11/22/33", got)
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

func TestStructureGameObjectBytesWritesParseableStructureObject(t *testing.T) {
	inventoryID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	objectID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	isEngram := false
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		0x10000003: "IntProperty",
		0x10000004: "None",
		0x10000005: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
		0x10000006: "StructureID",
		0x10000007: "MaxHealth",
		0x10000008: "Health",
		0x10000009: "TargetingTeam",
		0x1000000a: "FloatProperty",
		0x1000000e: "BoolProperty",
		0x10000013: "bIsEngram",
		0x1000001f: "ObjectProperty",
		0x10000023: "MyInventoryComponent",
		0x10000045: "CurrentItemCount",
		0x10000046: "MaxItemCount",
	})

	object, err := arkobject.ParseGameObject(objectID, StructureGameObjectBytes(StructureGameObjectOptions{
		StructureID:   123,
		TribeID:       555,
		MaxHealth:     10000,
		CurrentHealth: 9000,
		IsEngram:      &isEngram,
		InventoryID:   inventoryID,
		ItemCount:     12,
		MaxItemCount:  300,
	}), ctx, nil)
	if err != nil {
		t.Fatalf("ParseGameObject(structure) error = %v", err)
	}
	structure := arkobject.StructureFromObject(object, nil)
	if structure.ID != 123 || structure.Owner.TribeID != 555 {
		t.Fatalf("Structure identity/owner = %#v", structure)
	}
	if structure.MaxHealth != 10000 || structure.CurrentHealth != 9000 {
		t.Fatalf("Structure health = %#v", structure)
	}
	if structure.InventoryUUID == nil || *structure.InventoryUUID != inventoryID {
		t.Fatalf("InventoryUUID = %v, want %s", structure.InventoryUUID, inventoryID)
	}
	if structure.ItemCount != 12 || structure.MaxItemCount != 300 {
		t.Fatalf("inventory counts = current %d max %d", structure.ItemCount, structure.MaxItemCount)
	}
}

func TestEquipmentGameObjectBytesWritesParseableEquipmentObject(t *testing.T) {
	ownerInventoryID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	objectID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	isEquipped := true
	isBlueprint := false
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		0x10000003: "IntProperty",
		0x10000004: "None",
		0x1000000c: "ItemQuantity",
		0x1000000d: "bIsBlueprint",
		0x1000000e: "BoolProperty",
		0x1000000f: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
		0x1000000a: "FloatProperty",
		0x10000010: "ItemRating",
		0x10000011: "ItemQualityIndex",
		0x10000012: "SavedDurability",
		0x1000001a: "StrProperty",
		0x1000001b: "CrafterCharacterName",
		0x1000001c: "CrafterTribeName",
		0x1000001f: "ObjectProperty",
		0x10000022: "bEquippedItem",
		0x10000040: "ItemStatValues",
		0x10000041: "UInt16Property",
		0x10000044: "OwnerInventory",
	})

	object, err := arkobject.ParseGameObject(objectID, EquipmentGameObjectBytes(EquipmentGameObjectOptions{
		Quantity:             1,
		Rating:               7.5,
		Quality:              3,
		Durability:           0.75,
		IsEquipped:           &isEquipped,
		IsBlueprint:          &isBlueprint,
		OwnerInventoryID:     ownerInventoryID,
		CrafterCharacterName: "Survivor",
		CrafterTribeName:     "Porters",
		Stats: map[int32]uint16{
			2: 1000,
			3: 1234,
		},
	}), ctx, nil)
	if err != nil {
		t.Fatalf("ParseGameObject(equipment) error = %v", err)
	}
	item := arkobject.EquipmentItemFromObject(object, arkobject.EquipmentWeapon)
	if item.Quantity != 1 || item.Rating != 7.5 || item.Quality != 3 || item.CurrentDurability != 0.75 {
		t.Fatalf("Equipment fields = %#v", item)
	}
	if !item.IsEquipped || item.IsBlueprint {
		t.Fatalf("equipment flags = equipped %v blueprint %v", item.IsEquipped, item.IsBlueprint)
	}
	if item.OwnerInventory == nil || *item.OwnerInventory != ownerInventoryID {
		t.Fatalf("OwnerInventory = %v, want %s", item.OwnerInventory, ownerInventoryID)
	}
	if item.Stats.Internal[arkobject.EquipmentStatDurability] != 1000 || item.Stats.Internal[arkobject.EquipmentStatDamage] != 1234 {
		t.Fatalf("equipment stats = %#v", item.Stats.Internal)
	}
	if item.Crafter == nil || item.Crafter.CharacterName != "Survivor" || item.Crafter.TribeName != "Porters" {
		t.Fatalf("Crafter = %#v", item.Crafter)
	}
}

func TestStackableGameObjectBytesWritesParseableStackableObject(t *testing.T) {
	ownerInventoryID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	objectID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	isBlueprint := true
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		0x10000003: "IntProperty",
		0x10000004: "None",
		0x1000000b: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
		0x1000000c: "ItemQuantity",
		0x1000000d: "bIsBlueprint",
		0x1000000e: "BoolProperty",
		0x1000001f: "ObjectProperty",
		0x10000044: "OwnerInventory",
	})

	object, err := arkobject.ParseGameObject(objectID, StackableGameObjectBytes(StackableGameObjectOptions{
		Quantity:         100,
		IsBlueprint:      &isBlueprint,
		OwnerInventoryID: ownerInventoryID,
	}), ctx, nil)
	if err != nil {
		t.Fatalf("ParseGameObject(stackable) error = %v", err)
	}
	item := arkobject.StackableItemFromObject(object)
	if item.Quantity != 100 {
		t.Fatalf("Stackable quantity = %d, want 100", item.Quantity)
	}
	if item.OwnerInventory == nil || *item.OwnerInventory != ownerInventoryID {
		t.Fatalf("OwnerInventory = %v, want %s", item.OwnerInventory, ownerInventoryID)
	}
	if got, ok := object.Value("bIsBlueprint"); !ok || got != true {
		t.Fatalf("bIsBlueprint = %#v, %v; want true, true", got, ok)
	}
}

func TestDinoGameObjectBytesWritesParseableDinoObject(t *testing.T) {
	objectID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	isFemale := true
	isDead := false
	isBaby := true
	ctx := arkbinary.NewContext()
	ctx.SetNames(map[uint32]string{
		0x10000003: "IntProperty",
		0x10000004: "None",
		0x1000000e: "BoolProperty",
		0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		0x10000015: "DinoID1",
		0x10000016: "DinoID2",
		0x10000017: "bIsFemale",
		0x10000018: "TamedTimeStamp",
		0x10000019: "DoubleProperty",
		0x10000020: "bIsDead",
		0x10000021: "bIsBaby",
	})

	object, err := arkobject.ParseGameObject(objectID, DinoGameObjectBytes(DinoGameObjectOptions{
		ID1:      1001,
		ID2:      2002,
		IsFemale: &isFemale,
		IsDead:   &isDead,
		IsBaby:   &isBaby,
		Tamed:    true,
	}), ctx, nil)
	if err != nil {
		t.Fatalf("ParseGameObject(dino) error = %v", err)
	}
	dino := arkobject.DinoFromObject(object, nil)
	if dino.ID1 != 1001 || dino.ID2 != 2002 {
		t.Fatalf("Dino IDs = %d/%d, want 1001/2002", dino.ID1, dino.ID2)
	}
	if !dino.IsFemale || dino.IsDead || !dino.IsBaby || !dino.IsTamed {
		t.Fatalf("Dino flags = %#v", dino)
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
