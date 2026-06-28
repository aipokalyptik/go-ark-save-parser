package testfixtures

import (
	"bytes"
	"compress/zlib"
	"database/sql"
	"encoding/binary"
	"os"
	"sort"
	"testing"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

type SaveOptions struct {
	Header      []byte
	Objects     map[uuid.UUID][]byte
	Custom      map[string][]byte
	EmptyTables bool
}

type ActorTransform struct {
	UUID       uuid.UUID
	X          float64
	Y          float64
	Z          float64
	Pitch      float64
	Roll       float64
	Yaw        float64
	Quaternion float64
}

type CryopodDinoPayloadOptions struct {
	Health   int32
	Reversed bool
}

type StructureGameObjectOptions struct {
	ClassID          uint32
	NoneID           uint32
	IntPropertyID    uint32
	FloatPropertyID  uint32
	BoolPropertyID   uint32
	ObjectPropertyID uint32

	StructureIDNameID   uint32
	MaxHealthNameID     uint32
	CurrentHealthNameID uint32
	TribeIDNameID       uint32
	EngramNameID        uint32
	InventoryNameID     uint32
	ItemCountNameID     uint32
	MaxItemCountNameID  uint32
	StructureID         int32
	TribeID             int32
	MaxHealth           float32
	CurrentHealth       float32
	IsEngram            *bool
	InventoryID         uuid.UUID
	ItemCount           int32
	MaxItemCount        int32
}

type EquipmentGameObjectOptions struct {
	ClassID          uint32
	NoneID           uint32
	IntPropertyID    uint32
	FloatPropertyID  uint32
	BoolPropertyID   uint32
	StringPropertyID uint32
	ObjectPropertyID uint32
	UInt16PropertyID uint32

	QuantityNameID       uint32
	BlueprintNameID      uint32
	EngramNameID         uint32
	EquippedNameID       uint32
	RatingNameID         uint32
	QualityNameID        uint32
	DurabilityNameID     uint32
	StatsNameID          uint32
	CrafterNameID        uint32
	CrafterTribeNameID   uint32
	OwnerInventoryNameID uint32

	Quantity             int32
	Rating               float32
	Quality              int32
	Durability           float32
	IsBlueprint          *bool
	IsEngram             *bool
	IsEquipped           *bool
	Stats                map[int32]uint16
	CrafterCharacterName string
	CrafterTribeName     string
	OwnerInventoryID     uuid.UUID
}

type StackableGameObjectOptions struct {
	ClassID          uint32
	NoneID           uint32
	IntPropertyID    uint32
	BoolPropertyID   uint32
	ObjectPropertyID uint32

	QuantityNameID       uint32
	BlueprintNameID      uint32
	OwnerInventoryNameID uint32

	Quantity         int32
	IsBlueprint      *bool
	OwnerInventoryID uuid.UUID
}

type DinoGameObjectOptions struct {
	ClassID          uint32
	NoneID           uint32
	IntPropertyID    uint32
	BoolPropertyID   uint32
	DoublePropertyID uint32

	ID1NameID      uint32
	ID2NameID      uint32
	FemaleNameID   uint32
	TamedNameID    uint32
	DeadNameID     uint32
	BabyNameID     uint32
	ID1            int32
	ID2            int32
	IsFemale       *bool
	IsDead         *bool
	IsBaby         *bool
	Tamed          bool
	TamedTimestamp float64
}

func DinoGameObjectBytes(opts DinoGameObjectOptions) []byte {
	defaultDinoGameObjectOptions(&opts)
	var props bytes.Buffer
	WriteIntPropertyID(&props, opts.ID1NameID, opts.IntPropertyID, opts.ID1)
	WriteIntPropertyID(&props, opts.ID2NameID, opts.IntPropertyID, opts.ID2)
	if opts.IsFemale != nil {
		WriteBoolPropertyID(&props, opts.FemaleNameID, opts.BoolPropertyID, *opts.IsFemale)
	}
	if opts.IsDead != nil {
		WriteBoolPropertyID(&props, opts.DeadNameID, opts.BoolPropertyID, *opts.IsDead)
	}
	if opts.IsBaby != nil {
		WriteBoolPropertyID(&props, opts.BabyNameID, opts.BoolPropertyID, *opts.IsBaby)
	}
	if opts.Tamed {
		WriteDoublePropertyID(&props, opts.TamedNameID, opts.DoublePropertyID, opts.TamedTimestamp)
	}
	return ObjectBytesWithProperties(opts.ClassID, opts.NoneID, props.Bytes())
}

func defaultDinoGameObjectOptions(opts *DinoGameObjectOptions) {
	if opts.ClassID == 0 {
		opts.ClassID = 0x10000014
	}
	if opts.NoneID == 0 {
		opts.NoneID = 0x10000004
	}
	if opts.IntPropertyID == 0 {
		opts.IntPropertyID = 0x10000003
	}
	if opts.BoolPropertyID == 0 {
		opts.BoolPropertyID = 0x1000000e
	}
	if opts.DoublePropertyID == 0 {
		opts.DoublePropertyID = 0x10000019
	}
	if opts.ID1NameID == 0 {
		opts.ID1NameID = 0x10000015
	}
	if opts.ID2NameID == 0 {
		opts.ID2NameID = 0x10000016
	}
	if opts.FemaleNameID == 0 {
		opts.FemaleNameID = 0x10000017
	}
	if opts.TamedNameID == 0 {
		opts.TamedNameID = 0x10000018
	}
	if opts.DeadNameID == 0 {
		opts.DeadNameID = 0x10000020
	}
	if opts.BabyNameID == 0 {
		opts.BabyNameID = 0x10000021
	}
	if opts.Tamed && opts.TamedTimestamp == 0 {
		opts.TamedTimestamp = 42
	}
}

func StackableGameObjectBytes(opts StackableGameObjectOptions) []byte {
	defaultStackableGameObjectOptions(&opts)
	var props bytes.Buffer
	WriteIntPropertyID(&props, opts.QuantityNameID, opts.IntPropertyID, opts.Quantity)
	if opts.IsBlueprint != nil {
		WriteBoolPropertyID(&props, opts.BlueprintNameID, opts.BoolPropertyID, *opts.IsBlueprint)
	}
	if opts.OwnerInventoryID != uuid.Nil {
		WriteObjectReferencePropertyID(&props, opts.OwnerInventoryNameID, opts.ObjectPropertyID, opts.OwnerInventoryID)
	}
	return ObjectBytesWithProperties(opts.ClassID, opts.NoneID, props.Bytes())
}

func defaultStackableGameObjectOptions(opts *StackableGameObjectOptions) {
	if opts.ClassID == 0 {
		opts.ClassID = 0x1000000b
	}
	if opts.NoneID == 0 {
		opts.NoneID = 0x10000004
	}
	if opts.IntPropertyID == 0 {
		opts.IntPropertyID = 0x10000003
	}
	if opts.BoolPropertyID == 0 {
		opts.BoolPropertyID = 0x1000000e
	}
	if opts.ObjectPropertyID == 0 {
		opts.ObjectPropertyID = 0x1000001f
	}
	if opts.QuantityNameID == 0 {
		opts.QuantityNameID = 0x1000000c
	}
	if opts.BlueprintNameID == 0 {
		opts.BlueprintNameID = 0x1000000d
	}
	if opts.OwnerInventoryNameID == 0 {
		opts.OwnerInventoryNameID = 0x10000044
	}
}

func EquipmentGameObjectBytes(opts EquipmentGameObjectOptions) []byte {
	defaultEquipmentGameObjectOptions(&opts)
	var props bytes.Buffer
	WriteIntPropertyID(&props, opts.QuantityNameID, opts.IntPropertyID, opts.Quantity)
	if opts.Rating != 0 {
		WriteFloatPropertyID(&props, opts.RatingNameID, opts.FloatPropertyID, opts.Rating)
	}
	if opts.Quality != 0 {
		WriteIntPropertyID(&props, opts.QualityNameID, opts.IntPropertyID, opts.Quality)
	}
	if opts.Durability != 0 {
		WriteFloatPropertyID(&props, opts.DurabilityNameID, opts.FloatPropertyID, opts.Durability)
	}
	for _, position := range sortedInt32Keys(opts.Stats) {
		WritePositionedUInt16PropertyID(&props, opts.StatsNameID, opts.UInt16PropertyID, position, opts.Stats[position])
	}
	if opts.CrafterCharacterName != "" {
		WriteStringPropertyID(&props, opts.CrafterNameID, opts.StringPropertyID, opts.CrafterCharacterName)
	}
	if opts.CrafterTribeName != "" {
		WriteStringPropertyID(&props, opts.CrafterTribeNameID, opts.StringPropertyID, opts.CrafterTribeName)
	}
	if opts.IsEngram != nil {
		WriteBoolPropertyID(&props, opts.EngramNameID, opts.BoolPropertyID, *opts.IsEngram)
	}
	if opts.IsEquipped != nil {
		WriteBoolPropertyID(&props, opts.EquippedNameID, opts.BoolPropertyID, *opts.IsEquipped)
	}
	if opts.IsBlueprint != nil {
		WriteBoolPropertyID(&props, opts.BlueprintNameID, opts.BoolPropertyID, *opts.IsBlueprint)
	}
	if opts.OwnerInventoryID != uuid.Nil {
		WriteObjectReferencePropertyID(&props, opts.OwnerInventoryNameID, opts.ObjectPropertyID, opts.OwnerInventoryID)
	}
	return ObjectBytesWithProperties(opts.ClassID, opts.NoneID, props.Bytes())
}

func defaultEquipmentGameObjectOptions(opts *EquipmentGameObjectOptions) {
	if opts.ClassID == 0 {
		opts.ClassID = 0x1000000f
	}
	if opts.NoneID == 0 {
		opts.NoneID = 0x10000004
	}
	if opts.IntPropertyID == 0 {
		opts.IntPropertyID = 0x10000003
	}
	if opts.FloatPropertyID == 0 {
		opts.FloatPropertyID = 0x1000000a
	}
	if opts.BoolPropertyID == 0 {
		opts.BoolPropertyID = 0x1000000e
	}
	if opts.StringPropertyID == 0 {
		opts.StringPropertyID = 0x1000001a
	}
	if opts.ObjectPropertyID == 0 {
		opts.ObjectPropertyID = 0x1000001f
	}
	if opts.UInt16PropertyID == 0 {
		opts.UInt16PropertyID = 0x10000041
	}
	if opts.QuantityNameID == 0 {
		opts.QuantityNameID = 0x1000000c
	}
	if opts.BlueprintNameID == 0 {
		opts.BlueprintNameID = 0x1000000d
	}
	if opts.EngramNameID == 0 {
		opts.EngramNameID = 0x10000013
	}
	if opts.EquippedNameID == 0 {
		opts.EquippedNameID = 0x10000022
	}
	if opts.RatingNameID == 0 {
		opts.RatingNameID = 0x10000010
	}
	if opts.QualityNameID == 0 {
		opts.QualityNameID = 0x10000011
	}
	if opts.DurabilityNameID == 0 {
		opts.DurabilityNameID = 0x10000012
	}
	if opts.StatsNameID == 0 {
		opts.StatsNameID = 0x10000040
	}
	if opts.CrafterNameID == 0 {
		opts.CrafterNameID = 0x1000001b
	}
	if opts.CrafterTribeNameID == 0 {
		opts.CrafterTribeNameID = 0x1000001c
	}
	if opts.OwnerInventoryNameID == 0 {
		opts.OwnerInventoryNameID = 0x10000044
	}
}

func StructureGameObjectBytes(opts StructureGameObjectOptions) []byte {
	defaultStructureGameObjectOptions(&opts)
	var props bytes.Buffer
	WriteIntPropertyID(&props, opts.StructureIDNameID, opts.IntPropertyID, opts.StructureID)
	if opts.MaxHealth != 0 {
		WriteFloatPropertyID(&props, opts.MaxHealthNameID, opts.FloatPropertyID, opts.MaxHealth)
	}
	if opts.CurrentHealth != 0 {
		WriteFloatPropertyID(&props, opts.CurrentHealthNameID, opts.FloatPropertyID, opts.CurrentHealth)
	}
	WriteIntPropertyID(&props, opts.TribeIDNameID, opts.IntPropertyID, opts.TribeID)
	if opts.IsEngram != nil {
		WriteBoolPropertyID(&props, opts.EngramNameID, opts.BoolPropertyID, *opts.IsEngram)
	}
	if opts.InventoryID != uuid.Nil {
		WriteObjectReferencePropertyID(&props, opts.InventoryNameID, opts.ObjectPropertyID, opts.InventoryID)
	}
	if opts.ItemCount != 0 {
		WriteIntPropertyID(&props, opts.ItemCountNameID, opts.IntPropertyID, opts.ItemCount)
	}
	if opts.MaxItemCount != 0 {
		WriteIntPropertyID(&props, opts.MaxItemCountNameID, opts.IntPropertyID, opts.MaxItemCount)
	}
	return ObjectBytesWithProperties(opts.ClassID, opts.NoneID, props.Bytes())
}

func defaultStructureGameObjectOptions(opts *StructureGameObjectOptions) {
	if opts.ClassID == 0 {
		opts.ClassID = 0x10000005
	}
	if opts.NoneID == 0 {
		opts.NoneID = 0x10000004
	}
	if opts.IntPropertyID == 0 {
		opts.IntPropertyID = 0x10000003
	}
	if opts.FloatPropertyID == 0 {
		opts.FloatPropertyID = 0x1000000a
	}
	if opts.BoolPropertyID == 0 {
		opts.BoolPropertyID = 0x1000000e
	}
	if opts.ObjectPropertyID == 0 {
		opts.ObjectPropertyID = 0x1000001f
	}
	if opts.StructureIDNameID == 0 {
		opts.StructureIDNameID = 0x10000006
	}
	if opts.MaxHealthNameID == 0 {
		opts.MaxHealthNameID = 0x10000007
	}
	if opts.CurrentHealthNameID == 0 {
		opts.CurrentHealthNameID = 0x10000008
	}
	if opts.TribeIDNameID == 0 {
		opts.TribeIDNameID = 0x10000009
	}
	if opts.EngramNameID == 0 {
		opts.EngramNameID = 0x10000013
	}
	if opts.InventoryNameID == 0 {
		opts.InventoryNameID = 0x10000023
	}
	if opts.ItemCountNameID == 0 {
		opts.ItemCountNameID = 0x10000045
	}
	if opts.MaxItemCountNameID == 0 {
		opts.MaxItemCountNameID = 0x10000046
	}
}

func sortedInt32Keys(values map[int32]uint16) []int32 {
	keys := make([]int32, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}

func WriteSave(tb testing.TB, path string, opts SaveOptions) {
	tb.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		tb.Fatalf("open sqlite fixture: %v", err)
	}
	defer db.Close()
	MustExec(tb, db, `create table custom (key text primary key, value blob)`)
	MustExec(tb, db, `create table game (key blob primary key, value blob)`)
	header := opts.Header
	if header == nil {
		header = Header("Valguero_WP", map[uint32]string{1: "Blueprint'/Game/Test.Test_C'"})
	}
	MustExec(tb, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", header)
	for key, value := range opts.Custom {
		MustExec(tb, db, `insert into custom (key, value) values (?, ?)`, key, value)
	}
	if opts.EmptyTables {
		return
	}
	for id, raw := range opts.Objects {
		MustExec(tb, db, `insert into game (key, value) values (?, ?)`, id[:], raw)
	}
}

func WriteArchive(tb testing.TB, path string, className string) {
	tb.Helper()
	WriteArchiveWithProperties(tb, path, className, nil)
}

func WriteArchiveWithProperties(tb testing.TB, path string, className string, properties []byte) {
	tb.Helper()
	WriteArchiveWithPropertiesAndNames(tb, path, className, []string{"Object_0"}, properties)
}

func WriteArchiveWithPropertiesAndNames(tb testing.TB, path string, className string, names []string, properties []byte) {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	writeFile(tb, path, ArchiveBytesWithProperties(id, className, names, properties), "archive fixture")
}

func ArchiveBytesWithProperties(id uuid.UUID, className string, names []string, properties []byte) []byte {
	var buf bytes.Buffer
	WriteArchivePrefix(&buf, id, className, names)
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	if len(properties) > 0 {
		propertiesOffset := int32(buf.Len() - 1)
		binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
		buf.Write(properties)
	} else {
		binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(128))
	}
	return buf.Bytes()
}

func WritePlayerArchive(tb testing.TB, path string) {
	tb.Helper()
	WritePlayerArchiveWithOptions(tb, path, PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Survivor",
		PlayerName:    "PlatformName",
		TribeID:       777,
	})
}

