package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark>", os.Args[0])
	}
	api, closeAPI, err := arkapi.NewGeneralFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer closeAPI()

	classes, err := api.Classes()
	if err != nil {
		log.Fatal(err)
	}
	for _, className := range classes {
		fmt.Println(className)
	}
}
