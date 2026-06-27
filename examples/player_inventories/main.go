package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark>", os.Args[0])
	}
	playerAPI, closePlayers, err := arkapi.NewPlayerFromPath(os.Args[1], arkapi.PlayerPathOptions{Fallback: arkapi.PlayerPathFallbackPlayers})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := closePlayers(); err != nil {
			log.Fatal(err)
		}
	}()
	players, err := playerAPI.Players()
	if err != nil {
		log.Fatal(err)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	inventoryAPI := arkapi.NewPlayer(save)
	summary, inventoryFaults, err := inventoryAPI.PlayerInventorySummaryForPlayers(players)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("players=%d with_inventory=%d without_inventory=%d total_items=%d max_items=%d min_items=%d avg_items=%.2f faults=%d\n", summary.Players, summary.WithInventory, summary.WithoutInventory, summary.TotalItems, summary.MaxItems, summary.MinItems, summary.AverageItems, len(inventoryFaults))
}
