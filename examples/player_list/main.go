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
	summary, faults, err := arkapi.PlayerRosterSummaryFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("players=%d with_names=%d highest_level=%d faults=%d\n", summary.Players, summary.WithNames, summary.HighestLevel, len(faults))
}