type PlayerArchiveOptions struct {
	PlayerDataID        int32
	CharacterName       string
	PlayerName          string
	UniqueID            string
	TribeID             int32
	NumDeaths           int32
	ExtraCharacterLevel int32
	ExperiencePoints    float32
	TotalEngramPoints   int32
	UnlockedEngrams     []string
}

func WritePlayerArchiveWithOptions(tb testing.TB, path string, opts PlayerArchiveOptions) {
	tb.Helper()
	writeFile(tb, path, PlayerArchiveBytes(tb, opts), "player archive fixture")
}

func PlayerArchiveBytes(tb testing.TB, opts PlayerArchiveOptions) []byte {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	var buf bytes.Buffer
	WriteArchivePrefix(&buf, id, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C", []string{"PlayerData_0"})
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	buf.Write(playerProperties(opts))
	return buf.Bytes()
}

func PlayerGameObjectBytes(opts PlayerArchiveOptions) []byte {
	return GameObjectBytesWithNames(
		"/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C",
		[]string{"PlayerData_0"},
		playerProperties(opts),
	)
}

func PlayerPawnGameObjectBytes(playerDataID int32, inventoryID uuid.UUID) []byte {
	var props bytes.Buffer
	WriteNameIntProperty(&props, "LinkedPlayerDataID", playerDataID)
	WriteNameObjectPathProperty(&props, "MyInventoryComponent", inventoryID.String())
	writeNameVectorProperty(&props, "SavedBaseWorldLocation", 11, 22, 33)
	WriteArkString(&props, "None")
	return GameObjectBytesWithNames(
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/PlayerPawnTest.PlayerPawnTest_C'",
		[]string{"PlayerPawn_0"},
		props.Bytes(),
	)
}

func InventoryGameObjectBytes(inventoryID uuid.UUID, itemIDs ...uuid.UUID) []byte {
	values := make([]string, 0, len(itemIDs))
	for _, id := range itemIDs {
		values = append(values, id.String())
	}
	var props bytes.Buffer
	WriteNameObjectPathArrayProperty(&props, "InventoryItems", values)
	WriteArkString(&props, "None")
	return GameObjectBytesWithNames(
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Inventories/PrimalInventoryTest.PrimalInventoryTest_C'",
		[]string{inventoryID.String()},
		props.Bytes(),
	)
}

func writeNameVectorProperty(buf *bytes.Buffer, name string, x float64, y float64, z float64) {
	WriteArkString(buf, name)
	WriteArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, "Vector")
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, "/Script/CoreUObject")
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(24))
	buf.WriteByte(8)
	_ = binary.Write(buf, binary.LittleEndian, x)
	_ = binary.Write(buf, binary.LittleEndian, y)
	_ = binary.Write(buf, binary.LittleEndian, z)
}

