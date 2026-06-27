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
		log.Fatalf("usage: %s [--no-cryos] <save.ark> <out.json> [resolution]", os.Args[0])
	}
	resolution := 100
	includeCryos := true
	args := os.Args[1:]
	if args[0] == "--no-cryos" {
		includeCryos = false
		args = args[1:]
	}
	if len(args) < 2 || len(args) > 3 {
		log.Fatalf("usage: %s [--no-cryos] <save.ark> <out.json> [resolution]", os.Args[0])
	}
	if len(args) == 3 {
		value, err := strconv.Atoi(args[2])
		if err != nil {
			log.Fatal(err)
		}
		resolution = value
	}

	save, err := arksave.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	summary, _, err := api.HeatmapSummaryWithFaults(arkapi.DinoHeatmapOptions{
		MapName:           save.Context.MapName,
		Resolution:        resolution,
		IncludeCryopodded: includeCryos,
	})
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(args[1], data, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("cells=%d total=%d max=%d faults=%d wrote=%s\n", summary.NonzeroCells, summary.Total, summary.Max, summary.Faults, args[1])
}
