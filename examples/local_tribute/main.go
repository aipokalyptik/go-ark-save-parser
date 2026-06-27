package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <tribute-file-or-directory>", os.Args[0])
	}

	info, err := os.Stat(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() {
		summary, err := arkapi.TributeDirectorySummaryFromPath(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		var playerIDs int
		var tribeIDs int
		for _, file := range summary.Files {
			playerIDs += file.PlayerDataCount
			tribeIDs += file.TribeDataCount
		}
		fmt.Printf("tribute_files=%d player_data_ids=%d tribe_data_ids=%d\n", summary.Count, playerIDs, tribeIDs)
		return
	}

	file, err := arkapi.TributeSummaryFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tribute_file=%s player_data_ids=%d tribe_data_ids=%d\n", file.ID, file.PlayerDataCount, file.TribeDataCount)
}
