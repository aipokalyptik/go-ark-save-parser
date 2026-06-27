package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

type historyReport struct {
	Saves        int             `json:"saves"`
	InitialCount int             `json:"initial_count"`
	Changes      []historyChange `json:"changes"`
	FinalCount   int             `json:"final_count"`
}

type historyChange struct {
	Save    string `json:"save"`
	Added   int    `json:"added"`
	Removed int    `json:"removed"`
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <ark-files.json> <out.json>", os.Args[0])
	}
	paths, err := readManifest(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	report := historyReport{Saves: len(paths)}
	var previous map[string]struct{}
	for index, path := range paths {
		current, err := arkapi.EquipmentHistorySnapshotFromPath(path)
		if err != nil {
			log.Fatal(err)
		}
		if index == 0 {
			report.InitialCount = len(current)
		} else {
			added, removed := arkapi.DiffEquipmentHistorySnapshots(previous, current)
			if added != 0 || removed != 0 {
				report.Changes = append(report.Changes, historyChange{
					Save:    filepath.Base(path),
					Added:   added,
					Removed: removed,
				})
			}
		}
		previous = current
	}
	report.FinalCount = len(previous)
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(os.Args[2], data, 0o600); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("saves=%d initial=%d changes=%d final=%d wrote=%s\n", report.Saves, report.InitialCount, len(report.Changes), report.FinalCount, os.Args[2])
}

func readManifest(path string) ([]string, error) {
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
