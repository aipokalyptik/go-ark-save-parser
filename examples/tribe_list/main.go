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
	api, closeAPI, err := arkapi.NewPlayerFromPath(os.Args[1], arkapi.PlayerPathOptions{Fallback: arkapi.PlayerPathFallbackTribes})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := closeAPI(); err != nil {
			log.Fatal(err)
		}
	}()

	tribes, err := api.TribeDetails()
	if err != nil {
		log.Fatal(err)
	}
	withNames := 0
	members := 0
	dinos := int32(0)
	for _, tribe := range tribes {
		if tribe.Name != "" {
			withNames++
		}
		members += len(tribe.MemberIDs)
		dinos += tribe.NumDinos
	}
	fmt.Printf("tribes=%d with_names=%d members=%d dinos=%d\n", len(tribes), withNames, members, dinos)
}
