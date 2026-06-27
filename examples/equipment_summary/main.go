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
	weapons, _, err := api.CanonicalCountWithFaults(arkobject.EquipmentWeapon, arkapi.UpstreamWeaponBlueprints())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read weapons: %v\n", err)
		os.Exit(1)
	}
	armor, _, err := api.CanonicalCountWithFaults(arkobject.EquipmentArmor, arkapi.UpstreamArmorBlueprints())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read armor: %v\n", err)
		os.Exit(1)
	}
	saddles, _, err := api.CanonicalCountWithFaults(arkobject.EquipmentSaddle, arkapi.UpstreamSaddleBlueprints())
	if err != nil {
		fmt.Fprintf(os.Stderr, "read saddles: %v\n", err)
		os.Exit(1)
	}
	shields, _, err := api.CanonicalCountWithFaults(arkobject.EquipmentShield, arkapi.UpstreamShieldBlueprints())
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
