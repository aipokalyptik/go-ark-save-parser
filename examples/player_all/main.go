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
	summary, faults, err := arkapi.PlayerAllSummaryFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("players=%d tribes=%d highest_level=%d total_deaths=%d unlocked_engrams=%d faults=%d\n", summary.Players, summary.Tribes, summary.HighestLevel, summary.TotalDeaths, summary.UnlockedEngrams, len(faults))
}
