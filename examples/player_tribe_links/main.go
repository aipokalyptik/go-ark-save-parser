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

	summary, err := api.TribePlayerRelationSummary()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(
		"players=%d tribes=%d active_links=%d inactive_members=%d players_without_tribe=%d tribes_with_inactive=%d tribes_without_active=%d\n",
		summary.Players,
		summary.Tribes,
		summary.ActiveLinks,
		summary.InactiveMembers,
		summary.PlayersWithoutTribe,
		summary.TribesWithInactive,
		summary.TribesWithoutActive,
	)
}
