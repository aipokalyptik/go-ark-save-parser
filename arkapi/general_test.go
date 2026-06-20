package arkapi

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
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

func openSyntheticSave(t *testing.T) *arksave.Save {
	t.Helper()

	path := filepath.Join(t.TempDir(), "synthetic.ark")
	objectID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")

	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	mustExec(t, db, `create table custom (key text primary key, value blob)`)
	mustExec(t, db, `create table game (key blob primary key, value blob)`)
	mustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	mustExec(t, db, `insert into game (key, value) values (?, ?)`, objectID[:], syntheticObjectBytes(0x10000001))
	if err := db.Close(); err != nil {
		t.Fatalf("close fixture db: %v", err)
	}

	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	return save
}

func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
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
	_ = binary.Write(&buf, binary.LittleEndian, int32(61))
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
