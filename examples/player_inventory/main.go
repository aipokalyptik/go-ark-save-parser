package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <player-data-id>", os.Args[0])
	}
	playerDataID, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewPlayer(save)
	inventory, ok, err := api.PlayerInventoryByDataID(playerDataID)
	if err != nil {
		log.Fatal(err)
	}
	if !ok {
		fmt.Printf("player=%d inventory=missing items=0\n", playerDataID)
		return
	}
	location, hasLocation, err := api.PlayerLocationByDataID(playerDataID)
	if err != nil {
		log.Fatal(err)
	}
	if hasLocation {
		fmt.Printf("player=%d inventory=%s items=%d location=(%.2f,%.2f,%.2f)\n", playerDataID, inventory.UUID, inventory.NumberOfItems(), location.X, location.Y, location.Z)
		return
	}
	fmt.Printf("player=%d inventory=%s items=%d location=missing\n", playerDataID, inventory.UUID, inventory.NumberOfItems())
}
