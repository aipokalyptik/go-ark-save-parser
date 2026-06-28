package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <class-substring>", os.Args[0])
	}
	classSubstring := os.Args[2]
	api, closeAPI, err := arkapi.NewGeneralFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer closeAPI()

	summary, faults, err := api.ClassPropertySummaryWithFaults(classSubstring)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("objects=%d properties=%d faults=%d\n", summary.Objects, summary.Properties, len(faults))
}
