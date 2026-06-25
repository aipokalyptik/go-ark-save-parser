package arkapi

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
)

func TestExportTributeDataSummarizesTributeIDs(t *testing.T) {
	data := &arktribute.Data{
		ID:            "abc",
		Path:          "/tmp/abc.arktributetribe",
		PlayerDataIDs: []uint64{11, 22},
		TribeDataIDs:  []uint64{33},
	}

	info := ExportTributeData(data)
	if info.ID != "abc" || info.Path != "/tmp/abc.arktributetribe" {
		t.Fatalf("TributeDataInfo identity = %#v", info)
	}
	if info.PlayerDataCount != 2 || info.TribeDataCount != 1 {
		t.Fatalf("TributeDataInfo counts = %#v", info)
	}
	if !reflect.DeepEqual(info.PlayerDataIDs, []uint64{11, 22}) || !reflect.DeepEqual(info.TribeDataIDs, []uint64{33}) {
		t.Fatalf("TributeDataInfo IDs = %#v", info)
	}
}

func TestExportTributeDirectoryDataSummarizesFiles(t *testing.T) {
	entries := []*arktribute.Data{
		{ID: "a", Path: "/tmp/a.arktributetribe", PlayerDataIDs: []uint64{1}},
		{ID: "b", Path: "/tmp/b.arktributetribetribe", TribeDataIDs: []uint64{2, 3}},
	}

	info := ExportTributeDirectoryData(entries)
	if info.Count != 2 || len(info.Files) != 2 {
		t.Fatalf("TributeDirectoryInfo = %#v, want two files", info)
	}
	if info.Files[0].ID != "a" || info.Files[1].ID != "b" {
		t.Fatalf("TributeDirectoryInfo files = %#v", info.Files)
	}
}

func TestExportTributeDataJSONIsDeterministic(t *testing.T) {
	data := &arktribute.Data{
		ID:            "abc",
		Path:          "/tmp/abc.arktributetribe",
		PlayerDataIDs: []uint64{11},
	}

	raw, err := ExportTributeDataJSON(data)
	if err != nil {
		t.Fatalf("ExportTributeDataJSON() error = %v", err)
	}
	var decoded TributeDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ID != "abc" || decoded.PlayerDataCount != 1 || len(decoded.PlayerDataIDs) != 1 {
		t.Fatalf("decoded TributeDataInfo = %#v", decoded)
	}
}
