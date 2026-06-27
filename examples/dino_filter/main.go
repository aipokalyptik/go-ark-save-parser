package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		log.Fatalf("usage: %s [--no-cryos] <save.ark>", os.Args[0])
	}
	includeCryos := true
	savePath := os.Args[1]
	if os.Args[1] == "--no-cryos" {
		if len(os.Args) != 3 {
			log.Fatalf("usage: %s [--no-cryos] <save.ark>", os.Args[0])
		}
		includeCryos = false
		savePath = os.Args[2]
	}
	save, err := arksave.Open(savePath)
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	summary, _, err := api.PopulationSummaryWithFaults(includeCryos)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("dinos=%d tamed=%d wild=%d cryopodded=%d classes=%d\n", summary.Dinos, summary.Tamed, summary.Wild, summary.Cryopodded, summary.Classes)
}
