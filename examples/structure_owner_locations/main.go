package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 2 && len(os.Args) != 3 && len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: structure_owner_locations <save.ark> [map] [digits]")
		os.Exit(2)
	}
	mapName := ""
	if len(os.Args) >= 3 {
		mapName = os.Args[2]
	}
	digits := 1
	if len(os.Args) == 4 {
		value, err := strconv.Atoi(os.Args[3])
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse digits: %v\n", err)
			os.Exit(2)
		}
		digits = value
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	export, _, err := arkapi.NewStructure(save).OwnerLocationsWithFaults(mapName, digits, arkapi.NewPlayer(save))
	if err != nil {
		fmt.Fprintf(os.Stderr, "read structure owner locations: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf(
		"structures=%d owners=%d cells=%d named_cells=%d multi_structure_cells=%d faults=%d\n",
		export.Structures,
		export.Owners,
		export.Cells,
		export.NamedCells,
		export.MultiStructureCells,
		export.FaultCount,
	)
	encoded, err := json.MarshalIndent(export.OwnersByLocation, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "encode owner locations: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(encoded))
}
