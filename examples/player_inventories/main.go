package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
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
	players, faults, err := inventoryAPI.PlayersWithFaults()
	if err != nil {
		log.Fatal(err)
	}
	if len(players) == 0 {
		players, err = playersFromSaveDirectory(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	}

	inventories, inventoryFaults, err := playerInventories(save)
	if err != nil {
		log.Fatal(err)
	}
	faults = append(faults, inventoryFaults...)

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
			items = upstreamInventoryItemCount(save, inventory)
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
	fmt.Printf("players=%d with_inventory=%d without_inventory=%d total_items=%d max_items=%d min_items=%d avg_items=%.2f faults=%d\n", len(players), withInventory, withoutInventory, totalItems, maxItems, minItems, avgItems, len(faults))
}

func playersFromSaveDirectory(savePath string) ([]arkobject.Player, error) {
	api, err := arkapi.NewPlayerFromDirectory(filepath.Dir(savePath))
	if err != nil {
		return nil, err
	}
	return api.Players()
}

func playerInventories(save *arksave.Save) (map[uint64]arkobject.Inventory, []arksave.FaultyObjectInfo, error) {
	pawns, faults, err := save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return strings.Contains(info.ClassName, "PlayerPawn")
	})
	if err != nil {
		return nil, nil, err
	}
	out := make(map[uint64]arkobject.Inventory, len(pawns))
	for _, pawn := range pawns {
		dataID, ok := numericUint64(pawn.Object.Value("LinkedPlayerDataID"))
		if !ok {
			continue
		}
		if _, exists := out[dataID]; exists {
			continue
		}
		inventoryID, ok := objectReferenceUUID(pawn.Object.Value("MyInventoryComponent"))
		if !ok {
			continue
		}
		object, err := save.Object(inventoryID)
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: inventoryID, Err: err})
			continue
		}
		out[dataID] = arkobject.InventoryFromObject(object)
	}
	return out, faults, nil
}

func numericUint64(value any, ok bool) (uint64, bool) {
	if !ok {
		return 0, false
	}
	switch v := value.(type) {
	case uint64:
		return v, true
	case uint32:
		return uint64(v), true
	case int32:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	case int:
		if v < 0 {
			return 0, false
		}
		return uint64(v), true
	default:
		return 0, false
	}
}

func objectReferenceUUID(value any, ok bool) (uuid.UUID, bool) {
	if !ok {
		return uuid.Nil, false
	}
	ref, ok := value.(arkproperty.ObjectReference)
	if !ok {
		return uuid.Nil, false
	}
	switch v := ref.Value.(type) {
	case uuid.UUID:
		return v, true
	case string:
		id, err := uuid.Parse(v)
		return id, err == nil
	default:
		return uuid.Nil, false
	}
}

func upstreamInventoryItemCount(save *arksave.Save, inventory arkobject.Inventory) int {
	seen := make(map[uuid.UUID]struct{}, len(inventory.ItemUUIDs))
	hasExistingItem := false
	for _, id := range inventory.ItemUUIDs {
		seen[id] = struct{}{}
		if !hasExistingItem {
			if _, err := save.ObjectBinary(id); err == nil {
				hasExistingItem = true
			}
		}
	}
	if !hasExistingItem {
		return 0
	}
	return len(seen)
}
