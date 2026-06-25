package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dino_wild_tamed <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	dinos, _, err := arkapi.NewDino(save).WildTamedWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read wild-tamed dinos: %v\n", err)
		os.Exit(1)
	}

	maxLevel := int32(0)
	for _, dino := range dinos {
		if dino.Stats != nil && dino.Stats.CurrentLevel > maxLevel {
			maxLevel = dino.Stats.CurrentLevel
		}
	}
	fmt.Printf("wild_tamed=%d max_level=%d\n", len(dinos), maxLevel)
}
