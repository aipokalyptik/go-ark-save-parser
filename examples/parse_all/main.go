package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: parse_all <save.ark>")
		os.Exit(2)
	}

	summary, _, err := arkapi.GeneralParseSummaryFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("objects=%d parsed=%d faults=%d\n", summary.Objects, summary.Parsed, summary.Faults)
}
