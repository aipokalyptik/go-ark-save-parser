package arkapi

import "testing"

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
