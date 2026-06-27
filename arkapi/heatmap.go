package arkapi

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
