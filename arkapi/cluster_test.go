package arkapi

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestClusterAPIClassifiesAndCountsItems(t *testing.T) {
	api := NewCluster(&arkcluster.Data{
		ID:   "EOS_abc123",
		Path: "/tmp/EOS_abc123",
		Items: []arkcluster.Item{
			{
				Index:     0,
				Version:   7,
				Blueprint: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
				Quantity:  3,
				Rating:    1.5,
				Properties: arkproperty.Container{Properties: []arkproperty.Property{{
					Name:  "CustomItemDatas",
					Type:  arkproperty.TypeArray,
					Value: arkproperty.Array{Values: []any{arkproperty.Container{}}},
				}}},
			},
			{
				Index:                1,
				Version:              6,
				Blueprint:            "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C",
				Quantity:             2,
				Rating:               7.5,
				CrafterCharacterName: "Survivor",
				CrafterTribeName:     "Porters",
			},
			{
				Index:     2,
				Blueprint: "/Game/Test/PrimalItemResource_Custom.PrimalItemResource_Custom_C",
				Quantity:  4,
				Rating:    3,
			},
		},
	})

	counts := api.ItemCountsByType()
	if counts["dino"] != 1 || counts["equipment"] != 1 || counts["other"] != 1 {
		t.Fatalf("ItemCountsByType() = %#v, want one dino/equipment/other", counts)
	}
	if got := api.ItemsByType("equipment"); len(got) != 1 || got[0].Index != 1 {
		t.Fatalf("ItemsByType(equipment) = %#v, want item index 1", got)
	}
	if got := api.ItemsByTypedType(arkobject.ClusterItemTypeEquipment); len(got) != 1 || got[0].Index != 1 {
		t.Fatalf("ItemsByTypedType(equipment) = %#v, want item index 1", got)
	}
	if got := api.ItemsByType("missing"); len(got) != 0 {
		t.Fatalf("ItemsByType(missing) = %#v, want empty", got)
	}
	typed := api.ItemsTyped()
	if len(typed) != 3 || typed[0].Type != "dino" || typed[1].Type != "equipment" || typed[2].Type != "other" {
		t.Fatalf("ItemsTyped() = %#v, want dino/equipment/other projections", typed)
	}
	if typed[0].Type != "dino" || typed[0].ItemType() != arkobject.ClusterItemTypeDino || !typed[0].IsDinoUpload() || typed[0].IsEquipmentUpload() || typed[0].IsOtherUpload() {
		t.Fatalf("typed dino item helpers = %#v, want dino upload only", typed[0])
	}
	if typed[0].UnsupportedVersion() || !typed[1].UnsupportedVersion() {
		t.Fatalf("typed item version helpers = %#v/%#v, want only item 1 unsupported", typed[0], typed[1])
	}
	if typed[1].Type != "equipment" || typed[1].ItemType() != arkobject.ClusterItemTypeEquipment || !typed[1].IsEquipmentUpload() || typed[1].IsDinoUpload() || typed[1].IsOtherUpload() || typed[1].SupportedVersion() {
		t.Fatalf("typed equipment item helpers = %#v, want unsupported equipment upload only", typed[1])
	}
	crafted := typed[1].Crafter()
	if !typed[1].IsCrafted() || crafted.CharacterName != "Survivor" || crafted.TribeName != "Porters" {
		t.Fatalf("typed equipment crafter = %#v crafted=%v, want Survivor/Porters crafted item", crafted, typed[1].IsCrafted())
	}
	if typed[2].Type != "other" || typed[2].ItemType() != arkobject.ClusterItemTypeOther || !typed[2].IsOtherUpload() || typed[2].IsDinoUpload() || typed[2].IsEquipmentUpload() {
		t.Fatalf("typed other item helpers = %#v, want other upload only", typed[2])
	}
	if typed[2].IsCrafted() || typed[2].Crafter().Valid() {
		t.Fatalf("typed other item crafter = %#v crafted=%v, want no crafter metadata", typed[2].Crafter(), typed[2].IsCrafted())
	}
	if got := api.ItemsByTypeTyped("dino"); len(got) != 1 || got[0].Index != 0 || got[0].Type != "dino" {
		t.Fatalf("ItemsByTypeTyped(dino) = %#v, want typed item index 0", got)
	}
	summary := api.ItemSummary()
	if summary.Items != 3 || summary.DinoItems != 1 || summary.EquipmentItems != 1 || summary.OtherItems != 1 {
		t.Fatalf("ItemSummary() counts = %#v, want one dino/equipment/other across 3 items", summary)
	}
	if summary.SupportedVersionItems != 1 || summary.UnsupportedVersionItems != 1 {
		t.Fatalf("ItemSummary() version counts = %#v, want one supported and one unsupported", summary)
	}
	if summary.CraftedItems != 1 || summary.TotalQuantity != 9 || summary.AverageQuantity != 3 {
		t.Fatalf("ItemSummary() quantity aggregates = %#v, want one crafted item, total quantity 9, average quantity 3", summary)
	}
	if summary.TotalRating != 12 || summary.AverageRating != 4 || summary.MaxRating != 7.5 || summary.MaxQuality != 0 {
		t.Fatalf("ItemSummary() rating aggregates = %#v, want total rating 12, average rating 4, max rating 7.5", summary)
	}
}

