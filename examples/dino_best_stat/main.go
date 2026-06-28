package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	args := os.Args[1:]
	opts := arkapi.DinoBestStatOptions{}
	if len(args) > 0 && args[0] == "--no-cryos" {
		opts.ExcludeCryopods = true
		args = args[1:]
	}
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: dino_best_stat [--no-cryos] <save.ark>")
		os.Exit(2)
	}

	summary, _, err := arkapi.DinoBestStatSummaryFromPath(args[0], opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dinos: %v\n", err)
		os.Exit(1)
	}
	if !summary.Found {
		fmt.Println("no_match")
		return
	}
	fmt.Printf("uuid=%s blueprint=%q stat=%s points=%d level=%d\n", summary.UUID, summary.Dino.Blueprint, summary.Stat.String(), summary.Points, summary.Level)
}
