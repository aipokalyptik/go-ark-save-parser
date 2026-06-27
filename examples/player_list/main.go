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
	withNames := 0
	highestLevel := int32(0)
	for _, player := range players {
		if player.CharacterName != "" || player.PlayerName != "" {
			withNames++
		}
		if player.Level > highestLevel {
			highestLevel = player.Level
		}
	}
	fmt.Printf("players=%d with_names=%d highest_level=%d\n", len(players), withNames, highestLevel)
}
