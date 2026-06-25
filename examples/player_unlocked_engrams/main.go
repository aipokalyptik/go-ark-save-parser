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

	engrams, err := api.UnlockedEngrams()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Unlocked engrams:")
	for _, engram := range engrams {
		fmt.Println(engram)
	}
	first, last := "", ""
	if len(engrams) > 0 {
		first = engrams[0]
		last = engrams[len(engrams)-1]
	}
	fmt.Printf("unlocked_engrams=%d first=%s last=%s\n", len(engrams), first, last)
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
