package arkapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func TestProvidedDataReadOnlyE2E(t *testing.T) {
	savePath := providedSavePath(t)
	if savePath == "" {
		t.Skip("set ARK_E2E_SAVE or ARK_E2E_SAVE_DIR to run provided-data read-only E2E")
	}

	save, err := arksave.Open(savePath)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", savePath, err)
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

	if dir := providedSaveDir(); dir != "" {
		playerAPI, err := NewPlayerFromDirectory(dir)
		if err != nil {
			t.Fatalf("NewPlayerFromDirectory(%q) error = %v", dir, err)
		}
		players, err := playerAPI.Players()
		if err != nil {
			t.Fatalf("PlayerAPI.Players() error = %v", err)
		}
		if len(playerAPI.ProfilePaths()) > 0 && len(players) == 0 {
			t.Fatalf("PlayerAPI.Players() returned zero players from %d profiles", len(playerAPI.ProfilePaths()))
		}
		tribes, err := playerAPI.TribeDetails()
		if err != nil {
			t.Fatalf("PlayerAPI.TribeDetails() error = %v", err)
		}
		if len(playerAPI.TribePaths()) > 0 && len(tribes) == 0 {
			t.Fatalf("PlayerAPI.TribeDetails() returned zero tribes from %d tribe files", len(playerAPI.TribePaths()))
		}
	}
}

func providedSavePath(t *testing.T) string {
	t.Helper()
	if path := os.Getenv("ARK_E2E_SAVE"); path != "" {
		return path
	}
	dir := providedSaveDir()
	if dir == "" {
		return ""
	}
	var matches []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".ark") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("discover .ark files in %q: %v", dir, err)
	}
	if len(matches) == 0 {
		return ""
	}
	return matches[0]
}

func providedSaveDir() string {
	if dir := os.Getenv("ARK_E2E_SAVE_DIR"); dir != "" {
		return dir
	}
	return ""
}