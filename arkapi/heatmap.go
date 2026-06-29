package arkapi

import (
	"encoding/json"
	"math"
	"os"
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

func heatmapCellFromCoords(lat float64, long float64, resolution int) (int, int, bool) {
	if resolution <= 0 || math.IsNaN(lat) || math.IsNaN(long) || math.IsInf(lat, 0) || math.IsInf(long, 0) {
		return 0, 0, false
	}
	x := int(math.Floor(lat))
	y := int(math.Floor(long))
	if x < 0 || x >= resolution || y < 0 || y >= resolution {
		return 0, 0, false
	}
	return x, y, true
}

func DinoHeatmapSummaryFromPath(savePath string, opts DinoHeatmapOptions) (HeatmapSummary, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return HeatmapSummary{}, err
	}
	defer closeAPI()

	if opts.MapName == "" && api.save.Context != nil {
		opts.MapName = api.save.Context.MapName
	}
	summary, _, err := api.HeatmapSummaryWithFaults(opts)
	if err != nil {
		return HeatmapSummary{}, err
	}
	return summary, nil
}

func ExportDinoHeatmapSummaryJSONFromPath(savePath string, outputPath string, opts DinoHeatmapOptions) (HeatmapSummary, error) {
	summary, err := DinoHeatmapSummaryFromPath(savePath, opts)
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

func StructureHeatmapSummaryFromPath(savePath string, opts StructureHeatmapOptions) (HeatmapSummary, error) {
	api, closeAPI, err := NewStructureFromPath(savePath)
	if err != nil {
		return HeatmapSummary{}, err
	}
	defer closeAPI()

	if opts.MapName == "" && api.save.Context != nil {
		opts.MapName = api.save.Context.MapName
	}
	summary, _, err := api.HeatmapSummaryWithFaults(opts)
	if err != nil {
		return HeatmapSummary{}, err
	}
	return summary, nil
}

func StructureSelectedHeatmapSummaryFromPath(savePath string, opts StructureHeatmapOptions) (HeatmapSummary, error) {
	api, closeAPI, err := NewStructureFromPath(savePath)
	if err != nil {
		return HeatmapSummary{}, err
	}
	defer closeAPI()

	if opts.MapName == "" && api.save.Context != nil {
		opts.MapName = api.save.Context.MapName
	}
	summary, _, err := api.SelectedHeatmapSummaryWithFaults(opts)
	if err != nil {
		return HeatmapSummary{}, err
	}
	return summary, nil
}

func ExportStructureHeatmapSummaryJSONFromPath(savePath string, outputPath string, opts StructureHeatmapOptions) (HeatmapSummary, error) {
	summary, err := StructureHeatmapSummaryFromPath(savePath, opts)
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

func ExportStructureSelectedHeatmapSummaryJSONFromPath(savePath string, outputPath string, opts StructureHeatmapOptions) (HeatmapSummary, error) {
	summary, err := StructureSelectedHeatmapSummaryFromPath(savePath, opts)
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
