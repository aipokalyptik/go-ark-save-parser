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

	api, closeAPI, err := arkapi.NewDinoFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	_, dino, total, ok, err := api.MostMutatedTamed()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dinos: %v\n", err)
		os.Exit(1)
	}
	if !ok {
		fmt.Println("no_match")
		return
	}

	level := int32(0)
	if dino.Stats != nil {
		level = dino.Stats.CurrentLevel
	}
	fmt.Printf("has_most_mutated=1 mutations=%d level=%d\n", total/2, level)
}
