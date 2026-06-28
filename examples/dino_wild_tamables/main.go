package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dino_wild_tamables <save.ark>")
		os.Exit(2)
	}

	api, closeAPI, err := arkapi.NewDinoFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	summary, _, err := api.WildTamableSummaryWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read wild dinos: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wild_dinos=%d wild_tamables=%d\n", summary.WildDinos, summary.WildTamables)
}
