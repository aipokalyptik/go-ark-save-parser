package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
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

	save, err := arksave.Open(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	id, dino, stat, points, ok, _, err := api.BestDinoForStatFilteredWithFaults(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dinos: %v\n", err)
		os.Exit(1)
	}
	if !ok {
		fmt.Println("no_match")
		return
	}
	fmt.Printf("uuid=%s blueprint=%q stat=%s points=%d level=%d\n", id, dino.Blueprint, stat.String(), points, dino.Stats.CurrentLevel)
}
