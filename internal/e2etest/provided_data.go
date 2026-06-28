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
	ClusterPath  string
	ProfilePath  string
	ClusterCount int
	ProfileCount int
	TribePath    string
	TribeCount   int
	TributePath  string
	TributeCount int
}

func DiscoverProvidedData(t *testing.T) ProvidedData {
	t.Helper()
	data := ProvidedData{
		SavePath:    os.Getenv("ARK_E2E_SAVE"),
		Dir:         os.Getenv("ARK_E2E_SAVE_DIR"),
		ClusterPath: os.Getenv("ARK_E2E_CLUSTER"),
	}
	clusterDir := os.Getenv("ARK_E2E_CLUSTER_DIR")
	if clusterDir != "" {
		discoverClusterFiles(t, clusterDir, &data)
	}
	if data.ClusterPath != "" && data.ClusterCount == 0 {
		data.ClusterCount = 1
	}
	if data.Dir == "" {
		return data
	}

	var savePaths []string
	var clusterPaths []string
	var profilePaths []string
	var tribePaths []string
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
			profilePaths = append(profilePaths, path)
			data.ProfileCount++
		case ".arktribe":
			tribePaths = append(tribePaths, path)
			data.TribeCount++
		case ".arktributetribe", ".arktributetribetribe":
			tributePaths = append(tributePaths, path)
		case "":
			if isClusterFileName(filepath.Base(path)) {
				clusterPaths = append(clusterPaths, path)
			}
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
	sort.Strings(profilePaths)
	if len(profilePaths) > 0 {
		data.ProfilePath = profilePaths[0]
	}
	sort.Strings(tribePaths)
	if len(tribePaths) > 0 {
		data.TribePath = tribePaths[0]
	}
	sort.Strings(clusterPaths)
	data.ClusterCount += len(clusterPaths)
	if data.ClusterPath == "" && len(clusterPaths) > 0 {
		data.ClusterPath = clusterPaths[0]
	}
	sort.Strings(tributePaths)
	data.TributeCount = len(tributePaths)
	if len(tributePaths) > 0 {
		data.TributePath = tributePaths[0]
	}
	return data
}

func discoverClusterFiles(t *testing.T, dir string, data *ProvidedData) {
	t.Helper()
	var clusterPaths []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.ToLower(filepath.Ext(path)) != "" || !isClusterFileName(entry.Name()) {
			return nil
		}
		clusterPaths = append(clusterPaths, path)
		return nil
	})
	if err != nil {
		t.Fatalf("discover provided cluster files in %q: %v", dir, err)
	}
	sort.Strings(clusterPaths)
	data.ClusterCount += len(clusterPaths)
	if data.ClusterPath == "" && len(clusterPaths) > 0 {
		data.ClusterPath = clusterPaths[0]
	}
}

func isClusterFileName(name string) bool {
	if strings.HasPrefix(name, "EOS_") && len(name) > len("EOS_") {
		return true
	}
	if len(name) != 32 {
		return false
	}
	for _, r := range name {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

func RedactProvidedPaths(s string) string {
	for _, path := range []string{os.Getenv("ARK_E2E_SAVE"), os.Getenv("ARK_E2E_SAVE_DIR"), os.Getenv("ARK_E2E_CLUSTER"), os.Getenv("ARK_E2E_CLUSTER_DIR")} {
		if path != "" {
			s = strings.ReplaceAll(s, path, "[provided-data]")
		}
	}
	return s
}
