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
	dinos, _, err := api.AllWithFaults()
	if err != nil {
		log.Fatal(err)
	}
	if !includeCryos {
		for id, dino := range dinos {
			if dino.IsCryopodded {
				delete(dinos, id)
			}
		}
	}
	tamed := 0
	wild := 0
	cryopodded := 0
	for _, dino := range dinos {
		if dino.IsTamed {
			tamed++
		} else {
			wild++
		}
		if dino.IsCryopodded {
			cryopodded++
		}
	}
	classes := api.CountByClass(dinos)

	fmt.Printf("dinos=%d tamed=%d wild=%d cryopodded=%d classes=%d\n", len(dinos), tamed, wild, cryopodded, len(classes))
}
