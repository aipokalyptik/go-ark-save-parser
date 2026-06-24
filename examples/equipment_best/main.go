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
		fmt.Fprintln(os.Stderr, "usage: equipment_best <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewEquipment(save)
	weapons, _, err := api.FilteredWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:        []arkobject.EquipmentKind{arkobject.EquipmentWeapon},
		Blueprints:   arkapi.UpstreamWeaponBlueprints(),
		NoBlueprints: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read weapons: %v\n", err)
		os.Exit(1)
	}
	_, weapon, ok := api.BestWeaponDamage(weapons)
	if ok {
		fmt.Printf(
			"weapon_damage=%.1f weapon=%s weapon_crafted=%t\n",
			weapon.Stats.Damage,
			arkobject.ShortNameFromBlueprint(weapon.Blueprint),
			weapon.IsCrafted(),
		)
	} else {
		fmt.Println("weapon=no_match")
	}

	armor, _, err := api.FilteredWithFaults(arkapi.EquipmentFilterOptions{
		Kinds:        []arkobject.EquipmentKind{arkobject.EquipmentArmor},
		Blueprints:   arkapi.UpstreamArmorBlueprints(),
		NoBlueprints: true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read armor: %v\n", err)
		os.Exit(1)
	}
	_, armorItem, ok := api.BestActualDurability(armor)
	if ok {
		fmt.Printf(
			"armor_durability=%.1f armor=%s armor_crafted=%t\n",
			armorItem.Stats.Durability,
			arkobject.ShortNameFromBlueprint(armorItem.Blueprint),
			armorItem.IsCrafted(),
		)
	} else {
		fmt.Println("armor=no_match")
	}
}
