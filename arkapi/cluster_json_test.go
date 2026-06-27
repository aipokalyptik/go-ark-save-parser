package arkapi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
)

func TestExportClusterDataSummarizesUploads(t *testing.T) {
	data := &arkcluster.Data{
		ID:      "EOS_abc123",
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Items: []arkcluster.Item{{
			Index:                0,
			Version:              7,
			UploadTime:           12345,
			Blueprint:            "/Game/Test/Item.Item_C",
			Quantity:             2,
			Rating:               7.5,
			Quality:              2,
			CrafterCharacterName: "Survivor",
			CrafterTribeName:     "Porters",
		}},
		Dinos: []arkcluster.Dino{{
			Index:      0,
			Version:    7,
			UploadTime: 67890,
			RawSize:    128,
			ParseError: "unsupported archive version",
			Archive:    &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Game/Test/Dino.Dino_C"}}},
		}},
	}

	info := ExportClusterData(data)
	if info.ID != "EOS_abc123" || info.ArchiveVersion != 7 || info.ObjectCount != 1 {
		t.Fatalf("ClusterDataInfo metadata = %#v", info)
	}
	if info.ItemCount != 1 || len(info.Items) != 1 || info.Items[0].Blueprint != "/Game/Test/Item.Item_C" {
		t.Fatalf("ClusterDataInfo items = %#v", info.Items)
	}
	if info.Items[0].Rating != 7.5 || info.Items[0].Quality != 2 || info.Items[0].CrafterCharacterName != "Survivor" || info.Items[0].CrafterTribeName != "Porters" {
		t.Fatalf("ClusterDataInfo item metadata = %#v", info.Items[0])
	}
	if info.Items[0].Type != "other" {
		t.Fatalf("ClusterDataInfo item type = %q, want other", info.Items[0].Type)
	}
	if info.DinoCount != 1 || len(info.Dinos) != 1 || info.Dinos[0].ObjectCount != 1 {
		t.Fatalf("ClusterDataInfo dinos = %#v", info.Dinos)
	}
	if info.Dinos[0].ParseError != "unsupported archive version" {
		t.Fatalf("ClusterDataInfo dino ParseError = %q", info.Dinos[0].ParseError)
	}
}

func TestExportClusterDataIncludesUploadedDinoClassNames(t *testing.T) {
	data := &arkcluster.Data{
		ID:   "EOS_abc123",
		Path: "/tmp/EOS_abc123",
		Dinos: []arkcluster.Dino{{
			Index:   0,
			Version: 7,
			Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
				{ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C"},
				{ClassName: "/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C"},
				{ClassName: ""},
			}},
		}},
	}

	info := ExportClusterData(data)
	if len(info.Dinos) != 1 {
		t.Fatalf("Dinos length = %d, want 1", len(info.Dinos))
	}
	want := []string{
		"/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
		"/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C",
	}
	if !reflect.DeepEqual(info.Dinos[0].ClassNames, want) {
		t.Fatalf("ClassNames = %#v, want %#v", info.Dinos[0].ClassNames, want)
	}
	if info.Dinos[0].ParseStatus != "parsed" || !info.Dinos[0].ParsedArchive || !info.Dinos[0].SupportedVersion || info.Dinos[0].UnsupportedVersion {
		t.Fatalf("typed dino status fields = %#v, want parsed supported archive", info.Dinos[0])
	}
	if !reflect.DeepEqual(info.Dinos[0].StatusComponentClassNames, []string{"/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C"}) {
		t.Fatalf("StatusComponentClassNames = %#v, want status component class", info.Dinos[0].StatusComponentClassNames)
	}
}