func TestNewClusterFromPathOpensLocalClusterFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	testfixtures.WriteArchive(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

	api, err := NewClusterFromPath(path)
	if err != nil {
		t.Fatalf("NewClusterFromPath() error = %v", err)
	}
	summary := api.Summary()
	if summary.ID != "EOS_abc123" || summary.Path != path || summary.ArchiveVersion == 0 || summary.ObjectCount != 1 {
		t.Fatalf("NewClusterFromPath() summary = %#v, want opened cluster archive metadata", summary)
	}
}

func TestClusterItemsFromPathReturnsTypedLocalUploads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	writeClusterArchiveWithPayload(t, path, clusterItemPayload(t))

	items, err := ClusterItemsFromPath(path)
	if err != nil {
		t.Fatalf("ClusterItemsFromPath() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("ClusterItemsFromPath() items length = %d, want 1", len(items))
	}
	item := items[0]
	if item.Blueprint != "/Game/Test/PrimalItem_Test.PrimalItem_Test_C" || item.Quantity != 3 || item.Rating != 7.5 || item.Quality != 2 {
		t.Fatalf("ClusterItemsFromPath() item = %#v, want typed local item upload", item)
	}
	if !item.IsCrafted() || item.Crafter().CharacterName != "Survivor" || item.Crafter().TribeName != "Porters" {
		t.Fatalf("ClusterItemsFromPath() crafter = %#v crafted=%v, want Survivor/Porters crafted item", item.Crafter(), item.IsCrafted())
	}
}

func TestClusterDinosFromPathReturnsTypedLocalUploads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	writeClusterArchiveWithPayload(t, path, clusterMalformedDinoPayload(t))

	dinos, err := ClusterDinosFromPath(path)
	if err != nil {
		t.Fatalf("ClusterDinosFromPath() error = %v", err)
	}
	if len(dinos) != 1 {
		t.Fatalf("ClusterDinosFromPath() dinos length = %d, want 1", len(dinos))
	}
	dino := dinos[0]
	if dino.RawSize != len("not an archive") || !dino.HasParseError() || dino.ParseStatus() != arkobject.ClusterDinoParseStatusParseError {
		t.Fatalf("ClusterDinosFromPath() dino = %#v, want parse-error typed upload", dino)
	}
}

