package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: stackable_owned_by <save.ark> <blueprint> <tribe-id>")
		os.Exit(2)
	}
	tribeID64, err := strconv.ParseInt(os.Args[3], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse tribe id: %v\n", err)
		os.Exit(2)
	}

	summary, err := arkapi.StackableOwnedSummaryFromPath(os.Args[1], []string{os.Args[2]}, arkobject.ObjectOwner{TribeID: int32(tribeID64)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read stackables: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("tribe_id=%d items=%d total=%d\n", tribeID64, summary.Items, summary.TotalQuantity)
}