func TestExportClusterDataIncludesTypedDinoStatusFields(t *testing.T) {
	data := &arkcluster.Data{
		ID:   "EOS_abc123",
		Path: "/tmp/EOS_abc123",
		Items: []arkcluster.Item{{
			Index:     0,
			Version:   6,
			Blueprint: "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C",
		}},
		Dinos: []arkcluster.Dino{
			{
				Index:   0,
				Version: 6,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
					{ClassName: "/Game/Test/Dino.Dino_C"},
					{ClassName: "/Game/Test/Dino_AIController_BP.Dino_AIController_BP_C"},
					{ClassName: "/Game/Test/DinoInventoryComponent_BP.DinoInventoryComponent_BP_C"},
				}},
			},
			{
				Index:      1,
				Version:    7,
				ParseError: "unsupported embedded archive",
			},
			{
				Index:   2,
				Version: 7,
			},
		},
	}

	info := ExportClusterData(data)
	if len(info.Items) != 1 || info.Items[0].Type != "equipment" || info.Items[0].SupportedVersion || !info.Items[0].UnsupportedVersion {
		t.Fatalf("typed item status fields = %#v, want unsupported equipment item", info.Items)
	}
	if len(info.Dinos) != 3 {
		t.Fatalf("Dinos length = %d, want 3", len(info.Dinos))
	}
	if info.Dinos[0].ParseStatus != "unsupported_version" || !info.Dinos[0].ParsedArchive || info.Dinos[0].SupportedVersion || !info.Dinos[0].UnsupportedVersion {
		t.Fatalf("unsupported-version dino info = %#v", info.Dinos[0])
	}
	if len(info.Dinos[0].AIControllerClassNames) != 1 || len(info.Dinos[0].InventoryComponentClassNames) != 1 {
		t.Fatalf("component summaries = %#v, want AI and inventory component classes", info.Dinos[0])
	}
	if info.Dinos[1].ParseStatus != "parse_error" || info.Dinos[1].ParsedArchive || info.Dinos[1].ParseError == "" {
		t.Fatalf("parse-error dino info = %#v", info.Dinos[1])
	}
	if info.Dinos[2].ParseStatus != "unparsed" || info.Dinos[2].ParsedArchive || !info.Dinos[2].SupportedVersion {
		t.Fatalf("unparsed dino info = %#v", info.Dinos[2])
	}
}

func TestExportClusterDataClassifiesUploadedItems(t *testing.T) {
	data := &arkcluster.Data{
		ID:   "EOS_abc123",
		Path: "/tmp/EOS_abc123",
		Items: []arkcluster.Item{
			{
				Index:     0,
				Blueprint: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
				Properties: arkproperty.Container{Properties: []arkproperty.Property{{
					Name: "CustomItemDatas",
					Type: arkproperty.TypeArray,
					Value: arkproperty.Array{
						ElementType: arkproperty.TypeStruct,
						StructType:  "CustomItemData",
						Values:      []any{arkproperty.Container{}},
					},
				}}},
			},
			{
				Index:     1,
				Blueprint: "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C",
			},
			{
				Index:     2,
				Blueprint: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
			},
			{
				Index:     3,
				Blueprint: "/Game/Test/PrimalItemResource_Custom.PrimalItemResource_Custom_C",
			},
		},
	}

	info := ExportClusterData(data)
	if len(info.Items) != 4 {
		t.Fatalf("Items length = %d, want 4", len(info.Items))
	}
	want := []string{"dino", "equipment", "other", "other"}
	for i, wantType := range want {
		if info.Items[i].Type != wantType {
			t.Fatalf("item %d type = %q, want %q", i, info.Items[i].Type, wantType)
		}
	}
}

func TestExportClusterDataJSONIsDeterministic(t *testing.T) {
	data := &arkcluster.Data{
		ID:      "EOS_abc123",
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7},
		Items: []arkcluster.Item{{
			Index:      0,
			Version:    7,
			UploadTime: 12345,
			Blueprint:  "/Game/Test/Item.Item_C",
			Quantity:   2,
		}},
	}

	raw, err := ExportClusterDataJSON(data)
	if err != nil {
		t.Fatalf("ExportClusterDataJSON() error = %v", err)
	}
	var decoded ClusterDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ItemCount != 1 || len(decoded.Items) != 1 || decoded.Items[0].Quantity != 2 {
		t.Fatalf("decoded ClusterDataInfo = %#v", decoded)
	}
}
