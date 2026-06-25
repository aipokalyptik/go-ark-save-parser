package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: base_export_from_save <save.ark> <out.json>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	exported, err := arkapi.NewJSON(save).ExportDomain("bases")
	if err != nil {
		fmt.Fprintf(os.Stderr, "export bases: %v\n", err)
		os.Exit(1)
	}
	data, err := json.MarshalIndent(exported, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode bases: %v\n", err)
		os.Exit(1)
	}
	data = append(data, '\n')
	if err := os.WriteFile(os.Args[2], data, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("bases=%d faults=%d wrote=%s\n", exported.Count, exported.FaultCount, os.Args[2])
}