func playerProperties(opts PlayerArchiveOptions) []byte {
	var myData bytes.Buffer
	WriteNameIntProperty(&myData, "PlayerDataID", opts.PlayerDataID)
	WriteNameStringProperty(&myData, "PlayerCharacterName", opts.CharacterName)
	WriteNameStringProperty(&myData, "PlayerName", opts.PlayerName)
	if opts.UniqueID != "" {
		WriteNameStringProperty(&myData, "UniqueID", opts.UniqueID)
	}
	WriteNameIntProperty(&myData, "TribeID", opts.TribeID)
	if opts.NumDeaths != 0 {
		WriteNameIntProperty(&myData, "NumOfDeaths", opts.NumDeaths)
	}
	WriteArkString(&myData, "None")

	var props bytes.Buffer
	WriteNameIntProperty(&props, "SavedPlayerDataVersion", 17)
	if opts.ExtraCharacterLevel != 0 {
		WriteNameIntProperty(&props, "CharacterStatusComponent_ExtraCharacterLevel", opts.ExtraCharacterLevel)
	}
	if opts.ExperiencePoints != 0 {
		WriteNameFloatProperty(&props, "CharacterStatusComponent_ExperiencePoints", opts.ExperiencePoints)
	}
	if opts.TotalEngramPoints != 0 {
		WriteNameIntProperty(&props, "PlayerState_TotalEngramPoints", opts.TotalEngramPoints)
	}
	if len(opts.UnlockedEngrams) > 0 {
		WriteNameObjectPathArrayProperty(&props, "PlayerState_EngramBlueprints", opts.UnlockedEngrams)
	}
	WriteNameStructProperty(&props, "MyData", "PlayerDataStruct", myData.Bytes())
	WriteArkString(&props, "None")
	return props.Bytes()
}

