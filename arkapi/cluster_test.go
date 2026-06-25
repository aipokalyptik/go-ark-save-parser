package arkapi

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

func TestClusterAPIClassifiesAndCountsItems(t *testing.T) {
	api := NewCluster(&arkcluster.Data{
		ID:   "EOS_abc123",
		Path: "/tmp/EOS_abc123",
		Items: []arkcluster.Item{
			{
				Index:     0,
				Blueprint: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
				Properties: arkproperty.Container{Properties: []arkproperty.Property{{
					Name:  "CustomItemDatas",
					Type:  arkproperty.TypeArray,
					Value: arkproperty.Array{Values: []any{arkproperty.Container{}}},
				}}},
			},
			{
				Index:     1,
				Blueprint: "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C",
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
	if got := api.ItemsByType("missing"); len(got) != 0 {
		t.Fatalf("ItemsByType(missing) = %#v, want empty", got)
	}
	typed := api.ItemsTyped()
	if len(typed) != 3 || typed[0].Type != "dino" || typed[1].Type != "equipment" || typed[2].Type != "other" {
		t.Fatalf("ItemsTyped() = %#v, want dino/equipment/other projections", typed)
	}
	if got := api.ItemsByTypeTyped("dino"); len(got) != 1 || got[0].Index != 0 || got[0].Type != "dino" {
		t.Fatalf("ItemsByTypeTyped(dino) = %#v, want typed item index 0", got)
	}
}

func TestClusterAPISummarizesDinoParseStatus(t *testing.T) {
	api := NewCluster(&arkcluster.Data{
		ID:      "EOS_abc123",
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Dinos: []arkcluster.Dino{
			{
				Index: 0,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					{ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C"},
				}},
			},
			{
				Index:      1,
				ParseError: "unsupported embedded archive",
			},
		},
	})

	if got := api.ParseErrorCount(); got != 1 {
		t.Fatalf("ParseErrorCount() = %d, want 1", got)
	}
	if got := api.DinosByParseStatus(true); len(got) != 1 || got[0].Index != 0 {
		t.Fatalf("DinosByParseStatus(true) = %#v, want parsed dino 0", got)
	}
	if got := api.DinosByParseStatus(false); len(got) != 1 || got[0].Index != 1 {
		t.Fatalf("DinosByParseStatus(false) = %#v, want failed dino 1", got)
	}
	typed := api.DinosTyped()
	if len(typed) != 2 || typed[0].ObjectCount != 1 || len(typed[0].ClassNames) != 1 || typed[0].ClassNames[0] != "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C" {
		t.Fatalf("DinosTyped() = %#v, want class-name projection for parsed dino", typed)
	}
	if got := api.DinosByParseStatusTyped(false); len(got) != 1 || got[0].Index != 1 || got[0].ParseError != "unsupported embedded archive" {
		t.Fatalf("DinosByParseStatusTyped(false) = %#v, want failed typed dino 1", got)
	}
	summary := api.Summary()
	if summary.ID != "EOS_abc123" || summary.ItemCount != 0 || summary.DinoCount != 2 || summary.ParseErrorCount != 1 || summary.ObjectCount != 1 {
		t.Fatalf("Summary() = %#v", summary)
	}
}
