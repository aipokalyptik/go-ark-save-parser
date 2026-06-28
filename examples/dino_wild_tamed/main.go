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

	api, closeAPI, err := arkapi.NewDinoFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	dinos, _, err := api.WildTamedWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read wild-tamed dinos: %v\n", err)
		os.Exit(1)
	}

	maxLevel := int32(0)
	if level, ok := api.MaxCurrentLevel(dinos); ok {
		maxLevel = level
	}
	fmt.Printf("wild_tamed=%d max_level=%d\n", len(dinos), maxLevel)
}
