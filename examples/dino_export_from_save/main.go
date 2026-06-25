package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: dino_export_from_save <save.ark> <out-dir>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	exported, err := arkapi.NewDino(save).ExportBinary(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "export dino binaries: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("dinos=%d rows=%d faults=%d wrote=%s\n", exported.DinoCount, exported.RowCount, exported.FaultCount, os.Args[2])
}
