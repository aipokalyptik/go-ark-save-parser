package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestInspectCommandPrintsOfflineSaveSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)

	var out bytes.Buffer
	err := run([]string{"inspect", path}, &out)
	if err != nil {
		t.Fatalf("run(inspect) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Map: Valguero_WP",
		"Save version: 12",
		"Objects: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("inspect output %q does not contain %q", got, want)
		}
	}
}

func TestDinoAggregateCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"dinos", "dinoWildTamables", "dinoBabies", "dinoBestStat", "dinoBestBaseStat", "dinoMostMutated", "dinoWildTamed"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arksave.Open") {
			t.Fatalf("%s() still opens saves directly; use typed arkapi path helper", name)
		}
	}
}

func TestEquipmentAggregateCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"equipmentSummary", "equipmentSaddles", "equipmentBest", "equipmentRank", "equipmentAscendantWeaponBPs", "equipmentOwnedBy"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arksave.Open") {
			t.Fatalf("%s() still opens saves directly; use typed arkapi path helper", name)
		}
	}
}

func TestStructureAggregateCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"structureHealth", "structureOwnerCount", "structureOwners", "structureOwnerLocations"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arksave.Open") {
			t.Fatalf("%s() still opens saves directly; use typed arkapi path helper", name)
		}
	}
}

func TestBaseAggregateCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	body := functionBody(t, string(data), "baseComponents")
	if strings.Contains(body, "arksave.Open") {
		t.Fatalf("baseComponents() still opens saves directly; use typed arkapi path helper")
	}
}

func TestGeneralCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"parseSave", "mapSummary", "objectClasses", "objectSummary", "propertyPositions", "classLookup", "classPropertySummary", "propertyFilter"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arksave.Open") {
			t.Fatalf("%s() still opens saves directly; use typed arkapi path helper", name)
		}
		if strings.Contains(body, "NewGeneralFromPath") {
			t.Fatalf("%s() still owns GeneralAPI lifecycle; use typed arkapi path helper", name)
		}
	}
}

func TestJSONExportCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"exportJSON", "exportDomainJSON"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arksave.Open") || strings.Contains(body, "NewJSONFromPath") {
			t.Fatalf("%s() still owns JSON save lifecycle; use typed arkapi path helper", name)
		}
	}
}

func TestHeatmapCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"structureHeatmap", "dinoHeatmap"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arksave.Open") {
			t.Fatalf("%s() still opens saves directly; use typed arkapi path helper", name)
		}
	}
}

func TestPlayerTribeAggregateCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"playerRoster", "tribeRoster", "playerTribeLinks", "players", "tribes"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "NewPlayerFromPath") {
			t.Fatalf("%s() still owns player API lifecycle; use typed arkapi path helper", name)
		}
		if strings.Contains(body, "arkprofile.Open") {
			t.Fatalf("%s() still owns local profile/tribe file lifecycle; use typed arkapi path helper", name)
		}
	}
	body := functionBody(t, source, "tribesDirectory")
	if strings.Contains(body, "NewPlayerFromDirectory") {
		t.Fatalf("tribesDirectory() still owns player directory lifecycle; use typed arkapi path helper")
	}
	body = functionBody(t, source, "playersDirectory")
	if strings.Contains(body, "NewPlayerFromDirectory") {
		t.Fatalf("playersDirectory() still owns player directory lifecycle; use typed arkapi path helper")
	}
}

func TestClusterCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"cluster", "clusterSummary", "exportClusterJSON", "exportClusterDirectoryJSON"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arkcluster.Open") || strings.Contains(body, "arkcluster.OpenDirectoryWithFaults") {
			t.Fatalf("%s() still owns cluster file lifecycle; use typed arkapi path helper", name)
		}
	}
}

func TestTributeCommandsUseTypedPathHelpers(t *testing.T) {
	data, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("ReadFile(main.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"tribute", "exportTributeJSON", "exportTributeDirectoryJSON"} {
		body := functionBody(t, source, name)
		if strings.Contains(body, "arktribute.Open") || strings.Contains(body, "arktribute.OpenDirectoryWithFaults") {
			t.Fatalf("%s() still owns tribute file lifecycle; use typed arkapi path helper", name)
		}
	}
}

func functionBody(t *testing.T, source string, name string) string {
	t.Helper()
	start := strings.Index(source, "func "+name+"(")
	if start < 0 {
		t.Fatalf("function %s not found", name)
	}
	next := strings.Index(source[start+len("func "+name+"("):], "\nfunc ")
	if next < 0 {
		return source[start:]
	}
	return source[start : start+len("func "+name+"(")+next]
}

func TestParseCommandPrintsOfflineParseSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)

	var out bytes.Buffer
	err := run([]string{"parse", path}, &out)
	if err != nil {
		t.Fatalf("run(parse) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Save: " + path,
		"Map: Valguero_WP",
		"Save version: 12",
		"Objects: 1",
		"Parsed objects: 1",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("parse output %q does not contain %q", got, want)
		}
	}
}

func TestParseCommandRedactsPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)

	var out bytes.Buffer
	err := run([]string{"--redact", "parse", path}, &out)
	if err != nil {
		t.Fatalf("run(--redact parse) error = %v", err)
	}
	got := out.String()
	if strings.Contains(got, path) || !strings.Contains(got, "Save: [redacted]") {
		t.Fatalf("redacted parse output = %q", got)
	}
	if !strings.Contains(got, "Parsed objects: 1") || !strings.Contains(got, "Parse faults: 0") {
		t.Fatalf("redacted parse output missing aggregate counts: %q", got)
	}
}

func TestMapSummaryCommandPrintsOfflineSaveSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)

	var out bytes.Buffer
	err := run([]string{"map-summary", path}, &out)
	if err != nil {
		t.Fatalf("run(map-summary) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"map=Valguero_WP",
		"save_version=12",
		"objects=1",
		"names=",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("map-summary output %q does not contain %q", got, want)
		}
	}
}

func TestObjectClassesCommandPrintsSaveClasses(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)

	var out bytes.Buffer
	err := run([]string{"object-classes", path}, &out)
	if err != nil {
		t.Fatalf("run(object-classes) error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Blueprint'/Game/Test.Test_C'") {
		t.Fatalf("object-classes output %q missing synthetic class", got)
	}
}

func TestObjectSummaryCommandPrintsObjectSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)
	objectID := "00010203-0405-0607-0809-0a0b0c0d0e0f"

	var out bytes.Buffer
	err := run([]string{"object-summary", path, objectID}, &out)
	if err != nil {
		t.Fatalf("run(object-summary) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Exists: true",
		"Bytes:",
		"Properties: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("object-summary output %q does not contain %q", got, want)
		}
	}
}

func TestPropertyPositionsCommandPrintsPositionSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "synthetic.ark")
	createSyntheticSave(t, path)
	objectID := "00010203-0405-0607-0809-0a0b0c0d0e0f"

	var out bytes.Buffer
	err := run([]string{"property-positions", path, objectID}, &out)
	if err != nil {
		t.Fatalf("run(property-positions) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Exists: true",
		"Properties: 1",
		"Name offsets:",
		"Value offsets:",
		"Encoded: 1",
		"Positioned:",
		"Offsets OK:",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("property-positions output %q does not contain %q", got, want)
		}
	}
}

func TestClassLookupCommandPrintsAggregateSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"class-lookup", path, "Wall_Stone"}, &out)
	if err != nil {
		t.Fatalf("run(class-lookup) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Objects: 1",
		"Classes: 1",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("class-lookup output %q does not contain %q", got, want)
		}
	}
}

func TestClassPropertySummaryCommandPrintsAggregateSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"class-property-summary", path, "Wall_Stone"}, &out)
	if err != nil {
		t.Fatalf("run(class-property-summary) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Objects: 1",
		"Properties: 4",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("class-property-summary output %q does not contain %q", got, want)
		}
	}
}

func TestPropertyFilterCommandPrintsAggregateSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"property-filter", path, "Health", "MaxHealth"}, &out)
	if err != nil {
		t.Fatalf("run(property-filter) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Objects: 1",
		"Classes: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("property-filter output %q does not contain %q", got, want)
		}
	}
}

func TestStructureHealthCommandPrintsAggregateSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"structure-health", path}, &out)
	if err != nil {
		t.Fatalf("run(structure-health) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Structures: 1",
		"With health: 1",
		"Damaged: 1",
		"Fully repaired: 0",
		"Without max health: 0",
		"Average health: 90.0%",
		"Minimum health: 90.0%",
		"Maximum health: 90.0%",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("structure-health output %q does not contain %q", got, want)
		}
	}
}

func TestStructureOwnerCountCommandPrintsAggregateSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"structure-owner-count", path, "555"}, &out)
	if err != nil {
		t.Fatalf("run(structure-owner-count) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe ID: 555",
		"Structures: 1",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("structure-owner-count output %q does not contain %q", got, want)
		}
	}
}

func TestStructureOwnerCountCommandRedactsTribeID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"--redact", "structure-owner-count", path, "555"}, &out)
	if err != nil {
		t.Fatalf("run(--redact structure-owner-count) error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Tribe ID: [redacted]") || strings.Contains(got, "555") {
		t.Fatalf("redacted structure-owner-count output = %q", got)
	}
	if !strings.Contains(got, "Structures: 1") || !strings.Contains(got, "Parse faults: 0") {
		t.Fatalf("redacted structure-owner-count output missing aggregate counts: %q", got)
	}
}

func TestStructureOwnerCountCommandRejectsInvalidTribeID(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"structure-owner-count", path, "not-an-id"}, &out)
	if err == nil {
		t.Fatalf("run(structure-owner-count invalid id) error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "parse tribe id") {
		t.Fatalf("run(structure-owner-count invalid id) error = %v", err)
	}
}

func TestStructureOwnersCommandPrintsAggregateSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"structure-owners", path}, &out)
	if err != nil {
		t.Fatalf("run(structure-owners) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Structures: 1",
		"With tribe ID: 1",
		"With player ID: 0",
		"With tribe name: 0",
		"Unique tribes: 1",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("structure-owners output %q does not contain %q", got, want)
		}
	}
}

