package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark>", os.Args[0])
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	inventoryAPI := arkapi.NewPlayer(save)
	players, _, err := inventoryAPI.PlayersWithFaults()
	if err != nil {
		log.Fatal(err)
	}
	if len(players) == 0 {
		players, err = playersFromSaveDirectory(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	}

	inventories, inventoryFaults, err := inventoryAPI.PlayerInventoriesWithFaults()
	if err != nil {
		log.Fatal(err)
	}

	withInventory := 0
	withoutInventory := 0
	totalItems := 0
	maxItems := 0
	minItems := 0
	for i, player := range players {
		items := 0
		inventory, ok := inventories[player.PlayerDataID]
		if ok {
			withInventory++
			items = inventoryAPI.InventoryItemCount(inventory)
		} else {
			withoutInventory++
		}
		if i == 0 || items < minItems {
			minItems = items
		}
		if items > maxItems {
			maxItems = items
		}
		totalItems += items
	}

	avgItems := 0.0
	if len(players) > 0 {
		avgItems = float64(totalItems) / float64(len(players))
	}
	fmt.Printf("players=%d with_inventory=%d without_inventory=%d total_items=%d max_items=%d min_items=%d avg_items=%.2f faults=%d\n", len(players), withInventory, withoutInventory, totalItems, maxItems, minItems, avgItems, len(inventoryFaults))
}

func playersFromSaveDirectory(savePath string) ([]arkobject.Player, error) {
	api, err := arkapi.NewPlayerFromDirectory(filepath.Dir(savePath))
	if err != nil {
		return nil, err
	}
	return api.Players()
}
