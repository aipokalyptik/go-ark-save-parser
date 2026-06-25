package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

type exportManifest struct {
	Files []exportFile `json:"files"`
}

type exportFile struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <save.ark> <out-dir>", os.Args[0])
	}
	if err := os.MkdirAll(os.Args[2], 0o755); err != nil {
		log.Fatal(err)
	}
	save, err := arksave.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer save.Close()

	api := arkapi.NewJSON(save)
	manifest := exportManifest{}
	saveInfo, err := api.ExportSaveInfoJSON()
	if err != nil {
		log.Fatal(err)
	}
	if err := writeJSONFile(os.Args[2], "save-info.json", saveInfo); err != nil {
		log.Fatal(err)
	}
	manifest.Files = append(manifest.Files, exportFile{Name: "save-info.json", Count: 1})

	for _, domain := range []string{"dinos", "structures", "equipment", "stackables", "players", "tribes", "bases"} {
		exported, err := api.ExportDomain(domain)
		if err != nil {
			log.Fatal(err)
		}
		data, err := json.MarshalIndent(exported, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		name := domain + ".json"
		if err := writeJSONFile(os.Args[2], name, data); err != nil {
			log.Fatal(err)
		}
		manifest.Files = append(manifest.Files, exportFile{Name: name, Count: exported.Count})
	}

	manifestData, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	manifestData = append(manifestData, '\n')
	if err := writeJSONFile(os.Args[2], "manifest.json", manifestData); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("exports=%d wrote=%s\n", len(manifest.Files), os.Args[2])
}

func writeJSONFile(dir string, name string, data []byte) error {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}
	return os.WriteFile(filepath.Join(dir, name), data, 0o644)
}
