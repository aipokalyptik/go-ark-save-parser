package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
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

type equipmentIdentity struct {
	Blueprint   string  `json:"blueprint"`
	Kind        string  `json:"kind"`
	IsBlueprint bool    `json:"is_blueprint"`
	Rating      float64 `json:"rating"`
	Quality     int32   `json:"quality"`
	Damage      float64 `json:"damage,omitempty"`
	Armor       float64 `json:"armor,omitempty"`
	Durability  float64 `json:"durability,omitempty"`
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
		current, err := equipmentSnapshot(path)
		if err != nil {
			log.Fatal(err)
		}
		if index == 0 {
			report.InitialCount = len(current)
		} else {
			added, removed := diffSnapshot(previous, current)
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

func equipmentSnapshot(path string) (map[string]struct{}, error) {
	save, err := arksave.Open(path)
	if err != nil {
		return nil, err
	}
	defer save.Close()

	exported, err := arkapi.NewJSON(save).ExportDomain("equipment")
	if err != nil {
		return nil, err
	}
	items, ok := exported.Items.([]arkapi.EquipmentInfo)
	if !ok {
		return nil, fmt.Errorf("equipment export item type %T", exported.Items)
	}
	out := map[string]struct{}{}
	for _, item := range items {
		identity := equipmentIdentity{
			Blueprint:   item.Blueprint,
			Kind:        item.Kind,
			IsBlueprint: item.IsBlueprint,
			Rating:      item.Rating,
			Quality:     item.Quality,
		}
		if item.Stats != nil {
			identity.Damage = item.Stats.Damage
			identity.Armor = item.Stats.Armor
			identity.Durability = item.Stats.Durability
		}
		data, err := json.Marshal(identity)
		if err != nil {
			return nil, err
		}
		out[string(data)] = struct{}{}
	}
	return out, nil
}

func diffSnapshot(previous map[string]struct{}, current map[string]struct{}) (int, int) {
	added := 0
	for key := range current {
		if _, ok := previous[key]; !ok {
			added++
		}
	}
	removed := 0
	for key := range previous {
		if _, ok := current[key]; !ok {
			removed++
		}
	}
	return added, removed
}
