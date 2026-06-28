package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/google/uuid"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <object-uuid>", os.Args[0])
	}
	id, err := uuid.Parse(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	api, closeAPI, err := arkapi.NewGeneralFromPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer closeAPI()

	summary, err := api.ObjectSummary(id)
	if err != nil {
		log.Fatal(err)
	}
	if !summary.Exists {
		fmt.Println("has_object=0")
		return
	}
	fmt.Printf("has_object=1 bytes=%d properties=%d\n", summary.Bytes, summary.Properties)
}
