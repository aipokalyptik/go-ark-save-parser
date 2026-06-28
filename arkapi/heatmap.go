package arkapi

import (
	"encoding/json"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

type HeatmapSummary struct {
	Resolution   int `json:"resolution"`
	NonzeroCells int `json:"nonzero_cells"`
	Total        int `json:"total"`
	Max          int `json:"max"`
	Faults       int `json:"faults"`
}

func SummarizeHeatmap(heatmap [][]int, faults int) HeatmapSummary {
	summary := HeatmapSummary{Resolution: len(heatmap), Faults: faults}
	for _, row := range heatmap {
		for _, value := range row {
			if value == 0 {
				continue
			}
			summary.NonzeroCells++
			summary.Total += value
			if value > summary.Max {
				summary.Max = value
			}
		}
	}
	return summary
}

func ExportDinoHeatmapSummaryJSONFromPath(savePath string, outputPath string, opts DinoHeatmapOptions) (HeatmapSummary, error) {
	save, err := arksave.Open(savePath)
	if err != nil {
		return HeatmapSummary{}, err
	}
	defer save.Close()

	if opts.MapName == "" && save.Context != nil {
		opts.MapName = save.Context.MapName
	}
	summary, _, err := NewDino(save).HeatmapSummaryWithFaults(opts)
	if err != nil {
		return HeatmapSummary{}, err
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		return HeatmapSummary{}, err
	}
	if err := os.WriteFile(outputPath, append(data, '\n'), 0o644); err != nil {
		return HeatmapSummary{}, err
	}
	return summary, nil
}