func WriteTribeArchive(tb testing.TB, path string) {
	tb.Helper()
	WriteTribeArchiveWithOptions(tb, path, TribeArchiveOptions{
		Name:     "Porters",
		TribeID:  12345,
		OwnerID:  42,
		NumDinos: 7,
	})
}

type TribeArchiveOptions struct {
	Name      string
	TribeID   int32
	OwnerID   int32
	NumDinos  int32
	Members   []string
	MemberIDs []int32
	TribeLog  []string
}

type GameModeCustomBytesOptions struct {
	Player         PlayerArchiveOptions
	NextPlayer     PlayerArchiveOptions
	Tribe          TribeArchiveOptions
	TribeArchiveID uuid.UUID
}

func WriteTribeArchiveWithOptions(tb testing.TB, path string, opts TribeArchiveOptions) {
	tb.Helper()
	writeFile(tb, path, TribeArchiveBytes(tb, opts), "tribe archive fixture")
}

func TribeArchiveBytes(tb testing.TB, opts TribeArchiveOptions) []byte {
	tb.Helper()
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	var buf bytes.Buffer
	WriteArchivePrefix(&buf, id, "/Script/ShooterGame.PrimalTribeData", []string{"TribeData_0"})
	offsetPos := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	propertiesOffset := int32(buf.Len() - 1)
	binary.LittleEndian.PutUint32(buf.Bytes()[offsetPos:offsetPos+4], uint32(propertiesOffset))
	buf.Write(tribeProperties(opts))
	return buf.Bytes()
}

func TribeGameObjectBytes(opts TribeArchiveOptions) []byte {
	return GameObjectBytesWithNames("/Script/ShooterGame.PrimalTribeData", []string{"TribeData_0"}, tribeProperties(opts))
}

func tribeProperties(opts TribeArchiveOptions) []byte {
	var tribeData bytes.Buffer
	WriteNameStringProperty(&tribeData, "TribeName", opts.Name)
	WriteNameIntProperty(&tribeData, "TribeID", opts.TribeID)
	WriteNameIntProperty(&tribeData, "OwnerPlayerDataId", opts.OwnerID)
	WriteNameIntProperty(&tribeData, "NumTribeDinos", opts.NumDinos)
	if len(opts.Members) > 0 {
		WriteNameStringArrayProperty(&tribeData, "MembersPlayerName", opts.Members)
	}
	if len(opts.MemberIDs) > 0 {
		WriteNameIntArrayProperty(&tribeData, "MembersPlayerDataID", opts.MemberIDs)
	}
	if len(opts.TribeLog) > 0 {
		WriteNameStringArrayProperty(&tribeData, "TribeLog", opts.TribeLog)
	}
	WriteArkString(&tribeData, "None")

	var props bytes.Buffer
	WriteNameStructProperty(&props, "TribeData", "TribeDataStruct", tribeData.Bytes())
	WriteArkString(&props, "None")
	return props.Bytes()
}

func GameModeCustomBytes(tb testing.TB, opts GameModeCustomBytesOptions) []byte {
	tb.Helper()
	tribeArchiveID := opts.TribeArchiveID
	if tribeArchiveID == uuid.Nil {
		tribeArchiveID = uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee")
	}

	player := PlayerArchiveBytes(tb, opts.Player)
	nextPlayer := PlayerArchiveBytes(tb, opts.NextPlayer)
	tribe := TribeArchiveBytes(tb, opts.Tribe)

	const archiveUUIDOffset = 16
	copy(tribe[archiveUUIDOffset:archiveUUIDOffset+16], tribeArchiveID[:])

	var buf bytes.Buffer
	buf.WriteByte(1)
	buf.Write(player)
	buf.WriteByte(0)
	buf.Write(nextPlayer)
	buf.WriteByte(0)
	buf.Write(tribe)
	buf.Write(tribe[archiveUUIDOffset-1 : archiveUUIDOffset+15])
	return buf.Bytes()
}