func TestStructureOwnerLocationsCommandPrintsAggregateSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"structure-owner-locations", path, "Valguero", "1"}, &out)
	if err != nil {
		t.Fatalf("run(structure-owner-locations) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Structures: 1",
		"Owners: 1",
		"Cells: 1",
		"Named cells: 1",
		"Skipped without owner: 0",
		"Skipped without location: 0",
		"Parse faults: 0",
		`"owner": "555"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("structure-owner-locations output %q does not contain %q", got, want)
		}
	}
}

func TestStructureOwnerLocationsCommandRedactsOwnerBuckets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"--redact", "structure-owner-locations", path, "Valguero", "1"}, &out)
	if err != nil {
		t.Fatalf("run(--redact structure-owner-locations) error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `"owner": "[redacted]"`) || strings.Contains(got, `"owner": "555"`) {
		t.Fatalf("redacted structure-owner-locations output = %q", got)
	}
	if !strings.Contains(got, "Structures: 1") || !strings.Contains(got, "Parse faults: 0") {
		t.Fatalf("redacted structure-owner-locations output missing aggregate counts: %q", got)
	}
}

func TestStructureHeatmapCommandWritesSummaryJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "structures.ark")
	outPath := filepath.Join(dir, "structure-heatmap.json")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"structure-heatmap", path, outPath, "100", "1"}, &out)
	if err != nil {
		t.Fatalf("run(structure-heatmap) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cells: 1",
		"Total: 1",
		"Max: 1",
		"Parse faults: 0",
		"Wrote: " + outPath,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("structure-heatmap output %q does not contain %q", got, want)
		}
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outPath, err)
	}
	var summary arkapi.HeatmapSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatalf("Unmarshal heatmap summary error = %v", err)
	}
	if summary.Resolution != 100 || summary.NonzeroCells != 1 || summary.Total != 1 || summary.Max != 1 || summary.Faults != 0 {
		t.Fatalf("heatmap summary = %#v, want one populated structure cell", summary)
	}
}

func TestDinoHeatmapCommandWritesSummaryJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dinos.ark")
	outPath := filepath.Join(dir, "dino-heatmap.json")
	createSyntheticLocatedDinoSave(t, path)

	var out bytes.Buffer
	err := run([]string{"--no-cryos", "dino-heatmap", path, outPath, "100"}, &out)
	if err != nil {
		t.Fatalf("run(dino-heatmap) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cells: 1",
		"Total: 1",
		"Max: 1",
		"Parse faults: 0",
		"Wrote: " + outPath,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dino-heatmap output %q does not contain %q", got, want)
		}
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outPath, err)
	}
	var summary arkapi.HeatmapSummary
	if err := json.Unmarshal(data, &summary); err != nil {
		t.Fatalf("Unmarshal dino heatmap summary error = %v", err)
	}
	if summary.Resolution != 100 || summary.NonzeroCells != 1 || summary.Total != 1 || summary.Max != 1 || summary.Faults != 0 {
		t.Fatalf("dino heatmap summary = %#v, want one populated dino cell", summary)
	}
}

func TestBaseComponentsCommandPrintsComponentStats(t *testing.T) {
	path := filepath.Join(t.TempDir(), "structures.ark")
	createSyntheticStructureHealthSave(t, path)

	var out bytes.Buffer
	err := run([]string{"base-components", path}, &out)
	if err != nil {
		t.Fatalf("run(base-components) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Components: 1",
		"Total structures: 1",
		"Largest component: 1",
		"Components at least 10: 0",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("base-components output %q does not contain %q", got, want)
		}
	}
}

func TestDinosCommandPrintsPopulationSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dinos.ark")
	createSyntheticDinoSave(t, path)

	var out bytes.Buffer
	err := run([]string{"dinos", path}, &out)
	if err != nil {
		t.Fatalf("run(dinos) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Dinos: 1",
		"Tamed: 0",
		"Wild: 1",
		"Cryopodded: 0",
		"Classes: 1",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dinos output %q does not contain %q", got, want)
		}
	}
}

func TestDinoWildTamablesCommandPrintsSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dinos.ark")
	createSyntheticDinoSave(t, path)

	var out bytes.Buffer
	err := run([]string{"dino-wild-tamables", path}, &out)
	if err != nil {
		t.Fatalf("run(dino-wild-tamables) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Wild dinos: 1",
		"Wild tamables: 1",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dino-wild-tamables output %q does not contain %q", got, want)
		}
	}
}

func TestDinoBabiesCommandPrintsSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "baby-dinos.ark")
	createSyntheticBabyDinoSave(t, path)

	var out bytes.Buffer
	err := run([]string{"dino-babies", path}, &out)
	if err != nil {
		t.Fatalf("run(dino-babies) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Baby dinos: 2",
		"Tamed babies: 1",
		"Wild babies: 1",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dino-babies output %q does not contain %q", got, want)
		}
	}
}

func TestDinoBestStatCommandPrintsBestStat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dino-stats.ark")
	createSyntheticDinoStatsSave(t, path)

	var out bytes.Buffer
	err := run([]string{"dino-best-stat", path}, &out)
	if err != nil {
		t.Fatalf("run(dino-best-stat) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Best stat: health",
		"Points: 6",
		"Level: 12",
		"Blueprint: Raptor",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dino-best-stat output %q does not contain %q", got, want)
		}
	}
}

func TestDinoBestBaseStatCommandPrintsFilteredBaseStat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dino-base-stats.ark")
	createSyntheticTamedDinoStatsSave(t, path)

	var out bytes.Buffer
	err := run([]string{"dino-best-base-stat", path, "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'", "health"}, &out)
	if err != nil {
		t.Fatalf("run(dino-best-base-stat) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Best base stat: health",
		"Points: 5",
		"Level: 12",
		"Blueprint: Raptor",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dino-best-base-stat output %q does not contain %q", got, want)
		}
	}
}

func TestDinoMostMutatedCommandPrintsMostMutatedTamedDino(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dino-mutated.ark")
	createSyntheticTamedDinoStatsSave(t, path)

	var out bytes.Buffer
	err := run([]string{"dino-most-mutated", path}, &out)
	if err != nil {
		t.Fatalf("run(dino-most-mutated) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Most mutated: Raptor",
		"Total mutation points: 1",
		"Mutation pairs: 0",
		"Level: 12",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dino-most-mutated output %q does not contain %q", got, want)
		}
	}
}

func TestDinoWildTamedCommandPrintsSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dino-wild-tamed.ark")
	createSyntheticTamedDinoStatsSave(t, path)

	var out bytes.Buffer
	err := run([]string{"dino-wild-tamed", path}, &out)
	if err != nil {
		t.Fatalf("run(dino-wild-tamed) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Wild-tamed dinos: 1",
		"Max level: 12",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("dino-wild-tamed output %q does not contain %q", got, want)
		}
	}
}

func TestEquipmentSaddlesCommandPrintsSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "saddles.ark")
	createSyntheticSaddleEquipmentSave(t, path)

	var out bytes.Buffer
	err := run([]string{"equipment-saddles", path}, &out)
	if err != nil {
		t.Fatalf("run(equipment-saddles) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Item saddles: 1",
		"Cryopod saddles: 0",
		"Total saddles: 1",
		"Max armor: 23.2",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("equipment-saddles output %q does not contain %q", got, want)
		}
	}
}

func TestStackablesCommandPrintsSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stackables.ark")
	createSyntheticStackableSave(t, path)

	var out bytes.Buffer
	err := run([]string{"stackables", path}, &out)
	if err != nil {
		t.Fatalf("run(stackables) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Stackable items: 1",
		"Total quantity: 250",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stackables output %q does not contain %q", got, want)
		}
	}
}

func TestStackableOwnedByCommandPrintsOwnedSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "stackables.ark")
	createSyntheticOwnedStackableSave(t, path)
	blueprint := "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'"

	var out bytes.Buffer
	err := run([]string{"stackable-owned-by", path, blueprint, "555"}, &out)
	if err != nil {
		t.Fatalf("run(stackable-owned-by) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe ID: 555",
		"Blueprint: " + blueprint,
		"Items: 1",
		"Total quantity: 100",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("stackable-owned-by output %q does not contain %q", got, want)
		}
	}
}

func TestEquipmentOwnedByCommandPrintsOwnedSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "equipment-owned.ark")
	createSyntheticOwnedEquipmentSave(t, path)
	blueprint := "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'"

	var out bytes.Buffer
	err := run([]string{"equipment-owned-by", path, blueprint, "555"}, &out)
	if err != nil {
		t.Fatalf("run(equipment-owned-by) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe ID: 555",
		"Blueprint: " + blueprint,
		"Items: 1",
		"Max damage: 112.3",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("equipment-owned-by output %q does not contain %q", got, want)
		}
	}
}

func TestEquipmentBestCommandPrintsBestItems(t *testing.T) {
	path := filepath.Join(t.TempDir(), "equipment-best.ark")
	createSyntheticBestEquipmentSave(t, path)

	var out bytes.Buffer
	err := run([]string{"equipment-best", path}, &out)
	if err != nil {
		t.Fatalf("run(equipment-best) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Best weapon damage: 112.3",
		"Best weapon crafted: false",
		"Best armor durability: 31.2",
		"Best armor crafted: false",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("equipment-best output %q does not contain %q", got, want)
		}
	}
}

func TestEquipmentRankCommandPrintsRankStats(t *testing.T) {
	path := filepath.Join(t.TempDir(), "equipment-rank.ark")
	createSyntheticRankEquipmentSave(t, path)

	var out bytes.Buffer
	err := run([]string{"equipment-rank", path}, &out)
	if err != nil {
		t.Fatalf("run(equipment-rank) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Ranked: 2",
		"Best rating: 5.5",
		"Best average stat: 700.0",
		"Crafted: 0",
		"Blueprints: 1",
		"Classes: 2",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("equipment-rank output %q does not contain %q", got, want)
		}
	}
}

func TestEquipmentAscendantWeaponBPsCommandPrintsSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "equipment-ascendant.ark")
	createSyntheticAscendantWeaponBlueprintSave(t, path)

	var out bytes.Buffer
	err := run([]string{"equipment-ascendant-weapon-bps", path}, &out)
	if err != nil {
		t.Fatalf("run(equipment-ascendant-weapon-bps) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Items: 1",
		"Max damage: 112.3",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("equipment-ascendant-weapon-bps output %q does not contain %q", got, want)
		}
	}
}

func TestEquipmentHistoryCommandWritesReport(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "equipment.ark")
	manifestPath := filepath.Join(dir, "equipment-history-files.json")
	outPath := filepath.Join(dir, "equipment-history-report.json")
	createSyntheticEquipmentSave(t, savePath)

	manifest, err := json.Marshal([]string{savePath, savePath})
	if err != nil {
		t.Fatalf("Marshal(manifest) error = %v", err)
	}
	if err := os.WriteFile(manifestPath, append(manifest, '\n'), 0o600); err != nil {
		t.Fatalf("WriteFile(manifest) error = %v", err)
	}

	var out bytes.Buffer
	err = run([]string{"equipment-history", manifestPath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(equipment-history) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"saves=2",
		"initial=1",
		"changes=0",
		"final=1",
		"wrote=" + outPath,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("equipment-history output %q does not contain %q", got, want)
		}
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(report) error = %v", err)
	}
	var report arkapi.EquipmentHistoryReport
	if err := json.Unmarshal(raw, &report); err != nil {
		t.Fatalf("Unmarshal(report) error = %v; data = %s", err, raw)
	}
	if report.Saves != 2 || report.InitialCount != 1 || report.FinalCount != 1 || len(report.Changes) != 0 {
		t.Fatalf("equipment history report = %#v, want stable two-save report", report)
	}
	assertPrivateFileMode(t, outPath)
}

func TestEquipmentSummaryCommandPrintsSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "equipment.ark")
	createSyntheticEquipmentSave(t, path)

	var out bytes.Buffer
	err := run([]string{"equipment-summary", path}, &out)
	if err != nil {
		t.Fatalf("run(equipment-summary) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Items: 1",
		"Total quantity: 2",
		"Weapon items: 1",
		"Blueprints: 1",
		"Equipped: 1",
		"Crafted: 1",
		"Classes: 1",
		"Max quality: 3",
		"Max rating: 7.5",
		"Max damage: 112.3",
		"Max durability: 0.8",
		"Parse faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("equipment-summary output %q does not contain %q", got, want)
		}
	}
}

func TestPlayerInventoriesCommandPrintsInventorySummary(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "inventory.ark")
	createSyntheticPlayerInventorySave(t, path)

	var out bytes.Buffer
	err := run([]string{"player-inventories", path}, &out)
	if err != nil {
		t.Fatalf("run(player-inventories) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Players: 1",
		"With inventory: 1",
		"Without inventory: 0",
		"Total items: 2",
		"Max items: 2",
		"Min items: 2",
		"Average items: 2.00",
		"Inventory faults: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("player-inventories output %q does not contain %q", got, want)
		}
	}
}

func TestPlayerRosterCommandPrintsRosterSummary(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		ExtraCharacterLevel: 9,
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        43,
		CharacterName:       "",
		PlayerName:          "",
		ExtraCharacterLevel: 4,
	})

	var out bytes.Buffer
	err := run([]string{"player-roster", dir}, &out)
	if err != nil {
		t.Fatalf("run(player-roster) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Players: 2",
		"With names: 1",
		"Highest level: 10",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("player-roster output %q does not contain %q", got, want)
		}
	}
}

func TestTribeRosterCommandPrintsRosterSummary(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout"},
		MemberIDs: []int32{42, 43},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "789.arktribe"), testfixtures.TribeArchiveOptions{
		Name:     "",
		TribeID:  222,
		OwnerID:  43,
		NumDinos: 3,
		Members:  []string{"Builder"},
	})

	var out bytes.Buffer
	err := run([]string{"tribe-roster", dir}, &out)
	if err != nil {
		t.Fatalf("run(tribe-roster) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribes: 2",
		"With names: 1",
		"Members: 3",
		"Dinos: 10",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribe-roster output %q does not contain %q", got, want)
		}
	}
}

func TestPlayerTribeLinksCommandPrintsRelationSummary(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerTribeRelationDirectory(t, dir)

	var out bytes.Buffer
	err := run([]string{"player-tribe-links", dir}, &out)
	if err != nil {
		t.Fatalf("run(player-tribe-links) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Players: 3",
		"Tribes: 2",
		"Active links: 2",
		"Inactive members: 2",
		"Players without tribe: 1",
		"Tribes with inactive: 2",
		"Tribes without active: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("player-tribe-links output %q does not contain %q", got, want)
		}
	}
}

func TestRunRejectsNetworkCommands(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"rcon"}, &out)
	if err == nil {
		t.Fatalf("run(rcon) error = nil, want unsupported command")
	}
	if !strings.Contains(err.Error(), "offline-only") {
		t.Fatalf("run(rcon) error = %v, want offline-only message", err)
	}
}

func TestRunPrintsUsageForHelp(t *testing.T) {
	for _, args := range [][]string{
		{"help"},
		{"--help"},
		{"-h"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			var out bytes.Buffer
			err := run(args, &out)
			if err != nil {
				t.Fatalf("run(%v) error = %v", args, err)
			}
			got := out.String()
			for _, want := range []string{
				"Usage:",
				"arksave --help",
				"Offline-only scope: FTP and RCON are intentionally unsupported.",
			} {
				if !strings.Contains(got, want) {
					t.Fatalf("help output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestRunPrintsVersion(t *testing.T) {
	oldVersion, oldCommit, oldBuiltAt := version, commit, builtAt
	version, commit, builtAt = "test-version", "test-commit", "test-date"
	defer func() {
		version, commit, builtAt = oldVersion, oldCommit, oldBuiltAt
	}()

	for _, args := range [][]string{
		{"version"},
		{"--version"},
		{"-V"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			var out bytes.Buffer
			err := run(args, &out)
			if err != nil {
				t.Fatalf("run(%v) error = %v", args, err)
			}
			got := out.String()
			for _, want := range []string{
				"arksave version=test-version",
				"commit=test-commit",
				"built_at=test-date",
			} {
				if !strings.Contains(got, want) {
					t.Fatalf("version output %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestRunRejectsUnknownOption(t *testing.T) {
	var out bytes.Buffer
	err := run([]string{"--verbose", "inspect", "save.ark"}, &out)
	if err == nil {
		t.Fatalf("run(unknown option) error = nil, want unknown option")
	}
	if !strings.Contains(err.Error(), "unknown option") {
		t.Fatalf("run(unknown option) error = %v, want unknown option message", err)
	}
}

func TestArchiveSummaryPrintsPropertyErrorCount(t *testing.T) {
	var out bytes.Buffer
	err := printArchiveSummary(&out, "Archive", arkapi.LocalArchiveSummary{
		Path:                "/tmp/profile.arkprofile",
		ArchiveVersion:      7,
		ObjectCount:         2,
		PropertyParseErrors: 1,
		ClassNames:          []string{"/Game/Valid.Valid_C", "/Game/Broken.Broken_C"},
	}, runOptions{})
	if err != nil {
		t.Fatalf("printArchiveSummary() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Property parse errors: 1") {
		t.Fatalf("archive summary %q does not contain property error count", got)
	}
}

func TestPlayersCommandReturnsErrorWhenParsedProfileSummaryMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WriteArchive(t, path, "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C")

	var out bytes.Buffer
	err := run([]string{"players", path}, &out)
	if err == nil {
		t.Fatalf("run(players) error = nil, want missing normalized profile error")
	}
	if !strings.Contains(err.Error(), "parse player profile details") {
		t.Fatalf("run(players) error = %v, want normalized parse context", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player profile:",
		"Archive version: 7",
		"Objects: 1",
		"/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("players output %q does not contain %q", got, want)
		}
	}
}

func TestPlayersCommandPrintsParsedLocalProfileSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WritePlayerArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"players", path}, &out)
	if err != nil {
		t.Fatalf("run(players) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player profile:",
		"Character name: Survivor",
		"Player name: PlatformName",
		"Player data ID: 42",
		"Tribe ID: 777",
		"Deaths: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("players output %q does not contain %q", got, want)
		}
	}
}

func TestPlayersCommandRedactsLocalProfileDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "123.arkprofile")
	testfixtures.WritePlayerArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"players", "--redact", path}, &out)
	if err != nil {
		t.Fatalf("run(players --redact) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player profile: [redacted]",
		"Character name: [redacted]",
		"Player name: [redacted]",
		"Player data ID: [redacted]",
		"Tribe ID: [redacted]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted players output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "Survivor", "PlatformName", "/Game/"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted players output %q contains private detail %q", got, leaked)
		}
	}
}

func TestPlayersCommandPrintsDirectorySummary(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		TribeID:             777,
		NumDeaths:           3,
		ExtraCharacterLevel: 9,
		ExperiencePoints:    12.5,
		TotalEngramPoints:   14,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "456.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        43,
		CharacterName:       "Scout",
		PlayerName:          "OtherPlatform",
		TribeID:             888,
		NumDeaths:           1,
		ExtraCharacterLevel: 4,
		ExperiencePoints:    7.5,
		TotalEngramPoints:   6,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
			"Blueprint'/Game/Engrams/EngramC.EngramC_C'",
		},
	})

	var out bytes.Buffer
	err := run([]string{"players", dir}, &out)
	if err != nil {
		t.Fatalf("run(players directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player directory:",
		"Profiles: 2",
		"Players: 2",
		"Total deaths: 4",
		"Average deaths: 2.00",
		"Total level: 15",
		"Average level: 7.50",
		"Highest level: 10",
		"Total experience: 20.00",
		"Average experience: 10.00",
		"Highest experience: 12.50",
		"Total engram points: 20",
		"Unlocked engrams: 3",
		"  player id=42 character=Survivor platform=PlatformName tribe=777 level=10 deaths=3",
		"  player id=43 character=Scout platform=OtherPlatform tribe=888 level=5 deaths=1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("players directory output %q does not contain %q", got, want)
		}
	}
}

func TestPlayersCommandRedactsDirectoryDetails(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(dir, "123.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:        42,
		CharacterName:       "Survivor",
		PlayerName:          "PlatformName",
		TribeID:             777,
		NumDeaths:           3,
		ExtraCharacterLevel: 9,
		ExperiencePoints:    12.5,
		TotalEngramPoints:   14,
		UnlockedEngrams: []string{
			"Blueprint'/Game/Engrams/EngramB.EngramB_C'",
			"Blueprint'/Game/Engrams/EngramA.EngramA_C'",
		},
	})

	var out bytes.Buffer
	err := run([]string{"players", "--redact", dir}, &out)
	if err != nil {
		t.Fatalf("run(players --redact directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Player directory: [redacted]",
		"Profiles: 1",
		"Players: 1",
		"Total deaths: 3",
		"Unlocked engrams: 2",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted players directory output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{dir, "Survivor", "PlatformName", "/Game/", "player id=42"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted players directory output %q contains private detail %q", got, leaked)
		}
	}
}

func TestTribesCommandPrintsLocalTribeSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteTribeArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"tribes", path}, &out)
	if err != nil {
		t.Fatalf("run(tribes) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save:",
		"Archive version: 7",
		"Objects: 1",
		"/Script/ShooterGame.PrimalTribeData",
		"Tribe name: Porters",
		"Tribe ID: 12345",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes output %q does not contain %q", got, want)
		}
	}
}

func TestTribesCommandRedactsLocalTribeDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteTribeArchive(t, path)

	var out bytes.Buffer
	err := run([]string{"--redact", "tribes", path}, &out)
	if err != nil {
		t.Fatalf("run(--redact tribes) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save: [redacted]",
		"Tribe name: [redacted]",
		"Tribe ID: [redacted]",
		"Owner ID: [redacted]",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted tribes output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "Porters", "/Script/ShooterGame.PrimalTribeData"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted tribes output %q contains private detail %q", got, leaked)
		}
	}
}

func TestTribesCommandPrintsDirectorySummary(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "456.arktribe"), testfixtures.TribeArchiveOptions{
		Name:      "Porters",
		TribeID:   12345,
		OwnerID:   42,
		NumDinos:  7,
		Members:   []string{"Survivor", "Scout"},
		MemberIDs: []int32{42, 43},
	})
	testfixtures.WriteTribeArchiveWithOptions(t, filepath.Join(dir, "789.arktribe"), testfixtures.TribeArchiveOptions{
		Name:     "Builders",
		TribeID:  222,
		OwnerID:  43,
		NumDinos: 3,
		Members:  []string{"Builder"},
	})

	var out bytes.Buffer
	err := run([]string{"tribes", dir}, &out)
	if err != nil {
		t.Fatalf("run(tribes directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe directory:",
		"Tribe files: 2",
		"Tribes: 2",
		"Total members: 3",
		"Average members: 1.50",
		"Total dinos: 10",
		"Average dinos: 5.00",
		"  tribe id=222 name=Builders owner=43 members=1 dinos=3",
		"  tribe id=12345 name=Porters owner=42 members=2 dinos=7",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes directory output %q does not contain %q", got, want)
		}
	}
}

func TestTribesCommandReturnsErrorWhenParsedSummaryMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "456.arktribe")
	testfixtures.WriteArchive(t, path, "/Script/ShooterGame.PrimalTribeData")

	var out bytes.Buffer
	err := run([]string{"tribes", path}, &out)
	if err == nil {
		t.Fatalf("run(tribes) error = nil, want missing normalized tribe error")
	}
	if !strings.Contains(err.Error(), "parse tribe details") {
		t.Fatalf("run(tribes) error = %v, want normalized parse context", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribe save:",
		"Archive version: 7",
		"Objects: 1",
		"/Script/ShooterGame.PrimalTribeData",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribes output %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "Tribe name:") {
		t.Fatalf("tribes output %q includes summary despite missing TribeData", got)
	}
}

func TestClusterCommandPrintsLocalClusterSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	testfixtures.WriteArchive(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"cluster", path}, &out)
	if err != nil {
		t.Fatalf("run(cluster) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster file:",
		"Archive version: 7",
		"Objects: 1",
		"Items: 0",
		"Dinos: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster output %q does not contain %q", got, want)
		}
	}
}

func TestClusterCommandPrintsDirectoryAggregateSummary(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_abc123"), "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_def456"), "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"cluster", dir}, &out)
	if err != nil {
		t.Fatalf("run(cluster directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster directory:",
		"Files: 2",
		"Objects: 2",
		"Items: 0",
		"Dinos: 0",
		"Parse errors: 0",
		"Cluster file:",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster directory output %q does not contain %q", got, want)
		}
	}
}

func TestClusterCommandDirectoryKeepsValidFilesAndReportsFaults(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_abc123"), "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(filepath.Join(dir, "EOS_broken"), []byte("not a cluster archive"), 0o600); err != nil {
		t.Fatalf("write broken cluster file: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"cluster", dir}, &out)
	if err != nil {
		t.Fatalf("run(cluster directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster directory:",
		"Files: 1",
		"File faults: 1",
		"Objects: 1",
		"Cluster file:",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster directory output %q does not contain %q", got, want)
		}
	}
}

func TestClusterSummaryCommandPrintsTypedAggregate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	testfixtures.WriteArchive(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"cluster-summary", path}, &out)
	if err != nil {
		t.Fatalf("run(cluster-summary) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster file:",
		"Archive version: 7",
		"Objects: 1",
		"Items: 0",
		"Dinos: 0",
		"Parse errors: 0",
		"Dino item uploads: 0",
		"Equipment item uploads: 0",
		"Other item uploads: 0",
		"Parsed dinos: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster-summary output %q does not contain %q", got, want)
		}
	}
}

func TestPrintClusterTypedSummariesIncludesEmbeddedDinoAggregates(t *testing.T) {
	var out bytes.Buffer
	err := printClusterTypedSummaries(&out, arkapi.ClusterItemSummary{}, arkapi.ClusterDinoSummary{
		WithDinoID:          2,
		TamedDinos:          1,
		FemaleDinos:         1,
		BabyDinos:           1,
		DeadDinos:           1,
		WithStats:           1,
		TotalBaseLevel:      12,
		AverageBaseLevel:    12,
		MaxBaseLevel:        12,
		TotalCurrentLevel:   6,
		AverageCurrentLevel: 6,
		MaxCurrentLevel:     6,
	})
	if err != nil {
		t.Fatalf("printClusterTypedSummaries() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Dinos with IDs: 2",
		"Tamed dinos: 1",
		"Female dinos: 1",
		"Baby dinos: 1",
		"Dead dinos: 1",
		"Dinos with stats: 1",
		"Total base level: 12",
		"Average base level: 12.00",
		"Max base level: 12",
		"Total current level: 6",
		"Average current level: 6.00",
		"Max current level: 6",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster typed summary %q does not contain %q", got, want)
		}
	}
}

func TestClusterInfoDinoSummaryIncludesEmbeddedDinoAggregates(t *testing.T) {
	summary := clusterInfoDinoSummary(arkapi.ClusterDataInfo{Dinos: []arkapi.ClusterDinoInfo{
		{DinoID1: 1001, DinoID2: 2002, IsTamed: true, IsFemale: true, HasStats: true, BaseLevel: 12, CurrentLevel: 6},
		{DinoID1: 3003, DinoID2: 4004, IsBaby: true, IsDead: true, HasStats: true, BaseLevel: 8, CurrentLevel: 4},
		{},
	}})

	if summary.WithDinoID != 2 || summary.TamedDinos != 1 || summary.FemaleDinos != 1 || summary.BabyDinos != 1 || summary.DeadDinos != 1 || summary.WithStats != 2 {
		t.Fatalf("clusterInfoDinoSummary() identity/stat counts = %#v, want aggregate counts from exported cluster info", summary)
	}
	if summary.TotalBaseLevel != 20 || summary.MaxBaseLevel != 12 || summary.AverageBaseLevel != 10 || summary.TotalCurrentLevel != 10 || summary.MaxCurrentLevel != 6 || summary.AverageCurrentLevel != 5 {
		t.Fatalf("clusterInfoDinoSummary() level aggregates = %#v, want aggregate levels from exported cluster info", summary)
	}
}

func TestClusterSummaryCommandPrintsDirectoryTypedAggregate(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_abc123"), "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_def456"), "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"cluster-summary", dir}, &out)
	if err != nil {
		t.Fatalf("run(cluster-summary directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster directory:",
		"Files: 2",
		"Objects: 2",
		"Items: 0",
		"Dinos: 0",
		"Parse errors: 0",
		"Dino item uploads: 0",
		"Parsed dinos: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster-summary directory output %q does not contain %q", got, want)
		}
	}
}

func TestClusterSummaryCommandDirectoryKeepsValidFilesAndReportsFaults(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_abc123"), "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(filepath.Join(dir, "EOS_broken"), []byte("not a cluster archive"), 0o600); err != nil {
		t.Fatalf("write broken cluster file: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"cluster-summary", dir}, &out)
	if err != nil {
		t.Fatalf("run(cluster-summary directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster directory:",
		"Files: 1",
		"File faults: 1",
		"Objects: 1",
		"Dino item uploads: 0",
		"Parsed dinos: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster-summary directory output %q does not contain %q", got, want)
		}
	}
}

func TestClusterSummaryPrintsDinoParseErrors(t *testing.T) {
	var out bytes.Buffer
	err := printClusterSummary(&out, arkapi.ExportClusterData(&arkcluster.Data{
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Dinos: []arkcluster.Dino{{
			Index:      0,
			UploadTime: 12345,
			RawSize:    32,
			ParseError: "unsupported embedded dino archive",
		}},
	}), runOptions{})
	if err != nil {
		t.Fatalf("printClusterSummary() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "parse_error=unsupported embedded dino archive") {
		t.Fatalf("cluster summary %q does not contain dino parse error", got)
	}
}

func TestClusterSummaryPrintsDinoClassNames(t *testing.T) {
	var out bytes.Buffer
	err := printClusterSummary(&out, arkapi.ExportClusterData(&arkcluster.Data{
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Dinos: []arkcluster.Dino{{
			Index:      0,
			UploadTime: 12345,
			RawSize:    64,
			Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
				{ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C"},
				{ClassName: "/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C"},
			}},
		}},
	}), runOptions{})
	if err != nil {
		t.Fatalf("printClusterSummary() error = %v", err)
	}
	got := out.String()
	want := "dino[0] raw_bytes=64 objects=2 upload=12345 primary_class=/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C short=Raptor class_names=/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C,/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C"
	if !strings.Contains(got, want) {
		t.Fatalf("cluster summary %q does not contain %q", got, want)
	}
	if !strings.Contains(got, "primary_class=/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C short=Raptor") {
		t.Fatalf("cluster summary %q does not contain primary class and short dino name", got)
	}
}

func TestClusterSummaryPrintsEmbeddedDinoIdentityAndStats(t *testing.T) {
	dinoID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	statusID := uuid.MustParse("11112222-3333-4444-5555-666677778888")
	var out bytes.Buffer
	err := printClusterSummary(&out, arkapi.ExportClusterData(&arkcluster.Data{
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Dinos: []arkcluster.Dino{{
			Index:      0,
			Version:    7,
			UploadTime: 12345,
			RawSize:    64,
			Archive: &arkarchive.Archive{Objects: []arkarchive.Object{
				{
					UUID:      dinoID,
					ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
					Properties: []arkproperty.Property{
						{Name: "DinoID1", Type: arkproperty.TypeUInt32, Value: uint32(1001)},
						{Name: "DinoID2", Type: arkproperty.TypeUInt32, Value: uint32(2002)},
						{Name: "TamedTimeStamp", Type: arkproperty.TypeDouble, Value: float64(42)},
						{Name: "TamedName", Type: arkproperty.TypeString, Value: "Needle"},
						{Name: "bIsFemale", Type: arkproperty.TypeBool, Value: true},
						{Name: "MyCharacterStatusComponent", Type: arkproperty.TypeObject, Value: arkproperty.ObjectReference{Type: arkproperty.ObjectReferenceUUID, Value: statusID}},
					},
				},
				{
					UUID:      statusID,
					ClassName: "/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatus_BP.DinoCharacterStatus_BP_C",
					Properties: []arkproperty.Property{
						{Name: "BaseCharacterLevel", Type: arkproperty.TypeInt, Value: int32(12)},
						{Name: "NumberOfLevelUpPointsApplied", Type: arkproperty.TypeInt, Position: 0, Value: int32(5)},
						{Name: "NumberOfLevelUpPointsApplied", Type: arkproperty.TypeInt, Position: 8, Value: int32(2)},
					},
				},
			}},
		}},
	}), runOptions{})
	if err != nil {
		t.Fatalf("printClusterSummary() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"dino_id=1001/2002",
		"tamed_name=Needle",
		"tamed=true",
		"female=true",
		"base_level=12 current_level=8",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster summary %q does not contain %q", got, want)
		}
	}
}

func TestClusterSummaryPrintsDinoParseStatusCounts(t *testing.T) {
	var out bytes.Buffer
	err := printClusterSummary(&out, arkapi.ExportClusterData(&arkcluster.Data{
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Dinos: []arkcluster.Dino{
			{
				Index:   0,
				Version: 7,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{{
					ClassName: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
				}}},
			},
			{
				Index:   1,
				Version: 6,
				Archive: &arkarchive.Archive{Objects: []arkarchive.Object{{
					ClassName: "/Game/Test/UnsupportedVersion.UnsupportedVersion_C",
				}}},
			},
			{Index: 2, Version: 7, ParseError: "unsupported embedded archive"},
			{Index: 3, Version: 7},
		},
	}), runOptions{})
	if err != nil {
		t.Fatalf("printClusterSummary() error = %v", err)
	}
	got := out.String()
	want := "Dino parse statuses: parsed=1 unsupported_version=1 parse_error=1 unparsed=1"
	if !strings.Contains(got, want) {
		t.Fatalf("cluster summary %q does not contain %q", got, want)
	}
}

func TestClusterSummaryPrintsItemTypes(t *testing.T) {
	var out bytes.Buffer
	err := printClusterSummary(&out, arkapi.ExportClusterData(&arkcluster.Data{
		Path:    "/tmp/EOS_abc123",
		Archive: &arkarchive.Archive{Version: 7, Objects: []arkarchive.Object{{ClassName: "/Script/ShooterGame.ArkCloudInventoryData"}}},
		Items: []arkcluster.Item{
			{
				Index:      0,
				Blueprint:  "/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C",
				Quantity:   1,
				UploadTime: 12345,
			},
			{
				Index:     1,
				Blueprint: "/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C",
				Properties: arkproperty.Container{Properties: []arkproperty.Property{{
					Name:  "CustomItemDatas",
					Type:  arkproperty.TypeArray,
					Value: arkproperty.Array{Values: []any{arkproperty.Container{}}},
				}}},
				UploadTime: 67890,
			},
		},
	}), runOptions{})
	if err != nil {
		t.Fatalf("printClusterSummary() error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"item[0] type=equipment short=WeaponBow blueprint=/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C quantity=1 upload=12345",
		"item[1] type=dino short=Raptor blueprint=/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C quantity=0 upload=67890",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("cluster summary %q does not contain %q", got, want)
		}
	}
}

func TestClusterCommandRedactsPathAndUploadDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "EOS_abc123")
	testfixtures.WriteArchive(t, path, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"--redact", "cluster", path}, &out)
	if err != nil {
		t.Fatalf("run(--redact cluster) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Cluster file: [redacted]",
		"Archive version: 7",
		"Objects: 1",
		"Items: 0",
		"Dinos: 0",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted cluster output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "EOS_abc123", "blueprint="} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted cluster output %q contains private detail %q", got, leaked)
		}
	}
}

func TestTributeCommandPrintsLocalTributeSummary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.arktributetribe")
	testfixtures.WriteTributeFile(t, path, []uint64{11, 22}, []uint64{33})

	var out bytes.Buffer
	err := run([]string{"tribute", path}, &out)
	if err != nil {
		t.Fatalf("run(tribute) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribute file:",
		"Player data IDs: 2",
		"Tribe data IDs: 1",
		"player_data_id=11",
		"player_data_id=22",
		"tribe_data_id=33",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribute output %q does not contain %q", got, want)
		}
	}
}

func TestTributeCommandDirectoryKeepsValidFilesAndReportsFaults(t *testing.T) {
	dir := t.TempDir()
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "abc.arktributetribe"), []uint64{11}, []uint64{22})
	if err := os.WriteFile(filepath.Join(dir, "broken.arktributetribe"), []byte("not a tribute index"), 0o600); err != nil {
		t.Fatalf("write broken tribute file: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"tribute", dir}, &out)
	if err != nil {
		t.Fatalf("run(tribute directory) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribute directory:",
		"Files: 1",
		"File faults: 1",
		"Tribute file:",
		"Player data IDs: 1",
		"Tribe data IDs: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("tribute directory output %q does not contain %q", got, want)
		}
	}
}

func TestTributeCommandRedactsLocalTributeDetails(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.arktributetribe")
	testfixtures.WriteTributeFile(t, path, []uint64{11, 22}, []uint64{33})

	var out bytes.Buffer
	err := run([]string{"--redact", "tribute", path}, &out)
	if err != nil {
		t.Fatalf("run(--redact tribute) error = %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"Tribute file: [redacted]",
		"Player data IDs: 2",
		"Tribe data IDs: 1",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("redacted tribute output %q does not contain %q", got, want)
		}
	}
	for _, leaked := range []string{path, "player_data_id=", "tribe_data_id=", "abc.arktributetribe"} {
		if strings.Contains(got, leaked) {
			t.Fatalf("redacted tribute output %q contains private detail %q", got, leaked)
		}
	}
}

func TestExportJSONWritesSaveInfoToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "save_info.json")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"export-json", savePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-json output %q does not mention %q", out.String(), outPath)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(exported json) error = %v", err)
	}
	var decoded struct {
		MapName     string `json:"map_name"`
		SaveVersion int16  `json:"save_version"`
		ObjectCount int    `json:"object_count"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, data)
	}
	if decoded.MapName != "Valguero_WP" || decoded.SaveVersion != 12 || decoded.ObjectCount != 1 {
		t.Fatalf("exported json = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportJSONRedactsObjectDetailsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "save_info.json")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"--redact", "export-json", savePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(--redact export-json) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-json output %q mentions output path %q", out.String(), outPath)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted json) error = %v", err)
	}
	var decoded arkapi.SaveInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, data)
	}
	if decoded.ObjectCount != 1 || len(decoded.Objects) != 0 {
		t.Fatalf("redacted SaveInfo = %#v, want object count without object details", decoded)
	}
	if strings.Contains(string(data), "00010203") || strings.Contains(string(data), "Blueprint'/Game/Test.Test_C'") {
		t.Fatalf("redacted save info contains object detail: %s", data)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONWritesClusterSummaryToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	outPath := filepath.Join(dir, "cluster.json")
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"export-cluster-json", clusterPath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-cluster-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-cluster-json output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(cluster json) error = %v", err)
	}
	var decoded arkapi.ClusterDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ID != "EOS_abc123" || decoded.ArchiveVersion != 7 || decoded.ObjectCount != 1 {
		t.Fatalf("decoded ClusterDataInfo = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONRedactsIdentifiersWhenRequested(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	outPath := filepath.Join(dir, "cluster.json")
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"export-cluster-json", "--redact", clusterPath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-cluster-json --redact) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-cluster-json output %q mentions output path %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted cluster json) error = %v", err)
	}
	var decoded arkapi.ClusterDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ID != "[redacted]" || decoded.Path != "[redacted]" || decoded.ObjectCount != 1 || len(decoded.Items) != 0 || len(decoded.Dinos) != 0 {
		t.Fatalf("redacted ClusterDataInfo = %#v", decoded)
	}
	if strings.Contains(string(raw), clusterPath) || strings.Contains(string(raw), "EOS_abc123") {
		t.Fatalf("redacted cluster json contains private detail: %s", raw)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONWritesDirectorySummary(t *testing.T) {
	dir := t.TempDir()
	clusterPath := filepath.Join(dir, "EOS_abc123")
	outPath := filepath.Join(dir, "clusters.json")
	testfixtures.WriteArchive(t, clusterPath, "/Script/ShooterGame.ArkCloudInventoryData")
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_def456"), "/Script/ShooterGame.ArkCloudInventoryData")

	var out bytes.Buffer
	err := run([]string{"export-cluster-json", dir, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-cluster-json directory) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-cluster-json directory output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(cluster directory json) error = %v", err)
	}
	var decoded arkapi.ClusterDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(cluster directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 2 || len(decoded.Files) != 2 {
		t.Fatalf("decoded ClusterDirectoryInfo = %#v, want two files", decoded)
	}
	if decoded.Files[0].ID != "EOS_abc123" || decoded.Files[1].ID != "EOS_def456" {
		t.Fatalf("decoded cluster IDs = %#v", decoded.Files)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONRedactsDirectoryIdentifiersWhenRequested(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "clusters.json")
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_abc123"), "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(filepath.Join(dir, "EOS_broken"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken cluster file: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"--redact", "export-cluster-json", dir, outPath}, &out)
	if err != nil {
		t.Fatalf("run(--redact export-cluster-json directory) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-cluster-json directory output %q mentions output path %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted cluster directory json) error = %v", err)
	}
	var decoded arkapi.ClusterDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(redacted cluster directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 1 || len(decoded.Files) != 1 || decoded.Files[0].ID != redactedValue || decoded.Files[0].Path != redactedValue || len(decoded.Files[0].Items) != 0 || len(decoded.Files[0].Dinos) != 0 {
		t.Fatalf("redacted ClusterDirectoryInfo files = %#v", decoded)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Path != redactedValue || decoded.Faults[0].Error == "" {
		t.Fatalf("redacted ClusterDirectoryInfo faults = %#v", decoded.Faults)
	}
	for _, leaked := range []string{dir, outPath, "EOS_abc123", "EOS_broken"} {
		if strings.Contains(string(raw), leaked) {
			t.Fatalf("redacted cluster directory json contains private detail %q: %s", leaked, raw)
		}
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportClusterJSONDirectoryKeepsValidFilesAndReportsFaults(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "clusters.json")
	testfixtures.WriteArchive(t, filepath.Join(dir, "EOS_valid"), "/Script/ShooterGame.ArkCloudInventoryData")
	if err := os.WriteFile(filepath.Join(dir, "EOS_broken"), []byte("not an archive"), 0o600); err != nil {
		t.Fatalf("write broken cluster file: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"export-cluster-json", dir, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-cluster-json directory) error = %v", err)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(cluster directory json) error = %v", err)
	}
	var decoded arkapi.ClusterDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(cluster directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 1 || len(decoded.Files) != 1 || decoded.Files[0].ID != "EOS_valid" {
		t.Fatalf("decoded valid files = %#v, want one valid cluster file", decoded)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Path != filepath.Join(dir, "EOS_broken") || decoded.Faults[0].Error == "" {
		t.Fatalf("decoded faults = %#v, want broken cluster file fault", decoded.Faults)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportTributeJSONWritesSummaryToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	outPath := filepath.Join(dir, "tribute.json")
	testfixtures.WriteTributeFile(t, tributePath, []uint64{11, 22}, []uint64{33})

	var out bytes.Buffer
	err := run([]string{"export-tribute-json", tributePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-tribute-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-tribute-json output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(tribute json) error = %v", err)
	}
	var decoded arkapi.TributeDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ID != "abc" || decoded.PlayerDataCount != 2 || decoded.TribeDataCount != 1 {
		t.Fatalf("decoded TributeDataInfo = %#v", decoded)
	}
	if !reflect.DeepEqual(decoded.PlayerDataIDs, []uint64{11, 22}) || !reflect.DeepEqual(decoded.TribeDataIDs, []uint64{33}) {
		t.Fatalf("decoded tribute IDs = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportTributeJSONRedactsIdentifiersWhenRequested(t *testing.T) {
	dir := t.TempDir()
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	outPath := filepath.Join(dir, "tribute.json")
	testfixtures.WriteTributeFile(t, tributePath, []uint64{11, 22}, []uint64{33})

	var out bytes.Buffer
	err := run([]string{"--redact", "export-tribute-json", tributePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(--redact export-tribute-json) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-tribute-json output %q mentions output path %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted tribute json) error = %v", err)
	}
	var decoded arkapi.TributeDataInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.ID != "[redacted]" || decoded.Path != "[redacted]" || decoded.PlayerDataCount != 2 || decoded.TribeDataCount != 1 {
		t.Fatalf("redacted TributeDataInfo = %#v", decoded)
	}
	if decoded.PlayerDataIDs != nil || decoded.TribeDataIDs != nil {
		t.Fatalf("redacted tribute IDs = %#v/%#v, want hidden", decoded.PlayerDataIDs, decoded.TribeDataIDs)
	}
	if strings.Contains(string(raw), tributePath) || strings.Contains(string(raw), "abc.arktributetribe") || strings.Contains(string(raw), "11") {
		t.Fatalf("redacted tribute json contains private detail: %s", raw)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportTributeJSONWritesDirectorySummary(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "tributes.json")
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "abc.arktributetribe"), []uint64{11}, nil)
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "def.arktributetribetribe"), nil, []uint64{22})

	var out bytes.Buffer
	err := run([]string{"export-tribute-json", dir, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-tribute-json directory) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-tribute-json directory output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(tribute directory json) error = %v", err)
	}
	var decoded arkapi.TributeDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(tribute directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 2 || len(decoded.Files) != 2 {
		t.Fatalf("decoded TributeDirectoryInfo = %#v, want two files", decoded)
	}
	if decoded.Files[0].ID != "abc" || decoded.Files[1].ID != "def" {
		t.Fatalf("decoded tribute IDs = %#v", decoded.Files)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportTributeJSONRedactsDirectoryIdentifiersWhenRequested(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "tributes.json")
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "abc.arktributetribe"), []uint64{11}, nil)
	if err := os.WriteFile(filepath.Join(dir, "broken.arktributetribe"), []byte("not a tribute index"), 0o600); err != nil {
		t.Fatalf("write broken tribute file: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"--redact", "export-tribute-json", dir, outPath}, &out)
	if err != nil {
		t.Fatalf("run(--redact export-tribute-json directory) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-tribute-json directory output %q mentions output path %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted tribute directory json) error = %v", err)
	}
	var decoded arkapi.TributeDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(redacted tribute directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 1 || len(decoded.Files) != 1 || decoded.Files[0].ID != redactedValue || decoded.Files[0].Path != redactedValue {
		t.Fatalf("redacted TributeDirectoryInfo files = %#v", decoded)
	}
	if len(decoded.Files[0].PlayerDataIDs) != 0 || len(decoded.Files[0].TribeDataIDs) != 0 {
		t.Fatalf("redacted tribute IDs = %#v/%#v, want hidden", decoded.Files[0].PlayerDataIDs, decoded.Files[0].TribeDataIDs)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Path != redactedValue || decoded.Faults[0].Error == "" {
		t.Fatalf("redacted TributeDirectoryInfo faults = %#v", decoded.Faults)
	}
	for _, leaked := range []string{dir, outPath, "abc.arktributetribe", "broken.arktributetribe", "11"} {
		if strings.Contains(string(raw), leaked) {
			t.Fatalf("redacted tribute directory json contains private detail %q: %s", leaked, raw)
		}
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportTributeJSONDirectoryKeepsValidFilesAndReportsFaults(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "tributes.json")
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "abc.arktributetribe"), []uint64{11}, nil)
	if err := os.WriteFile(filepath.Join(dir, "broken.arktributetribe"), []byte("not a tribute index"), 0o600); err != nil {
		t.Fatalf("write broken tribute file: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"export-tribute-json", dir, outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-tribute-json directory) error = %v", err)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(tribute directory json) error = %v", err)
	}
	var decoded arkapi.TributeDirectoryInfo
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(tribute directory) error = %v; data = %s", err, raw)
	}
	if decoded.Count != 1 || len(decoded.Files) != 1 || decoded.Files[0].ID != "abc" {
		t.Fatalf("decoded valid files = %#v, want one valid tribute file", decoded)
	}
	if len(decoded.Faults) != 1 || decoded.Faults[0].Path != filepath.Join(dir, "broken.arktributetribe") || decoded.Faults[0].Error == "" {
		t.Fatalf("decoded faults = %#v, want broken tribute file fault", decoded.Faults)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportDomainJSONWritesDomainSummaryToExplicitPath(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "stackables.json")
	createSyntheticEmptySave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"export-domain-json", savePath, "stackables", outPath}, &out)
	if err != nil {
		t.Fatalf("run(export-domain-json) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("export-domain-json output %q does not mention %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(domain json) error = %v", err)
	}
	var decoded arkapi.DomainExport
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.Domain != "stackables" {
		t.Fatalf("decoded DomainExport = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportDomainJSONRedactsItemsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "stackables.json")
	createSyntheticEmptySave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"--redact", "export-domain-json", savePath, "stackables", outPath}, &out)
	if err != nil {
		t.Fatalf("run(--redact export-domain-json) error = %v", err)
	}
	if strings.Contains(out.String(), outPath) {
		t.Fatalf("redacted export-domain-json output %q mentions output path %q", out.String(), outPath)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile(redacted domain json) error = %v", err)
	}
	var decoded arkapi.DomainExport
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded.Domain != "stackables" || decoded.Count != 0 || decoded.Items != nil {
		t.Fatalf("redacted DomainExport = %#v", decoded)
	}
	assertPrivateFileMode(t, outPath)
}

func TestExportDomainJSONAcceptsPlayerAndTribeDomains(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	createSyntheticEmptySave(t, savePath)

	for _, domain := range []string{"players", "tribes"} {
		outPath := filepath.Join(dir, domain+".json")
		var out bytes.Buffer
		if err := run([]string{"--redact", "export-domain-json", savePath, domain, outPath}, &out); err != nil {
			t.Fatalf("run(--redact export-domain-json %s) error = %v", domain, err)
		}
		if strings.Contains(out.String(), outPath) {
			t.Fatalf("redacted export-domain-json %s output %q mentions output path %q", domain, out.String(), outPath)
		}
		raw, err := os.ReadFile(outPath)
		if err != nil {
			t.Fatalf("ReadFile(%s domain json) error = %v", domain, err)
		}
		var decoded arkapi.DomainExport
		if err := json.Unmarshal(raw, &decoded); err != nil {
			t.Fatalf("json.Unmarshal(%s) error = %v; data = %s", domain, err, raw)
		}
		if decoded.Domain != domain || decoded.Count != 0 || decoded.Items != nil {
			t.Fatalf("redacted %s DomainExport = %#v", domain, decoded)
		}
		assertPrivateFileMode(t, outPath)
	}
}

func TestMutateCopyCommandWritesExplicitOutput(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "copy.ark")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"mutate", "copy", savePath, outPath}, &out)
	if err != nil {
		t.Fatalf("run(mutate copy) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("mutate copy output %q does not mention %q", out.String(), outPath)
	}
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("stat mutated copy: %v", err)
	}
}

func TestMutateCommandsRedactOutputDetailsWhenRequested(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	copyPath := filepath.Join(dir, "copy.ark")
	removedPath := filepath.Join(dir, "removed.ark")
	removedClassPath := filepath.Join(dir, "removed-class.ark")
	objectPath := filepath.Join(dir, "object.ark")
	importBasePath := filepath.Join(dir, "import-base.ark")
	importStructurePath := filepath.Join(dir, "import-structure.ark")
	importDinoPath := filepath.Join(dir, "import-dino.ark")
	importEquipmentPath := filepath.Join(dir, "import-equipment.ark")
	propertyPath := filepath.Join(dir, "property.ark")
	baseExportDir := filepath.Join(dir, "base-export", "base_aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	structureExportDir := filepath.Join(dir, "structure-export")
	dinoExportDir := filepath.Join(dir, "dino-export", "dino_bbbbbbbb-bbbb-cccc-dddd-eeeeffffffff")
	equipmentExportDir := filepath.Join(dir, "equipment-export")
	createSyntheticSave(t, savePath)
	objectID := "00010203-0405-0607-0809-0a0b0c0d0e0f"
	if err := os.MkdirAll(baseExportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(base export) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(baseExportDir, "str_aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff.bin"), testfixtures.GenericObjectBytes(1, 2), 0o600); err != nil {
		t.Fatalf("write base export row: %v", err)
	}
	if err := os.MkdirAll(structureExportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(structure export) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(structureExportDir, "manifest.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write structure export manifest: %v", err)
	}
	if err := os.WriteFile(filepath.Join(structureExportDir, "str_dddddddd-bbbb-cccc-dddd-eeeeffffffff.bin"), testfixtures.GenericObjectBytes(1, 2), 0o600); err != nil {
		t.Fatalf("write structure export row: %v", err)
	}
	if err := os.MkdirAll(dinoExportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(dino export) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(dinoExportDir, "dino_bbbbbbbb-bbbb-cccc-dddd-eeeeffffffff.bin"), testfixtures.GenericObjectBytes(1, 2), 0o600); err != nil {
		t.Fatalf("write dino export row: %v", err)
	}
	if err := os.MkdirAll(equipmentExportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(equipment export) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(equipmentExportDir, "item_cccccccc-bbbb-cccc-dddd-eeeeffffffff.bin"), testfixtures.GenericObjectBytes(1, 2), 0o600); err != nil {
		t.Fatalf("write equipment export row: %v", err)
	}

	var copyOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "copy", savePath, copyPath}, &copyOut); err != nil {
		t.Fatalf("run(--redact mutate copy) error = %v", err)
	}
	if strings.Contains(copyOut.String(), copyPath) || !strings.Contains(copyOut.String(), "[redacted]") {
		t.Fatalf("redacted mutate copy output = %q", copyOut.String())
	}

	var removeOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "remove-object", savePath, removedPath, objectID}, &removeOut); err != nil {
		t.Fatalf("run(--redact mutate remove-object) error = %v", err)
	}
	got := removeOut.String()
	if strings.Contains(got, removedPath) || strings.Contains(got, objectID) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate remove-object output = %q", got)
	}

	var removeClassOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "remove-class-contains", savePath, removedClassPath, "/Game/Test"}, &removeClassOut); err != nil {
		t.Fatalf("run(--redact mutate remove-class-contains) error = %v", err)
	}
	got = removeClassOut.String()
	if strings.Contains(got, removedClassPath) || strings.Contains(got, "/Game/Test") || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate remove-class-contains output = %q", got)
	}

	var objectOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "put-object-hex", savePath, objectPath, objectID, "090807"}, &objectOut); err != nil {
		t.Fatalf("run(--redact mutate put-object-hex) error = %v", err)
	}
	got = objectOut.String()
	if strings.Contains(got, objectPath) || strings.Contains(got, objectID) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate put-object-hex output = %q", got)
	}

	var replacement bytes.Buffer
	testfixtures.WriteIntPropertyID(&replacement, 0x10000004, 0x10000003, 2002)
	var propertyOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "replace-object-property-hex", savePath, propertyPath, objectID, "DinoID1", "0", hex.EncodeToString(replacement.Bytes())}, &propertyOut); err != nil {
		t.Fatalf("run(--redact mutate replace-object-property-hex) error = %v", err)
	}
	got = propertyOut.String()
	if strings.Contains(got, propertyPath) || strings.Contains(got, objectID) || strings.Contains(got, "DinoID1") || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate replace-object-property-hex output = %q", got)
	}

	var importBaseOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "import-base-binary", savePath, importBasePath, filepath.Join(dir, "base-export")}, &importBaseOut); err != nil {
		t.Fatalf("run(--redact mutate import-base-binary) error = %v", err)
	}
	got = importBaseOut.String()
	if strings.Contains(got, importBasePath) || strings.Contains(got, filepath.Join(dir, "base-export")) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate import-base-binary output = %q", got)
	}

	var importStructureOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "import-structure-binary", savePath, importStructurePath, structureExportDir}, &importStructureOut); err != nil {
		t.Fatalf("run(--redact mutate import-structure-binary) error = %v", err)
	}
	got = importStructureOut.String()
	if strings.Contains(got, importStructurePath) || strings.Contains(got, structureExportDir) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate import-structure-binary output = %q", got)
	}

	var importDinoOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "import-dino-binary", savePath, importDinoPath, filepath.Join(dir, "dino-export")}, &importDinoOut); err != nil {
		t.Fatalf("run(--redact mutate import-dino-binary) error = %v", err)
	}
	got = importDinoOut.String()
	if strings.Contains(got, importDinoPath) || strings.Contains(got, filepath.Join(dir, "dino-export")) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate import-dino-binary output = %q", got)
	}

	var importEquipmentOut bytes.Buffer
	if err := run([]string{"--redact", "mutate", "import-equipment-binary", savePath, importEquipmentPath, equipmentExportDir}, &importEquipmentOut); err != nil {
		t.Fatalf("run(--redact mutate import-equipment-binary) error = %v", err)
	}
	got = importEquipmentOut.String()
	if strings.Contains(got, importEquipmentPath) || strings.Contains(got, equipmentExportDir) || !strings.Contains(got, "[redacted]") {
		t.Fatalf("redacted mutate import-equipment-binary output = %q", got)
	}
}

func TestMutateRemoveObjectCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "removed.ark")
	createSyntheticSave(t, savePath)
	objectID := "00010203-0405-0607-0809-0a0b0c0d0e0f"

	var out bytes.Buffer
	err := run([]string{"mutate", "remove-object", savePath, outPath, objectID}, &out)
	if err != nil {
		t.Fatalf("run(mutate remove-object) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), objectID) {
		t.Fatalf("mutate remove-object output %q missing path or uuid", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	ids, err := save.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs(mutated output) error = %v", err)
	}
	_ = save.Close()
	if len(ids) != 0 {
		t.Fatalf("mutated output ObjectIDs length = %d, want 0", len(ids))
	}
}

func TestMutateRemoveClassContainsCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "removed-class.ark")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"mutate", "remove-class-contains", savePath, outPath, "/Game/Test"}, &out)
	if err != nil {
		t.Fatalf("run(mutate remove-class-contains) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), "/Game/Test") || !strings.Contains(out.String(), "removed=1") {
		t.Fatalf("mutate remove-class-contains output %q missing path, class substring, or removal count", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	ids, err := save.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs(mutated output) error = %v", err)
	}
	_ = save.Close()
	if len(ids) != 0 {
		t.Fatalf("mutated output ObjectIDs length = %d, want 0", len(ids))
	}
}

func TestMutatePutCustomCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "custom.ark")
	createSyntheticSave(t, savePath)

	var out bytes.Buffer
	err := run([]string{"mutate", "put-custom", savePath, outPath, "Extra", "090807"}, &out)
	if err != nil {
		t.Fatalf("run(mutate put-custom) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), "Extra") {
		t.Fatalf("mutate put-custom output %q missing path or key", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.CustomValue("Extra")
	if err != nil {
		t.Fatalf("CustomValue(Extra) error = %v", err)
	}
	_ = save.Close()
	if !bytes.Equal(got, []byte{9, 8, 7}) {
		t.Fatalf("CustomValue(Extra) = % x, want 09 08 07", got)
	}
}

func TestMutatePutObjectHexCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "object.ark")
	createSyntheticSave(t, savePath)
	objectID := uuid.MustParse("00010203-0405-0607-0809-0a0b0c0d0e0f")

	var out bytes.Buffer
	err := run([]string{"mutate", "put-object-hex", savePath, outPath, objectID.String(), "090807"}, &out)
	if err != nil {
		t.Fatalf("run(mutate put-object-hex) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), objectID.String()) {
		t.Fatalf("mutate put-object-hex output %q missing path or uuid", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", objectID, err)
	}
	_ = save.Close()
	if !bytes.Equal(got, []byte{9, 8, 7}) {
		t.Fatalf("ObjectBinary(%s) = % x, want 09 08 07", objectID, got)
	}
}

func TestMutateReplaceObjectPropertyHexCommandWritesReparseableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "property.ark")
	createSyntheticSave(t, savePath)
	objectID := uuid.MustParse("00010203-0405-0607-0809-0a0b0c0d0e0f")
	var replacement bytes.Buffer
	testfixtures.WriteIntPropertyID(&replacement, 0x10000004, 0x10000003, 2002)

	var out bytes.Buffer
	err := run([]string{"mutate", "replace-object-property-hex", savePath, outPath, objectID.String(), "DinoID1", "0", hex.EncodeToString(replacement.Bytes())}, &out)
	if err != nil {
		t.Fatalf("run(mutate replace-object-property-hex) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), objectID.String()) || !strings.Contains(out.String(), "DinoID1") {
		t.Fatalf("mutate replace-object-property-hex output %q missing path, uuid, or property", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	object, err := save.ParsedObject(objectID)
	if err != nil {
		t.Fatalf("ParsedObject(%s) error = %v", objectID, err)
	}
	_ = save.Close()
	got, ok := object.Container().Value("DinoID1")
	if !ok || got != int32(2002) {
		t.Fatalf("DinoID1 = %#v, %v; want int32(2002)", got, ok)
	}
}

func TestMutateImportBaseBinaryCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "imported-base.ark")
	exportDir := filepath.Join(dir, "base-export", "base_aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	createSyntheticSave(t, savePath)
	objectID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(exportDir, "manifest.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write structure export manifest: %v", err)
	}
	want := testfixtures.GenericObjectBytes(1, 2)
	if err := os.WriteFile(filepath.Join(exportDir, "str_"+objectID.String()+".bin"), want, 0o600); err != nil {
		t.Fatalf("write base export row: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"mutate", "import-base-binary", savePath, outPath, filepath.Join(dir, "base-export")}, &out)
	if err != nil {
		t.Fatalf("run(mutate import-base-binary) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), "rows=1") {
		t.Fatalf("mutate import-base-binary output %q missing path or row count", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", objectID, err)
	}
	_ = save.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("ObjectBinary(%s) = % x, want exported row", objectID, got)
	}
}

func TestMutateImportStructureBinaryCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "imported-structure.ark")
	exportDir := filepath.Join(dir, "structure-export")
	createSyntheticSave(t, savePath)
	objectID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(exportDir, "manifest.json"), []byte("{}\n"), 0o600); err != nil {
		t.Fatalf("write structure export manifest: %v", err)
	}
	want := testfixtures.GenericObjectBytes(1, 2)
	if err := os.WriteFile(filepath.Join(exportDir, "str_"+objectID.String()+".bin"), want, 0o600); err != nil {
		t.Fatalf("write structure export row: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"mutate", "import-structure-binary", savePath, outPath, exportDir}, &out)
	if err != nil {
		t.Fatalf("run(mutate import-structure-binary) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), "rows=1") {
		t.Fatalf("mutate import-structure-binary output %q missing path or row count", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", objectID, err)
	}
	_ = save.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("ObjectBinary(%s) = % x, want exported row", objectID, got)
	}
}

func TestMutateImportDinoBinaryCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "imported-dino.ark")
	exportDir := filepath.Join(dir, "dino-export", "dino_aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	createSyntheticSave(t, savePath)
	objectID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	want := testfixtures.GenericObjectBytes(1, 2)
	if err := os.WriteFile(filepath.Join(exportDir, "dino_"+objectID.String()+".bin"), want, 0o600); err != nil {
		t.Fatalf("write dino export row: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"mutate", "import-dino-binary", savePath, outPath, filepath.Join(dir, "dino-export")}, &out)
	if err != nil {
		t.Fatalf("run(mutate import-dino-binary) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), "rows=1") {
		t.Fatalf("mutate import-dino-binary output %q missing path or row count", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", objectID, err)
	}
	_ = save.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("ObjectBinary(%s) = % x, want exported row", objectID, got)
	}
}

func TestMutateImportEquipmentBinaryCommandWritesReopenableCopy(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "synthetic.ark")
	outPath := filepath.Join(dir, "imported-equipment.ark")
	exportDir := filepath.Join(dir, "equipment-export")
	createSyntheticSave(t, savePath)
	objectID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	if err := os.MkdirAll(exportDir, 0o700); err != nil {
		t.Fatalf("MkdirAll(exportDir) error = %v", err)
	}
	want := testfixtures.GenericObjectBytes(1, 2)
	if err := os.WriteFile(filepath.Join(exportDir, "item_"+objectID.String()+".bin"), want, 0o600); err != nil {
		t.Fatalf("write equipment export row: %v", err)
	}

	var out bytes.Buffer
	err := run([]string{"mutate", "import-equipment-binary", savePath, outPath, exportDir}, &out)
	if err != nil {
		t.Fatalf("run(mutate import-equipment-binary) error = %v", err)
	}
	if !strings.Contains(out.String(), outPath) || !strings.Contains(out.String(), "rows=1") {
		t.Fatalf("mutate import-equipment-binary output %q missing path or row count", out.String())
	}
	save, err := arksave.Open(outPath)
	if err != nil {
		t.Fatalf("Open(mutated output) error = %v", err)
	}
	got, err := save.ObjectBinary(objectID)
	if err != nil {
		t.Fatalf("ObjectBinary(%s) error = %v", objectID, err)
	}
	_ = save.Close()
	if !bytes.Equal(got, want) {
		t.Fatalf("ObjectBinary(%s) = % x, want exported row", objectID, got)
	}
}

func createSyntheticSave(t *testing.T, path string) {
	t.Helper()
	objectID := uuid.MustParse("00010203-0405-0607-0809-0a0b0c0d0e0f")
	var props bytes.Buffer
	testfixtures.WriteIntPropertyID(&props, 0x10000004, 0x10000003, 1001)
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000001: "Blueprint'/Game/Test.Test_C'",
			0x10000002: "None",
			0x10000003: "IntProperty",
			0x10000004: "DinoID1",
		}),
		Objects: map[uuid.UUID][]byte{
			objectID: testfixtures.ObjectBytesWithProperties(0x10000001, 0x10000002, props.Bytes()),
		},
	})
}

func createSyntheticStructureHealthSave(t *testing.T, path string) {
	t.Helper()
	structureID := uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000001: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			0x10000002: "None",
			0x10000003: "IntProperty",
			0x10000004: "StructureID",
			0x10000005: "TargetingTeam",
			0x10000006: "MaxHealth",
			0x10000007: "FloatProperty",
			0x10000008: "Health",
		}),
		Objects: map[uuid.UUID][]byte{
			structureID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
				ClassID:             0x10000001,
				NoneID:              0x10000002,
				IntPropertyID:       0x10000003,
				StructureIDNameID:   0x10000004,
				TribeIDNameID:       0x10000005,
				MaxHealthNameID:     0x10000006,
				FloatPropertyID:     0x10000007,
				CurrentHealthNameID: 0x10000008,
				StructureID:         101,
				TribeID:             555,
				MaxHealth:           10000,
				CurrentHealth:       9000,
			}),
		},
		Custom: map[string][]byte{
			"ActorTransforms": testfixtures.ActorTransforms(testfixtures.ActorTransform{
				UUID:       structureID,
				X:          11,
				Y:          22,
				Z:          33,
				Quaternion: 1,
			}),
		},
	})
}

