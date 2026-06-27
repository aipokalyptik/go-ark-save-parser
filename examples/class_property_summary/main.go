package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <class-substring>", os.Args[0])
	}
	classSubstring := os.Args[2]
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	summary, faults, err := arkapi.NewGeneral(save).ClassPropertySummaryWithFaults(classSubstring)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("objects=%d properties=%d faults=%d\n", summary.Objects, summary.Properties, len(faults))
}
