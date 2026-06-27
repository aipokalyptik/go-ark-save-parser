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

	var healthOut bytes.Buffer
	if err := run([]string{"structure-health", data.SavePath}, &healthOut); err != nil {
		t.Fatalf("run(structure-health) error = %v", err)
	}
	for _, want := range []string{"Structures:", "With health:", "Parse faults:"} {
		if !strings.Contains(healthOut.String(), want) {
			t.Fatalf("structure-health output missing %q", want)
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

	for _, domain := range []string{"stackables"} {
		outPath := filepath.Join(t.TempDir(), domain+".json")
		var domainOut bytes.Buffer
		if err := run([]string{"--redact", "export-domain-json", data.SavePath, domain, outPath}, &domainOut); err != nil {
			t.Fatalf("run(export-domain-json %s) error = %v", domain, err)
		}
		if !strings.Contains(domainOut.String(), "Wrote "+domain+" JSON export: [redacted]") {
			t.Fatalf("export-domain-json %s output was not redacted: %q", domain, domainOut.String())
		}
		exportData, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("read %s domain JSON export: %v", domain, err)
		}
		var export struct {
			Domain string          `json:"domain"`
			Items  json.RawMessage `json:"items"`
		}
		if err := json.Unmarshal(exportData, &export); err != nil {
			t.Fatalf("unmarshal %s domain JSON export: %v", domain, err)
		}
		if export.Domain != domain || len(export.Items) == 0 {
			t.Fatalf("%s domain JSON export = %#v, want matching domain with items field", domain, export)
		}
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
	if data.TributeCount > 0 {
		var tributeOut bytes.Buffer
		if err := run([]string{"--redact", "tribute", data.Dir}, &tributeOut); err != nil {
			t.Fatalf("run(tribute directory) error = %v", err)
		}
		for _, want := range []string{"Tribute file: [redacted]", "Player data IDs:", "Tribe data IDs:"} {
			if !strings.Contains(tributeOut.String(), want) {
				t.Fatalf("tribute directory output missing %q", want)
			}
		}

		tributeExportPath := filepath.Join(t.TempDir(), "tribute.json")
		var tributeExportOut bytes.Buffer
		if err := run([]string{"--redact", "export-tribute-json", data.Dir, tributeExportPath}, &tributeExportOut); err != nil {
			t.Fatalf("run(export-tribute-json directory) error = %v", err)
		}
		if !strings.Contains(tributeExportOut.String(), "Wrote tribute JSON export: [redacted]") {
			t.Fatalf("export-tribute-json output was not redacted: %q", tributeExportOut.String())
		}
		exportData, err := os.ReadFile(tributeExportPath)
		if err != nil {
			t.Fatalf("read tribute JSON export: %v", err)
		}
		var info struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(exportData, &info); err != nil {
			t.Fatalf("unmarshal tribute JSON export: %v", err)
		}
		if info.Count == 0 {
			t.Fatalf("exported tribute JSON count = 0")
		}
	}
}
