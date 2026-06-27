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
