package e2etest

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverProvidedDataRecordsRepresentativeLocalFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{
		"b.arkprofile",
		"a.arkprofile",
		"200.arktribe",
		"100.arktribe",
		"c.arktributetribe",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("fixture"), 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	t.Setenv("ARK_E2E_SAVE_DIR", dir)

	data := DiscoverProvidedData(t)
	if data.ProfileCount != 2 || filepath.Base(data.ProfilePath) != "a.arkprofile" {
		t.Fatalf("profile discovery = count %d path %q", data.ProfileCount, data.ProfilePath)
	}
	if data.TribeCount != 2 || filepath.Base(data.TribePath) != "100.arktribe" {
		t.Fatalf("tribe discovery = count %d path %q", data.TribeCount, data.TribePath)
	}
	if data.TributeCount != 1 || filepath.Base(data.TributePath) != "c.arktributetribe" {
		t.Fatalf("tribute discovery = count %d path %q", data.TributeCount, data.TributePath)
	}
}
