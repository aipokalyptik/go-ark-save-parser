package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestProvidedDataReadOnlyE2E(t *testing.T) {
	data := providedData(t)
	if data.savePath == "" {
		t.Skip("set ARK_E2E_SAVE or ARK_E2E_SAVE_DIR to run provided-data CLI read-only E2E")
	}

	var inspectOut bytes.Buffer
	if err := run([]string{"inspect", data.savePath}, &inspectOut); err != nil {
		t.Fatalf("run(inspect) error = %v", err)
	}
	for _, want := range []string{"Map:", "Save version:", "Objects:"} {
		if !strings.Contains(inspectOut.String(), want) {
			t.Fatalf("inspect output missing %q", want)
		}
	}

	exportPath := filepath.Join(t.TempDir(), "save-info.json")
	var exportOut bytes.Buffer
	if err := run([]string{"--redact", "export-json", data.savePath, exportPath}, &exportOut); err != nil {
		t.Fatalf("run(export-json) error = %v", err)
	}
	if !strings.Contains(exportOut.String(), "Wrote JSON export: [redacted]") {
		t.Fatalf("export-json output was not redacted: %q", exportOut.String())
	}
	exportData, err := os.ReadFile(exportPath)
	if err != nil {
		t.Fatalf("read exported JSON: %v", err)
	}
	var info struct {
		MapName     string `json:"map_name"`
		ObjectCount int    `json:"object_count"`
	}
	if err := json.Unmarshal(exportData, &info); err != nil {
		t.Fatalf("unmarshal exported JSON: %v", err)
	}
	if info.MapName == "" {
		t.Fatalf("exported JSON map_name is empty")
	}
	if info.ObjectCount == 0 {
		t.Fatalf("exported JSON object_count = 0")
	}

	if data.dir == "" {
		return
	}
	if data.profileCount > 0 {
		var playersOut bytes.Buffer
		if err := run([]string{"--redact", "players", data.dir}, &playersOut); err != nil {
			t.Fatalf("run(players directory) error = %v", err)
		}
		for _, want := range []string{"Player directory: [redacted]", "Profiles:", "Players:"} {
			if !strings.Contains(playersOut.String(), want) {
				t.Fatalf("players directory output missing %q", want)
			}
		}
	}
	if data.tribeCount > 0 {
		var tribesOut bytes.Buffer
		if err := run([]string{"--redact", "tribes", data.dir}, &tribesOut); err != nil {
			t.Fatalf("run(tribes directory) error = %v", err)
		}
		for _, want := range []string{"Tribe directory: [redacted]", "Tribe files:", "Tribes:"} {
			if !strings.Contains(tribesOut.String(), want) {
				t.Fatalf("tribes directory output missing %q", want)
			}
		}
	}
}

type providedDataSet struct {
	savePath     string
	dir          string
	profileCount int
	tribeCount   int
}

func providedData(t *testing.T) providedDataSet {
	t.Helper()
	data := providedDataSet{savePath: os.Getenv("ARK_E2E_SAVE"), dir: os.Getenv("ARK_E2E_SAVE_DIR")}
	if data.dir == "" {
		return data
	}

	var savePaths []string
	err := filepath.WalkDir(data.dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".ark":
			savePaths = append(savePaths, path)
		case ".arkprofile":
			data.profileCount++
		case ".arktribe":
			data.tribeCount++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("discover provided data files in %q: %v", data.dir, err)
	}
	sort.Strings(savePaths)
	if data.savePath == "" && len(savePaths) > 0 {
		data.savePath = savePaths[0]
	}
	return data
}
