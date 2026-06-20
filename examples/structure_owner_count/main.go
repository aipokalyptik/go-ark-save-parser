package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: structure_owner_count <save.ark> <tribe-id>")
		os.Exit(2)
	}
	tribeID64, err := strconv.ParseInt(os.Args[2], 10, 32)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse tribe id: %v\n", err)
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	api := arkapi.NewStructure(save)
	owned, err := api.OwnedBy(arkobject.ObjectOwner{TribeID: int32(tribeID64)})
	if err != nil {
		fmt.Fprintf(os.Stderr, "read structures: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("tribe_id=%d structures=%d\n", tribeID64, len(owned))
}