func createSyntheticDinoSave(t *testing.T, path string) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			0x10000015: "DinoID1",
			0x10000016: "DinoID2",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("10111213-1415-1617-1819-1a1b1c1d1e1f"): testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
				ID1: 1001,
				ID2: 2002,
			}),
		},
	})
}

func createSyntheticLocatedDinoSave(t *testing.T, path string) {
	t.Helper()
	dinoID := uuid.MustParse("10111213-1415-1617-1819-1a1b1c1d1e1f")
	transform := arkobject.MapCoords{Lat: 12.4, Long: 34.6}.AsActorTransform("Valguero")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			0x10000015: "DinoID1",
			0x10000016: "DinoID2",
		}),
		Objects: map[uuid.UUID][]byte{
			dinoID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
				ID1: 1001,
				ID2: 2002,
			}),
		},
		Custom: map[string][]byte{
			"ActorTransforms": testfixtures.ActorTransforms(testfixtures.ActorTransform{
				UUID:       dinoID,
				X:          transform.X,
				Y:          transform.Y,
				Z:          transform.Z,
				Quaternion: 1,
			}),
		},
	})
}

func createSyntheticBabyDinoSave(t *testing.T, path string) {
	t.Helper()
	trueValue := true
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000e: "BoolProperty",
			0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			0x10000015: "DinoID1",
			0x10000016: "DinoID2",
			0x10000018: "TamedTimeStamp",
			0x10000019: "DoubleProperty",
			0x10000021: "bIsBaby",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("20212223-2425-2627-2829-2a2b2c2d2e2f"): testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
				ID1:    1001,
				ID2:    2002,
				IsBaby: &trueValue,
				Tamed:  true,
			}),
			uuid.MustParse("30313233-3435-3637-3839-3a3b3c3d3e3f"): testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
				ID1:    3003,
				ID2:    4004,
				IsBaby: &trueValue,
			}),
		},
	})
}

