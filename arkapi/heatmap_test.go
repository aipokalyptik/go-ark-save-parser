package arkapi

import (
	"encoding/json"
	"os"
	"path/filepath"
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
