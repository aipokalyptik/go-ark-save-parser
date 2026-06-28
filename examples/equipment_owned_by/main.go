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
		fmt.Fprintln(os.Stderr, "usage: equipment_owned_by <save.ark> <blueprint> <tribe-id>")
		os.Exit(2)
	}

	tribeID64, err := strconv.ParseInt(os.Args[3], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse tribe id: %v\n", err)
		os.Exit(2)
	}

	api, closeAPI, err := arkapi.NewEquipmentFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	summary, _, err := api.OwnedSummaryWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:          []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		Blueprints:     []string{os.Args[2]},
		OnlyBlueprints: true,
	}, arkobject.ObjectOwner{TribeID: int32(tribeID64)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read owned equipment: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("tribe_id=%d items=%d max_damage=%.1f\n", tribeID64, summary.Items, summary.MaxDamage)
}
