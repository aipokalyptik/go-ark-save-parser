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

func TestArkObjectDinoModelIsSplitByResponsibility(t *testing.T) {
	for _, path := range []string{
		"dino.go",
		"dino_colors.go",
		"dino_lineage.go",
		"dino_traits.go",
		"object_property_values.go",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected arkobject dino model split file %s: %v", path, err)
		}
	}

	contents, err := os.ReadFile("dino.go")
	if err != nil {
		t.Fatalf("read dino.go: %v", err)
	}
	if lines := strings.Count(string(contents), "\n"); lines > 160 {
		t.Fatalf("dino.go has %d lines, want <= 160 after responsibility split", lines)
	}
}

func TestArkObjectPlayerAndTribeModelsAreSplitByResponsibility(t *testing.T) {
	for _, path := range []string{
		"player.go",
		"tribe.go",
		"profile_property_values.go",
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected arkobject player/tribe split file %s: %v", path, err)
		}
	}

	for _, path := range []string{"player.go", "tribe.go"} {
		contents, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if lines := strings.Count(string(contents), "\n"); lines > 140 {
			t.Fatalf("%s has %d lines, want <= 140 after responsibility split", path, lines)
		}
	}
}