func WriteTributeFile(tb testing.TB, path string, playerIDs []uint64, tribeIDs []uint64) {
	tb.Helper()
	writeFile(tb, path, TributeBytes(tb, playerIDs, tribeIDs), "tribute fixture")
}

func WriteSparseFile(tb testing.TB, path string, size int64) {
	tb.Helper()
	file, err := os.Create(path)
	if err != nil {
		tb.Fatalf("create sparse file: %v", err)
	}
	if err := file.Truncate(size); err != nil {
		_ = file.Close()
		tb.Fatalf("truncate sparse file: %v", err)
	}
	if err := file.Close(); err != nil {
		tb.Fatalf("close sparse file: %v", err)
	}
}

func TributeBytes(tb testing.TB, playerIDs []uint64, tribeIDs []uint64) []byte {
	tb.Helper()
	var buf bytes.Buffer
	WriteIDList(tb, &buf, playerIDs)
	WriteIDList(tb, &buf, tribeIDs)
	return buf.Bytes()
}

func WriteIDList(tb testing.TB, buf *bytes.Buffer, ids []uint64) {
	tb.Helper()
	if err := binary.Write(buf, binary.LittleEndian, int32(len(ids))); err != nil {
		tb.Fatalf("write ID list count: %v", err)
	}
	for _, id := range ids {
		if err := binary.Write(buf, binary.LittleEndian, id); err != nil {
			tb.Fatalf("write ID list value: %v", err)
		}
	}
}

func Header(mapName string, names map[uint32]string) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, int16(12))
	nameOffsetPosition := buf.Len()
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, float64(1234.5))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(77))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0))
	for buf.Len() < 30 {
		buf.WriteByte(0)
	}
	WriteArkString(&buf, mapName)
	nameOffset := int32(buf.Len())
	binary.LittleEndian.PutUint32(buf.Bytes()[nameOffsetPosition:nameOffsetPosition+4], uint32(nameOffset))
	_ = binary.Write(&buf, binary.LittleEndian, int32(len(names)))
	keys := make([]uint32, 0, len(names))
	for key := range names {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, key := range keys {
		_ = binary.Write(&buf, binary.LittleEndian, key)
		WriteArkString(&buf, names[key])
	}
	return buf.Bytes()
}

func GenericObjectBytes(classNameID uint32, noneNameID uint32) []byte {
	return ObjectBytesWithProperties(classNameID, noneNameID, nil)
}

func TruncatedObjectBytes(classNameID uint32) []byte {
	var buf bytes.Buffer
	WriteUInt32(&buf, classNameID)
	WriteInt32(&buf, 0)
	return buf.Bytes()
}

func TruncatedObjectWithPropertiesBytes(classNameID uint32, noneNameID uint32, properties []byte, truncateBytes int) []byte {
	raw := ObjectBytesWithProperties(classNameID, noneNameID, properties)
	if truncateBytes <= 0 {
		return raw
	}
	if truncateBytes >= len(raw) {
		return raw[:0]
	}
	return raw[:len(raw)-truncateBytes]
}

func ObjectBytesWithProperties(classNameID uint32, noneNameID uint32, properties []byte) []byte {
	return ObjectBytesWithNamePayload(classNameID, nil, 0, properties, noneNameID)
}

func ObjectBytesWithNamePayload(classNameID uint32, names []byte, unknown int16, properties []byte, noneNameID uint32) []byte {
	var buf bytes.Buffer
	WriteNameID(&buf, classNameID)
	WriteUInt32(&buf, 0)
	nameCount := int32(0)
	if len(names) > 0 {
		nameCount = 1
	}
	_ = binary.Write(&buf, binary.LittleEndian, nameCount)
	buf.Write(names)
	WriteInt32(&buf, 0)
	_ = binary.Write(&buf, binary.LittleEndian, unknown)
	buf.Write(properties)
	_ = binary.Write(&buf, binary.LittleEndian, noneNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func GameObjectBytesWithNames(className string, names []string, properties []byte) []byte {
	var buf bytes.Buffer
	WriteArkString(&buf, className)
	WriteUInt32(&buf, 0)
	WriteInt32(&buf, int32(len(names)))
	for _, name := range names {
		WriteArkString(&buf, name)
	}
	WriteInt32(&buf, -1)
	WriteInt16(&buf, 0)
	buf.Write(properties)
	return buf.Bytes()
}

func ActorTransforms(transforms ...ActorTransform) []byte {
	var buf bytes.Buffer
	for _, transform := range transforms {
		buf.Write(transform.UUID[:])
		_ = binary.Write(&buf, binary.LittleEndian, transform.X)
		_ = binary.Write(&buf, binary.LittleEndian, transform.Y)
		_ = binary.Write(&buf, binary.LittleEndian, transform.Z)
		_ = binary.Write(&buf, binary.LittleEndian, transform.Pitch)
		_ = binary.Write(&buf, binary.LittleEndian, transform.Roll)
		_ = binary.Write(&buf, binary.LittleEndian, transform.Yaw)
		_ = binary.Write(&buf, binary.LittleEndian, transform.Quaternion)
	}
	buf.Write(uuid.Nil[:])
	return buf.Bytes()
}

func CryopodDinoPayload(tb testing.TB, dinoID uuid.UUID, statusID uuid.UUID, opts CryopodDinoPayloadOptions) []byte {
	tb.Helper()
	var decoded bytes.Buffer
	WriteInt32(&decoded, 0)
	WriteInt32(&decoded, 0)
	WriteUInt32(&decoded, 2)
	var dinoOffsetPos int
	var statusOffsetPos int
	if opts.Reversed {
		statusOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, statusID, "Status", []string{"S0"})
		dinoOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, dinoID, "Dino", []string{"D0"})
	} else {
		dinoOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, dinoID, "Dino", []string{"D0"})
		statusOffsetPos = writeCryopodEmbeddedObjectHeader(&decoded, statusID, "Status", []string{"S0"})
	}

	dinoPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000001, 1001)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000002, 2002)
	writeCryopodEmbeddedNameDoubleProperty(&decoded, 0x10000003, 42)
	writeCryopodEmbeddedNone(&decoded)
	statusPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeCryopodEmbeddedNameIntProperty(&decoded, 0x10000005, 12)
	if opts.Health > 0 {
		writeCryopodEmbeddedNamePositionedIntProperty(&decoded, 0x10000007, 0, opts.Health)
	}
	writeCryopodEmbeddedNone(&decoded)

	binary.LittleEndian.PutUint32(decoded.Bytes()[dinoOffsetPos:dinoOffsetPos+4], uint32(dinoPropsOffset))
	binary.LittleEndian.PutUint32(decoded.Bytes()[statusOffsetPos:statusOffsetPos+4], uint32(statusPropsOffset))

	namesOffset := decoded.Len()
	WriteUInt32(&decoded, 8)
	WriteArkString(&decoded, "None")
	WriteArkString(&decoded, "DinoID1")
	WriteArkString(&decoded, "DinoID2")
	WriteArkString(&decoded, "TamedTimeStamp")
	WriteArkString(&decoded, "IntProperty")
	WriteArkString(&decoded, "BaseCharacterLevel")
	WriteArkString(&decoded, "DoubleProperty")
	WriteArkString(&decoded, "NumberOfLevelUpPointsApplied")

	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(decoded.Bytes()); err != nil {
		tb.Fatalf("zlib write: %v", err)
	}
	if err := writer.Close(); err != nil {
		tb.Fatalf("zlib close: %v", err)
	}

	var payload bytes.Buffer
	WriteUInt32(&payload, 0x0407)
	WriteUInt32(&payload, uint32(decoded.Len()))
	WriteUInt32(&payload, uint32(namesOffset))
	payload.Write(compressed.Bytes())
	return payload.Bytes()
}

