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

func main() {
	if len(os.Args) < 3 || len(os.Args) > 5 {
		log.Fatalf("usage: %s <save.ark> <out.json> [resolution] [min-in-cell]", os.Args[0])
	}
	resolution := 100
	if len(os.Args) >= 4 {
		value, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal(err)
		}
		resolution = value
	}
	minInCell := 1
	if len(os.Args) == 5 {
		value, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatal(err)
		}
		minInCell = value
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewStructure(save)
	structures, faults, err := api.AllWithFaults()
	if err != nil {
		log.Fatal(err)
	}
	heatmap, err := api.Heatmap(save.Context.MapName, resolution, structures, nil, nil, minInCell)
	if err != nil {
		log.Fatal(err)
	}
	summary := arkapi.SummarizeHeatmap(heatmap, len(faults))
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
