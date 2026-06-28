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
	summary, faults, err := arkapi.TribeRosterSummaryFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("tribes=%d with_names=%d members=%d dinos=%d faults=%d\n", summary.Tribes, summary.WithNames, summary.Members, summary.Dinos, len(faults))
}
