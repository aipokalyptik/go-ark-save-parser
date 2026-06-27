package arkapi

import (
	"encoding/json"

	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
)

type TributeDataInfo struct {
	ID              string   `json:"id"`
	Path            string   `json:"path"`
	PlayerDataCount int      `json:"player_data_count"`
	TribeDataCount  int      `json:"tribe_data_count"`
	PlayerDataIDs   []uint64 `json:"player_data_ids,omitempty"`
	TribeDataIDs    []uint64 `json:"tribe_data_ids,omitempty"`
}

type TributeDirectoryInfo struct {
	Count int               `json:"count"`
	Files []TributeDataInfo `json:"files"`
}

func ExportTributeData(data *arktribute.Data) TributeDataInfo {
	return TributeDataInfo{
		ID:              data.ID,
		Path:            data.Path,
		PlayerDataCount: len(data.PlayerDataIDs),
		TribeDataCount:  len(data.TribeDataIDs),
		PlayerDataIDs:   append([]uint64(nil), data.PlayerDataIDs...),
		TribeDataIDs:    append([]uint64(nil), data.TribeDataIDs...),
	}
}

func ExportTributeDirectoryData(entries []*arktribute.Data) TributeDirectoryInfo {
	info := TributeDirectoryInfo{
		Count: len(entries),
		Files: make([]TributeDataInfo, 0, len(entries)),
	}
	for _, entry := range entries {
		info.Files = append(info.Files, ExportTributeData(entry))
	}
	return info
}

func TributeSummaryFromPath(path string) (TributeDataInfo, error) {
	data, err := arktribute.Open(path)
	if err != nil {
		return TributeDataInfo{}, err
	}
	return ExportTributeData(data), nil
}

func TributeDirectorySummaryFromPath(path string) (TributeDirectoryInfo, error) {
	entries, err := arktribute.OpenDirectory(path)
	if err != nil {
		return TributeDirectoryInfo{}, err
	}
	return ExportTributeDirectoryData(entries), nil
}

func ExportTributeDataJSON(data *arktribute.Data) ([]byte, error) {
	return json.MarshalIndent(ExportTributeData(data), "", "  ")
}

func ExportTributeDirectoryDataJSON(entries []*arktribute.Data) ([]byte, error) {
	return json.MarshalIndent(ExportTributeDirectoryData(entries), "", "  ")
}
