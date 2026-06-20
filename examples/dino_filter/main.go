package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark>", os.Args[0])
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	dinos, _, err := api.AllWithFaults()
	if err != nil {
		log.Fatal(err)
	}
	tamed := 0
	wild := 0
	for _, dino := range dinos {
		if dino.IsTamed {
			tamed++
		} else {
			wild++
		}
	}
	classes := api.CountByClass(dinos)

	fmt.Printf("dinos=%d tamed=%d wild=%d classes=%d\n", len(dinos), tamed, wild, len(classes))
}
