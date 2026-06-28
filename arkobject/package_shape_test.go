package arkobject

import (
	"os"
	"strings"
	"testing"
)

func TestArkObjectEquipmentModelIsSplitByResponsibility(t *testing.T) {
	for _, path := range []string{
		"equipment.go",
		"equipment_stats.go",
		"equipment_defaults.go",
		"equipment_properties.go",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected arkobject equipment model split file %s: %v", path, err)
		}
	}

	contents, err := os.ReadFile("equipment.go")
	if err != nil {
		t.Fatalf("read equipment.go: %v", err)
	}
	if lines := strings.Count(string(contents), "\n"); lines > 180 {
		t.Fatalf("equipment.go has %d lines, want <= 180 after responsibility split", lines)
	}
}
