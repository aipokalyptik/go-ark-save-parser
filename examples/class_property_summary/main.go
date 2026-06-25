package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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

	objects, faults, err := save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return strings.Contains(info.ClassName, classSubstring)
	})
	if err != nil {
		log.Fatal(err)
	}
	properties := map[string]struct{}{}
	for _, info := range objects {
		for _, property := range info.Object.Properties {
			properties[property.Name] = struct{}{}
		}
	}
	fmt.Printf("objects=%d properties=%d faults=%d\n", len(objects), len(properties), len(faults))
}
