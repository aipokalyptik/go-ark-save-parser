package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <out-dir>", os.Args[0])
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewJSON(save)
	manifest, err := api.ExportAllDomains(os.Args[2], []string{"dinos", "structures", "equipment", "stackables", "players", "tribes", "bases"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("exports=%d wrote=%s\n", len(manifest.Files), os.Args[2])
}
