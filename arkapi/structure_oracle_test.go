package arkapi

import (
	"encoding/hex"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func TestOracleStructureParsesWirelessExchangeRefs(t *testing.T) {
	path := os.Getenv("ARK_ORACLE_SAVE")
	if path == "" {
		t.Skip("ARK_ORACLE_SAVE not set")
	}
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}
	defer save.Close()

	id := uuid.MustParse("00143429-3576-aa48-b2de-4462d0209e90")
	object, err := save.ParsedObject(id)
	if err != nil {
		raw, rawErr := save.ObjectBinary(id)
		if rawErr == nil {
			pos := errorPosition(err.Error())
			start := pos - 24
			if start < 0 {
				start = 0
			}
			end := pos + 48
			if end > len(raw) {
				end = len(raw)
			}
			t.Logf("bytes[%d:%d]=%s", start, end, hex.EncodeToString(raw[start:end]))
		}
		t.Fatalf("ParsedObject(%s) error = %v", id, err)
	}
	if _, ok := object.Value("WirelessExchangeRefs"); !ok {
		t.Fatalf("WirelessExchangeRefs property missing from parsed object")
	}
	structure := arkobject.StructureFromObject(object, nil)
	if structure.ID == 0 {
		t.Fatalf("StructureID = 0, want parsed non-zero structure")
	}
}

func TestOracleStructureSelectionDiagnostics(t *testing.T) {
	if os.Getenv("ARK_ORACLE_STRUCTURE_DIAGNOSTICS") == "" {
		t.Skip("ARK_ORACLE_STRUCTURE_DIAGNOSTICS not set")
	}
	path := os.Getenv("ARK_ORACLE_SAVE")
	if path == "" {
		t.Skip("ARK_ORACLE_SAVE not set")
	}
	save, err := arksave.Open(path)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", path, err)
	}
	defer save.Close()

	infos, err := save.ObjectClassInfos()
	if err != nil {
		t.Fatalf("ObjectClassInfos() error = %v", err)
	}
	selected := 0
	for _, info := range infos {
		if isStructureBlueprint(info.ClassName) {
			selected++
		}
	}
	objects, faults, err := save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isStructureBlueprint(info.ClassName)
	})
	if err != nil {
		t.Fatalf("ParsedObjectsWithFaults() error = %v", err)
	}
	withStructureID := 0
	withInventory := 0
	engrams := 0
	for _, info := range objects {
		if _, ok := info.Object.Value("StructureID"); ok {
			withStructureID++
		}
		if _, ok := info.Object.Value("MyInventoryComponent"); ok {
			withInventory++
		}
		if _, ok := info.Object.Value("bIsEngram"); ok {
			engrams++
		}
	}
	nonBlueprintContainers := 0
	for _, info := range infos {
		if isStructureBlueprint(info.ClassName) {
			continue
		}
		if strings.Contains(info.ClassName, "PlayerPawn") ||
			strings.Contains(info.ClassName, "/Dinos/") ||
			strings.Contains(info.ClassName, "Character_BP") {
			continue
		}
		object, err := save.ParsedObject(info.UUID)
		if err != nil {
			continue
		}
		if _, ok := object.Value("MyInventoryComponent"); ok {
			nonBlueprintContainers++
		}
	}
	t.Logf("objects=%d selected=%d parsed=%d faults=%d with_structure_id=%d with_inventory=%d engrams=%d non_blueprint_containers=%d",
		len(infos), selected, len(objects), len(faults), withStructureID, withInventory, engrams, nonBlueprintContainers)
	errorCounts := map[string]int{}
	classCounts := map[string]int{}
	for _, fault := range faults {
		errorCounts[fault.Err.Error()]++
		classCounts[fault.ClassName]++
	}
	logTopCounts(t, "fault error", errorCounts, 10)
	logTopCounts(t, "fault class", classCounts, 10)
	if len(faults) > 0 {
		fault := faults[0]
		raw, err := save.ObjectBinary(fault.UUID)
		if err == nil {
			pos := errorPosition(fault.Err.Error())
			start := pos - 32
			if start < 0 {
				start = 0
			}
			end := pos + 48
			if end > len(raw) {
				end = len(raw)
			}
			t.Logf("first fault uuid=%s class=%q err=%v bytes[%d:%d]=%s", fault.UUID, fault.ClassName, fault.Err, start, end, hex.EncodeToString(raw[start:end]))
		}
	}
}

func logTopCounts(t *testing.T, label string, counts map[string]int, limit int) {
	t.Helper()
	type entry struct {
		name  string
		count int
	}
	entries := make([]entry, 0, len(counts))
	for name, count := range counts {
		entries = append(entries, entry{name: name, count: count})
	}
	sort.Slice(entries, func(i int, j int) bool {
		if entries[i].count == entries[j].count {
			return entries[i].name < entries[j].name
		}
		return entries[i].count > entries[j].count
	})
	if len(entries) > limit {
		entries = entries[:limit]
	}
	for _, entry := range entries {
		t.Logf("%s count=%d value=%q", label, entry.count, entry.name)
	}
}

func errorPosition(message string) int {
	match := regexp.MustCompile(`position ([0-9]+)`).FindStringSubmatch(message)
	if len(match) != 2 {
		return 0
	}
	var out int
	for _, ch := range match[1] {
		out = out*10 + int(ch-'0')
	}
	return out
}
