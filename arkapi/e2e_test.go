package arkapi

import (
	"encoding/json"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/e2etest"
)

func TestProvidedDataReadOnlyE2E(t *testing.T) {
	data := e2etest.DiscoverProvidedData(t)
	if data.SavePath == "" && data.ClusterPath == "" {
		t.Skip("set ARK_E2E_SAVE or ARK_E2E_SAVE_DIR to run provided-data read-only E2E")
	}

	if data.SavePath != "" {
		save, err := arksave.Open(data.SavePath)
		if err != nil {
			t.Fatalf("Open(%q) error = %v", data.SavePath, err)
		}
		defer save.Close()

		info, err := NewJSON(save).ExportSaveInfo()
		if err != nil {
			t.Fatalf("ExportSaveInfo() error = %v", err)
		}
		if info.MapName == "" {
			t.Fatalf("ExportSaveInfo() MapName is empty")
		}
		if info.ObjectCount == 0 {
			t.Fatalf("ExportSaveInfo() ObjectCount = 0")
		}

		ids, err := NewGeneral(save).ObjectIDs()
		if err != nil {
			t.Fatalf("ObjectIDs() error = %v", err)
		}
		if len(ids) == 0 {
			t.Fatalf("ObjectIDs() returned no objects")
		}

		structures, err := save.ObjectClassInfosWithAnyProperty([]string{"MyInventoryComponent"})
		if err != nil {
			t.Fatalf("ObjectClassInfosWithAnyProperty(MyInventoryComponent) error = %v", err)
		}
		if len(structures) == 0 {
			t.Fatalf("ObjectClassInfosWithAnyProperty(MyInventoryComponent) returned no objects")
		}

		jsonAPI := NewJSON(save)
		for _, domain := range []string{"dinos", "equipment", "stackables"} {
			export, err := jsonAPI.ExportDomain(domain)
			if err != nil {
				t.Fatalf("ExportDomain(%q) error = %v", domain, err)
			}
			if export.Domain != domain {
				t.Fatalf("ExportDomain(%q).Domain = %q", domain, export.Domain)
			}
			if export.Count < 0 {
				t.Fatalf("ExportDomain(%q).Count = %d", domain, export.Count)
			}
			if _, err := json.Marshal(export); err != nil {
				t.Fatalf("json.Marshal(ExportDomain(%q)) error = %v", domain, err)
			}
		}
	}

	if data.Dir != "" {
		playerAPI, err := NewPlayerFromDirectory(data.Dir)
		if err != nil {
			t.Fatalf("NewPlayerFromDirectory(%q) error = %v", data.Dir, err)
		}
		players, err := playerAPI.Players()
		if err != nil {
			t.Fatalf("PlayerAPI.Players() error = %v", err)
		}
		if data.ProfileCount > 0 && len(players) == 0 {
			t.Fatalf("PlayerAPI.Players() returned zero players from %d profiles", data.ProfileCount)
		}
		tribes, err := playerAPI.TribeDetails()
		if err != nil {
			t.Fatalf("PlayerAPI.TribeDetails() error = %v", err)
		}
		if data.TribeCount > 0 && len(tribes) == 0 {
			t.Fatalf("PlayerAPI.TribeDetails() returned zero tribes from %d tribe files", data.TribeCount)
		}
	}

	if data.ClusterPath != "" {
		cluster, err := arkcluster.Open(data.ClusterPath)
		if err != nil {
			t.Fatalf("arkcluster.Open(%q) error = %v", data.ClusterPath, err)
		}
		clusterAPI := NewCluster(cluster)
		summary := clusterAPI.Summary()
		if summary.ID == "" {
			t.Fatalf("ClusterAPI.Summary() ID is empty")
		}
		if len(clusterAPI.ItemsTyped()) != len(cluster.Items) {
			t.Fatalf("ClusterAPI.ItemsTyped() length = %d, want %d", len(clusterAPI.ItemsTyped()), len(cluster.Items))
		}
		if len(clusterAPI.DinosTyped()) != len(cluster.Dinos) {
			t.Fatalf("ClusterAPI.DinosTyped() length = %d, want %d", len(clusterAPI.DinosTyped()), len(cluster.Dinos))
		}
		if summary.ParseErrorCount != clusterAPI.ParseErrorCount() {
			t.Fatalf("ClusterAPI.Summary().ParseErrorCount = %d, want %d", summary.ParseErrorCount, clusterAPI.ParseErrorCount())
		}
	}
}
