package main

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: equipment_saddles <save.ark>")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	summary, _, err := arkapi.NewEquipment(save).SaddleSummaryWithFaults()
	if err != nil {
		fmt.Fprintf(os.Stderr, "read saddles: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(
		"item_saddles=%d cryopod_saddles=%d total_saddles=%d max_armor=%.1f\n",
		summary.ItemSaddles,
		summary.CryopodSaddles,
		summary.TotalSaddles,
		summary.MaxArmor,
	)
}
