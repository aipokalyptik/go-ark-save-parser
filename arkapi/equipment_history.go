package arkapi

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type EquipmentHistoryReport struct {
	Saves        int                      `json:"saves"`
	InitialCount int                      `json:"initial_count"`
	Changes      []EquipmentHistoryChange `json:"changes"`
	FinalCount   int                      `json:"final_count"`
}

type EquipmentHistoryChange struct {
	Save    string `json:"save"`
	Added   int    `json:"added"`
	Removed int    `json:"removed"`
}

func EquipmentHistoryReportFromManifestPath(path string) (EquipmentHistoryReport, error) {
	paths, err := equipmentHistoryManifestPaths(path)
	if err != nil {
		return EquipmentHistoryReport{}, err
	}
	report := EquipmentHistoryReport{Saves: len(paths)}
	var previous map[string]struct{}
	for index, savePath := range paths {
		current, err := EquipmentHistorySnapshotFromPath(savePath)
		if err != nil {
			return EquipmentHistoryReport{}, err
		}
		if index == 0 {
			report.InitialCount = len(current)
		} else {
			added, removed := DiffEquipmentHistorySnapshots(previous, current)
			if added != 0 || removed != 0 {
				report.Changes = append(report.Changes, EquipmentHistoryChange{
					Save:    filepath.Base(savePath),
					Added:   added,
					Removed: removed,
				})
			}
		}
		previous = current
	}
	report.FinalCount = len(previous)
	return report, nil
}

func equipmentHistoryManifestPaths(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		return nil, err
	}
	return paths, nil
}
