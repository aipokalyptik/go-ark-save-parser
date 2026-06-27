package arkapi

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
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
				CrafterCharacterName: "Survivor",
				CrafterTribeName:     "Porters",
			},
			{
				Index:     2,
				Blueprint: "/Game/Test/PrimalItemResource_Custom.PrimalItemResource_Custom_C",
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
	if summary.CraftedItems != 1 || summary.TotalQuantity != 0 || summary.MaxRating != 0 || summary.MaxQuality != 0 {
		t.Fatalf("ItemSummary() aggregates = %#v, want one crafted item and zero quantity/rating/quality", summary)
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
	if summary.ItemSummary.DinoItems != 1 || summary.ItemSummary.EquipmentItems != 1 || summary.ItemSummary.OtherItems != 1 || summary.ItemSummary.TotalQuantity != 9 || summary.ItemSummary.CraftedItems != 1 {
		t.Fatalf("ClusterDirectorySummary() item summary = %#v", summary.ItemSummary)
	}
	if summary.DinoSummary.ParsedDinos != 1 || summary.DinoSummary.ParseErrorDinos != 1 || summary.DinoSummary.TotalEmbeddedObjects != 1 {
		t.Fatalf("ClusterDirectorySummary() dino summary = %#v", summary.DinoSummary)
	}
}
