package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: structure_health <save.ark>")
		os.Exit(2)
	}

	api, closeAPI, err := arkapi.NewStructureFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	summary, faults, err := api.HealthSummaryWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read structure health: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(
		"structures=%d with_health=%d damaged=%d repaired=%d without_max_health=%d avg_health=%.1f min_health=%.1f max_health=%.1f faults=%d\n",
		summary.Structures,
		summary.WithHealth,
		summary.Damaged,
		summary.FullyRepaired,
		summary.WithoutMaxHealth,
		summary.AverageHealthPercent,
		summary.MinimumHealthPercent,
		summary.MaximumHealthPercent,
		len(faults),
	)
}
