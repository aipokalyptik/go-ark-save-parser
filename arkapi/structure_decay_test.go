package arkapi

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
	"github.com/google/uuid"
)

func TestStructureDemolishableReportFiltersAndSortsEligibleStructures(t *testing.T) {
	save := openSyntheticDemolishableStructureSave(t)
	defer save.Close()

	periodsPath := filepath.Join(t.TempDir(), "periods.json")
	if err := os.WriteFile(periodsPath, []byte(`{"substring":{"Stone":100,"Wood":500}}`), 0o600); err != nil {
		t.Fatalf("write periods override: %v", err)
	}
	settingsPath := filepath.Join(t.TempDir(), "GameUserSettings.ini")
	if err := os.WriteFile(settingsPath, []byte("[ServerSettings]\nPvEStructureDecayPeriodMultiplier=2.0\n"), 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	report, faults, err := NewStructure(save).DemolishableReport(StructureDemolishableOptions{
		MapName:              "Valguero",
		GameUserSettingsPath: settingsPath,
		DecayPeriodsPath:     periodsPath,
	})
	if err != nil {
		t.Fatalf("DemolishableReport() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("DemolishableReport() faults = %#v, want none", faults)
	}
	if report.Summary.TotalStructures != 5 || report.Summary.EligibleStructures != 2 || report.Summary.UnknownTimestampStructures != 1 {
		t.Fatalf("DemolishableReport() summary = %#v, want total 5 eligible 2 unknown timestamp 1", report.Summary)
	}
	if report.Summary.DecayMultiplier != 2 {
		t.Fatalf("DecayMultiplier = %f, want 2", report.Summary.DecayMultiplier)
	}
	if len(report.Structures) != 2 {
		t.Fatalf("eligible structures length = %d, want 2: %#v", len(report.Structures), report.Structures)
	}
	if report.Structures[0].Owner.SortKey != "Alpha" || report.Structures[0].StructureID != 101 || report.Structures[0].RemainingSeconds != 0 {
		t.Fatalf("first structure = %#v, want Alpha stone wall eligible with zero remaining", report.Structures[0])
	}
	if report.Structures[1].Owner.SortKey != "Beta" || report.Structures[1].StructureID != 103 || report.Structures[1].ElapsedSeconds != 1100 {
		t.Fatalf("second structure = %#v, want Beta wood wall eligible after owner/location sort", report.Structures[1])
	}
	if report.Structures[0].Location == nil || report.Structures[1].Location == nil {
		t.Fatalf("eligible locations = %#v / %#v, want map locations", report.Structures[0].Location, report.Structures[1].Location)
	}
}

func TestStructureDemolishableReportFlagMultiplierOverridesSettings(t *testing.T) {
	save := openSyntheticDemolishableStructureSave(t)
	defer save.Close()

	settingsPath := filepath.Join(t.TempDir(), "GameUserSettings.ini")
	if err := os.WriteFile(settingsPath, []byte("[ServerSettings]\nPvEStructureDecayPeriodMultiplier=99.0\n"), 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	report, _, err := NewStructure(save).DemolishableReport(StructureDemolishableOptions{
		MapName:              "Valguero",
		GameUserSettingsPath: settingsPath,
		DecayMultiplier:      1,
		DecayPeriods: StructureDecayPeriods{
			Substring: map[string]float64{"Stone": 100, "Wood": 500},
		},
	})
	if err != nil {
		t.Fatalf("DemolishableReport() error = %v", err)
	}
	if report.Summary.DecayMultiplier != 1 {
		t.Fatalf("DecayMultiplier = %f, want explicit override 1", report.Summary.DecayMultiplier)
	}
	if report.Summary.EligibleStructures != 3 {
		t.Fatalf("EligibleStructures = %d, want 3 with multiplier override", report.Summary.EligibleStructures)
	}
}

func TestStructureDemolishableReportGroupsEligibleBases(t *testing.T) {
	save := openSyntheticDemolishableStructureSave(t)
	defer save.Close()

	report, _, err := NewStructure(save).DemolishableReport(StructureDemolishableOptions{
		MapName: "Valguero",
		DecayPeriods: StructureDecayPeriods{
			Substring: map[string]float64{"Stone": 100, "Wood": 500},
		},
		GroupBases: true,
	})
	if err != nil {
		t.Fatalf("DemolishableReport(group bases) error = %v", err)
	}
	if len(report.Bases) != 2 {
		t.Fatalf("Bases length = %d, want 2: %#v", len(report.Bases), report.Bases)
	}
	if report.Bases[0].Owner.SortKey != "Alpha" || report.Bases[0].EligibleStructures != 2 || report.Bases[0].TotalStructures != 2 {
		t.Fatalf("first base = %#v, want Alpha connected stone base with two eligible structures", report.Bases[0])
	}
	if report.Bases[1].Owner.SortKey != "Beta" || report.Bases[1].EligibleStructures != 1 {
		t.Fatalf("second base = %#v, want Beta eligible wood structure", report.Bases[1])
	}
}

func TestStructureDecayClassifiesDefaultPeriodsAndUnknownFallback(t *testing.T) {
	tests := []struct {
		name   string
		class  string
		tier   string
		period float64
		source string
	}{
		{name: "thatch", class: "Blueprint'/Game/Structures/Thatch/PrimalStructureWall_Thatch.PrimalStructureWall_Thatch_C'", tier: "thatch", period: 4 * 24 * 60 * 60, source: "built_in"},
		{name: "wood", class: "Blueprint'/Game/Structures/Wood/PrimalStructureWall_Wood.PrimalStructureWall_Wood_C'", tier: "wood", period: 8 * 24 * 60 * 60, source: "built_in"},
		{name: "stone", class: "Blueprint'/Game/Structures/Stone/PrimalStructure_Wall_Stone.PrimalStructure_Wall_Stone_C'", tier: "stone", period: 12 * 24 * 60 * 60, source: "built_in"},
		{name: "metal", class: "Blueprint'/Game/Structures/Metal/PrimalStructureWall_Metal.PrimalStructureWall_Metal_C'", tier: "metal", period: 16 * 24 * 60 * 60, source: "built_in"},
		{name: "greenhouse", class: "Blueprint'/Game/Structures/Greenhouse/PrimalStructure_GreenhouseWall.PrimalStructure_GreenhouseWall_C'", tier: "greenhouse", period: 16 * 24 * 60 * 60, source: "built_in"},
		{name: "vault", class: "Blueprint'/Game/Structures/Storage/PrimalStructureItemContainer_StorageBox_Huge.PrimalStructureItemContainer_StorageBox_Huge_C'", tier: "vault", period: 16 * 24 * 60 * 60, source: "built_in"},
		{name: "tek", class: "Blueprint'/Game/Structures/Tek/PrimalStructureWall_Tek.PrimalStructureWall_Tek_C'", tier: "tek", period: 20 * 24 * 60 * 60, source: "built_in"},
		{name: "cryofridge", class: "Blueprint'/Game/Extinction/CoreBlueprints/Items/Structures/CryoFridge/PrimalStructure_CryoFridge.PrimalStructure_CryoFridge_C'", tier: "tek_utility", period: 40 * 24 * 60 * 60, source: "built_in"},
		{name: "unknown", class: "Blueprint'/Game/Mods/Custom/PrimalStructure_Custom.PrimalStructure_Custom_C'", tier: "unknown", period: 40 * 24 * 60 * 60, source: "unknown_fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyStructureDecay(tt.class, StructureDecayPeriods{})
			if got.Tier != tt.tier || got.PeriodSeconds != tt.period || got.Source != tt.source {
				t.Fatalf("ClassifyStructureDecay(%q) = %#v, want tier %s period %.0f source %s", tt.class, got, tt.tier, tt.period, tt.source)
			}
		})
	}
}

func openSyntheticDemolishableStructureSave(t *testing.T) *arksave.Save {
	t.Helper()

	alphaOldID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000001")
	alphaLinkedID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000002")
	alphaFreshID := uuid.MustParse("aaaaaaaa-0000-0000-0000-000000000003")
	betaOldID := uuid.MustParse("bbbbbbbb-0000-0000-0000-000000000001")
	unknownTimeID := uuid.MustParse("cccccccc-0000-0000-0000-000000000001")
	alphaLoc := arkobject.MapCoords{Lat: 10, Long: 20}.AsActorTransform("Valguero")
	alphaLinkedLoc := arkobject.MapCoords{Lat: 10.2, Long: 20.2}.AsActorTransform("Valguero")
	alphaFreshLoc := arkobject.MapCoords{Lat: 10.4, Long: 20.4}.AsActorTransform("Valguero")
	betaLoc := arkobject.MapCoords{Lat: 60, Long: 70}.AsActorTransform("Valguero")
	unknownLoc := arkobject.MapCoords{Lat: 80, Long: 10}.AsActorTransform("Valguero")

	return openSyntheticSaveWith(t, "structures.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransformsFor(map[uuid.UUID][3]float64{
			alphaOldID:    {alphaLoc.X, alphaLoc.Y, alphaLoc.Z},
			alphaLinkedID: {alphaLinkedLoc.X, alphaLinkedLoc.Y, alphaLinkedLoc.Z},
			alphaFreshID:  {alphaFreshLoc.X, alphaFreshLoc.Y, alphaFreshLoc.Z},
			betaOldID:     {betaLoc.X, betaLoc.Y, betaLoc.Z},
			unknownTimeID: {unknownLoc.X, unknownLoc.Y, unknownLoc.Z},
		}),
	}, map[uuid.UUID][]byte{
		alphaOldID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			StructureID:          101,
			TribeID:              555,
			OwnerName:            "Alpha",
			LastEnterStasisTime:  1034.5,
			LinkedStructureUUIDs: []uuid.UUID{alphaLinkedID},
		}),
		alphaLinkedID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			StructureID:          102,
			TribeID:              555,
			OwnerName:            "Alpha",
			LastEnterStasisTime:  1080,
			LinkedStructureUUIDs: []uuid.UUID{alphaOldID},
		}),
		alphaFreshID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			StructureID:         104,
			TribeID:             555,
			OwnerName:           "Alpha",
			LastEnterStasisTime: 1200,
		}),
		betaOldID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			ClassID:             0x10000060,
			StructureID:         103,
			TribeID:             777,
			OwnerName:           "Beta",
			LastEnterStasisTime: 134.5,
		}),
		unknownTimeID: testfixtures.StructureGameObjectBytes(testfixtures.StructureGameObjectOptions{
			StructureID: 105,
			TribeID:     888,
			OwnerName:   "Gamma",
		}),
	})
}
