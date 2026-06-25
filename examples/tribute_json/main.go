package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <tribute-file-or-directory>", os.Args[0])
	}

	info, err := os.Stat(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() {
		entries, err := arktribute.OpenDirectory(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		data, err := arkapi.ExportTributeDirectoryDataJSON(entries)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(data))
		return
	}

	tribute, err := arktribute.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	data, err := arkapi.ExportTributeDataJSON(tribute)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(data))
}