func CryopodSaddlePayload() []byte {
	var payload bytes.Buffer
	WriteUInt32(&payload, 8)
	WriteUInt32(&payload, 7)
	WriteUInt32(&payload, 0)
	WriteUInt32(&payload, 0)
	writeCryopodPathObjectProperty(&payload, "ItemArchetype", "BlueprintGeneratedClass /Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C")
	WriteArkString(&payload, "None")
	return payload.Bytes()
}

func WriteCustomItemDatasPropertyID(buf *bytes.Buffer, payloads ...[]byte) {
	elements := make([][]byte, 0, len(payloads))
	for _, payload := range payloads {
		byteValues := make([]byte, len(payload))
		copy(byteValues, payload)

		var bytesElement bytes.Buffer
		WriteByteArrayPropertyID(&bytesElement, 0x1000004f, 0x1000001e, 0x10000050, byteValues)
		WriteUInt32(&bytesElement, 0x10000004)
		WriteInt32(&bytesElement, 0)
		elements = append(elements, bytesElement.Bytes())
	}

	var customDataBytes bytes.Buffer
	WriteStructArrayPropertyID(&customDataBytes, 0x1000004d, 0x1000001e, 0x10000049, 0x1000004e, elements)
	WriteUInt32(&customDataBytes, 0x10000004)
	WriteInt32(&customDataBytes, 0)

	var customItemData bytes.Buffer
	WriteStructPropertyID(&customItemData, 0x1000004b, 0x10000049, 0x1000004c, customDataBytes.Bytes())
	WriteUInt32(&customItemData, 0x10000004)
	WriteInt32(&customItemData, 0)

	WriteStructArrayPropertyID(buf, 0x10000048, 0x1000001e, 0x10000049, 0x1000004a, [][]byte{customItemData.Bytes()})
}

func MinimalEmbeddedCryopodPayload(tb testing.TB, dinoID uuid.UUID, statusID uuid.UUID) []byte {
	tb.Helper()

	var decoded bytes.Buffer
	WriteInt32(&decoded, 0)
	WriteInt32(&decoded, 0)
	WriteUInt32(&decoded, 2)
	dinoOffsetPos := writeMinimalEmbeddedObjectHeader(&decoded, dinoID, "Dino", []string{"D0"})
	statusOffsetPos := writeMinimalEmbeddedObjectHeader(&decoded, statusID, "Status", []string{"S0"})

	dinoPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeMinimalEmbeddedNameIntProperty(&decoded, 0x10000001, 1001)
	writeMinimalEmbeddedNone(&decoded)
	statusPropsOffset := decoded.Len()
	decoded.WriteByte(0)
	writeMinimalEmbeddedNameIntProperty(&decoded, 0x10000003, 12)
	writeMinimalEmbeddedNone(&decoded)

	binary.LittleEndian.PutUint32(decoded.Bytes()[dinoOffsetPos:dinoOffsetPos+4], uint32(dinoPropsOffset))
	binary.LittleEndian.PutUint32(decoded.Bytes()[statusOffsetPos:statusOffsetPos+4], uint32(statusPropsOffset))

	namesOffset := decoded.Len()
	WriteUInt32(&decoded, 4)
	WriteArkString(&decoded, "None")
	WriteArkString(&decoded, "DinoID1")
	WriteArkString(&decoded, "IntProperty")
	WriteArkString(&decoded, "BaseCharacterLevel")

	var compressed bytes.Buffer
	writer := zlib.NewWriter(&compressed)
	if _, err := writer.Write(decoded.Bytes()); err != nil {
		tb.Fatalf("zlib write: %v", err)
	}
	if err := writer.Close(); err != nil {
		tb.Fatalf("zlib close: %v", err)
	}

	var payload bytes.Buffer
	WriteUInt32(&payload, 0x0407)
	WriteUInt32(&payload, uint32(decoded.Len()))
	WriteUInt32(&payload, uint32(namesOffset))
	payload.Write(compressed.Bytes())
	return payload.Bytes()
}

func WriteArkString(buf *bytes.Buffer, value string) {
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+1))
	buf.WriteString(value)
	buf.WriteByte(0)
}

func WriteStringArray(buf *bytes.Buffer, values []string) {
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		WriteArkString(buf, value)
	}
}

