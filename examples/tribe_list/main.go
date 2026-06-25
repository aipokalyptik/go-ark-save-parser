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
	tribes, _, err := api.TribeDetailsWithFaults()
	if err != nil {
		_ = save.Close()
		return nil, func() {}, err
	}
	if len(tribes) == 0 {
		_ = save.Close()
		api, err := arkapi.NewPlayerFromDirectory(filepath.Dir(path))
		return api, func() {}, err
	}
	return api, func() { _ = save.Close() }, nil
}