func TestClusterAPISummarizesDinoParseStatus(t *testing.T) {
	api := NewCluster(&arkcluster.Data{
		ID:      "EOS_abc123",
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Dinos: []arkcluster.Dino{
			{
				Index:   0,
				Version: 7,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					{ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C"},
					{ClassName: "/Game/PrimalEarth/CoreBlueprints/CharacterStatusComponent_BP.CharacterStatusComponent_BP_C"},
					{ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_AIController_BP.Raptor_AIController_BP_C"},
					{ClassName: "/Game/PrimalEarth/CoreBlueprints/InventoryComponent_BP.InventoryComponent_BP_C"},
				}},
			},
			{
				Index:   1,
				Version: 6,
				Archive: &arkarchive.Archive{},
			},
			{
				Index:      2,
				Version:    7,
				ParseError: "unsupported embedded archive",
			},
		},
	})

	if got := api.ParseErrorCount(); got != 1 {
		t.Fatalf("ParseErrorCount() = %d, want 1", got)
	}
	if got := api.DinosByParseStatus(true); len(got) != 2 || got[0].Index != 0 || got[1].Index != 1 {
		t.Fatalf("DinosByParseStatus(true) = %#v, want parsed dinos 0 and 1", got)
	}
	if got := api.DinosByParseStatus(false); len(got) != 1 || got[0].Index != 2 {
		t.Fatalf("DinosByParseStatus(false) = %#v, want unparsed dino 2", got)
	}
	typed := api.DinosTyped()
	if len(typed) != 3 || typed[0].ObjectCount != 4 || len(typed[0].ClassNames) != 4 || typed[0].ClassNames[0] != "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C" {
		t.Fatalf("DinosTyped() = %#v, want class-name projection for parsed dino", typed)
	}
	if !typed[0].Parsed() || typed[0].HasParseError() || !typed[1].Parsed() || !typed[1].UnsupportedVersion() || !typed[2].HasParseError() {
		t.Fatalf("typed dino helpers = %#v, want parsed/unsupported/error helpers", typed)
	}
	if len(typed[0].StatusComponentClassNames) != 1 || len(typed[0].AIControllerClassNames) != 1 || len(typed[0].InventoryComponentClassNames) != 1 {
		t.Fatalf("typed dino component class names = %#v, want status/AI/inventory component summaries", typed[0])
	}
	if got := api.DinosByParseStatusTyped(false); len(got) != 1 || got[0].Index != 2 || got[0].ParseError != "unsupported embedded archive" {
		t.Fatalf("DinosByParseStatusTyped(false) = %#v, want unparsed typed dino 2", got)
	}
	summary := api.Summary()
	if summary.ID != "EOS_abc123" || summary.ItemCount != 0 || summary.DinoCount != 3 || summary.ParseErrorCount != 1 || summary.ObjectCount != 1 {
		t.Fatalf("Summary() = %#v", summary)
	}
	dinoSummary := api.DinoSummary()
	if dinoSummary.Dinos != 3 || dinoSummary.ParsedDinos != 2 || dinoSummary.ParseErrorDinos != 1 || dinoSummary.UnsupportedVersionDinos != 1 {
		t.Fatalf("DinoSummary() counts = %#v, want 3 dinos, 2 parsed, 1 parse error, 1 unsupported version", dinoSummary)
	}
	if dinoSummary.WithStatusComponent != 1 || dinoSummary.WithAIController != 1 || dinoSummary.WithInventoryComponent != 1 || dinoSummary.TotalEmbeddedObjects != 4 || dinoSummary.MaxEmbeddedObjects != 4 {
		t.Fatalf("DinoSummary() component counts = %#v, want one component-bearing dino with four embedded objects", dinoSummary)
	}
}