func WriteInt32(buf *bytes.Buffer, value int32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameID(buf *bytes.Buffer, id uint32) {
	WriteUInt32(buf, id)
	WriteInt32(buf, 0)
}

func WriteInt16(buf *bytes.Buffer, value int16) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteUInt32(buf *bytes.Buffer, value uint32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteFloat32(buf *bytes.Buffer, value float32) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteFloat64(buf *bytes.Buffer, value float64) {
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameIntProperty(buf *bytes.Buffer, name string, value int32) {
	WriteArkString(buf, name)
	WriteArkString(buf, "IntProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteIntPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value int32) {
	writePropertyIDHeader(buf, name, propertyType, 4, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteFloatPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value float32) {
	writePropertyIDHeader(buf, name, propertyType, 4, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteDoublePropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value float64) {
	writePropertyIDHeader(buf, name, propertyType, 8, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteStringPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value string) {
	writePropertyIDHeader(buf, name, propertyType, int32(len(value)+5), 0, false)
	WriteArkString(buf, value)
}

func WriteBoolPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, value bool) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, propertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	if value {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
}

func WriteObjectReferencePropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, id uuid.UUID) {
	writePropertyIDHeader(buf, name, propertyType, 18, 0, false)
	_ = binary.Write(buf, binary.LittleEndian, int16(0))
	buf.Write(id[:])
}

func WriteObjectReferenceArrayPropertyID(buf *bytes.Buffer, name uint32, arrayPropertyType uint32, objectPropertyType uint32, values []uuid.UUID) {
	bodySize := int32(4 + len(values)*18)
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, arrayPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, bodySize)
	_ = binary.Write(buf, binary.LittleEndian, objectPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, bodySize)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, int32(len(values)))
	for _, id := range values {
		_ = binary.Write(buf, binary.LittleEndian, int16(0))
		buf.Write(id[:])
	}
}

func WriteNameArrayPropertyID(buf *bytes.Buffer, name uint32, arrayPropertyType uint32, namePropertyType uint32, values []uint32) {
	bodySize := uint32(4 + len(values)*8)
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, arrayPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	_ = binary.Write(buf, binary.LittleEndian, namePropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, bodySize)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		_ = binary.Write(buf, binary.LittleEndian, value)
		_ = binary.Write(buf, binary.LittleEndian, int32(0))
	}
}

func WriteByteArrayPropertyID(buf *bytes.Buffer, name uint32, arrayPropertyType uint32, bytePropertyType uint32, values []byte) {
	bodySize := 4 + len(values)
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, arrayPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	_ = binary.Write(buf, binary.LittleEndian, bytePropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(bodySize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	buf.Write(values)
}

func WriteStructPropertyID(buf *bytes.Buffer, name uint32, structPropertyType uint32, structType uint32, body []byte) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, structPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
}

func WriteVectorPropertyID(buf *bytes.Buffer, name uint32, structPropertyType uint32, vectorType uint32, coreObjectName uint32, x float64, y float64, z float64) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, structPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, vectorType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, coreObjectName)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(24))
	buf.WriteByte(8)
	_ = binary.Write(buf, binary.LittleEndian, x)
	_ = binary.Write(buf, binary.LittleEndian, y)
	_ = binary.Write(buf, binary.LittleEndian, z)
}

func WriteStructArrayPropertyID(buf *bytes.Buffer, name uint32, arrayPropertyType uint32, structPropertyType uint32, structType uint32, elements [][]byte) {
	bodySize := 4
	for _, element := range elements {
		bodySize += len(element)
	}
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, arrayPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	_ = binary.Write(buf, binary.LittleEndian, structPropertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	_ = binary.Write(buf, binary.LittleEndian, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(bodySize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(elements)))
	for _, element := range elements {
		buf.Write(element)
	}
}

func WritePositionedIntPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, value int32) {
	writePropertyIDHeader(buf, name, propertyType, 4, position, false)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WritePositionedFloatPropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, value float32) {
	writePropertyIDHeader(buf, name, propertyType, 4, position, true)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WritePositionedInt8PropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, value int8) {
	writePropertyIDHeader(buf, name, propertyType, 1, position, true)
	_ = binary.Write(buf, binary.LittleEndian, position)
	buf.WriteByte(byte(value))
}

func WritePositionedNamePropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, valueName uint32) {
	writePropertyIDHeader(buf, name, propertyType, 8, position, true)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, valueName)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
}

