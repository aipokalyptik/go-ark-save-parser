package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
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
	"CLoth",
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
	items, err := rankedEquipmentCandidates(api)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read ranked equipment candidates: %v\n", err)
		os.Exit(1)
	}

	stats := rankEquipment(items, 3)
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

func rankedEquipmentCandidates(api *arkapi.EquipmentAPI) (map[uuid.UUID]arkobject.EquipmentItem, error) {
	out := map[uuid.UUID]arkobject.EquipmentItem{}
	for _, group := range []struct {
		kind       arkobject.EquipmentKind
		blueprints []string
	}{
		{kind: arkobject.EquipmentWeapon, blueprints: arkapi.UpstreamWeaponBlueprints()},
		{kind: arkobject.EquipmentArmor, blueprints: arkapi.UpstreamArmorBlueprints()},
		{kind: arkobject.EquipmentShield, blueprints: arkapi.UpstreamShieldBlueprints()},
		{kind: arkobject.EquipmentSaddle, blueprints: arkapi.UpstreamSaddleBlueprints()},
	} {
		items, _, err := api.FilteredWithFaults(arkapi.EquipmentFilterOptions{
			Kinds:      []arkobject.EquipmentKind{group.kind},
			Blueprints: group.blueprints,
		})
		if err != nil {
			return nil, err
		}
		for id, item := range items {
			out[id] = item
		}
	}
	return out, nil
}

type equipmentRankStats struct {
	Ranked          int
	BestRating      float64
	BestAverageStat float64
	Crafted         int
	Blueprints      int
	Classes         int
}

func rankEquipment(items map[uuid.UUID]arkobject.EquipmentItem, qualityLimit float64) equipmentRankStats {
	stats := equipmentRankStats{}
	classes := map[string]struct{}{}
	for _, item := range items {
		if item.Rating <= qualityLimit || item.IsCrafted() || ignoredEquipment(item) {
			continue
		}
		stats.Ranked++
		if item.Rating > stats.BestRating {
			stats.BestRating = item.Rating
		}
		average := item.AverageStat()
		if average > stats.BestAverageStat {
			stats.BestAverageStat = average
		}
		if item.IsCrafted() {
			stats.Crafted++
		}
		if item.IsBlueprint {
			stats.Blueprints++
		}
		classes[item.Blueprint] = struct{}{}
	}
	stats.Classes = len(classes)
	return stats
}

func ignoredEquipment(item arkobject.EquipmentItem) bool {
	shortName := arkobject.ShortNameFromBlueprint(item.Blueprint)
	for _, part := range ignoredEquipmentNameParts {
		if strings.Contains(shortName, part) {
			return true
		}
	}
	return false
}
