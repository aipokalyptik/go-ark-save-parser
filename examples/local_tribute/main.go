package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
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
		files, err := arktribute.OpenDirectory(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		var playerIDs int
		var tribeIDs int
		for _, file := range files {
			playerIDs += len(file.PlayerDataIDs)
			tribeIDs += len(file.TribeDataIDs)
		}
		fmt.Printf("tribute_files=%d player_data_ids=%d tribe_data_ids=%d\n", len(files), playerIDs, tribeIDs)
		return
	}

	file, err := arktribute.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tribute_file=%s player_data_ids=%d tribe_data_ids=%d\n", file.ID, len(file.PlayerDataIDs), len(file.TribeDataIDs))
}
