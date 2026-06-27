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
	api, err := arkapi.NewPlayerFromDirectory(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	files := api.LocalFileSummary()
	fmt.Printf("profiles=%d tribes=%d clusters=%d tributes=%d\n", files.Profiles, files.Tribes, files.Clusters, files.Tributes)

	players, err := api.Players()
	if err != nil {
		log.Printf("players: %v", err)
	} else {
		fmt.Printf("parsed_players=%d\n", len(players))
	}

	tribes, err := api.TribeSummaries()
	if err != nil {
		log.Printf("tribes: %v", err)
	} else {
		fmt.Printf("parsed_tribes=%d\n", len(tribes))
	}

	links, err := api.TribePlayerLinkCount()
	if err != nil {
		log.Printf("tribe player map: %v", err)
	} else {
		fmt.Printf("tribe_player_links=%d\n", links)
	}

	totalDeaths, err := api.TotalDeaths()
	if err != nil {
		log.Printf("total deaths: %v", err)
	} else {
		fmt.Printf("total_deaths=%d\n", totalDeaths)
	}

	if _, level, ok, err := api.PlayerWithHighestLevel(); err != nil {
		log.Printf("highest level: %v", err)
	} else if ok {
		fmt.Printf("highest_level=%d\n", level)
	}

	if _, experience, ok, err := api.PlayerWithHighestExperience(); err != nil {
		log.Printf("highest experience: %v", err)
	} else if ok {
		fmt.Printf("highest_experience=%.2f\n", experience)
	}

	unlockedEngrams, err := api.UnlockedEngrams()
	if err != nil {
		log.Printf("unlocked engrams: %v", err)
	} else {
		fmt.Printf("unlocked_engrams=%d\n", len(unlockedEngrams))
	}
}
