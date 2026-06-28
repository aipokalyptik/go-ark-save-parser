package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <directory-with-arkprofile-arktribe-cluster-files>", os.Args[0])
	}
	summary, faults, err := arkapi.LocalProfileSummaryFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	faultByOperation := map[string]error{}
	for _, fault := range faults {
		faultByOperation[fault.Operation] = fault.Err
	}

	fmt.Printf("profiles=%d tribes=%d clusters=%d tributes=%d\n", summary.Files.Profiles, summary.Files.Tribes, summary.Files.Clusters, summary.Files.Tributes)
	if err, ok := faultByOperation["players"]; ok {
		log.Printf("players: %v", err)
	} else if summary.HasParsedPlayers {
		fmt.Printf("parsed_players=%d\n", summary.ParsedPlayers)
	}
	if err, ok := faultByOperation["tribes"]; ok {
		log.Printf("tribes: %v", err)
	} else if summary.HasParsedTribes {
		fmt.Printf("parsed_tribes=%d\n", summary.ParsedTribes)
	}
	if err, ok := faultByOperation["tribe player map"]; ok {
		log.Printf("tribe player map: %v", err)
	} else if summary.HasTribePlayerLinks {
		fmt.Printf("tribe_player_links=%d\n", summary.TribePlayerLinks)
	}
	if err, ok := faultByOperation["total deaths"]; ok {
		log.Printf("total deaths: %v", err)
	} else if summary.HasTotalDeaths {
		fmt.Printf("total_deaths=%d\n", summary.TotalDeaths)
	}
	if err, ok := faultByOperation["highest level"]; ok {
		log.Printf("highest level: %v", err)
	} else if summary.HasHighestLevel {
		fmt.Printf("highest_level=%d\n", summary.HighestLevel)
	}
	if err, ok := faultByOperation["highest experience"]; ok {
		log.Printf("highest experience: %v", err)
	} else if summary.HasHighestExperience {
		fmt.Printf("highest_experience=%.2f\n", summary.HighestExperience)
	}
	if err, ok := faultByOperation["average deaths"]; ok {
		log.Printf("average deaths: %v", err)
	}
	if err, ok := faultByOperation["average level"]; ok {
		log.Printf("average level: %v", err)
	}
	if err, ok := faultByOperation["average experience"]; ok {
		log.Printf("average experience: %v", err)
	}
	if summary.HasAverageDeaths && summary.HasAverageLevel && summary.HasAverageExperience {
		fmt.Printf("average_deaths=%.2f average_level=%.2f average_experience=%.2f\n", summary.AverageDeaths, summary.AverageLevel, summary.AverageExperience)
	}
	if err, ok := faultByOperation["unlocked engrams"]; ok {
		log.Printf("unlocked engrams: %v", err)
	} else if summary.HasUnlockedEngrams {
		fmt.Printf("unlocked_engrams=%d\n", summary.UnlockedEngrams)
	}
}
