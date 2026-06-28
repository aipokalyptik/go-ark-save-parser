package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <player-data-id>", os.Args[0])
	}
	playerDataID, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		log.Fatal(err)
	}
	lookup, err := arkapi.PlayerInventoryLookupFromPath(os.Args[1], playerDataID)
	if err != nil {
		log.Fatal(err)
	}
	if !lookup.Found {
		fmt.Printf("player=%d inventory=missing items=0\n", playerDataID)
		return
	}
	if lookup.HasLocation {
		fmt.Printf("player=%d inventory=%s items=%d location=(%.2f,%.2f,%.2f)\n", playerDataID, lookup.InventoryUUID, lookup.Items, lookup.Location.X, lookup.Location.Y, lookup.Location.Z)
		return
	}
	fmt.Printf("player=%d inventory=%s items=%d location=missing\n", playerDataID, lookup.InventoryUUID, lookup.Items)
}
