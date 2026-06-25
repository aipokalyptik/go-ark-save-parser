package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark-or-save-directory>", os.Args[0])
	}
	api, closeSave, err := openPlayerAPI(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer closeSave()

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

func openPlayerAPI(path string) (*arkapi.PlayerAPI, func(), error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, func() {}, err
	}
	if info.IsDir() {
		api, err := arkapi.NewPlayerFromDirectory(path)
		return api, func() {}, err
	}
	save, err := arksave.Open(path)
	if err != nil {
		return nil, func() {}, err
	}
	api := arkapi.NewPlayer(save)
	players, _, err := api.PlayersWithFaults()
	if err != nil {
		_ = save.Close()
		return nil, func() {}, err
	}
	if len(players) == 0 {
		_ = save.Close()
		api, err := arkapi.NewPlayerFromDirectory(filepath.Dir(path))
		return api, func() {}, err
	}
	return api, func() { _ = save.Close() }, nil
}
