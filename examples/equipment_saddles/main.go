package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: equipment_saddles <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	equipmentAPI := arkapi.NewEquipment(save)
	itemSaddles, _, err := equipmentAPI.FilteredWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:      []arkobject.EquipmentKind{arkobject.EquipmentSaddle},
		Blueprints: arkapi.UpstreamSaddleBlueprints(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read item saddles: %v\n", err)
		os.Exit(1)
	}

	cryopodSaddles, _, err := arkapi.NewDino(save).SaddlesFromCryopodsWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read cryopod saddles: %v\n", err)
		os.Exit(1)
	}

	maxArmor := float64(0)
	if _, saddle, ok := equipmentAPI.BestArmor(itemSaddles); ok {
		maxArmor = saddle.Stats.Armor
	}
	if _, saddle, ok := equipmentAPI.BestArmor(cryopodSaddles); ok && saddle.Stats.Armor > maxArmor {
		maxArmor = saddle.Stats.Armor
	}

	fmt.Printf(
		"item_saddles=%d cryopod_saddles=%d total_saddles=%d max_armor=%.1f\n",
		len(itemSaddles),
		len(cryopodSaddles),
		len(itemSaddles)+len(cryopodSaddles),
		maxArmor,
	)
}
