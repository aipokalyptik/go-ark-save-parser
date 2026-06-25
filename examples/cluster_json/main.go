package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <local-cluster-file>", os.Args[0])
	}
	cluster, err := arkcluster.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	data, err := arkapi.ExportClusterDataJSON(cluster)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
