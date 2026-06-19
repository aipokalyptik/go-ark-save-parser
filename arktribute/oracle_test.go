package arktribute

import (
	"os"
	"testing"
)

func TestOracleTributeParsesLocalIndex(t *testing.T) {
	path := os.Getenv("ARK_ORACLE_TRIBUTE")
	if path == "" {
		t.Skip("set ARK_ORACLE_TRIBUTE to a private .arktributetribe path to run oracle integration test")
	}
	data, err := Open(path)
	if err != nil {
		t.Fatalf("Open(oracle tribute) error = %v", err)
	}
	if len(data.PlayerDataIDs)+len(data.TribeDataIDs) == 0 {
		t.Fatalf("oracle tribute parsed zero aggregate IDs")
	}
}
