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

	info, err := arkapi.NewJSON(save).ExportSaveInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("map=%s save_version=%d objects=%d names=%d\n", info.MapName, info.SaveVersion, info.ObjectCount, info.NameCount)
}