func createSyntheticDinoStatsSave(t *testing.T, path string) {
	t.Helper()
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	statusID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000a: "FloatProperty",
			0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			0x10000015: "DinoID1",
			0x10000016: "DinoID2",
			0x1000001f: "ObjectProperty",
			0x10000035: "MyCharacterStatusComponent",
			0x10000036: "Blueprint'/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatusComponent_BP.DinoCharacterStatusComponent_BP_C'",
			0x10000037: "BaseCharacterLevel",
			0x10000038: "NumberOfLevelUpPointsApplied",
			0x10000039: "NumberOfLevelUpPointsAppliedTamed",
			0x1000003a: "NumberOfMutationsAppliedTamed",
			0x1000003b: "CurrentStatusValues",
			0x1000003c: "DinoImprintingQuality",
		}),
		Objects: map[uuid.UUID][]byte{
			dinoID:   testfixtures.DinoStatsFixtureObjectBytes(statusID, false),
			statusID: testfixtures.DinoStatusComponentFixtureBytes(5),
		},
	})
}

func createSyntheticTamedDinoStatsSave(t *testing.T, path string) {
	t.Helper()
	dinoID := uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff")
	statusID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000a: "FloatProperty",
			0x10000014: "Blueprint'/Game/PrimalEarth/Dinos/Raptor/Raptor_Character_BP.Raptor_Character_BP_C'",
			0x10000015: "DinoID1",
			0x10000016: "DinoID2",
			0x10000018: "TamedTimeStamp",
			0x10000019: "DoubleProperty",
			0x1000001f: "ObjectProperty",
			0x10000035: "MyCharacterStatusComponent",
			0x10000036: "Blueprint'/Game/PrimalEarth/CoreBlueprints/DinoCharacterStatusComponent_BP.DinoCharacterStatusComponent_BP_C'",
			0x10000037: "BaseCharacterLevel",
			0x10000038: "NumberOfLevelUpPointsApplied",
			0x10000039: "NumberOfLevelUpPointsAppliedTamed",
			0x1000003a: "NumberOfMutationsAppliedTamed",
			0x1000003b: "CurrentStatusValues",
			0x1000003c: "DinoImprintingQuality",
		}),
		Objects: map[uuid.UUID][]byte{
			dinoID:   testfixtures.DinoStatsFixtureObjectBytes(statusID, true),
			statusID: testfixtures.DinoStatusComponentFixtureBytes(5),
		},
	})
}

