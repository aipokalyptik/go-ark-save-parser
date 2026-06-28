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

	summary, err := arkapi.ExportStructureHeatmapSummaryJSONFromPath(os.Args[1], os.Args[2], arkapi.StructureHeatmapOptions{
		Resolution:   resolution,
		MinInSection: minInCell,
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("cells=%d total=%d max=%d faults=%d wrote=%s\n", summary.NonzeroCells, summary.Total, summary.Max, summary.Faults, os.Args[2])
}
