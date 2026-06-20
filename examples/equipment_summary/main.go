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
	items, err := api.All()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read equipment: %v\n", err)
		os.Exit(1)
	}

	var weapons, armor, saddles, shields int
	for _, item := range items {
		switch item.Kind {
		case arkobject.EquipmentWeapon:
			weapons++
		case arkobject.EquipmentArmor:
			armor++
		case arkobject.EquipmentSaddle:
			saddles++
		case arkobject.EquipmentShield:
			shields++
		}
	}

	fmt.Printf(
		"items=%d weapons=%d armor=%d saddles=%d shields=%d\n",
		len(items),
		weapons,
		armor,
		saddles,
		shields,
	)
}
