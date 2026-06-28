package arkapi

import (
	"bytes"
	"database/sql"
	"errors"
	"path/filepath"
	"strings"
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

func TestGeneralSaveInfoSummarizesLocalSave(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	info, err := NewGeneral(save).SaveInfo()
	if err != nil {
		t.Fatalf("SaveInfo() error = %v", err)
	}
	if info.MapName != "Valguero_WP" || info.SaveVersion != 12 || info.ObjectCount != 1 {
		t.Fatalf("SaveInfo() = %#v, want synthetic save info", info)
	}
}

func TestNewGeneralFromPathOpensLocalSave(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	api, closeAPI, err := NewGeneralFromPath(save.Path())
	if err != nil {
		t.Fatalf("NewGeneralFromPath() error = %v", err)
	}
	defer closeAPI()

	classes, err := api.Classes()
	if err != nil {
		t.Fatalf("Classes() error = %v", err)
	}
	if len(classes) != 1 || classes[0] != "Blueprint'/Game/Test.Test_C'" {
		t.Fatalf("Classes() = %#v, want synthetic class", classes)
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

func TestGeneralClassesReturnsSortedSaveClasses(t *testing.T) {
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000005),
		uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000001),
	})
	defer save.Close()

	classes, err := NewGeneral(save).Classes()
	if err != nil {
		t.Fatalf("Classes() error = %v", err)
	}
	want := []string{
		"Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
		"Blueprint'/Game/Test.Test_C'",
	}
	if len(classes) != len(want) || classes[0] != want[0] || classes[1] != want[1] {
		t.Fatalf("Classes() = %#v, want %#v", classes, want)
	}
}

func TestGeneralClassesFromPathReturnsSortedSaveClasses(t *testing.T) {
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000005),
		uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff"): syntheticObjectBytes(0x10000001),
	})
	defer save.Close()

	classes, err := GeneralClassesFromPath(save.Path())
	if err != nil {
		t.Fatalf("GeneralClassesFromPath() error = %v", err)
	}
	want := []string{
		"Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
		"Blueprint'/Game/Test.Test_C'",
	}
	if len(classes) != len(want) || classes[0] != want[0] || classes[1] != want[1] {
		t.Fatalf("GeneralClassesFromPath() = %#v, want %#v", classes, want)
	}
}

func TestGeneralParseSummaryCountsObjectsParsedAndFaults(t *testing.T) {
	validID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		validID:  syntheticObjectBytes(0x10000001),
		faultyID: syntheticObjectBytes(0x10000001)[:40],
	})
	defer save.Close()

	summary, faults, err := NewGeneral(save).ParseSummaryWithFaults()
	if err != nil {
		t.Fatalf("ParseSummaryWithFaults() error = %v", err)
	}
	if summary.Objects != 2 || summary.Parsed != 1 || summary.Faults != 1 {
		t.Fatalf("ParseSummaryWithFaults() summary = %#v, want 2 objects, 1 parsed, 1 fault", summary)
	}
	if len(faults) != 1 || faults[0].UUID != faultyID || faults[0].Err == nil {
		t.Fatalf("ParseSummaryWithFaults() faults = %#v, want faulty row", faults)
	}
}

func TestGeneralParseSummaryFromPathCountsObjectsParsedAndFaults(t *testing.T) {
	validID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		validID:  syntheticObjectBytes(0x10000001),
		faultyID: syntheticObjectBytes(0x10000001)[:40],
	})
	defer save.Close()

	summary, faults, err := GeneralParseSummaryFromPath(save.Path())
	if err != nil {
		t.Fatalf("GeneralParseSummaryFromPath() error = %v", err)
	}
	if summary.Objects != 2 || summary.Parsed != 1 || summary.Faults != 1 {
		t.Fatalf("GeneralParseSummaryFromPath() summary = %#v, want 2 objects, 1 parsed, 1 fault", summary)
	}
	if len(faults) != 1 || faults[0].UUID != faultyID || faults[0].Err == nil {
		t.Fatalf("GeneralParseSummaryFromPath() faults = %#v, want faulty row", faults)
	}
}

