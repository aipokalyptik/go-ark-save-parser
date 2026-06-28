package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dino_wild_tamed <save.ark>")
		os.Exit(2)
	}

	summary, _, err := arkapi.DinoWildTamedSummaryFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read wild-tamed dinos: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wild_tamed=%d max_level=%d\n", summary.Dinos, summary.MaxLevel)
}
