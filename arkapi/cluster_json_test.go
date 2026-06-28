package arkapi

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
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
	if info.Items[0].ShortName != "Item" {
		t.Fatalf("ClusterDataInfo item short name = %q, want Item", info.Items[0].ShortName)
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
	if info.Dinos[0].PrimaryClassName != "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C" || info.Dinos[0].ShortName != "Raptor" {
		t.Fatalf("ClusterDataInfo dino primary/short name = %q/%q, want raptor class/Raptor", info.Dinos[0].PrimaryClassName, info.Dinos[0].ShortName)
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

func TestExportClusterDirectoryDataIncludesAggregateSummary(t *testing.T) {
	entries := []*arkcluster.Data{
		{
			ID:      "EOS_one",
			Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
			Items: []arkcluster.Item{{
				Index:     0,
				Version:   7,
				Quantity:  2,
				Blueprint: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
				Properties: arkproperty.Container{Properties: []arkproperty.Property{{
					Name:  "CustomItemDatas",
					Type:  arkproperty.TypeArray,
					Value: arkproperty.Array{Values: []any{arkproperty.Container{}}},
				}}},
			}},
			Dinos: []arkcluster.Dino{{Index: 0, Version: 7, Archive: &arkarchive.Archive{Objects: []arkarchive.Object{{ClassName: "/Game/Test/Dino.Dino_C"}}}}},
		},
		{
			ID:    "EOS_two",
			Items: []arkcluster.Item{{Index: 0, Blueprint: "/Game/Test/PrimalItemResource_Custom.PrimalItemResource_Custom_C", Quantity: 5}},
			Dinos: []arkcluster.Dino{{Index: 0, Version: 7, ParseError: "unsupported embedded archive"}},
		},
	}

	info := ExportClusterDirectoryData(entries)
	if info.Count != 2 || len(info.Files) != 2 {
		t.Fatalf("ClusterDirectoryInfo = %#v, want two files", info)
	}
	if info.Summary.Files != 2 || info.Summary.Items != 2 || info.Summary.Dinos != 2 || info.Summary.ParseErrors != 1 || info.Summary.Objects != 1 {
		t.Fatalf("ClusterDirectoryInfo.Summary totals = %#v", info.Summary)
	}
	if info.Summary.ItemSummary.TotalQuantity != 7 || info.Summary.ItemSummary.DinoItems != 1 || info.Summary.ItemSummary.OtherItems != 1 {
		t.Fatalf("ClusterDirectoryInfo.Summary item summary = %#v", info.Summary.ItemSummary)
	}
	if info.Summary.DinoSummary.ParsedDinos != 1 || info.Summary.DinoSummary.ParseErrorDinos != 1 {
		t.Fatalf("ClusterDirectoryInfo.Summary dino summary = %#v", info.Summary.DinoSummary)
	}
	raw, err := ExportClusterDirectoryDataJSON(entries)
	if err != nil {
		t.Fatalf("ExportClusterDirectoryDataJSON() error = %v", err)
	}
	var decoded ClusterDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.Summary.ItemSummary.TotalQuantity != 7 {
		t.Fatalf("decoded summary = %#v, want total quantity 7", decoded.Summary)
	}
}

func TestExportClusterDirectoryDataWithFaultsReportsBrokenFiles(t *testing.T) {
	entries := []*arkcluster.Data{{
		ID:      "EOS_valid",
		Path:    "/tmp/EOS_valid",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
	}}
	faults := []arkcluster.FileFault{{
		Path: "/tmp/EOS_broken",
		Err:  errors.New("not an archive"),
	}}

	info := ExportClusterDirectoryDataWithFaults(entries, faults)
	if info.Count != 1 || len(info.Files) != 1 || info.Summary.Files != 1 {
		t.Fatalf("ClusterDirectoryInfo = %#v, want one valid file", info)
	}
	if len(info.Faults) != 1 || info.Faults[0].Path != "/tmp/EOS_broken" || info.Faults[0].Error != "not an archive" {
		t.Fatalf("ClusterDirectoryInfo.Faults = %#v, want broken file fault", info.Faults)
	}

	raw, err := ExportClusterDirectoryDataWithFaultsJSON(entries, faults)
	if err != nil {
		t.Fatalf("ExportClusterDirectoryDataWithFaultsJSON() error = %v", err)
	}
	var decoded ClusterDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Error != "not an archive" {
		t.Fatalf("decoded faults = %#v, want one serialized fault", decoded.Faults)
	}
}

func TestClusterDirectorySummaryFromPathReadsDirectoryWithFaults(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "EOS_valid")
	brokenPath := filepath.Join(dir, "EOS_broken")
	testfixtures.WriteArchive(t, validPath, "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(brokenPath, []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken cluster file: %v", err)
	}

	info, err := ClusterDirectorySummaryFromPath(dir)
	if err != nil {
		t.Fatalf("ClusterDirectorySummaryFromPath() error = %v", err)
	}
	if info.Count != 1 || len(info.Files) != 1 || info.Files[0].ID != "EOS_valid" {
		t.Fatalf("ClusterDirectorySummaryFromPath() = %#v, want one valid cluster file", info)
	}
	if info.Summary.Files != 1 || info.Summary.Objects != 1 {
		t.Fatalf("ClusterDirectorySummaryFromPath() summary = %#v, want one archive object", info.Summary)
	}
	if len(info.Faults) != 1 || info.Faults[0].Path != brokenPath || info.Faults[0].Error == "" {
		t.Fatalf("ClusterDirectorySummaryFromPath() faults = %#v, want broken file fault", info.Faults)
	}
}

func TestExportClusterPathJSONReadsDirectoryWithFaults(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_valid"), "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(filepath.Join(dir, "EOS_broken"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken cluster file: %v", err)
	}

	raw, err := ExportClusterPathJSON(dir)
	if err != nil {
		t.Fatalf("ExportClusterPathJSON(directory) error = %v", err)
	}
	var decoded ClusterDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 1 || len(decoded.Files) != 1 || decoded.Files[0].ID != "EOS_valid" {
		t.Fatalf("decoded directory = %#v, want one valid cluster file", decoded)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Error == "" {
		t.Fatalf("decoded faults = %#v, want malformed file fault", decoded.Faults)
	}
}
