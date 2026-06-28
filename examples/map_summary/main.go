package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark>", os.Args[0])
	}
	info, err := arkapi.ExportSaveInfoFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("map=%s save_version=%d objects=%d names=%d\n", info.MapName, info.SaveVersion, info.ObjectCount, info.NameCount)
}
