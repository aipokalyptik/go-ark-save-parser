package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: equipment_best <save.ark>")
		os.Exit(2)
	}

	summary, _, err := arkapi.EquipmentBestSummaryFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "read equipment: %v\n", err)
		os.Exit(1)
	}
	if summary.WeaponFound {
		fmt.Printf(
			"weapon_damage=%.1f weapon=%s weapon_crafted=%t\n",
			summary.Weapon.Stats.Damage,
			arkobject.ShortNameFromBlueprint(summary.Weapon.Blueprint),
			summary.Weapon.IsCrafted(),
		)
	} else {
		fmt.Println("weapon=no_match")
	}

	if summary.ArmorFound {
		fmt.Printf(
			"armor_durability=%.1f armor=%s armor_crafted=%t\n",
			summary.Armor.Stats.Durability,
			arkobject.ShortNameFromBlueprint(summary.Armor.Blueprint),
			summary.Armor.IsCrafted(),
		)
	} else {
		fmt.Println("armor=no_match")
	}
}