func TestClusterAPISummarizesEmbeddedDinoIdentityAndStats(t *testing.T) {
	api := NewCluster(&arkcluster.Data{
		Dinos: []arkcluster.Dino{
			{
				Index:   0,
				Version: 7,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					clusterDinoObjectForSummaryTest(1001, 2002, true, true, false, false, true),
					clusterStatusObjectForSummaryTest(),
				}},
			},
			{
				Index:   1,
				Version: 7,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					clusterDinoObjectForSummaryTest(3003, 4004, false, false, true, true, false),
				}},
			},
			{Index: 2, Version: 7},
		},
	})

	summary := api.DinoSummary()
	if summary.WithDinoID != 2 || summary.TamedDinos != 1 || summary.FemaleDinos != 1 || summary.BabyDinos != 1 || summary.DeadDinos != 1 || summary.WithStats != 1 {
		t.Fatalf("DinoSummary() identity/stat counts = %#v, want 2 IDs, 1 tamed/female/baby/dead/stats", summary)
	}
	if summary.TotalBaseLevel != 12 || summary.MaxBaseLevel != 12 || summary.AverageBaseLevel != 12 || summary.TotalCurrentLevel != 6 || summary.MaxCurrentLevel != 6 || summary.AverageCurrentLevel != 6 {
		t.Fatalf("DinoSummary() level aggregates = %#v, want base/current level aggregates from embedded status", summary)
	}
}

func writeClusterArchiveWithPayload(t *testing.T, path string, payload []byte) {
	t.Helper()

	var props bytes.Buffer
	testfixtures.WriteNameStructProperty(&props, "MyArkData", "ArkInventoryData", payload)
	testfixtures.WriteArkString(&props, "None")
	testfixtures.WriteArchiveWithProperties(t, path, "/Script/ShooterGame.ArkCloudInventoryData", props.Bytes())
}

func clusterItemPayload(t *testing.T) []byte {
	t.Helper()

	var item bytes.Buffer
	testfixtures.WriteNameDoubleProperty(&item, "Version", 7)
	testfixtures.WriteNameDoubleProperty(&item, "UploadTime", 12345)
	testfixtures.WriteNameObjectPathProperty(&item, "ItemArchetype", "BlueprintGeneratedClass /Game/Test/PrimalItem_Test.PrimalItem_Test_C")
	testfixtures.WriteNameIntProperty(&item, "ItemQuantity", 3)
	testfixtures.WriteNameFloatProperty(&item, "ItemRating", 7.5)
	testfixtures.WriteNameIntProperty(&item, "ItemQualityIndex", 2)
	testfixtures.WriteNameStringProperty(&item, "CrafterCharacterName", "Survivor")
	testfixtures.WriteNameStringProperty(&item, "CrafterTribeName", "Porters")
	testfixtures.WriteArkString(&item, "None")

	var payload bytes.Buffer
	testfixtures.WriteNameStructArrayProperty(&payload, "ArkItems", "ArkTributeInventoryItem", [][]byte{item.Bytes()})
	testfixtures.WriteArkString(&payload, "None")
	return payload.Bytes()
}

func clusterMalformedDinoPayload(t *testing.T) []byte {
	t.Helper()

	var dino bytes.Buffer
	testfixtures.WriteNameDoubleProperty(&dino, "Version", 7)
	testfixtures.WriteNameDoubleProperty(&dino, "UploadTime", 12345)
	testfixtures.WriteNameByteArrayProperty(&dino, "DinoData", []byte("not an archive"))
	testfixtures.WriteArkString(&dino, "None")

	var payload bytes.Buffer
	testfixtures.WriteNameStructArrayProperty(&payload, "ArkTamedDinosData", "ArkTributeDinoData", [][]byte{dino.Bytes()})
	testfixtures.WriteArkString(&payload, "None")
	return payload.Bytes()
}

