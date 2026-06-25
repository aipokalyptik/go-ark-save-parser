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
