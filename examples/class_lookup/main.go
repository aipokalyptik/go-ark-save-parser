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

	substrings := os.Args[2:]
	summary, _, err := arkapi.GeneralClassLookupSummaryFromPath(os.Args[1], substrings)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("objects=%d classes=%d\n", summary.Objects, summary.Classes)
}
