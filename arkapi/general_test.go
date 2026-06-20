package arkapi

import (
	"bytes"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestGeneralObjectIDsReturnsSaveObjectIDs(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewGeneral(save)
	ids, err := api.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs() error = %v", err)
	}

	wantID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	if len(ids) != 1 || ids[0] != wantID {
		t.Fatalf("ObjectIDs() = %v, want [%s]", ids, wantID)
	}
}

func TestGeneralObjectReturnsParsedSaveObject(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewGeneral(save)
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	obj, err := api.Object(id)
	if err != nil {
		t.Fatalf("Object() error = %v", err)
	}

	if obj.UUID != id {
		t.Fatalf("Object().UUID = %s, want %s", obj.UUID, id)
	}
	if obj.Blueprint != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Object().Blueprint = %q", obj.Blueprint)
	}
	if len(obj.Properties) != 1 || obj.Properties[0].Name != "Health" || obj.Properties[0].Type != arkproperty.TypeInt {
		t.Fatalf("Object().Properties = %#v, want Health Int property", obj.Properties)
	}
}

func TestGeneralObjectsReturnsParsedSaveObjects(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api := NewGeneral(save)
	objects, err := api.Objects()
	if err != nil {
		t.Fatalf("Objects() error = %v", err)
	}

	wantID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	if len(objects) != 1 {
		t.Fatalf("Objects() length = %d, want 1", len(objects))
	}
	if objects[0].UUID != wantID {
		t.Fatalf("Objects()[0].UUID = %s, want %s", objects[0].UUID, wantID)
	}
	if objects[0].Blueprint != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Objects()[0].Blueprint = %q", objects[0].Blueprint)
	}
}

func TestGeneralObjectsWithAnyPropertyFiltersByPropertyName(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	objects, err := NewGeneral(save).ObjectsWithAnyProperty([]string{"Health"})
	if err != nil {
		t.Fatalf("ObjectsWithAnyProperty() error = %v", err)
	}
	if len(objects) != 1 || objects[0].Properties[0].Name != "Health" {
		t.Fatalf("ObjectsWithAnyProperty(Health) = %#v, want one Health object", objects)
	}

	missing, err := NewGeneral(save).ObjectsWithAnyProperty([]string{"TamerString"})
	if err != nil {
		t.Fatalf("ObjectsWithAnyProperty(missing) error = %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("ObjectsWithAnyProperty(missing) = %#v, want empty", missing)
	}
}

func TestGeneralObjectsWithAnyPropertyWithFaultsReportsParseFaults(t *testing.T) {
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000001),
		uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000001)[:40],
	})
	defer save.Close()

	objects, faults, err := NewGeneral(save).ObjectsWithAnyPropertyWithFaults([]string{"Health"})
	if err != nil {
		t.Fatalf("ObjectsWithAnyPropertyWithFaults() error = %v", err)
	}
	if len(objects) != 1 {
		t.Fatalf("objects length = %d, want 1", len(objects))
	}
	if len(faults) != 1 || faults[0].Err == nil {
		t.Fatalf("faults = %#v, want one parse fault", faults)
	}
}

func TestGeneralObjectsWithFaultsReportsParseFaults(t *testing.T) {
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000001),
		uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000001)[:40],
	})
	defer save.Close()

	objects, faults, err := NewGeneral(save).ObjectsWithFaults()
	if err != nil {
		t.Fatalf("ObjectsWithFaults() error = %v", err)
	}
	if len(objects) != 1 || objects[0].UUID != uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff") {
		t.Fatalf("objects = %#v, want one valid object", objects)
	}
	if len(faults) != 1 || faults[0].UUID != uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff") || faults[0].Err == nil {
		t.Fatalf("faults = %#v, want one parse fault", faults)
	}
}

func openSyntheticSave(t *testing.T) *arksave.Save {
	t.Helper()
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	return openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		objectID: syntheticObjectBytes(0x10000001),
	})
}

func openSyntheticSaveWith(t *testing.T, name string, custom map[string][]byte, objects map[uuid.UUID][]byte) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), name)
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header:  syntheticHeader(),
		Custom:  custom,
		Objects: objects,
	})
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func syntheticObjectBytes(classNameID uint32) []byte {
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.LittleEndian, classNameID)
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int16(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000002))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000003))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	_ = binary.Write(&buf, binary.LittleEndian, int32(4))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	buf.WriteByte(0)
	_ = binary.Write(&buf, binary.LittleEndian, int32(250))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	_ = binary.Write(&buf, binary.LittleEndian, int32(0))
	return buf.Bytes()
}

