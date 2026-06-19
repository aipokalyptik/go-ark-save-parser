package arkapi

import (
	"encoding/json"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
)

func TestExportClusterDataSummarizesUploads(t *testing.T) {
	data := &arkcluster.Data{
		ID:      "EOS_abc123",
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Items: []arkcluster.Item{{
			Index:      0,
			Version:    7,
			UploadTime: 12345,
			Blueprint:  "/Game/Test/Item.Item_C",
			Quantity:   2,
		}},
		Dinos: []arkcluster.Dino{{
			Index:      0,
			Version:    7,
			UploadTime: 67890,
			RawSize:    128,
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
	if info.DinoCount != 1 || len(info.Dinos) != 1 || info.Dinos[0].ObjectCount != 1 {
		t.Fatalf("ClusterDataInfo dinos = %#v", info.Dinos)
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
