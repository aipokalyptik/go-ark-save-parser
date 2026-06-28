package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: dino_export_from_save <save.ark> <out-dir>")
		os.Exit(2)
	}

	exported, err := arkapi.ExportDinoBinaryFromPath(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "export dino binaries: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("dinos=%d rows=%d faults=%d wrote=%s\n", exported.DinoCount, exported.RowCount, exported.FaultCount, os.Args[2])
}
