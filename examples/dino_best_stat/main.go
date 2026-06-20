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
		fmt.Fprintln(os.Stderr, "usage: dino_best_stat <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	id, dino, stat, points, ok, err := api.BestDinoForStatFiltered(arkapi.DinoBestStatOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read dinos: %v\n", err)
		os.Exit(1)
	}
	if !ok {
		fmt.Println("no_match")
		return
	}
	fmt.Printf("uuid=%s blueprint=%q stat=%s points=%d level=%d\n", id, dino.Blueprint, dinoStatName(stat), points, dino.Stats.CurrentLevel)
}

func dinoStatName(stat arkobject.DinoStat) string {
	switch stat {
	case arkobject.DinoStatHealth:
		return "health"
	case arkobject.DinoStatStamina:
		return "stamina"
	case arkobject.DinoStatTorpidity:
		return "torpidity"
	case arkobject.DinoStatOxygen:
		return "oxygen"
	case arkobject.DinoStatFood:
		return "food"
	case arkobject.DinoStatWater:
		return "water"
	case arkobject.DinoStatTemperature:
		return "temperature"
	case arkobject.DinoStatWeight:
		return "weight"
	case arkobject.DinoStatMeleeDamage:
		return "melee_damage"
	case arkobject.DinoStatMovementSpeed:
		return "movement_speed"
	case arkobject.DinoStatFortitude:
		return "fortitude"
	case arkobject.DinoStatCraftingSpeed:
		return "crafting_speed"
	default:
		return "unknown"
	}
}