func TestGeneralParseSummaryFromPathLabelsOpenAndParseErrors(t *testing.T) {
	_, _, err := GeneralParseSummaryFromPath(filepath.Join(t.TempDir(), "missing.ark"))
	if err == nil || !strings.HasPrefix(err.Error(), "open save: ") {
		t.Fatalf("GeneralParseSummaryFromPath(missing) error = %v, want open save label", err)
	}

	path := filepath.Join(t.TempDir(), "no-game-table.ark")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite fixture: %v", err)
	}
	testfixtures.MustExec(t, db, `create table custom (key text primary key, value blob)`)
	testfixtures.MustExec(t, db, `insert into custom (key, value) values (?, ?)`, "SaveHeader", syntheticHeader())
	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite fixture: %v", err)
	}

	_, _, err = GeneralParseSummaryFromPath(path)
	if err == nil || !strings.HasPrefix(err.Error(), "parse objects: ") {
		t.Fatalf("GeneralParseSummaryFromPath(no game table) error = %v, want parse objects label", err)
	}
}

func TestGeneralObjectSummaryReportsBytesAndProperties(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	summary, err := NewGeneral(save).ObjectSummary(id)
	if err != nil {
		t.Fatalf("ObjectSummary() error = %v", err)
	}
	if !summary.Exists {
		t.Fatalf("ObjectSummary().Exists = false, want true")
	}
	if summary.Bytes == 0 {
		t.Fatalf("ObjectSummary().Bytes = 0, want raw byte count")
	}
	if summary.Properties != 1 {
		t.Fatalf("ObjectSummary().Properties = %d, want 1", summary.Properties)
	}
}

func TestGeneralObjectSummaryReportsMissingObject(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	summary, err := NewGeneral(save).ObjectSummary(uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff"))
	if err != nil {
		t.Fatalf("ObjectSummary(missing) error = %v", err)
	}
	if summary.Exists || summary.Bytes != 0 || summary.Properties != 0 {
		t.Fatalf("ObjectSummary(missing) = %#v, want empty missing summary", summary)
	}
}

func TestGeneralObjectSummaryReturnsParseErrors(t *testing.T) {
	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		id: syntheticObjectBytes(0x10000001)[:40],
	})
	defer save.Close()

	_, err := NewGeneral(save).ObjectSummary(id)
	if err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("ObjectSummary(broken) error = %v, want parse error", err)
	}
}

func TestGeneralClassPropertySummaryCountsUniquePropertiesAndFaults(t *testing.T) {
	firstID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	secondID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	faultyID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		firstID:  syntheticObjectBytes(0x10000001),
		secondID: syntheticObjectBytesWithExtraProperty(0x10000001),
		faultyID: syntheticObjectBytes(0x10000001)[:40],
	})
	defer save.Close()

	summary, faults, err := NewGeneral(save).ClassPropertySummaryWithFaults("Test_C")
	if err != nil {
		t.Fatalf("ClassPropertySummaryWithFaults() error = %v", err)
	}
	if summary.Objects != 2 || summary.Properties != 2 {
		t.Fatalf("ClassPropertySummaryWithFaults() summary = %#v, want 2 objects and 2 unique properties", summary)
	}
	if len(faults) != 1 || faults[0].UUID != faultyID || faults[0].Err == nil {
		t.Fatalf("ClassPropertySummaryWithFaults() faults = %#v, want faulty object", faults)
	}
}

func TestGeneralClassPropertySummaryReturnsEmptyForNoMatch(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	summary, faults, err := NewGeneral(save).ClassPropertySummaryWithFaults("DoesNotExist")
	if err != nil {
		t.Fatalf("ClassPropertySummaryWithFaults(no match) error = %v", err)
	}
	if summary.Objects != 0 || summary.Properties != 0 {
		t.Fatalf("ClassPropertySummaryWithFaults(no match) summary = %#v, want empty", summary)
	}
	if len(faults) != 0 {
		t.Fatalf("ClassPropertySummaryWithFaults(no match) faults = %#v, want none", faults)
	}
}

func TestGeneralClassLookupSummaryCountsMatchedStructureClasses(t *testing.T) {
	wallID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	testID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	engramID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	noStructureID := uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff")
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		wallID:        syntheticLookupStructureObjectBytes(0x10000005, false),
		testID:        syntheticLookupStructureObjectBytes(0x10000001, false),
		engramID:      syntheticLookupStructureObjectBytes(0x10000005, true),
		noStructureID: syntheticObjectBytes(0x10000005),
	})
	defer save.Close()

	summary, faults, err := NewGeneral(save).ClassLookupSummaryWithFaults([]string{"Wall_Stone", "Test_C"})
	if err != nil {
		t.Fatalf("ClassLookupSummaryWithFaults() error = %v", err)
	}
	if summary.Objects != 2 || summary.Classes != 2 {
		t.Fatalf("ClassLookupSummaryWithFaults() = %#v, want 2 objects and 2 classes", summary)
	}
	if len(faults) != 0 {
		t.Fatalf("ClassLookupSummaryWithFaults() faults = %#v, want none", faults)
	}
}