func createSyntheticEquipmentSave(t *testing.T, path string) {
	t.Helper()
	trueValue := true
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000a: "FloatProperty",
			0x1000000c: "ItemQuantity",
			0x1000000d: "bIsBlueprint",
			0x1000000e: "BoolProperty",
			0x1000000f: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
			0x10000010: "ItemRating",
			0x10000011: "ItemQualityIndex",
			0x10000012: "SavedDurability",
			0x1000001a: "StrProperty",
			0x1000001b: "CrafterCharacterName",
			0x1000001c: "CrafterTribeName",
			0x10000022: "bEquippedItem",
			0x10000040: "ItemStatValues",
			0x10000041: "UInt16Property",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("20212223-2425-2627-2829-2a2b2c2d2e2f"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				Quantity:             2,
				Rating:               7.5,
				Quality:              3,
				Durability:           0.75,
				IsBlueprint:          &trueValue,
				IsEquipped:           &trueValue,
				CrafterCharacterName: "Survivor",
				CrafterTribeName:     "Porters",
				Stats: map[int32]uint16{
					2: 1000,
					3: 1234,
				},
			}),
		},
	})
}

func createSyntheticSaddleEquipmentSave(t *testing.T, path string) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000a: "FloatProperty",
			0x1000000c: "ItemQuantity",
			0x1000000e: "BoolProperty",
			0x1000000f: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_TurtleSaddle.PrimalItemArmor_TurtleSaddle_C'",
			0x10000010: "ItemRating",
			0x10000011: "ItemQualityIndex",
			0x10000012: "SavedDurability",
			0x10000040: "ItemStatValues",
			0x10000041: "UInt16Property",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("40414243-4445-4647-4849-4a4b4c4d4e4f"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				Quantity: 1,
				Rating:   4.5,
				Quality:  3,
				Stats: map[int32]uint16{
					1: 800,
				},
			}),
		},
	})
}

