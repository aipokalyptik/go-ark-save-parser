package arkapi

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func TestProvidedDataReadOnlyE2E(t *testing.T) {
	data := providedData(t)
	if data.savePath == "" {
		t.Skip("set ARK_E2E_SAVE or ARK_E2E_SAVE_DIR to run provided-data read-only E2E")
	}

	save, err := arksave.Open(data.savePath)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", data.savePath, err)
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

	if data.dir != "" {
		playerAPI, err := NewPlayerFromDirectory(data.dir)
		if err != nil {
			t.Fatalf("NewPlayerFromDirectory(%q) error = %v", data.dir, err)
		}
		players, err := playerAPI.Players()
		if err != nil {
			t.Fatalf("PlayerAPI.Players() error = %v", err)
		}
		if data.profileCount > 0 && len(players) == 0 {
			t.Fatalf("PlayerAPI.Players() returned zero players from %d profiles", data.profileCount)
		}
		tribes, err := playerAPI.TribeDetails()
		if err != nil {
			t.Fatalf("PlayerAPI.TribeDetails() error = %v", err)
		}
		if data.tribeCount > 0 && len(tribes) == 0 {
			t.Fatalf("PlayerAPI.TribeDetails() returned zero tribes from %d tribe files", data.tribeCount)
		}
	}
}

type providedDataSet struct {
	savePath     string
	dir          string
	profileCount int
	tribeCount   int
}

func providedData(t *testing.T) providedDataSet {
	t.Helper()
	data := providedDataSet{savePath: os.Getenv("ARK_E2E_SAVE"), dir: os.Getenv("ARK_E2E_SAVE_DIR")}
	if data.dir == "" {
		return data
	}

	var savePaths []string
	err := filepath.WalkDir(data.dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".ark":
			savePaths = append(savePaths, path)
		case ".arkprofile":
			data.profileCount++
		case ".arktribe":
			data.tribeCount++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("discover provided data files in %q: %v", data.dir, err)
	}
	sort.Strings(savePaths)
	if data.savePath == "" && len(savePaths) > 0 {
		data.savePath = savePaths[0]
	}
	return data
}
