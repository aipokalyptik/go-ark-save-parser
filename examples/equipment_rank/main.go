package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
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

	stats, _, err := arkapi.EquipmentRankStatsFromPath(os.Args[1], arkapi.EquipmentRankOptions{
		MinRating:        3,
		ExcludeCrafted:   true,
		IgnoredNameParts: ignoredEquipmentNameParts,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read ranked equipment candidates: %v\n", err)
		os.Exit(1)
	}
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
