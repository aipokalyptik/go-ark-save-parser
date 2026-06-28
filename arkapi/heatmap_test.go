package arkapi

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSummarizeHeatmapCountsNonzeroCellsTotalsAndMax(t *testing.T) {
	heatmap := [][]int{
		{0, 2, 0},
		{3, 0, 5},
	}

	summary := SummarizeHeatmap(heatmap, 7)
	want := HeatmapSummary{
		Resolution:   2,
		NonzeroCells: 3,
		Total:        10,
		Max:          5,
		Faults:       7,
	}
	if summary != want {
		t.Fatalf("SummarizeHeatmap() = %#v, want %#v", summary, want)
	}
}

func TestHeatmapPathHelpersUseTypedConstructors(t *testing.T) {
	data, err := os.ReadFile("heatmap.go")
	if err != nil {
		t.Fatalf("ReadFile(heatmap.go) error = %v", err)
	}
	source := string(data)
	for _, name := range []string{"DinoHeatmapSummaryFromPath", "StructureHeatmapSummaryFromPath"} {
		body := heatmapFunctionBody(t, source, name)
		if strings.Contains(body, "arksave.Open") {
			t.Fatalf("%s() still opens saves directly; use typed arkapi path constructor", name)
		}
	}
}

func heatmapFunctionBody(t *testing.T, source string, name string) string {
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

func TestHeatmapSummaryFromPathReturnsTypedSummariesWithoutWritingJSON(t *testing.T) {
	dinoSave := openSyntheticDinoHeatmapSaveWithMalformedCryopod(t)
	defer dinoSave.Close()
	dinoSummary, err := DinoHeatmapSummaryFromPath(dinoSave.Path(), DinoHeatmapOptions{
		MapName:           "Valguero",
		Resolution:        100,
		IncludeCryopodded: false,
	})
	if err != nil {
		t.Fatalf("DinoHeatmapSummaryFromPath() error = %v", err)
	}
	if dinoSummary.Total != 1 || dinoSummary.NonzeroCells != 1 || dinoSummary.Faults != 0 {
		t.Fatalf("DinoHeatmapSummaryFromPath() = %#v, want one direct dino without cryopod faults", dinoSummary)
	}

	structureSave := openSyntheticStructureOwnerLocationSave(t)
	defer structureSave.Close()
	structureSummary, err := StructureHeatmapSummaryFromPath(structureSave.Path(), StructureHeatmapOptions{
		MapName:      "Valguero",
		Resolution:   100,
		MinInSection: 2,
	})
	if err != nil {
		t.Fatalf("StructureHeatmapSummaryFromPath() error = %v", err)
	}
	if structureSummary.Total != 2 || structureSummary.NonzeroCells != 1 || structureSummary.Max != 2 || structureSummary.Faults != 0 {
		t.Fatalf("StructureHeatmapSummaryFromPath() = %#v, want one cell with two structures", structureSummary)
	}
}

func TestExportDinoHeatmapSummaryJSONFromPathWritesSummary(t *testing.T) {
	save := openSyntheticDinoHeatmapSaveWithMalformedCryopod(t)
	defer save.Close()
	outputPath := filepath.Join(t.TempDir(), "dino-heatmap.json")

	summary, err := ExportDinoHeatmapSummaryJSONFromPath(save.Path(), outputPath, DinoHeatmapOptions{
		MapName:           "Valguero",
		Resolution:        100,
		IncludeCryopodded: false,
	})
	if err != nil {
		t.Fatalf("ExportDinoHeatmapSummaryJSONFromPath() error = %v", err)
	}
	if summary.Total != 1 || summary.NonzeroCells != 1 || summary.Faults != 0 {
		t.Fatalf("ExportDinoHeatmapSummaryJSONFromPath() summary = %#v, want one direct dino without cryopod faults", summary)
	}
	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outputPath, err)
	}
	var decoded HeatmapSummary
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded != summary {
		t.Fatalf("decoded summary = %#v, want returned summary %#v", decoded, summary)
	}
}

func TestExportStructureHeatmapSummaryJSONFromPathWritesSummary(t *testing.T) {
	save := openSyntheticStructureOwnerLocationSave(t)
	defer save.Close()
	outputPath := filepath.Join(t.TempDir(), "structure-heatmap.json")

	summary, err := ExportStructureHeatmapSummaryJSONFromPath(save.Path(), outputPath, StructureHeatmapOptions{
		MapName:      "Valguero",
		Resolution:   100,
		MinInSection: 2,
	})
	if err != nil {
		t.Fatalf("ExportStructureHeatmapSummaryJSONFromPath() error = %v", err)
	}
	if summary.Total != 2 || summary.NonzeroCells != 1 || summary.Max != 2 || summary.Faults != 0 {
		t.Fatalf("ExportStructureHeatmapSummaryJSONFromPath() summary = %#v, want one cell with two structures", summary)
	}
	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outputPath, err)
	}
	var decoded HeatmapSummary
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded != summary {
		t.Fatalf("decoded summary = %#v, want returned summary %#v", decoded, summary)
	}
}

func TestExportStructureSelectedHeatmapSummaryJSONFromPathWritesSummary(t *testing.T) {
	save := openSyntheticStructureSaveWithFault(t)
	defer save.Close()
	outputPath := filepath.Join(t.TempDir(), "structure-selected-heatmap.json")

	summary, err := ExportStructureSelectedHeatmapSummaryJSONFromPath(save.Path(), outputPath, StructureHeatmapOptions{
		MapName:      "Valguero",
		Resolution:   100,
		MinInSection: 2,
	})
	if err != nil {
		t.Fatalf("ExportStructureSelectedHeatmapSummaryJSONFromPath() error = %v", err)
	}
	if summary.Faults != 0 {
		t.Fatalf("ExportStructureSelectedHeatmapSummaryJSONFromPath() summary = %#v, want selected-property heatmap without full-parse faults", summary)
	}
	fullSummary, err := StructureHeatmapSummaryFromPath(save.Path(), StructureHeatmapOptions{
		MapName:      "Valguero",
		Resolution:   100,
		MinInSection: 1,
	})
	if err != nil {
		t.Fatalf("StructureHeatmapSummaryFromPath() error = %v", err)
	}
	if fullSummary.Faults != 1 {
		t.Fatalf("StructureHeatmapSummaryFromPath() faults = %d, want full-parse helper to report one fault", fullSummary.Faults)
	}
	raw, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", outputPath, err)
	}
	var decoded HeatmapSummary
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; data = %s", err, raw)
	}
	if decoded != summary {
		t.Fatalf("decoded summary = %#v, want returned summary %#v", decoded, summary)
	}
}