func createSyntheticStackableSave(t *testing.T, path string) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000c: "ItemQuantity",
			0x1000000b: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("50515253-5455-5657-5859-5a5b5c5d5e5f"): testfixtures.StackableGameObjectBytes(testfixtures.StackableGameObjectOptions{
				Quantity: 250,
			}),
		},
	})
}

func createSyntheticOwnedStackableSave(t *testing.T, path string) {
	t.Helper()
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	otherInventoryID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x10000005: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			0x10000006: "StructureID",
			0x10000009: "TargetingTeam",
			0x1000000b: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Resources/PrimalItemResource_Stone.PrimalItemResource_Stone_C'",
			0x1000000c: "ItemQuantity",
			0x1000001f: "ObjectProperty",
			0x10000023: "MyInventoryComponent",
			0x10000044: "OwnerInventory",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
				StructureID: 101,
				TribeID:     555,
				InventoryID: inventoryID,
			}),
			uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): testfixtures.StackableGameObjectBytes(testfixtures.StackableGameObjectOptions{
				Quantity:         100,
				OwnerInventoryID: inventoryID,
			}),
			uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000"): testfixtures.StackableGameObjectBytes(testfixtures.StackableGameObjectOptions{
				Quantity:         200,
				OwnerInventoryID: otherInventoryID,
			}),
		},
	})
}

