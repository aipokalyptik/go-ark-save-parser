package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
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

	summary, err := arkapi.ExportDinoHeatmapSummaryJSONFromPath(args[0], args[1], arkapi.DinoHeatmapOptions{
		Resolution:        resolution,
		IncludeCryopodded: includeCryos,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("cells=%d total=%d max=%d faults=%d skipped_coordinates=%d wrote=%s\n", summary.NonzeroCells, summary.Total, summary.Max, summary.Faults, summary.SkippedCoordinates, args[1])
}
