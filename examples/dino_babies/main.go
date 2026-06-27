package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dino_babies <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	counts, _, err := api.BabySummaryWithFaults(arkapi.BabyFilterOptions{
		IncludeTamed:      true,
		IncludeCryopodded: true,
		IncludeWild:       true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read babies: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wild_babies=%d tamed_babies=%d\n", counts.Wild, counts.Tamed)
}
