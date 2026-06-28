package arkapi

import (
	"encoding/json"
	"os"

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
	Count  int                    `json:"count"`
	Faults []TributeFileFaultInfo `json:"faults,omitempty"`
	Files  []TributeDataInfo      `json:"files"`
}

type TributeFileFaultInfo struct {
	Path  string `json:"path"`
	Error string `json:"error"`
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
	return ExportTributeDirectoryDataWithFaults(entries, nil)
}

func ExportTributeDirectoryDataWithFaults(entries []*arktribute.Data, faults []arktribute.FileFault) TributeDirectoryInfo {
	info := TributeDirectoryInfo{
		Count:  len(entries),
		Files:  make([]TributeDataInfo, 0, len(entries)),
		Faults: make([]TributeFileFaultInfo, 0, len(faults)),
	}
	for _, entry := range entries {
		info.Files = append(info.Files, ExportTributeData(entry))
	}
	for _, fault := range faults {
		if fault.Err == nil {
			continue
		}
		info.Faults = append(info.Faults, TributeFileFaultInfo{
			Path:  fault.Path,
			Error: fault.Err.Error(),
		})
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
	entries, faults, err := arktribute.OpenDirectoryWithFaults(path)
	if err != nil {
		return TributeDirectoryInfo{}, err
	}
	return ExportTributeDirectoryDataWithFaults(entries, faults), nil
}

func ExportTributePathJSON(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		summary, err := TributeDirectorySummaryFromPath(path)
		if err != nil {
			return nil, err
		}
		return json.MarshalIndent(summary, "", "  ")
	}
	summary, err := TributeSummaryFromPath(path)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(summary, "", "  ")
}

func ExportTributeDataJSON(data *arktribute.Data) ([]byte, error) {
	return json.MarshalIndent(ExportTributeData(data), "", "  ")
}

func ExportTributeDirectoryDataJSON(entries []*arktribute.Data) ([]byte, error) {
	return ExportTributeDirectoryDataWithFaultsJSON(entries, nil)
}

func ExportTributeDirectoryDataWithFaultsJSON(entries []*arktribute.Data, faults []arktribute.FileFault) ([]byte, error) {
	return json.MarshalIndent(ExportTributeDirectoryDataWithFaults(entries, faults), "", "  ")
}
