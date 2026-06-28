package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
)

func main() {
	if len(os.Args) != 5 && len(os.Args) != 6 {
		fmt.Fprintln(os.Stderr, "usage: structure_at_location <save.ark> <map> <lat> <lon> [radius]")
		os.Exit(2)
	}

	lat, err := strconv.ParseFloat(os.Args[3], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse lat: %v\n", err)
		os.Exit(2)
	}
	lon, err := strconv.ParseFloat(os.Args[4], 64)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse lon: %v\n", err)
		os.Exit(2)
	}
	radius := 0.3
	if len(os.Args) == 6 {
		radius, err = strconv.ParseFloat(os.Args[5], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "parse radius: %v\n", err)
			os.Exit(2)
		}
	}

	api, closeAPI, err := arkapi.NewStructureFromPath(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer closeAPI()

	summary, _, err := api.AtLocationSummaryWithFaults(os.Args[2], arkobject.MapCoords{Lat: lat, Long: lon}, radius, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "find structures: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("structures=%d connected=%d\n", summary.Structures, summary.Connected)
}
