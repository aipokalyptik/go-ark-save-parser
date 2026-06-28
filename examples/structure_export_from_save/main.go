package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: structure_export_from_save <save.ark> <out-dir>")
		os.Exit(2)
	}

	exported, err := arkapi.ExportStructureBinaryFromPath(os.Args[1], os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "export structure binaries: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("structures=%d rows=%d faults=%d wrote=%s\n", exported.StructureCount, exported.RowCount, exported.FaultCount, os.Args[2])
}
