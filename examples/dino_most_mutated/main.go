package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dino_most_mutated <save.ark>")
		os.Exit(2)
	}

	summary, err := arkapi.DinoMostMutatedSummaryFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dinos: %v\n", err)
		os.Exit(1)
	}
	if !summary.Found {
		fmt.Println("no_match")
		return
	}

	fmt.Printf("has_most_mutated=1 mutations=%d level=%d\n", summary.MutationPairs, summary.Level)
}
