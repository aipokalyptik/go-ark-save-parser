package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: class_lookup <save.ark> <class-substring> [class-substring...]")
		os.Exit(2)
	}

	api, closeAPI, err := arkapi.NewGeneralFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	substrings := os.Args[2:]
	summary, _, err := api.ClassLookupSummaryWithFaults(substrings)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lookup class substring: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("objects=%d classes=%d\n", summary.Objects, summary.Classes)
}