func TestGeneralClassLookupSummaryReturnsEmptyForNoSubstrings(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	summary, faults, err := NewGeneral(save).ClassLookupSummaryWithFaults(nil)
	if err != nil {
		t.Fatalf("ClassLookupSummaryWithFaults(nil) error = %v", err)
	}
	if summary.Objects != 0 || summary.Classes != 0 {
		t.Fatalf("ClassLookupSummaryWithFaults(nil) = %#v, want empty", summary)
	}
	if len(faults) != 0 {
		t.Fatalf("ClassLookupSummaryWithFaults(nil) faults = %#v, want none", faults)
	}
}

func TestGeneralPropertyFilterSummaryCountsObjectsAndClasses(t *testing.T) {
	firstID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	secondID := uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff")
	missingID := uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff")
	save := openSyntheticSaveWith(t, "synthetic.ark", nil, map[uuid.UUID][]byte{
		firstID:   syntheticObjectBytes(0x10000001),
		secondID:  syntheticObjectBytesWithExtraProperty(0x10000005),
		missingID: testfixtures.ObjectBytesWithProperties(0x1000000b, 0x10000004, nil),
	})
	defer save.Close()

	summary, err := NewGeneral(save).PropertyFilterSummary([]string{"Health", "MaxHealth"})
	if err != nil {
		t.Fatalf("PropertyFilterSummary() error = %v", err)
	}
	if summary.Objects != 2 || summary.Classes != 2 {
		t.Fatalf("PropertyFilterSummary() = %#v, want 2 objects and 2 classes", summary)
	}
}

func TestGeneralPropertyFilterSummaryReturnsEmptyForNoMatch(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	summary, err := NewGeneral(save).PropertyFilterSummary([]string{"DoesNotExist"})
	if err != nil {
		t.Fatalf("PropertyFilterSummary(no match) error = %v", err)
	}
	if summary.Objects != 0 || summary.Classes != 0 {
		t.Fatalf("PropertyFilterSummary(no match) = %#v, want empty", summary)
	}
}

func TestGeneralPropertyPositionSummaryCountsMetadata(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	id := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	summary, err := NewGeneral(save).PropertyPositionSummary(id)
	if err != nil {
		t.Fatalf("PropertyPositionSummary() error = %v", err)
	}
	if !summary.Exists {
		t.Fatalf("PropertyPositionSummary().Exists = false, want true")
	}
	if summary.Properties != 1 || summary.NameOffsets != 1 || summary.ValueOffsets != 1 || summary.Encoded != 1 || summary.OffsetsOK != 1 {
		t.Fatalf("PropertyPositionSummary() = %#v, want one fully positioned property", summary)
	}
}

