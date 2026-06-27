package e2etest

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type ProvidedData struct {
	SavePath     string
	Dir          string
	ProfileCount int
	TribeCount   int
	TributePath  string
	TributeCount int
}

func DiscoverProvidedData(t *testing.T) ProvidedData {
	t.Helper()
	data := ProvidedData{SavePath: os.Getenv("ARK_E2E_SAVE"), Dir: os.Getenv("ARK_E2E_SAVE_DIR")}
	if data.Dir == "" {
		return data
	}

	var savePaths []string
	var tributePaths []string
	err := filepath.WalkDir(data.Dir, func(path string, entry os.DirEntry, err error) error {
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
			data.ProfileCount++
		case ".arktribe":
			data.TribeCount++
		case ".arktributetribe", ".arktributetribetribe":
			tributePaths = append(tributePaths, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("discover provided data files in %q: %v", data.Dir, err)
	}
	sort.Strings(savePaths)
	if data.SavePath == "" && len(savePaths) > 0 {
		data.SavePath = savePaths[0]
	}
	sort.Strings(tributePaths)
	data.TributeCount = len(tributePaths)
	if len(tributePaths) > 0 {
		data.TributePath = tributePaths[0]
	}
	return data
}

func RedactProvidedPaths(s string) string {
	for _, path := range []string{os.Getenv("ARK_E2E_SAVE"), os.Getenv("ARK_E2E_SAVE_DIR")} {
		if path != "" {
			s = strings.ReplaceAll(s, path, "[provided-data]")
		}
	}
	return s
}
