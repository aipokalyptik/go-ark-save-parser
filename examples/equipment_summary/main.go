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

	summary, _, err := arkapi.EquipmentSummaryIncludingCryopodSaddlesFromPath(os.Args[1], arkapi.EquipmentFilterOptions{
		Blueprints: arkapi.CanonicalEquipmentBlueprints(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read equipment: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"items=%d total_quantity=%d avg_quantity=%.2f total_rating=%.2f avg_rating=%.2f weapons=%d armor=%d saddles=%d cryopod_saddles=%d shields=%d with_custom_data=%d custom_data_entries=%d\n",
		summary.Items,
		summary.TotalQuantity,
		summary.AverageQuantity,
		summary.TotalRating,
		summary.AverageRating,
		summary.ByKind[arkobject.EquipmentWeapon],
		summary.ByKind[arkobject.EquipmentArmor],
		summary.ByKind[arkobject.EquipmentSaddle],
		summary.CryopodSaddles,
		summary.ByKind[arkobject.EquipmentShield],
		summary.WithCustomData,
		summary.CustomDataEntries,
	)
}
