package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <ark-files.json> <out.json>", os.Args[0])
	}
	report, err := arkapi.EquipmentHistoryReportFromManifestPath(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(os.Args[2], data, 0o600); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("saves=%d initial=%d changes=%d final=%d wrote=%s\n", report.Saves, report.InitialCount, len(report.Changes), report.FinalCount, os.Args[2])
}