func TestClusterAPIDinoParseStatusCounts(t *testing.T) {
	api := NewCluster(&arkcluster.Data{
		Dinos: []arkcluster.Dino{
			{
				Index:   0,
				Version: 7,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					{ClassName: "/Game/Test/ParsedDino.ParsedDino_C"},
				}},
			},
			{
				Index:   1,
				Version: 6,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					{ClassName: "/Game/Test/UnsupportedVersion.UnsupportedVersion_C"},
				}},
			},
			{
				Index:      2,
				Version:    7,
				ParseError: "unsupported embedded archive",
			},
			{
				Index:   3,
				Version: 7,
			},
		},
	})

	typed := api.DinosTyped()
	if got := typed[0].ParseStatus(); got != arkobject.ClusterDinoParseStatusParsed {
		t.Fatalf("typed[0].ParseStatus() = %q, want parsed", got)
	}
	if got := typed[1].ParseStatus(); got != arkobject.ClusterDinoParseStatusUnsupportedVersion {
		t.Fatalf("typed[1].ParseStatus() = %q, want unsupported_version", got)
	}
	if got := typed[2].ParseStatus(); got != arkobject.ClusterDinoParseStatusParseError {
		t.Fatalf("typed[2].ParseStatus() = %q, want parse_error", got)
	}
	if got := typed[3].ParseStatus(); got != arkobject.ClusterDinoParseStatusUnparsed {
		t.Fatalf("typed[3].ParseStatus() = %q, want unparsed", got)
	}
	counts := api.DinoParseStatusCounts()
	for status, want := range map[arkobject.ClusterDinoParseStatus]int{
		arkobject.ClusterDinoParseStatusParsed:             1,
		arkobject.ClusterDinoParseStatusUnsupportedVersion: 1,
		arkobject.ClusterDinoParseStatusParseError:         1,
		arkobject.ClusterDinoParseStatusUnparsed:           1,
	} {
		if got := counts[status.String()]; got != want {
			t.Fatalf("DinoParseStatusCounts()[%q] = %d, want %d; all counts %#v", status, got, want, counts)
		}
	}
}

func TestClusterDirectorySummaryAggregatesFiles(t *testing.T) {
	entries := []*arkcluster.Data{
		{
			ID:      "EOS_one",
			Path:    "/tmp/EOS_one",
			Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
			Items: []arkcluster.Item{
				{Index: 0, Version: 7, Quantity: 3, Blueprint: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C", Properties: arkproperty.Container{Properties: []arkproperty.Property{{
					Name:  "CustomItemDatas",
					Type:  arkproperty.TypeArray,
					Value: arkproperty.Array{Values: []any{arkproperty.Container{}}},
				}}}},
				{Index: 1, Version: 6, Quantity: 2, Blueprint: "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C", CrafterCharacterName: "Survivor"},
			},
			Dinos: []arkcluster.Dino{{Index: 0, Version: 7, Archive: &arkarchive.Archive{Objects: []arkarchive.Object{{ClassName: "/Game/Test/Dino.Dino_C"}}}}},
		},
		{
			ID:    "EOS_two",
			Path:  "/tmp/EOS_two",
			Items: []arkcluster.Item{{Index: 0, Blueprint: "/Game/Test/PrimalItemResource_Custom.PrimalItemResource_Custom_C", Quantity: 4}},
			Dinos: []arkcluster.Dino{{Index: 0, Version: 7, ParseError: "unsupported embedded archive"}},
		},
	}

	summary := ClusterDirectorySummary(entries)
	if summary.Files != 2 || summary.Items != 3 || summary.Dinos != 2 || summary.ParseErrors != 1 || summary.Objects != 1 {
		t.Fatalf("ClusterDirectorySummary() totals = %#v", summary)
	}
	if summary.ItemSummary.DinoItems != 1 || summary.ItemSummary.EquipmentItems != 1 || summary.ItemSummary.OtherItems != 1 || summary.ItemSummary.TotalQuantity != 9 || summary.ItemSummary.AverageQuantity != 3 || summary.ItemSummary.CraftedItems != 1 {
		t.Fatalf("ClusterDirectorySummary() item summary = %#v", summary.ItemSummary)
	}
	if summary.DinoSummary.ParsedDinos != 1 || summary.DinoSummary.ParseErrorDinos != 1 || summary.DinoSummary.TotalEmbeddedObjects != 1 {
		t.Fatalf("ClusterDirectorySummary() dino summary = %#v", summary.DinoSummary)
	}
}

