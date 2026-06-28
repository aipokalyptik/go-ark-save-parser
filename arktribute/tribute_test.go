package arktribute

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/internal/safefile"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
)

func TestParseReadsLocalTributeIDLists(t *testing.T) {
	raw := testfixtures.TributeBytes(t, []uint64{11, 22}, []uint64{33})

	playerIDs, tribeIDs, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !sameUint64s(playerIDs, []uint64{11, 22}) || !sameUint64s(tribeIDs, []uint64{33}) {
		t.Fatalf("Parse() playerIDs=%v tribeIDs=%v", playerIDs, tribeIDs)
	}
}

func TestParseRejectsMalformedLocalTributeData(t *testing.T) {
	tests := map[string][]byte{
		"short":    {1, 0, 0},
		"negative": {0xff, 0xff, 0xff, 0xff},
		"trailing": append(testfixtures.TributeBytes(t, nil, nil), 0),
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			if _, _, err := Parse(raw); err == nil {
				t.Fatalf("Parse(%s) error = nil, want malformed data error", name)
			}
		})
	}
}

func TestDiscoverFindsLocalTributeFilesOnly(t *testing.T) {
	dir := t.TempDir()
	tributePath := filepath.Join(dir, "abc.arktributetribe")
	testfixtures.WriteTributeFile(t, tributePath, []uint64{1}, []uint64{2})
	testfixtures.WriteTributeFile(t, filepath.Join(dir, "def.arktributetribetribe"), nil, nil)
	if err := os.WriteFile(filepath.Join(dir, "EOS_abc123"), []byte("cluster"), 0o600); err != nil {
		t.Fatalf("write cluster placeholder: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "123.arkprofile"), []byte("profile"), 0o600); err != nil {
		t.Fatalf("write profile placeholder: %v", err)
	}

	files, err := Discover(dir)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(files) != 2 || files[0].Path != tributePath || files[0].ID != "abc" {
		t.Fatalf("Discover() = %#v", files)
	}
}

func TestOpenLoadsLocalTributeFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.arktributetribe")
	testfixtures.WriteTributeFile(t, path, []uint64{11, 22}, []uint64{33})

	data, err := Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if data.ID != "abc" || data.Path != path {
		t.Fatalf("Data identity = %#v", data)
	}
	if !sameUint64s(data.PlayerDataIDs, []uint64{11, 22}) || !sameUint64s(data.TribeDataIDs, []uint64{33}) {
		t.Fatalf("Open() PlayerDataIDs=%v TribeDataIDs=%v", data.PlayerDataIDs, data.TribeDataIDs)
	}
}

func TestOpenRejectsLocalTributeFileAboveConfiguredLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "abc.arktributetribe")
	testfixtures.WriteSparseFile(t, path, 1024)

	_, err := OpenWithOptions(path, Options{MaxFileSize: 16})
	if !errors.Is(err, safefile.ErrFileTooLarge) {
		t.Fatalf("OpenWithOptions() error = %v, want ErrFileTooLarge", err)
	}
}

func TestOpenDirectoryWithFaultsKeepsValidTributeFiles(t *testing.T) {
	dir := t.TempDir()
	validPath := filepath.Join(dir, "abc.arktributetribe")
	brokenPath := filepath.Join(dir, "broken.arktributetribe")
	testfixtures.WriteTributeFile(t, validPath, []uint64{11}, []uint64{22})
	if err := os.WriteFile(brokenPath, []byte("not a tribute index"), 0o600); err != nil {
		t.Fatalf("write broken tribute file: %v", err)
	}

	entries, faults, err := OpenDirectoryWithFaults(dir)
	if err != nil {
		t.Fatalf("OpenDirectoryWithFaults() error = %v", err)
	}
	if len(entries) != 1 || entries[0].ID != "abc" {
		t.Fatalf("OpenDirectoryWithFaults() entries = %#v, want valid tribute only", entries)
	}
	if len(faults) != 1 || faults[0].Path != brokenPath || faults[0].Err == nil {
		t.Fatalf("OpenDirectoryWithFaults() faults = %#v, want broken path fault", faults)
	}
}

func sameUint64s(got []uint64, want []uint64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
