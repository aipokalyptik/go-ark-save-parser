package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
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
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	object, err := save.Object(id)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("has_object=0")
		return
	}
	if err != nil {
		log.Fatal(err)
	}

	nameOffsets := 0
	valueOffsets := 0
	encoded := 0
	positioned := 0
	offsetsOK := 0
	for _, property := range object.Properties {
		if property.NameOffset > 0 {
			nameOffsets++
		}
		if property.ValueOffset > 0 {
			valueOffsets++
		}
		if len(property.EncodedBytes) > 0 {
			encoded++
		}
		if property.Position != 0 {
			positioned++
		}
		if property.NameOffset >= 0 && property.ValueOffset > property.NameOffset && len(property.EncodedBytes) > 0 {
			offsetsOK++
		}
	}
	fmt.Printf(
		"has_object=1 properties=%d name_offsets=%d value_offsets=%d encoded=%d positioned=%d offsets_ok=%d\n",
		len(object.Properties),
		nameOffsets,
		valueOffsets,
		encoded,
		positioned,
		offsetsOK,
	)
}
