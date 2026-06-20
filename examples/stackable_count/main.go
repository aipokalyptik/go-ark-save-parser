package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: stackable_count <save.ark> <blueprint> [blueprint...]")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewStackable(save)
	items, err := api.ByClass(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stackables: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("items=%d total=%d\n", len(items), api.Count(items))
}
