package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkmutation"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <out.ark>", os.Args[0])
	}
	if err := arkmutation.CopySave(os.Args[1], os.Args[2]); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote copy: %s\n", os.Args[2])
}
