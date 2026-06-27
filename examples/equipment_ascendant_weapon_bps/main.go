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
		fmt.Fprintln(os.Stderr, "usage: equipment_ascendant_weapon_bps <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewEquipment(save)
	items, _, err := api.FilteredWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:          []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		Blueprints:     arkapi.UpstreamWeaponBlueprints(),
		OnlyBlueprints: true,
		MinQuality:     arkapi.AscendantQualityIndex,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read equipment: %v\n", err)
		os.Exit(1)
	}

	maxDamage := float64(0)
	if _, item, ok := api.BestWeaponDamage(items); ok {
		maxDamage = item.Stats.Damage
	}
	fmt.Printf("items=%d max_damage=%.1f\n", len(items), maxDamage)
}
