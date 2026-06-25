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

	raw, err := save.ObjectBinary(id)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("has_object=0")
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	object, err := save.Object(id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("has_object=1 bytes=%d properties=%d\n", len(raw), len(object.Properties))
}
