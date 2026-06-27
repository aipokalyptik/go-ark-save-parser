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
		fmt.Fprintln(os.Stderr, "usage: equipment_summary <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewEquipment(save)
	summary, _, err := api.SummaryIncludingCryopodSaddlesWithFaults(arkapi.EquipmentFilterOptions{
		Blueprints: canonicalEquipmentBlueprints(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read equipment: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"items=%d weapons=%d armor=%d saddles=%d cryopod_saddles=%d shields=%d\n",
		summary.Items,
		summary.ByKind[arkobject.EquipmentWeapon],
		summary.ByKind[arkobject.EquipmentArmor],
		summary.ByKind[arkobject.EquipmentSaddle],
		summary.CryopodSaddles,
		summary.ByKind[arkobject.EquipmentShield],
	)
}

func canonicalEquipmentBlueprints() []string {
	blueprints := []string{}
	blueprints = append(blueprints, arkapi.UpstreamWeaponBlueprints()...)
	blueprints = append(blueprints, arkapi.UpstreamArmorBlueprints()...)
	blueprints = append(blueprints, arkapi.UpstreamSaddleBlueprints()...)
	blueprints = append(blueprints, arkapi.UpstreamShieldBlueprints()...)
	return blueprints
}
