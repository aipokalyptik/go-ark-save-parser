package arkapi

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
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

func TestExportTributeDirectoryDataWithFaultsReportsBrokenFiles(t *testing.T) {
	entries := []*arktribute.Data{{
		ID:            "a",
		Path:          "/tmp/a.arktributetribe",
		PlayerDataIDs: []uint64{1},
	}}
	faults := []arktribute.FileFault{{
		Path: "/tmp/broken.arktributetribe",
		Err:  errors.New("bad tribute"),
	}}

	info := ExportTributeDirectoryDataWithFaults(entries, faults)
	if info.Count != 1 || len(info.Files) != 1 {
		t.Fatalf("TributeDirectoryInfo = %#v, want one valid file", info)
	}
	if len(info.Faults) != 1 || info.Faults[0].Path != "/tmp/broken.arktributetribe" || info.Faults[0].Error != "bad tribute" {
		t.Fatalf("TributeDirectoryInfo.Faults = %#v, want broken file fault", info.Faults)
	}

	raw, err := ExportTributeDirectoryDataWithFaultsJSON(entries, faults)
	if err != nil {
		t.Fatalf("ExportTributeDirectoryDataWithFaultsJSON() error = %v", err)
	}
	var decoded TributeDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Error != "bad tribute" {
		t.Fatalf("decoded faults = %#v, want one serialized fault", decoded.Faults)
	}
}

func TestTributeSummaryFromPathReadsFileOrDirectory(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "abc.arktributetribe")
	second := filepath.Join(dir, "def.arktributetribetribe")
	testfixtures.WriteTributeFile(t, first, []uint64{11, 22}, []uint64{33})
	testfixtures.WriteTributeFile(t, second, []uint64{44}, []uint64{55, 66})

	fileInfo, err := TributeSummaryFromPath(first)
	if err != nil {
		t.Fatalf("TributeSummaryFromPath(file) error = %v", err)
	}
	if fileInfo.ID != "abc" || fileInfo.PlayerDataCount != 2 || fileInfo.TribeDataCount != 1 {
		t.Fatalf("TributeSummaryFromPath(file) = %#v", fileInfo)
	}

	dirInfo, err := TributeDirectorySummaryFromPath(dir)
	if err != nil {
		t.Fatalf("TributeDirectorySummaryFromPath() error = %v", err)
	}
	if dirInfo.Count != 2 || len(dirInfo.Files) != 2 {
		t.Fatalf("TributeDirectorySummaryFromPath() = %#v, want two files", dirInfo)
	}
	if dirInfo.Files[0].PlayerDataCount+dirInfo.Files[1].PlayerDataCount != 3 ||
		dirInfo.Files[0].TribeDataCount+dirInfo.Files[1].TribeDataCount != 3 {
		t.Fatalf("TributeDirectorySummaryFromPath() counts = %#v", dirInfo)
	}
}

func TestExportTributePathJSONReadsDirectoryWithFaults(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "abc.arktributetribe"), []uint64{11}, nil)
	if err := os.WriteFile(filepath.Join(dir, "broken.arktributetribe"), []byte("not a tribute index"), 0o600); err != nil {
		t.Fatalf("write broken tribute file: %v", err)
	}

	raw, err := ExportTributePathJSON(dir)
	if err != nil {
		t.Fatalf("ExportTributePathJSON(directory) error = %v", err)
	}
	var decoded TributeDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 1 || len(decoded.Files) != 1 || decoded.Files[0].PlayerDataCount != 1 {
		t.Fatalf("decoded directory = %#v, want one valid tribute file", decoded)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Error == "" {
		t.Fatalf("decoded directory faults = %#v, want malformed file fault", decoded.Faults)
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
