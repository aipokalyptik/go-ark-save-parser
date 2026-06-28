package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: stackable_count <save.ark> <blueprint> [blueprint...]")
		os.Exit(2)
	}

	summary, err := arkapi.StackableSummaryFromPath(os.Args[1], os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stackables: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("items=%d total=%d\n", summary.Items, summary.TotalQuantity)
}