func TestClusterDirectorySummaryAggregatesEmbeddedDinoIdentityAndStats(t *testing.T) {
	entries := []*arkcluster.Data{
		{
			ID: "EOS_one",
			Dinos: []arkcluster.Dino{{
				Index:   0,
				Version: 7,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					clusterDinoObjectForSummaryTest(1001, 2002, true, false, false, false, true),
					clusterStatusObjectForSummaryTest(),
				}},
			}},
		},
		{
			ID: "EOS_two",
			Dinos: []arkcluster.Dino{{
				Index:   0,
				Version: 7,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					clusterDinoObjectForSummaryTest(3003, 4004, false, true, true, true, false),
				}},
			}},
		},
	}

	summary := ClusterDirectorySummary(entries)
	if summary.DinoSummary.WithDinoID != 2 || summary.DinoSummary.TamedDinos != 1 || summary.DinoSummary.FemaleDinos != 1 || summary.DinoSummary.BabyDinos != 1 || summary.DinoSummary.DeadDinos != 1 || summary.DinoSummary.WithStats != 1 {
		t.Fatalf("ClusterDirectorySummary() dino identity/stat summary = %#v, want aggregated embedded dino counts", summary.DinoSummary)
	}
	if summary.DinoSummary.TotalBaseLevel != 12 || summary.DinoSummary.MaxBaseLevel != 12 || summary.DinoSummary.AverageBaseLevel != 12 || summary.DinoSummary.TotalCurrentLevel != 6 || summary.DinoSummary.MaxCurrentLevel != 6 || summary.DinoSummary.AverageCurrentLevel != 6 {
		t.Fatalf("ClusterDirectorySummary() level aggregates = %#v, want base/current level aggregates from embedded status", summary.DinoSummary)
	}
}

func clusterDinoObjectForSummaryTest(id1, id2 uint32, tamed bool, female bool, baby bool, dead bool, withStatus bool) arkarchive.Object {
	properties := []arkproperty.Property{
		{Name: "DinoID1", Type: arkproperty.TypeUInt32, Value: id1},
		{Name: "DinoID2", Type: arkproperty.TypeUInt32, Value: id2},
	}
	if tamed {
		properties = append(properties, arkproperty.Property{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)})
	}
	if female {
		properties = append(properties, arkproperty.Property{Name: "bIsFemale", Type: arkproperty.TypeBool, Value: true})
	}
	if baby {
		properties = append(properties, arkproperty.Property{Name: "bIsBaby", Type: arkproperty.TypeBool, Value: true})
	}
	if dead {
		properties = append(properties, arkproperty.Property{Name: "bIsDead", Type: arkproperty.TypeBool, Value: true})
	}
	if withStatus {
		properties = append(properties, arkproperty.Property{Name: "MyCharacterStatusComponent", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{Type: arkproperty.ObjectReferenceUUID, Value: clusterStatusIDForSummaryTest()}})
	}
	return arkarchive.Object{
		UUID:       uuid.New(),
		ClassName:  "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
		Properties: properties,
	}
}

func clusterStatusObjectForSummaryTest() arkarchive.Object {
	return arkarchive.Object{
		UUID:      clusterStatusIDForSummaryTest(),
		ClassName: "/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C",
		Properties: []arkproperty.Property{
			{Name: "BaseCharacterLevel", Type: arkproperty.TypeInt, Value: int32(12)},
			{Name: "NumberOfLevelUpPointsApplied", Type: arkproperty.TypeInt, Position: 0, Value: int32(5)},
		},
	}
}

func clusterStatusIDForSummaryTest() uuid.UUID {
	return uuid.MustParse("11112222-3333-4444-5555-666677778888")
}
