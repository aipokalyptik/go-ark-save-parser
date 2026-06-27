package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: equipment_owned_by <save.ark> <blueprint> <tribe-id>")
		os.Exit(2)
	}

	tribeID64, err := strconv.ParseInt(os.Args[3], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse tribe id: %v\n", err)
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
		Blueprints:     []string{os.Args[2]},
		OnlyBlueprints: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read equipment: %v\n", err)
		os.Exit(1)
	}
	owned, err := api.FilterOwnedBy(items, arkobject.ObjectOwner{TribeID: int32(tribeID64)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "filter owner: %v\n", err)
		os.Exit(1)
	}
	maxDamage := float64(0)
	if _, item, ok := api.BestWeaponDamage(owned); ok {
		maxDamage = item.Stats.Damage
	}
	fmt.Printf("tribe_id=%d items=%d max_damage=%.1f\n", tribeID64, len(owned), maxDamage)
}
