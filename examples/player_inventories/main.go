package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark>", os.Args[0])
	}
	summary, inventoryFaults, err := arkapi.PlayerInventorySummaryFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("players=%d with_inventory=%d without_inventory=%d total_items=%d max_items=%d min_items=%d avg_items=%.2f faults=%d\n", summary.Players, summary.WithInventory, summary.WithoutInventory, summary.TotalItems, summary.MaxItems, summary.MinItems, summary.AverageItems, len(inventoryFaults))
}
