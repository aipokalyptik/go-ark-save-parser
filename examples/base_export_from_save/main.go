package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: base_export_from_save <save.ark> <out-dir>")
		os.Exit(2)
	}

	exported, err := arkapi.ExportBaseBinaryFromPath(os.Args[1], "", os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "export base binaries: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("bases=%d structures=%d faults=%d wrote=%s\n", exported.BaseCount, exported.StructureCount, exported.FaultCount, os.Args[2])
}