func syntheticHeader() []byte {
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
	writeArkString(&buf, "Valguero_WP")
	nameOffset := int32(buf.Len())
	binary.LittleEndian.PutUint32(buf.Bytes()[nameOffsetPosition:nameOffsetPosition+4], uint32(nameOffset))
	_ = binary.Write(&buf, binary.LittleEndian, int32(81))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000000))
	writeArkString(&buf, "None")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000001))
	writeArkString(&buf, "Blueprint'/Game/Test.Test_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000002))
	writeArkString(&buf, "Health")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000003))
	writeArkString(&buf, "IntProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000004))
	writeArkString(&buf, "None")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000005))
	writeArkString(&buf, "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000006))
	writeArkString(&buf, "StructureID")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000007))
	writeArkString(&buf, "MaxHealth")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000008))
	writeArkString(&buf, "Health")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000009))
	writeArkString(&buf, "TargetingTeam")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000a))
	writeArkString(&buf, "FloatProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000b))
	writeArkString(&buf, "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000c))
	writeArkString(&buf, "ItemQuantity")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000d))
	writeArkString(&buf, "bIsBlueprint")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000e))
	writeArkString(&buf, "BoolProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000000f))
	writeArkString(&buf, "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000010))
	writeArkString(&buf, "ItemRating")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000011))
	writeArkString(&buf, "ItemQualityIndex")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000012))
	writeArkString(&buf, "SavedDurability")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000013))
	writeArkString(&buf, "bIsEngram")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000014))
	writeArkString(&buf, "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000015))
	writeArkString(&buf, "DinoID1")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000016))
	writeArkString(&buf, "DinoID2")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000017))
	writeArkString(&buf, "bIsFemale")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000018))
	writeArkString(&buf, "TamedTimeStamp")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000019))
	writeArkString(&buf, "DoubleProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000001a))
	writeArkString(&buf, "StrProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000001b))
	writeArkString(&buf, "CrafterCharacterName")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000001c))
	writeArkString(&buf, "CrafterTribeName")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000001d))
	writeArkString(&buf, "LinkedStructures")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000001e))
	writeArkString(&buf, "ArrayProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000001f))
	writeArkString(&buf, "ObjectProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000020))
	writeArkString(&buf, "bIsDead")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000021))
	writeArkString(&buf, "bIsBaby")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000022))
	writeArkString(&buf, "bEquippedItem")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000023))
	writeArkString(&buf, "MyInventoryComponent")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000024))
	writeArkString(&buf, "TamedName")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000025))
	writeArkString(&buf, "bNeutered")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000026))
	writeArkString(&buf, "TribeName")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000027))
	writeArkString(&buf, "TamingTeamID")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000028))
	writeArkString(&buf, "TamerString")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000029))
	writeArkString(&buf, "OwningPlayerName")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000002a))
	writeArkString(&buf, "ImprinterName")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000002b))
	writeArkString(&buf, "ImprinterPlayerUniqueNetId")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000002c))
	writeArkString(&buf, "OwningPlayerID")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000002d))
	writeArkString(&buf, "BabyAge")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000002e))
	writeArkString(&buf, "ColorSetIndices")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000002f))
	writeArkString(&buf, "ColorSetNames")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000030))
	writeArkString(&buf, "UploadedFromServerName")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000031))
	writeArkString(&buf, "Black")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000032))
	writeArkString(&buf, "Int8Property")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000033))
	writeArkString(&buf, "NameProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000034))
	writeArkString(&buf, "Blue")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000035))
	writeArkString(&buf, "MyCharacterStatusComponent")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000036))
	writeArkString(&buf, "Blueprint'/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatusComponent_BP.DinoCharacterStatusComponent_BP_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000037))
	writeArkString(&buf, "BaseCharacterLevel")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000038))
	writeArkString(&buf, "NumberOfLevelUpPointsApplied")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000039))
	writeArkString(&buf, "NumberOfLevelUpPointsAppliedTamed")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000003a))
	writeArkString(&buf, "NumberOfMutationsAppliedTamed")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000003b))
	writeArkString(&buf, "CurrentStatusValues")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000003c))
	writeArkString(&buf, "DinoImprintingQuality")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000003d))
	writeArkString(&buf, "GeneTraits")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000003e))
	writeArkString(&buf, "MutableMelee[2]")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000003f))
	writeArkString(&buf, "Robust")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000040))
	writeArkString(&buf, "ItemStatValues")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000041))
	writeArkString(&buf, "UInt16Property")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000042))
	writeArkString(&buf, "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000043))
	writeArkString(&buf, "/ArkOmega/Buffs/Variants/Other/PrimalItemResource_Crystal_Poop.PrimalItemResource_Crystal_Poop_C")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000044))
	writeArkString(&buf, "OwnerInventory")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000045))
	writeArkString(&buf, "CurrentItemCount")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000046))
	writeArkString(&buf, "MaxItemCount")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000047))
	writeArkString(&buf, "Blueprint'/Game/Extinction/CoreBlueprints/Weapons/PrimalItem_WeaponEmptyCryopod.PrimalItem_WeaponEmptyCryopod_C'")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000048))
	writeArkString(&buf, "CustomItemDatas")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000049))
	writeArkString(&buf, "StructProperty")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000004a))
	writeArkString(&buf, "CustomItemData")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000004b))
	writeArkString(&buf, "CustomDataBytes")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000004c))
	writeArkString(&buf, "CustomItemByteArrays")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000004d))
	writeArkString(&buf, "ByteArrays")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000004e))
	writeArkString(&buf, "CustomItemByteArray")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x1000004f))
	writeArkString(&buf, "Bytes")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(0x10000050))
	writeArkString(&buf, "ByteProperty")
	return buf.Bytes()
}

func writeArkString(buf *bytes.Buffer, s string) {
	if s == "" {
		_ = binary.Write(buf, binary.LittleEndian, int32(0))
		return
	}
	_ = binary.Write(buf, binary.LittleEndian, int32(len(s)+1))
	buf.WriteString(s)
	buf.WriteByte(0)
}
