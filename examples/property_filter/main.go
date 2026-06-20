package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <save.ark> <property> [property...]", os.Args[0])
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	objects, err := save.ObjectClassInfosWithAnyProperty(os.Args[2:])
	if err != nil {
		log.Fatal(err)
	}
	classes := map[string]struct{}{}
	for _, object := range objects {
		classes[object.ClassName] = struct{}{}
	}
	fmt.Printf("objects=%d classes=%d\n", len(objects), len(classes))
}
