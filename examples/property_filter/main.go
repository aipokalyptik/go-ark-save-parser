package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <save.ark> <property> [property...]", os.Args[0])
	}
	api, closeAPI, err := arkapi.NewGeneralFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer closeAPI()

	summary, err := api.PropertyFilterSummary(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("objects=%d classes=%d\n", summary.Objects, summary.Classes)
}
