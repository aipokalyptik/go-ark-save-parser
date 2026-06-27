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
	relations, err := api.TribePlayerRelations()
	if err != nil {
		log.Fatal(err)
	}

	tribeIDs := map[int32]struct{}{}
	for _, tribe := range tribes {
		tribeIDs[tribe.TribeID] = struct{}{}
	}
	playersWithoutTribe := 0
	for _, player := range players {
		if _, ok := tribeIDs[player.TribeID]; !ok {
			playersWithoutTribe++
		}
	}

	activeLinks := 0
	inactiveMembers := 0
	tribesWithInactive := 0
	tribesWithoutActive := 0
	for _, relation := range relations {
		activeLinks += len(relation.ActivePlayers)
		inactiveMembers += len(relation.InactiveMemberIDs)
		if len(relation.InactiveMemberIDs) > 0 {
			tribesWithInactive++
		}
		if len(relation.ActivePlayers) == 0 {
			tribesWithoutActive++
		}
	}

	fmt.Printf(
		"players=%d tribes=%d active_links=%d inactive_members=%d players_without_tribe=%d tribes_with_inactive=%d tribes_without_active=%d\n",
		len(players),
		len(tribes),
		activeLinks,
		inactiveMembers,
		playersWithoutTribe,
		tribesWithInactive,
		tribesWithoutActive,
	)
}
