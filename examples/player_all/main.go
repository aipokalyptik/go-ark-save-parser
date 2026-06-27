package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark-or-save-directory>", os.Args[0])
	}
	api, closeAPI, err := arkapi.NewPlayerFromPath(os.Args[1], arkapi.PlayerPathOptions{Fallback: arkapi.PlayerPathFallbackPlayers})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := closeAPI(); err != nil {
			log.Fatal(err)
		}
	}()

	players, err := api.Players()
	if err != nil {
		log.Fatal(err)
	}
	tribes, err := api.TribeDetails()
	if err != nil {
		log.Fatal(err)
	}
	totalDeaths, err := api.TotalDeaths()
	if err != nil {
		log.Fatal(err)
	}
	unlockedEngrams, err := api.UnlockedEngrams()
	if err != nil {
		log.Fatal(err)
	}
	_, highestLevel, hasLevel, err := api.PlayerWithHighestLevel()
	if err != nil {
		log.Fatal(err)
	}
	if !hasLevel {
		highestLevel = 0
	}
	fmt.Printf("players=%d tribes=%d highest_level=%d total_deaths=%d unlocked_engrams=%d\n", len(players), len(tribes), highestLevel, totalDeaths, len(unlockedEngrams))
}
