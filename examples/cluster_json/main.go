package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <local-cluster-file-or-directory>", os.Args[0])
	}
	data, err := arkapi.ExportClusterPathJSON(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
