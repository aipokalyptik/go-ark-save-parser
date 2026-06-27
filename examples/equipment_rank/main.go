package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

var ignoredEquipmentNameParts = []string{
	"WeaponCrossbow",
	"WeaponMetalHatchet",
	"WeaponMetalPick",
	"WeaponBow",
	"Chitin",
	"Hide",
	"WeaponPike",
	"WeaponGun",
	"Cloth",
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: equipment_rank <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewEquipment(save)
	items, _, err := api.RankedCandidatesWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read ranked equipment candidates: %v\n", err)
		os.Exit(1)
	}

	stats := api.RankStats(items, arkapi.EquipmentRankOptions{
		MinRating:        3,
		ExcludeCrafted:   true,
		IgnoredNameParts: ignoredEquipmentNameParts,
	})
	fmt.Printf(
		"ranked=%d best_rating=%.1f best_average_stat=%.1f crafted=%d blueprints=%d classes=%d\n",
		stats.Ranked,
		stats.BestRating,
		stats.BestAverageStat,
		stats.Crafted,
		stats.Blueprints,
		stats.Classes,
	)
}
