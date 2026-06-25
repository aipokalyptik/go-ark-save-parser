package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: parse_all <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	ids, err := save.ObjectIDs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "list objects: %v\n", err)
		os.Exit(1)
	}
	objects, faults, err := arkapi.NewGeneral(save).ObjectsWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse objects: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("objects=%d parsed=%d faults=%d\n", len(ids), len(objects), len(faults))
}
