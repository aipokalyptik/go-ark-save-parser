package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

type heatmapSummary struct {
	Resolution   int `json:"resolution"`
	NonzeroCells int `json:"nonzero_cells"`
	Total        int `json:"total"`
	Max          int `json:"max"`
	Faults       int `json:"faults"`
}

func main() {
	if len(os.Args) < 3 || len(os.Args) > 4 {
		log.Fatalf("usage: %s <save.ark> <out.json> [resolution]", os.Args[0])
	}
	resolution := 100
	if len(os.Args) == 4 {
		value, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
		resolution = value
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	dinos, faults, err := api.AllWithFaults()
	if err != nil {
		log.Fatal(err)
	}
	heatmap, err := api.Heatmap(save.Context.MapName, resolution, dinos, nil, false)
	if err != nil {
		log.Fatal(err)
	}
	summary := summarizeHeatmap(heatmap, len(faults))
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(os.Args[2], data, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("cells=%d total=%d max=%d faults=%d wrote=%s\n", summary.NonzeroCells, summary.Total, summary.Max, summary.Faults, os.Args[2])
}

func summarizeHeatmap(heatmap [][]int, faults int) heatmapSummary {
	summary := heatmapSummary{Resolution: len(heatmap), Faults: faults}
	for _, row := range heatmap {
		for _, value := range row {
			if value == 0 {
				continue
			}
			summary.NonzeroCells++
			summary.Total += value
			if value > summary.Max {
				summary.Max = value
			}
		}
	}
	return summary
}
