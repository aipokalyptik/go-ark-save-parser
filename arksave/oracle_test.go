package arksave

import (
	"os"
	"testing"
)

func TestOracleSaveEnumeratesObjects(t *testing.T) {
	path := os.Getenv("ARK_ORACLE_SAVE")
	if path == "" {
		t.Skip("set ARK_ORACLE_SAVE to a private .ark path to run oracle integration test")
	}

	save, err := Open(path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}
	defer save.Close()

	ids, err := save.ObjectIDs()
	if err != nil {
		t.Fatalf("ObjectIDs() error = %v", err)
	}
	if len(ids) == 0 {
		t.Fatalf("ObjectIDs() returned zero objects")
	}
	if save.Context.MapName == "" {
		t.Fatalf("MapName is empty")
	}
}
