package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
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

	var objectClassesOut bytes.Buffer
	if err := run([]string{"object-classes", data.SavePath}, &objectClassesOut); err != nil {
		t.Fatalf("run(object-classes) error = %v", err)
	}
	if !strings.Contains(objectClassesOut.String(), "Blueprint") {
		t.Fatalf("object-classes output missing Blueprint class marker")
	}

	save, err := arksave.Open(data.SavePath)
	if err != nil {
		t.Fatalf("open provided save for object-summary id: %v", err)
	}
	objectIDs, err := save.ObjectIDs()
	if closeErr := save.Close(); closeErr != nil {
		t.Fatalf("close provided save after object id lookup: %v", closeErr)
	}
	if err != nil {
		t.Fatalf("ObjectIDs() error = %v", err)
	}
	if len(objectIDs) == 0 {
		t.Fatalf("provided save has no objects")
	}
	var objectSummaryOut bytes.Buffer
	if err := run([]string{"object-summary", data.SavePath, objectIDs[0].String()}, &objectSummaryOut); err != nil {
		t.Fatalf("run(object-summary) error = %v", err)
	}
	for _, want := range []string{"Exists: true", "Bytes:", "Properties:"} {
		if !strings.Contains(objectSummaryOut.String(), want) {
			t.Fatalf("object-summary output missing %q", want)
		}
	}

	var classLookupOut bytes.Buffer
	if err := run([]string{"class-lookup", data.SavePath, "PrimalStructure"}, &classLookupOut); err != nil {
		t.Fatalf("run(class-lookup) error = %v", err)
	}
	for _, want := range []string{"Objects:", "Classes:", "Parse faults:"} {
		if !strings.Contains(classLookupOut.String(), want) {
			t.Fatalf("class-lookup output missing %q", want)
		}
	}

	var classPropertySummaryOut bytes.Buffer
	if err := run([]string{"class-property-summary", data.SavePath, "PrimalStructure"}, &classPropertySummaryOut); err != nil {
		t.Fatalf("run(class-property-summary) error = %v", err)
	}
	for _, want := range []string{"Objects:", "Properties:", "Parse faults:"} {
		if !strings.Contains(classPropertySummaryOut.String(), want) {
			t.Fatalf("class-property-summary output missing %q", want)
		}
	}

	var propertyFilterOut bytes.Buffer
	if err := run([]string{"property-filter", data.SavePath, "Health", "MaxHealth"}, &propertyFilterOut); err != nil {
		t.Fatalf("run(property-filter) error = %v", err)
	}
	for _, want := range []string{"Objects:", "Classes:"} {
		if !strings.Contains(propertyFilterOut.String(), want) {
			t.Fatalf("property-filter output missing %q", want)
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

	var ownerCountOut bytes.Buffer
	if err := run([]string{"--redact", "structure-owner-count", data.SavePath, "555"}, &ownerCountOut); err != nil {
		t.Fatalf("run(structure-owner-count) error = %v", err)
	}
	for _, want := range []string{"Tribe ID: [redacted]", "Structures:", "Parse faults:"} {
		if !strings.Contains(ownerCountOut.String(), want) {
			t.Fatalf("structure-owner-count output missing %q", want)
		}
	}

	var ownersOut bytes.Buffer
	if err := run([]string{"structure-owners", data.SavePath}, &ownersOut); err != nil {
		t.Fatalf("run(structure-owners) error = %v", err)
	}
	for _, want := range []string{"Structures:", "With tribe ID:", "Unique tribes:", "Parse faults:"} {
		if !strings.Contains(ownersOut.String(), want) {
			t.Fatalf("structure-owners output missing %q", want)
		}
	}

	var ownerLocationsOut bytes.Buffer
	if err := run([]string{"--redact", "structure-owner-locations", data.SavePath}, &ownerLocationsOut); err != nil {
		t.Fatalf("run(structure-owner-locations) error = %v", err)
	}
	for _, want := range []string{"Structures:", "Owners:", "Cells:", "Parse faults:"} {
		if !strings.Contains(ownerLocationsOut.String(), want) {
			t.Fatalf("structure-owner-locations output missing %q", want)
		}
	}

	heatmapPath := filepath.Join(t.TempDir(), "structure-heatmap.json")
	var heatmapOut bytes.Buffer
	if err := run([]string{"structure-heatmap", data.SavePath, heatmapPath, "100", "1"}, &heatmapOut); err != nil {
		t.Fatalf("run(structure-heatmap) error = %v", err)
	}
	for _, want := range []string{"Cells:", "Total:", "Max:", "Parse faults:", "Wrote:"} {
		if !strings.Contains(heatmapOut.String(), want) {
			t.Fatalf("structure-heatmap output missing %q", want)
		}
	}
	heatmapData, err := os.ReadFile(heatmapPath)
	if err != nil {
		t.Fatalf("read structure heatmap JSON: %v", err)
	}
	var heatmap struct {
		Resolution   int `json:"resolution"`
		NonzeroCells int `json:"nonzero_cells"`
		Total        int `json:"total"`
		Max          int `json:"max"`
		Faults       int `json:"faults"`
	}
	if err := json.Unmarshal(heatmapData, &heatmap); err != nil {
		t.Fatalf("unmarshal structure heatmap JSON: %v", err)
	}
	if heatmap.Resolution != 100 || heatmap.Total < 0 || heatmap.NonzeroCells < 0 || heatmap.Max < 0 || heatmap.Faults < 0 {
		t.Fatalf("structure heatmap JSON = %#v, want valid aggregate summary", heatmap)
	}

	var baseComponentsOut bytes.Buffer
	if err := run([]string{"base-components", data.SavePath}, &baseComponentsOut); err != nil {
		t.Fatalf("run(base-components) error = %v", err)
	}
	for _, want := range []string{"Components:", "Total structures:", "Parse faults:"} {
		if !strings.Contains(baseComponentsOut.String(), want) {
			t.Fatalf("base-components output missing %q", want)
		}
	}

	var dinosOut bytes.Buffer
	if err := run([]string{"dinos", data.SavePath}, &dinosOut); err != nil {
		t.Fatalf("run(dinos) error = %v", err)
	}
	for _, want := range []string{"Dinos:", "Tamed:", "Wild:", "Parse faults:"} {
		if !strings.Contains(dinosOut.String(), want) {
			t.Fatalf("dinos output missing %q", want)
		}
	}

	var wildTamablesOut bytes.Buffer
	if err := run([]string{"dino-wild-tamables", data.SavePath}, &wildTamablesOut); err != nil {
		t.Fatalf("run(dino-wild-tamables) error = %v", err)
	}
	for _, want := range []string{"Wild dinos:", "Wild tamables:", "Parse faults:"} {
		if !strings.Contains(wildTamablesOut.String(), want) {
			t.Fatalf("dino-wild-tamables output missing %q", want)
		}
	}

	var babyDinosOut bytes.Buffer
	if err := run([]string{"dino-babies", data.SavePath}, &babyDinosOut); err != nil {
		t.Fatalf("run(dino-babies) error = %v", err)
	}
	for _, want := range []string{"Baby dinos:", "Tamed babies:", "Wild babies:", "Parse faults:"} {
		if !strings.Contains(babyDinosOut.String(), want) {
			t.Fatalf("dino-babies output missing %q", want)
		}
	}

	var bestStatOut bytes.Buffer
	if err := run([]string{"dino-best-stat", data.SavePath}, &bestStatOut); err != nil {
		t.Fatalf("run(dino-best-stat) error = %v", err)
	}
	for _, want := range []string{"Best stat:", "Parse faults:"} {
		if !strings.Contains(bestStatOut.String(), want) {
			t.Fatalf("dino-best-stat output missing %q", want)
		}
	}

	var mostMutatedOut bytes.Buffer
	if err := run([]string{"dino-most-mutated", data.SavePath}, &mostMutatedOut); err != nil {
		t.Fatalf("run(dino-most-mutated) error = %v", err)
	}
	if !strings.Contains(mostMutatedOut.String(), "Most mutated:") {
		t.Fatalf("dino-most-mutated output missing summary")
	}

	var wildTamedOut bytes.Buffer
	if err := run([]string{"dino-wild-tamed", data.SavePath}, &wildTamedOut); err != nil {
		t.Fatalf("run(dino-wild-tamed) error = %v", err)
	}
	for _, want := range []string{"Wild-tamed dinos:", "Max level:", "Parse faults:"} {
		if !strings.Contains(wildTamedOut.String(), want) {
			t.Fatalf("dino-wild-tamed output missing %q", want)
		}
	}

	var equipmentSummaryOut bytes.Buffer
	if err := run([]string{"equipment-summary", data.SavePath}, &equipmentSummaryOut); err != nil {
		t.Fatalf("run(equipment-summary) error = %v", err)
	}
	for _, want := range []string{"Items:", "Weapon items:", "Blueprints:", "Parse faults:"} {
		if !strings.Contains(equipmentSummaryOut.String(), want) {
			t.Fatalf("equipment-summary output missing %q", want)
		}
	}

	var equipmentSaddlesOut bytes.Buffer
	if err := run([]string{"equipment-saddles", data.SavePath}, &equipmentSaddlesOut); err != nil {
		t.Fatalf("run(equipment-saddles) error = %v", err)
	}
	for _, want := range []string{"Item saddles:", "Cryopod saddles:", "Total saddles:", "Parse faults:"} {
		if !strings.Contains(equipmentSaddlesOut.String(), want) {
			t.Fatalf("equipment-saddles output missing %q", want)
		}
	}

	var equipmentBestOut bytes.Buffer
	if err := run([]string{"equipment-best", data.SavePath}, &equipmentBestOut); err != nil {
		t.Fatalf("run(equipment-best) error = %v", err)
	}
	for _, want := range []string{"Best weapon", "Best armor", "Parse faults:"} {
		if !strings.Contains(equipmentBestOut.String(), want) {
			t.Fatalf("equipment-best output missing %q", want)
		}
	}

	var equipmentRankOut bytes.Buffer
	if err := run([]string{"equipment-rank", data.SavePath}, &equipmentRankOut); err != nil {
		t.Fatalf("run(equipment-rank) error = %v", err)
	}
	for _, want := range []string{"Ranked:", "Best rating:", "Parse faults:"} {
		if !strings.Contains(equipmentRankOut.String(), want) {
			t.Fatalf("equipment-rank output missing %q", want)
		}
	}

	var stackablesOut bytes.Buffer
	if err := run([]string{"stackables", data.SavePath}, &stackablesOut); err != nil {
		t.Fatalf("run(stackables) error = %v", err)
	}
	for _, want := range []string{"Stackable items:", "Total quantity:", "Parse faults:"} {
		if !strings.Contains(stackablesOut.String(), want) {
			t.Fatalf("stackables output missing %q", want)
		}
	}

	var stackableOwnedOut bytes.Buffer
	if err := run([]string{
		"--redact",
		"stackable-owned-by",
		data.SavePath,
		"Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
		"555",
	}, &stackableOwnedOut); err != nil {
		t.Fatalf("run(stackable-owned-by) error = %v", err)
	}
	for _, want := range []string{"Tribe ID: [redacted]", "Blueprint: [redacted]", "Items:", "Total quantity:"} {
		if !strings.Contains(stackableOwnedOut.String(), want) {
			t.Fatalf("stackable-owned-by output missing %q", want)
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

	for _, domain := range e2etest.DomainJSONExportDomains() {
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
			Count  int             `json:"count"`
			Domain string          `json:"domain"`
			Items  json.RawMessage `json:"items"`
		}
		if err := json.Unmarshal(exportData, &export); err != nil {
			t.Fatalf("unmarshal %s domain JSON export: %v", domain, err)
		}
		if export.Domain != domain || export.Count < 0 || !json.Valid(export.Items) {
			t.Fatalf("%s domain JSON export = %#v, want matching domain with valid items field", domain, export)
		}
	}

	if data.ClusterPath != "" {
		var clusterSummaryOut bytes.Buffer
		if err := run([]string{"--redact", "cluster-summary", data.ClusterPath}, &clusterSummaryOut); err != nil {
			t.Fatalf("run(cluster-summary) error = %v", err)
		}
		for _, want := range []string{"Cluster file: [redacted]", "Items:", "Dinos:", "Dino item uploads:", "Parsed dinos:"} {
			if !strings.Contains(clusterSummaryOut.String(), want) {
				t.Fatalf("cluster-summary output missing %q", want)
			}
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

		var inventoryOut bytes.Buffer
		if err := run([]string{"player-inventories", data.SavePath}, &inventoryOut); err != nil {
			t.Fatalf("run(player-inventories) error = %v", err)
		}
		for _, want := range []string{"Players:", "With inventory:", "Total items:", "Inventory faults:"} {
			if !strings.Contains(inventoryOut.String(), want) {
				t.Fatalf("player-inventories output missing %q", want)
			}
		}

		var rosterOut bytes.Buffer
		if err := run([]string{"player-roster", data.SavePath}, &rosterOut); err != nil {
			t.Fatalf("run(player-roster) error = %v", err)
		}
		for _, want := range []string{"Players:", "With names:", "Highest level:"} {
			if !strings.Contains(rosterOut.String(), want) {
				t.Fatalf("player-roster output missing %q", want)
			}
		}
	}
	if data.ProfilePath != "" {
		var playerOut bytes.Buffer
		if err := run([]string{"--redact", "players", data.ProfilePath}, &playerOut); err != nil {
			t.Fatalf("run(players profile) error = %v", err)
		}
		for _, want := range []string{"Player profile: [redacted]", "Character name:", "Player data ID:", "Deaths:"} {
			if !strings.Contains(playerOut.String(), want) {
				t.Fatalf("players profile output missing %q", want)
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

		var rosterOut bytes.Buffer
		if err := run([]string{"tribe-roster", data.SavePath}, &rosterOut); err != nil {
			t.Fatalf("run(tribe-roster) error = %v", err)
		}
		for _, want := range []string{"Tribes:", "With names:", "Members:", "Dinos:"} {
			if !strings.Contains(rosterOut.String(), want) {
				t.Fatalf("tribe-roster output missing %q", want)
			}
		}
	}
	if data.ProfileCount > 0 && data.TribeCount > 0 {
		var linksOut bytes.Buffer
		if err := run([]string{"player-tribe-links", data.Dir}, &linksOut); err != nil {
			t.Fatalf("run(player-tribe-links) error = %v", err)
		}
		for _, want := range []string{"Players:", "Tribes:", "Active links:", "Inactive members:"} {
			if !strings.Contains(linksOut.String(), want) {
				t.Fatalf("player-tribe-links output missing %q", want)
			}
		}
	}
	if data.TribePath != "" {
		var tribeOut bytes.Buffer
		if err := run([]string{"--redact", "tribes", data.TribePath}, &tribeOut); err != nil {
			t.Fatalf("run(tribe file) error = %v", err)
		}
		for _, want := range []string{"Tribe save: [redacted]", "Tribe name:", "Members:", "Dinos:"} {
			if !strings.Contains(tribeOut.String(), want) {
				t.Fatalf("tribe file output missing %q", want)
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
	if data.TributePath != "" {
		tributeExportPath := filepath.Join(t.TempDir(), "single-tribute.json")
		var tributeExportOut bytes.Buffer
		if err := run([]string{"--redact", "export-tribute-json", data.TributePath, tributeExportPath}, &tributeExportOut); err != nil {
			t.Fatalf("run(export-tribute-json file) error = %v", err)
		}
		if !strings.Contains(tributeExportOut.String(), "Wrote tribute JSON export: [redacted]") {
			t.Fatalf("single export-tribute-json output was not redacted: %q", tributeExportOut.String())
		}
		exportData, err := os.ReadFile(tributeExportPath)
		if err != nil {
			t.Fatalf("read single tribute JSON export: %v", err)
		}
		var info struct {
			PlayerDataCount int `json:"player_data_count"`
			TribeDataCount  int `json:"tribe_data_count"`
		}
		if err := json.Unmarshal(exportData, &info); err != nil {
			t.Fatalf("unmarshal single tribute JSON export: %v", err)
		}
		if info.PlayerDataCount == 0 && info.TribeDataCount == 0 {
			t.Fatalf("single tribute JSON export has no IDs: %#v", info)
		}
	}
}
