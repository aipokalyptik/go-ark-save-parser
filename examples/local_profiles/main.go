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
	fmt.Printf("profiles=%d tribes=%d clusters=%d\n", len(api.ProfilePaths()), len(api.TribePaths()), len(api.ClusterPaths()))

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
}
