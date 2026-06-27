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

	summary, err := api.PlayerAllSummary()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("players=%d tribes=%d highest_level=%d total_deaths=%d unlocked_engrams=%d\n", summary.Players, summary.Tribes, summary.HighestLevel, summary.TotalDeaths, summary.UnlockedEngrams)
}
