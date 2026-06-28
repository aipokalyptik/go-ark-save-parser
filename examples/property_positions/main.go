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

	summary, err := api.PropertyPositionSummary(id)
	if err != nil {
		log.Fatal(err)
	}
	if !summary.Exists {
		fmt.Println("has_object=0")
		return
	}
	fmt.Printf(
		"has_object=1 properties=%d name_offsets=%d value_offsets=%d encoded=%d positioned=%d offsets_ok=%d\n",
		summary.Properties,
		summary.NameOffsets,
		summary.ValueOffsets,
		summary.Encoded,
		summary.Positioned,
		summary.OffsetsOK,
	)
}