func TestGeneralPropertyPositionSummaryReportsMissingObject(t *testing.T) {
	save := openSyntheticSave(t)
	defer save.Close()

	summary, err := NewGeneral(save).PropertyPositionSummary(uuid.MustParse("11112233-4455-6677-8899-aabbccddeeff"))
	if err != nil {
		t.Fatalf("PropertyPositionSummary(missing) error = %v", err)
	}
	if summary.Exists || summary.Properties != 0 || summary.Encoded != 0 {
		t.Fatalf("PropertyPositionSummary(missing) = %#v, want empty missing summary", summary)
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
	return testfixtures.ObjectBytesWithIntProperty(classNameID, 0x10000004, 0x10000002, 0x10000003, 250)
}

func syntheticObjectBytesWithExtraProperty(classNameID uint32) []byte {
	var buf bytes.Buffer
	testfixtures.WriteIntPropertyID(&buf, 0x10000002, 0x10000003, 250)
	testfixtures.WriteFloatPropertyID(&buf, 0x10000007, 0x1000000a, 123.5)
	return testfixtures.ObjectBytesWithProperties(classNameID, 0x10000004, buf.Bytes())
}

func syntheticLookupStructureObjectBytes(classNameID uint32, isEngram bool) []byte {
	var buf bytes.Buffer
	testfixtures.WriteIntPropertyID(&buf, 0x10000006, 0x10000003, 101)
	if isEngram {
		testfixtures.WriteBoolPropertyID(&buf, 0x10000013, 0x1000000e, true)
	}
	return testfixtures.ObjectBytesWithProperties(classNameID, 0x10000004, buf.Bytes())
}

func syntheticHeader() []byte {
	return testfixtures.Header("Valguero_WP", map[uint32]string{
		0x10000000: "None",
		0x10000001: "Blueprint'/Game/Test.Test_C'",
		0x10000002: "Health",
		0x10000003: "IntProperty",
		0x10000004: "None",
		0x10000005: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
		0x10000006: "StructureID",
		0x10000007: "MaxHealth",
		0x10000008: "Health",
		0x10000009: "TargetingTeam",
		0x1000000a: "FloatProperty",
		0x1000000b: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
		0x1000000c: "ItemQuantity",
		0x1000000d: "bIsBlueprint",
		0x1000000e: "BoolProperty",
		0x1000000f: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
		0x10000010: "ItemRating",
		0x10000011: "ItemQualityIndex",
		0x10000012: "SavedDurability",
		0x10000013: "bIsEngram",
		0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
		0x10000015: "DinoID1",
		0x10000016: "DinoID2",
		0x10000017: "bIsFemale",
		0x10000018: "TamedTimeStamp",
		0x10000019: "DoubleProperty",
		0x1000001a: "StrProperty",
		0x1000001b: "CrafterCharacterName",
		0x1000001c: "CrafterTribeName",
		0x1000001d: "LinkedStructures",
		0x1000001e: "ArrayProperty",
		0x1000001f: "ObjectProperty",
		0x10000020: "bIsDead",
		0x10000021: "bIsBaby",
		0x10000022: "bEquippedItem",
		0x10000023: "MyInventoryComponent",
		0x10000024: "TamedName",
		0x10000025: "bNeutered",
		0x10000026: "TribeName",
		0x10000027: "TamingTeamID",
		0x10000028: "TamerString",
		0x10000029: "OwningPlayerName",
		0x1000002a: "ImprinterName",
		0x1000002b: "ImprinterPlayerUniqueNetId",
		0x1000002c: "OwningPlayerID",
		0x1000002d: "BabyAge",
		0x1000002e: "ColorSetIndices",
		0x1000002f: "ColorSetNames",
		0x10000030: "UploadedFromServerName",
		0x10000031: "Black",
		0x10000032: "Int8Property",
		0x10000033: "NameProperty",
		0x10000034: "Blue",
		0x10000035: "MyCharacterStatusComponent",
		0x10000036: "Blueprint'/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatusComponent_BP.DinoCharacterStatusComponent_BP_C'",
		0x10000037: "BaseCharacterLevel",
		0x10000038: "NumberOfLevelUpPointsApplied",
		0x10000039: "NumberOfLevelUpPointsAppliedTamed",
		0x1000003a: "NumberOfMutationsAppliedTamed",
		0x1000003b: "CurrentStatusValues",
		0x1000003c: "DinoImprintingQuality",
		0x1000003d: "GeneTraits",
		0x1000003e: "MutableMelee[2]",
		0x1000003f: "Robust",
		0x10000040: "ItemStatValues",
		0x10000041: "UInt16Property",
		0x10000042: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'",
		0x10000043: "/ArkOmega/Buffs/Variants/Other/PrimalItemResource_Crystal_Poop.PrimalItemResource_Crystal_Poop_C",
		0x10000044: "OwnerInventory",
		0x10000045: "CurrentItemCount",
		0x10000046: "MaxItemCount",
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
		0x10000051: "Blueprint'/Game/Structures/Storage/PrimalStructureItemContainer_StorageBox_Huge.PrimalStructureItemContainer_StorageBox_Huge_C'",
		0x10000052: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Shields/PrimalItemArmor_WoodShield.PrimalItemArmor_WoodShield_C'",
		0x10000053: "Blueprint'/Game/Extinction/CoreBlueprints/Items/Saddle/PrimalItemArmor_GachaSaddle.PrimalItemArmor_GachaSaddle_C'",
	})
}
