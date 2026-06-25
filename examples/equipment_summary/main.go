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
	weapons, err := countCanonical(api, arkobject.EquipmentWeapon, arkapi.UpstreamWeaponBlueprints())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read weapons: %v\n", err)
		os.Exit(1)
	}
	armor, err := countCanonical(api, arkobject.EquipmentArmor, arkapi.UpstreamArmorBlueprints())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read armor: %v\n", err)
		os.Exit(1)
	}
	saddles, err := countCanonical(api, arkobject.EquipmentSaddle, arkapi.UpstreamSaddleBlueprints())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read saddles: %v\n", err)
		os.Exit(1)
	}
	shields, err := countCanonical(api, arkobject.EquipmentShield, arkapi.UpstreamShieldBlueprints())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read shields: %v\n", err)
		os.Exit(1)
	}
	cryopodSaddles, _, err := arkapi.NewDino(save).SaddlesFromCryopodsWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read cryopod saddles: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"items=%d weapons=%d armor=%d saddles=%d cryopod_saddles=%d shields=%d\n",
		weapons+armor+saddles+shields,
		weapons,
		armor,
		saddles,
		len(cryopodSaddles),
		shields,
	)
}

func countCanonical(api *arkapi.EquipmentAPI, kind arkobject.EquipmentKind, blueprints []string) (int, error) {
	items, _, err := api.FilteredWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:      []arkobject.EquipmentKind{kind},
		Blueprints: blueprints,
	})
	if err != nil {
		return 0, err
	}
	return len(items), nil
}
