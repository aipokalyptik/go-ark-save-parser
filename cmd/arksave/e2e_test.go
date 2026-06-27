package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/internal/e2etest"
)

func TestProvidedDataReadOnlyE2E(t *testing.T) {
	data := e2etest.DiscoverProvidedData(t)
	if data.SavePath == "" {
		t.Skip("set ARK_E2E_SAVE or ARK_E2E_SAVE_DIR to run provided-data CLI read-only E2E")
	}

	var inspectOut bytes.Buffer
	if err := run([]string{"inspect", data.SavePath}, &inspectOut); err != nil {
		t.Fatalf("run(inspect) error = %v", err)
	}
	for _, want := range []string{"Map:", "Save version:", "Objects:"} {
		if !strings.Contains(inspectOut.String(), want) {
			t.Fatalf("inspect output missing %q", want)
		}
	}

	exportPath := filepath.Join(t.TempDir(), "save-info.json")
	var exportOut bytes.Buffer
	if err := run([]string{"--redact", "export-json", data.SavePath, exportPath}, &exportOut); err != nil {
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

	if data.Dir == "" {
		return
	}
	if data.ProfileCount > 0 {
		var playersOut bytes.Buffer
		if err := run([]string{"--redact", "players", data.Dir}, &playersOut); err != nil {
			t.Fatalf("run(players directory) error = %v", err)
		}
		for _, want := range []string{"Player directory: [redacted]", "Profiles:", "Players:"} {
			if !strings.Contains(playersOut.String(), want) {
				t.Fatalf("players directory output missing %q", want)
			}
		}
	}
	if data.TribeCount > 0 {
		var tribesOut bytes.Buffer
		if err := run([]string{"--redact", "tribes", data.Dir}, &tribesOut); err != nil {
			t.Fatalf("run(tribes directory) error = %v", err)
		}
		for _, want := range []string{"Tribe directory: [redacted]", "Tribe files:", "Tribes:"} {
			if !strings.Contains(tribesOut.String(), want) {
				t.Fatalf("tribes directory output missing %q", want)
			}
		}
	}
}
