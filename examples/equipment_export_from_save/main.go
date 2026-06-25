package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: equipment_export_from_save <save.ark> <out-dir>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	exported, err := arkapi.NewEquipment(save).ExportBinary(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "export equipment binaries: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("items=%d rows=%d faults=%d wrote=%s\n", exported.ItemCount, exported.RowCount, exported.FaultCount, os.Args[2])
}
