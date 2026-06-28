package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: structure_owners <save.ark>")
		os.Exit(2)
	}

	api, closeAPI, err := arkapi.NewStructureFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	summary, faults, err := api.OwnerSummaryWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read structure owners: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(
		"structures=%d with_tribe_id=%d with_player_id=%d with_tribe_name=%d with_player_name=%d with_original_placer_id=%d unique_tribes=%d unique_players=%d unique_original_placers=%d faults=%d\n",
		summary.Structures,
		summary.WithTribeID,
		summary.WithPlayerID,
		summary.WithTribeName,
		summary.WithPlayerName,
		summary.WithOriginalPlacerID,
		summary.UniqueTribes,
		summary.UniquePlayers,
		summary.UniqueOriginalPlacers,
		len(faults),
	)
}
