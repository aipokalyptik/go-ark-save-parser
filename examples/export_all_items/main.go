package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <out-dir>", os.Args[0])
	}
	manifest, err := arkapi.ExportAllDomainsFromPath(os.Args[1], os.Args[2], []string{"dinos", "structures", "equipment", "stackables", "players", "tribes", "bases"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("exports=%d wrote=%s\n", len(manifest.Files), os.Args[2])
}
