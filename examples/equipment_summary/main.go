package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: equipment_summary <save.ark>")
		os.Exit(2)
	}

	api, closeAPI, err := arkapi.NewEquipmentFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	summary, _, err := api.SummaryIncludingCryopodSaddlesWithFaults(arkapi.EquipmentFilterOptions{
		Blueprints: arkapi.CanonicalEquipmentBlueprints(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read equipment: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"items=%d weapons=%d armor=%d saddles=%d cryopod_saddles=%d shields=%d with_custom_data=%d custom_data_entries=%d\n",
		summary.Items,
		summary.ByKind[arkobject.EquipmentWeapon],
		summary.ByKind[arkobject.EquipmentArmor],
		summary.ByKind[arkobject.EquipmentSaddle],
		summary.CryopodSaddles,
		summary.ByKind[arkobject.EquipmentShield],
		summary.WithCustomData,
		summary.CustomDataEntries,
	)
}