func createSyntheticOwnedEquipmentSave(t *testing.T, path string) {
	t.Helper()
	trueValue := true
	inventoryID := uuid.MustParse("99999999-aaaa-bbbb-cccc-ddddeeeeffff")
	otherInventoryID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x10000005: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'",
			0x10000006: "StructureID",
			0x10000009: "TargetingTeam",
			0x1000000a: "FloatProperty",
			0x1000000c: "ItemQuantity",
			0x1000000d: "bIsBlueprint",
			0x1000000e: "BoolProperty",
			0x1000000f: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
			0x10000010: "ItemRating",
			0x1000001f: "ObjectProperty",
			0x10000023: "MyInventoryComponent",
			0x10000040: "ItemStatValues",
			0x10000041: "UInt16Property",
			0x10000044: "OwnerInventory",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("aaaaaaaa-bbbb-cccc-dddd-eeeeffffffff"): testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
				StructureID: 101,
				TribeID:     555,
				InventoryID: inventoryID,
			}),
			uuid.MustParse("bbbbbbbb-cccc-dddd-eeee-ffffffffffff"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				Quantity:         1,
				Rating:           7.5,
				IsBlueprint:      &trueValue,
				Stats:            map[int32]uint16{3: 1234},
				OwnerInventoryID: inventoryID,
			}),
			uuid.MustParse("cccccccc-dddd-eeee-ffff-000000000000"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				Quantity:         1,
				Rating:           1.5,
				IsBlueprint:      &trueValue,
				Stats:            map[int32]uint16{3: 2000},
				OwnerInventoryID: otherInventoryID,
			}),
		},
	})
}

func createSyntheticAscendantWeaponBlueprintSave(t *testing.T, path string) {
	t.Helper()
	trueValue := true
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000a: "FloatProperty",
			0x1000000c: "ItemQuantity",
			0x1000000d: "bIsBlueprint",
			0x1000000e: "BoolProperty",
			0x1000000f: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
			0x10000010: "ItemRating",
			0x10000011: "ItemQualityIndex",
			0x10000040: "ItemStatValues",
			0x10000041: "UInt16Property",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("dddddddd-eeee-ffff-0000-111111111111"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				Quantity:    1,
				Rating:      9.5,
				Quality:     arkapi.AscendantQualityIndex,
				IsBlueprint: &trueValue,
				Stats:       map[int32]uint16{3: 1234},
			}),
		},
	})
}

func createSyntheticBestEquipmentSave(t *testing.T, path string) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000a: "FloatProperty",
			0x1000000c: "ItemQuantity",
			0x1000000d: "bIsBlueprint",
			0x1000000e: "BoolProperty",
			0x1000000f: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponBow.PrimalItem_WeaponBow_C'",
			0x10000010: "ItemRating",
			0x10000011: "ItemQualityIndex",
			0x10000012: "SavedDurability",
			0x10000013: "bIsEngram",
			0x10000040: "ItemStatValues",
			0x10000041: "UInt16Property",
			0x10000042: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Cloth/PrimalItemArmor_ClothShirt.PrimalItemArmor_ClothShirt_C'",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("60616263-6465-6667-6869-6a6b6c6d6e6f"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				Quantity: 1,
				Stats: map[int32]uint16{
					3: 1234,
				},
			}),
			uuid.MustParse("70717273-7475-7677-7879-7a7b7c7d7e7f"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				ClassID:  0x10000042,
				Quantity: 1,
				Stats: map[int32]uint16{
					2: 1000,
				},
			}),
		},
	})
}

func createSyntheticRankEquipmentSave(t *testing.T, path string) {
	t.Helper()
	trueValue := true
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", map[uint32]string{
			0x10000003: "IntProperty",
			0x10000004: "None",
			0x1000000a: "FloatProperty",
			0x1000000c: "ItemQuantity",
			0x1000000d: "bIsBlueprint",
			0x1000000e: "BoolProperty",
			0x10000010: "ItemRating",
			0x10000011: "ItemQualityIndex",
			0x10000012: "SavedDurability",
			0x1000001a: "StrProperty",
			0x1000001b: "CrafterCharacterName",
			0x10000040: "ItemStatValues",
			0x10000041: "UInt16Property",
			0x10000050: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Weapons/PrimalItem_WeaponSword.PrimalItem_WeaponSword_C'",
			0x10000051: "Blueprint'/Game/PrimalEarth/CoreBlueprints/Items/Armor/Saddles/PrimalItemArmor_TurtleSaddle.PrimalItemArmor_TurtleSaddle_C'",
		}),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("80818283-8485-8687-8889-8a8b8c8d8e8f"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				ClassID:  0x10000050,
				Quantity: 1,
				Rating:   4.2,
				Stats: map[int32]uint16{
					2: 100,
					3: 200,
				},
			}),
			uuid.MustParse("90919293-9495-9697-9899-9a9b9c9d9e9f"): testfixtures.EquipmentGameObjectBytes(testfixtures.EquipmentGameObjectOptions{
				ClassID:     0x10000051,
				Quantity:    1,
				Rating:      5.5,
				IsBlueprint: &trueValue,
				Stats: map[int32]uint16{
					1: 800,
					2: 600,
				},
			}),
		},
	})
}

func createSyntheticPlayerInventorySave(t *testing.T, path string) {
	t.Helper()
	inventoryID := uuid.MustParse("33333333-4455-6677-8899-aabbccddeeff")
	firstItemID := uuid.MustParse("44444444-4455-6677-8899-aabbccddeeff")
	secondItemID := uuid.MustParse("55555555-4455-6677-8899-aabbccddeeff")
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header: testfixtures.Header("Valguero_WP", nil),
		Objects: map[uuid.UUID][]byte{
			uuid.MustParse("22222233-4455-6677-8899-aabbccddeeff"): testfixtures.PlayerPawnGameObjectBytes(42, inventoryID),
			inventoryID: testfixtures.InventoryGameObjectBytes(
				inventoryID,
				firstItemID,
				secondItemID,
			),
		},
	})
	testfixtures.WritePlayerArchiveWithOptions(t, filepath.Join(filepath.Dir(path), "42.arkprofile"), testfixtures.PlayerArchiveOptions{
		PlayerDataID:  42,
		CharacterName: "Fallback",
	})
}

func createSyntheticEmptySave(t *testing.T, path string) {
	t.Helper()
	testfixtures.WriteSave(t, path, testfixtures.SaveOptions{
		Header:      testfixtures.Header("Valguero_WP", map[uint32]string{1: "Blueprint'/Game/Test.Test_C'"}),
		EmptyTables: true,
	})
}

func assertPrivateFileMode(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat exported file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("exported file mode = %o, want 600", got)
	}
}
