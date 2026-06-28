package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: base_components <save.ark>")
		os.Exit(2)
	}

	stats, err := arkapi.BaseComponentStatsFromPath(os.Args[1], "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read base components: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("bases=%d total_structures=%d largest=%d min10=%d faults=%d\n", stats.Components, stats.TotalStructures, stats.LargestComponent, stats.ComponentsAtLeast10, stats.Faults)
}
