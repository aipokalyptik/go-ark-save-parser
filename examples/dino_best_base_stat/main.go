package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("usage: %s <save.ark> <dino-blueprint> <stat>", os.Args[0])
	}
	stat, ok := parseDinoStat(os.Args[3])
	if !ok {
		log.Fatalf("unknown stat %q", os.Args[3])
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewDino(save)
	_, dino, gotStat, points, found, _, err := api.BestDinoForStatFilteredWithFaults(arkapi.DinoBestStatOptions{
		Blueprints:      []string{os.Args[2]},
		Stats:           []arkobject.DinoStat{stat},
		OnlyTamed:       true,
		ExcludeCryopods: true,
		BaseStat:        true,
	})
	if err != nil {
		log.Fatal(err)
	}
	if !found {
		fmt.Println("has_result=0")
		return
	}
	level := int32(0)
	if dino.Stats != nil {
		level = dino.Stats.CurrentLevel
	}
	fmt.Printf("has_result=1 stat=%s points=%d level=%d\n", dinoStatName(gotStat), points, level)
}

func parseDinoStat(value string) (arkobject.DinoStat, bool) {
	switch value {
	case "health":
		return arkobject.DinoStatHealth, true
	case "stamina":
		return arkobject.DinoStatStamina, true
	case "torpidity":
		return arkobject.DinoStatTorpidity, true
	case "oxygen":
		return arkobject.DinoStatOxygen, true
	case "food":
		return arkobject.DinoStatFood, true
	case "water":
		return arkobject.DinoStatWater, true
	case "temperature":
		return arkobject.DinoStatTemperature, true
	case "weight":
		return arkobject.DinoStatWeight, true
	case "melee_damage":
		return arkobject.DinoStatMeleeDamage, true
	case "movement_speed":
		return arkobject.DinoStatMovementSpeed, true
	case "fortitude":
		return arkobject.DinoStatFortitude, true
	case "crafting_speed":
		return arkobject.DinoStatCraftingSpeed, true
	default:
		return 0, false
	}
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