func WritePositionedUInt16PropertyID(buf *bytes.Buffer, name uint32, propertyType uint32, position int32, value uint16) {
	writePropertyIDHeader(buf, name, propertyType, 2, position, true)
	_ = binary.Write(buf, binary.LittleEndian, position)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writePropertyIDHeader(buf *bytes.Buffer, name uint32, propertyType uint32, size int32, position int32, positioned bool) {
	_ = binary.Write(buf, binary.LittleEndian, name)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, propertyType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, size)
	_ = binary.Write(buf, binary.LittleEndian, position)
	if positioned {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
}

func WriteNameStringProperty(buf *bytes.Buffer, name string, value string) {
	WriteArkString(buf, name)
	WriteArkString(buf, "StrProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(len(value)+5))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	WriteArkString(buf, value)
}

func WriteNameFloatProperty(buf *bytes.Buffer, name string, value float32) {
	WriteArkString(buf, name)
	WriteArkString(buf, "FloatProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameDoubleProperty(buf *bytes.Buffer, name string, value float64) {
	WriteArkString(buf, name)
	WriteArkString(buf, "DoubleProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(8))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func WriteNameObjectPathProperty(buf *bytes.Buffer, name string, value string) {
	var body bytes.Buffer
	_ = binary.Write(&body, binary.LittleEndian, int32(1))
	WriteArkString(&body, value)

	WriteArkString(buf, name)
	WriteArkString(buf, "ObjectProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(body.Len()))
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	buf.Write(body.Bytes())
}

func WriteNameByteArrayProperty(buf *bytes.Buffer, name string, values []byte) {
	WriteArkString(buf, name)
	WriteArkString(buf, "ArrayProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(4+len(values)))
	WriteArkString(buf, "ByteProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	_ = binary.Write(buf, binary.LittleEndian, uint32(4+len(values)))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	buf.Write(values)
}

func WriteNameStringArrayProperty(buf *bytes.Buffer, name string, values []string) {
	bodySize := 4
	for _, value := range values {
		bodySize += 4 + len(value) + 1
	}
	writeNameArrayHeader(buf, name, "StrProperty", bodySize)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		WriteArkString(buf, value)
	}
}

func WriteNameIntArrayProperty(buf *bytes.Buffer, name string, values []int32) {
	bodySize := 4 + len(values)*4
	writeNameArrayHeader(buf, name, "IntProperty", bodySize)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		_ = binary.Write(buf, binary.LittleEndian, value)
	}
}

func WriteNameObjectPathArrayProperty(buf *bytes.Buffer, name string, values []string) {
	bodySize := 4
	for _, value := range values {
		bodySize += 4
		bodySize += 4 + len(value) + 1
	}
	writeNameArrayHeader(buf, name, "ObjectProperty", bodySize)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(values)))
	for _, value := range values {
		_ = binary.Write(buf, binary.LittleEndian, int32(1))
		WriteArkString(buf, value)
	}
}

func writeNameArrayHeader(buf *bytes.Buffer, name string, elementType string, bodySize int) {
	WriteArkString(buf, name)
	WriteArkString(buf, "ArrayProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	WriteArkString(buf, elementType)
	_ = binary.Write(buf, binary.LittleEndian, int32(0))
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	buf.WriteByte(0)
}

func WriteNameStructProperty(buf *bytes.Buffer, name string, structType string, body []byte) {
	WriteArkString(buf, name)
	WriteArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(body)))
	buf.WriteByte(0)
	buf.Write(body)
}

func WriteNameStructArrayProperty(buf *bytes.Buffer, name string, structType string, elements [][]byte) {
	bodySize := 4
	for _, element := range elements {
		bodySize += len(element)
	}
	WriteArkString(buf, name)
	WriteArkString(buf, "ArrayProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(bodySize))
	WriteArkString(buf, "StructProperty")
	_ = binary.Write(buf, binary.LittleEndian, int32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(1))
	WriteArkString(buf, structType)
	_ = binary.Write(buf, binary.LittleEndian, uint32(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(bodySize))
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(elements)))
	for _, element := range elements {
		buf.Write(element)
	}
}

func MustExec(tb testing.TB, db *sql.DB, query string, args ...any) {
	tb.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		tb.Fatalf("exec %q: %v", query, err)
	}
}

func WriteArchivePrefix(buf *bytes.Buffer, id uuid.UUID, className string, names []string) {
	WriteInt32(buf, 7)
	WriteInt32(buf, 0)
	WriteInt32(buf, 0)
	WriteInt32(buf, 1)
	buf.Write(id[:])
	WriteArkString(buf, className)
	WriteUInt32(buf, 0)
	WriteStringArray(buf, names)
	WriteUInt32(buf, 0)
	WriteInt32(buf, -1)
	WriteUInt32(buf, 0)
}

func writeCryopodEmbeddedObjectHeader(buf *bytes.Buffer, id uuid.UUID, className string, names []string) int {
	buf.Write(id[:])
	WriteArkString(buf, className)
	WriteUInt32(buf, 0)
	WriteStringArray(buf, names)
	WriteUInt32(buf, 0)
	WriteInt32(buf, 0)
	WriteUInt32(buf, 0)
	offsetPos := buf.Len()
	WriteInt32(buf, 0)
	WriteUInt32(buf, 0)
	return offsetPos
}

func writeCryopodEmbeddedNameIntProperty(buf *bytes.Buffer, nameID uint32, value int32) {
	writeCryopodEmbeddedName(buf, nameID)
	writeCryopodEmbeddedName(buf, 0x10000004)
	WriteInt32(buf, 4)
	WriteInt32(buf, 0)
	buf.WriteByte(0)
	WriteInt32(buf, value)
}

func writeCryopodEmbeddedNamePositionedIntProperty(buf *bytes.Buffer, nameID uint32, position int32, value int32) {
	writeCryopodEmbeddedName(buf, nameID)
	writeCryopodEmbeddedName(buf, 0x10000004)
	WriteInt32(buf, 4)
	WriteInt32(buf, position)
	buf.WriteByte(0)
	WriteInt32(buf, value)
}

func writeCryopodEmbeddedNameDoubleProperty(buf *bytes.Buffer, nameID uint32, value float64) {
	writeCryopodEmbeddedName(buf, nameID)
	writeCryopodEmbeddedName(buf, 0x10000006)
	WriteInt32(buf, 8)
	WriteInt32(buf, 0)
	buf.WriteByte(0)
	_ = binary.Write(buf, binary.LittleEndian, value)
}

func writeCryopodEmbeddedNone(buf *bytes.Buffer) {
	writeCryopodEmbeddedName(buf, 0x10000000)
}

func writeCryopodEmbeddedName(buf *bytes.Buffer, nameID uint32) {
	WriteUInt32(buf, nameID)
	WriteInt32(buf, 0)
}

func writeCryopodPathObjectProperty(buf *bytes.Buffer, name string, path string) {
	WriteArkString(buf, name)
	WriteArkString(buf, "ObjectProperty")
	var body bytes.Buffer
	WriteInt32(&body, 1)
	WriteArkString(&body, path)
	WriteInt32(buf, int32(body.Len()))
	WriteInt32(buf, 0)
	buf.WriteByte(0)
	buf.Write(body.Bytes())
}

func writeMinimalEmbeddedObjectHeader(buf *bytes.Buffer, id uuid.UUID, className string, names []string) int {
	buf.Write(id[:])
	WriteArkString(buf, className)
	WriteUInt32(buf, 0)
	WriteStringArray(buf, names)
	WriteUInt32(buf, 0)
	WriteInt32(buf, 0)
	WriteUInt32(buf, 0)
	offsetPos := buf.Len()
	WriteInt32(buf, 0)
	WriteUInt32(buf, 0)
	return offsetPos
}

func writeMinimalEmbeddedNameIntProperty(buf *bytes.Buffer, nameID uint32, value int32) {
	writeMinimalEmbeddedName(buf, nameID)
	writeMinimalEmbeddedName(buf, 0x10000002)
	WriteInt32(buf, 4)
	WriteInt32(buf, 0)
	buf.WriteByte(0)
	WriteInt32(buf, value)
}

func writeMinimalEmbeddedNone(buf *bytes.Buffer) {
	writeMinimalEmbeddedName(buf, 0x10000000)
}

func writeMinimalEmbeddedName(buf *bytes.Buffer, nameID uint32) {
	WriteUInt32(buf, nameID)
	WriteInt32(buf, 0)
}

func writeFile(tb testing.TB, path string, data []byte, label string) {
	tb.Helper()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		tb.Fatalf("write %s: %v", label, err)
	}
}
