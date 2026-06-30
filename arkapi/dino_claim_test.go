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

func TestDinoClaimableReportFiltersAndSortsClaimableDinos(t *testing.T) {
	save := openSyntheticClaimableDinoSave(t)
	defer save.Close()

	settingsPath := filepath.Join(t.TempDir(), "GameUserSettings.ini")
	if err := os.WriteFile(settingsPath, []byte("[ServerSettings]\nPvEDinoDecayPeriodMultiplier=2.0\n"), 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	report, faults, err := NewDino(save).ClaimableReport(DinoClaimableOptions{
		MapName:              "Valguero",
		GameUserSettingsPath: settingsPath,
		ClaimPeriodSeconds:   100,
	})
	if err != nil {
		t.Fatalf("ClaimableReport() error = %v", err)
	}
	if len(faults) != 0 {
		t.Fatalf("ClaimableReport() faults = %#v, want none", faults)
	}
	if report.Summary.TotalDinos != 9 || report.Summary.OwnedDinos != 6 || report.Summary.ClaimableDinos != 4 || report.Summary.UnknownTimestampDinos != 1 || report.Summary.SystemTeamDinos != 3 || report.Summary.WildSystemDinos != 1 || report.Summary.UnclaimedDinos != 1 || report.Summary.AbandonedDinos != 1 {
		t.Fatalf("ClaimableReport() summary = %#v, want total 9 owned 6 claimable 4 unknown 1 system 3 split by category", report.Summary)
	}
	if report.Summary.ClaimMultiplier != 2 {
		t.Fatalf("ClaimMultiplier = %f, want 2", report.Summary.ClaimMultiplier)
	}
	if len(report.Dinos) != 4 {
		t.Fatalf("claimable dinos length = %d, want 4: %#v", len(report.Dinos), report.Dinos)
	}
	if report.Dinos[0].Owner.SortKey != "Alpha" || report.Dinos[0].DinoID1 != 1001 || report.Dinos[0].RemainingSeconds != 0 {
		t.Fatalf("first dino = %#v, want Alpha raptor claimable with zero remaining", report.Dinos[0])
	}
	if report.Dinos[1].Owner.SortKey != "Beta" || report.Dinos[1].DinoID1 != 3001 || report.Dinos[1].ElapsedSeconds != 1100 {
		t.Fatalf("second dino = %#v, want Beta fallback dino sorted after Alpha", report.Dinos[1])
	}
	if report.Dinos[0].ClaimReferenceSource != "last_in_ally_range_time_serialized" || report.Dinos[0].ClaimReferenceTime != 1034.5 {
		t.Fatalf("first dino claim clock = source %q time %f, want ally-range 1034.5", report.Dinos[0].ClaimReferenceSource, report.Dinos[0].ClaimReferenceTime)
	}
	if report.Dinos[1].ClaimReferenceSource != "tamed_time_stamp" || report.Dinos[1].ClaimReferenceTime != 134.5 {
		t.Fatalf("second dino claim clock = source %q time %f, want tamed timestamp fallback 134.5", report.Dinos[1].ClaimReferenceSource, report.Dinos[1].ClaimReferenceTime)
	}
}

func TestDinoClaimableReportFlagMultiplierOverridesSettings(t *testing.T) {
	save := openSyntheticClaimableDinoSave(t)
	defer save.Close()

	settingsPath := filepath.Join(t.TempDir(), "GameUserSettings.ini")
	if err := os.WriteFile(settingsPath, []byte("[ServerSettings]\nPvEDinoDecayPeriodMultiplier=99.0\n"), 0o600); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	report, _, err := NewDino(save).ClaimableReport(DinoClaimableOptions{
		MapName:              "Valguero",
		GameUserSettingsPath: settingsPath,
		ClaimMultiplier:      1,
		ClaimPeriodSeconds:   100,
	})
	if err != nil {
		t.Fatalf("ClaimableReport() error = %v", err)
	}
	if report.Summary.ClaimMultiplier != 1 {
		t.Fatalf("ClaimMultiplier = %f, want explicit override 1", report.Summary.ClaimMultiplier)
	}
	if report.Summary.ClaimableDinos != 5 {
		t.Fatalf("ClaimableDinos = %d, want 5 with multiplier override", report.Summary.ClaimableDinos)
	}
}

func TestDinoClaimableReportUsesSelectedOwnershipAndAllyRangeWithoutTamedTimestamp(t *testing.T) {
	save := openSyntheticClaimableDinoSave(t)
	defer save.Close()

	report, _, err := NewDino(save).ClaimableReport(DinoClaimableOptions{
		MapName:            "Valguero",
		ClaimMultiplier:    1,
		ClaimPeriodSeconds: 100,
	})
	if err != nil {
		t.Fatalf("ClaimableReport() error = %v", err)
	}
	var found *DinoClaimableRow
	for i := range report.Dinos {
		if report.Dinos[i].DinoID1 == 6001 {
			found = &report.Dinos[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("claimable report did not include owned ally-range dino without TamedTimeStamp: %#v", report)
	}
	if found.ShortName != "Raptor" || found.TamedName != "NoTimestamp" {
		t.Fatalf("claimable dino identity = short %q name %q, want Raptor / NoTimestamp", found.ShortName, found.TamedName)
	}
	if found.Owner.TribeName != "Delta" || found.Owner.TargetTeam != 1000000999 {
		t.Fatalf("claimable dino owner = %#v, want Delta target team 1000000999", found.Owner)
	}
	if found.Location == nil || found.Location.Lat == 0 || found.Location.Long == 0 {
		t.Fatalf("claimable dino location = %#v, want map coordinates", found.Location)
	}
	if found.ClaimReferenceSource != "last_in_ally_range_time_serialized" || found.ClaimReferenceTime != 1000 {
		t.Fatalf("claim clock = source %q time %f, want ally-range 1000", found.ClaimReferenceSource, found.ClaimReferenceTime)
	}
}

func TestDinoClaimableReportUsesLastInAllyRangeSerializedAlias(t *testing.T) {
	save := openSyntheticClaimableDinoSave(t)
	defer save.Close()

	report, _, err := NewDino(save).ClaimableReport(DinoClaimableOptions{
		MapName:            "Valguero",
		ClaimMultiplier:    1,
		ClaimPeriodSeconds: 100,
	})
	if err != nil {
		t.Fatalf("ClaimableReport() error = %v", err)
	}
	var found *DinoClaimableRow
	for i := range report.Dinos {
		if report.Dinos[i].DinoID1 == 7001 {
			found = &report.Dinos[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("claimable report did not include dino using LastInAllyRangeSerialized alias: %#v", report)
	}
	if found.ClaimReferenceSource != "last_in_ally_range_serialized" || found.ClaimReferenceTime != 950 {
		t.Fatalf("claim clock = source %q time %f, want LastInAllyRangeSerialized alias 950", found.ClaimReferenceSource, found.ClaimReferenceTime)
	}
	if found.LastInAllyRangeSerialized != 950 {
		t.Fatalf("LastInAllyRangeSerialized = %f, want 950", found.LastInAllyRangeSerialized)
	}
}

func TestDinoClaimableReportExcludesReservedSystemTeams(t *testing.T) {
	tests := []struct {
		name string
		team int32
		want bool
	}{
		{name: "wild low team", team: 5, want: false},
		{name: "last non-player team", team: 49_999, want: false},
		{name: "first player team", team: 50_000, want: true},
		{name: "tribe team", team: 1_000_000_000, want: true},
		{name: "unclaimed breeding team", team: 2_000_000_000, want: false},
		{name: "abandoned team", team: -2147483648, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dino := arkobject.Dino{
				Owner: arkobject.DinoOwner{
					TribeName:  "Named",
					TargetTeam: tt.team,
				},
			}
			if got := isOwnedClaimableCandidate(dino); got != tt.want {
				t.Fatalf("isOwnedClaimableCandidate(TargetTeam=%d) = %v, want %v", tt.team, got, tt.want)
			}
		})
	}
}

func TestDinoClaimableReportCanIncludeSystemTeamCategories(t *testing.T) {
	save := openSyntheticClaimableDinoSave(t)
	defer save.Close()

	report, _, err := NewDino(save).ClaimableReport(DinoClaimableOptions{
		MapName:               "Valguero",
		ClaimMultiplier:       1,
		ClaimPeriodSeconds:    100,
		IncludeWildDinos:      true,
		IncludeUnclaimedDinos: true,
		IncludeAbandonedDinos: true,
		IncludeIneligible:     true,
	})
	if err != nil {
		t.Fatalf("ClaimableReport() error = %v", err)
	}
	if report.Summary.IncludedSystemDinos != 3 || report.Summary.SystemTeamDinos != 3 {
		t.Fatalf("system summary = %#v, want 3 included system dinos and 3 total system dinos", report.Summary)
	}
	wantCategories := map[uint32]string{
		4001: "wild_system",
		8001: "unclaimed_bred",
		9001: "abandoned",
	}
	for dinoID, category := range wantCategories {
		row := findDinoClaimableRow(report.Dinos, dinoID)
		if row == nil {
			t.Fatalf("included system report missing dino %d in category %s: %#v", dinoID, category, report.Dinos)
		}
		if row.OwnershipCategory != category {
			t.Fatalf("dino %d ownership category = %q, want %q", dinoID, row.OwnershipCategory, category)
		}
	}
}

func TestDinoClaimableReportBredAndUnclaimedFlagsAreAliases(t *testing.T) {
	for _, tt := range []struct {
		name string
		opts DinoClaimableOptions
	}{
		{name: "bred", opts: DinoClaimableOptions{IncludeBredDinos: true}},
		{name: "unclaimed", opts: DinoClaimableOptions{IncludeUnclaimedDinos: true}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			save := openSyntheticClaimableDinoSave(t)
			defer save.Close()

			opts := tt.opts
			opts.MapName = "Valguero"
			opts.ClaimMultiplier = 1
			opts.ClaimPeriodSeconds = 100
			report, _, err := NewDino(save).ClaimableReport(opts)
			if err != nil {
				t.Fatalf("ClaimableReport() error = %v", err)
			}
			row := findDinoClaimableRow(report.Dinos, 8001)
			if row == nil || row.OwnershipCategory != "unclaimed_bred" {
				t.Fatalf("unclaimed/bred row = %#v in report %#v, want included unclaimed_bred dino", row, report.Dinos)
			}
		})
	}
}

func TestDinoClaimableFieldDebugCountsCandidateProperties(t *testing.T) {
	save := openSyntheticClaimableDinoSave(t)
	defer save.Close()

	debug, err := NewDino(save).ClaimableFieldDebug()
	if err != nil {
		t.Fatalf("ClaimableFieldDebug() error = %v", err)
	}
	if debug.TotalDinoCandidates != 9 || debug.FaultCount != 0 {
		t.Fatalf("ClaimableFieldDebug() summary = %#v, want 9 candidates and no faults", debug)
	}
	if debug.FieldCounts["LastInAllyRangeTimeSerialized"] != 3 {
		t.Fatalf("LastInAllyRangeTimeSerialized count = %d, want 3: %#v", debug.FieldCounts["LastInAllyRangeTimeSerialized"], debug.FieldCounts)
	}
	if debug.FieldCounts["TamedTimeStamp"] != 4 {
		t.Fatalf("TamedTimeStamp count = %d, want 4: %#v", debug.FieldCounts["TamedTimeStamp"], debug.FieldCounts)
	}
	if debug.FieldCounts["TargetingTeam"] != 9 {
		t.Fatalf("TargetingTeam count = %d, want 9: %#v", debug.FieldCounts["TargetingTeam"], debug.FieldCounts)
	}
	if debug.FieldCounts["LastInAllyRangeSerialized"] != 4 {
		t.Fatalf("LastInAllyRangeSerialized count = %d, want 4: %#v", debug.FieldCounts["LastInAllyRangeSerialized"], debug.FieldCounts)
	}
}

func findDinoClaimableRow(rows []DinoClaimableRow, dinoID uint32) *DinoClaimableRow {
	for i := range rows {
		if rows[i].DinoID1 == dinoID {
			return &rows[i]
		}
	}
	return nil
}

func openSyntheticClaimableDinoSave(t *testing.T) *arksave.Save {
	t.Helper()

	alphaOldID := uuid.MustParse("11111111-0000-0000-0000-000000000001")
	alphaFreshID := uuid.MustParse("11111111-0000-0000-0000-000000000002")
	betaOldID := uuid.MustParse("22222222-0000-0000-0000-000000000001")
	wildID := uuid.MustParse("33333333-0000-0000-0000-000000000001")
	unknownTimeID := uuid.MustParse("44444444-0000-0000-0000-000000000001")
	noTamedTimestampID := uuid.MustParse("55555555-0000-0000-0000-000000000001")
	aliasTimerID := uuid.MustParse("66666666-0000-0000-0000-000000000001")
	unclaimedID := uuid.MustParse("77777777-0000-0000-0000-000000000001")
	abandonedID := uuid.MustParse("88888888-0000-0000-0000-000000000001")
	alphaLoc := arkobject.MapCoords{Lat: 10, Long: 20}.AsActorTransform("Valguero")
	alphaFreshLoc := arkobject.MapCoords{Lat: 10.4, Long: 20.4}.AsActorTransform("Valguero")
	betaLoc := arkobject.MapCoords{Lat: 60, Long: 70}.AsActorTransform("Valguero")
	wildLoc := arkobject.MapCoords{Lat: 5, Long: 5}.AsActorTransform("Valguero")
	unknownLoc := arkobject.MapCoords{Lat: 80, Long: 10}.AsActorTransform("Valguero")
	noTamedTimestampLoc := arkobject.MapCoords{Lat: 30, Long: 40}.AsActorTransform("Valguero")
	aliasTimerLoc := arkobject.MapCoords{Lat: 35, Long: 45}.AsActorTransform("Valguero")
	unclaimedLoc := arkobject.MapCoords{Lat: 6, Long: 7}.AsActorTransform("Valguero")
	abandonedLoc := arkobject.MapCoords{Lat: 8, Long: 9}.AsActorTransform("Valguero")

	return openSyntheticSaveWith(t, "claimable-dinos.ark", map[string][]byte{
		"ActorTransforms": syntheticStructureActorTransformsFor(map[uuid.UUID][3]float64{
			alphaOldID:         {alphaLoc.X, alphaLoc.Y, alphaLoc.Z},
			alphaFreshID:       {alphaFreshLoc.X, alphaFreshLoc.Y, alphaFreshLoc.Z},
			betaOldID:          {betaLoc.X, betaLoc.Y, betaLoc.Z},
			wildID:             {wildLoc.X, wildLoc.Y, wildLoc.Z},
			unknownTimeID:      {unknownLoc.X, unknownLoc.Y, unknownLoc.Z},
			noTamedTimestampID: {noTamedTimestampLoc.X, noTamedTimestampLoc.Y, noTamedTimestampLoc.Z},
			aliasTimerID:       {aliasTimerLoc.X, aliasTimerLoc.Y, aliasTimerLoc.Z},
			unclaimedID:        {unclaimedLoc.X, unclaimedLoc.Y, unclaimedLoc.Z},
			abandonedID:        {abandonedLoc.X, abandonedLoc.Y, abandonedLoc.Z},
		}),
	}, map[uuid.UUID][]byte{
		alphaOldID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                           1001,
			ID2:                           2001,
			Tamed:                         true,
			TamedTimestamp:                1000,
			LastInAllyRangeTimeSerialized: 1034.5,
			TribeName:                     "Alpha",
			TamingTeamID:                  1000000555,
			TamerString:                   "Alpha",
			OwningPlayerName:              "Alice",
			OwningPlayerID:                42,
			TargetingTeam:                 1000000555,
		}),
		alphaFreshID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                           2001,
			ID2:                           2002,
			Tamed:                         true,
			TamedTimestamp:                900,
			LastInAllyRangeTimeSerialized: 1100,
			TribeName:                     "Alpha",
			TamingTeamID:                  1000000555,
			TamerString:                   "Alpha",
			OwningPlayerName:              "Alice",
			OwningPlayerID:                42,
			TargetingTeam:                 1000000555,
		}),
		betaOldID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:            3001,
			ID2:            3002,
			Tamed:          true,
			TamedTimestamp: 134.5,
			TribeName:      "Beta",
			TamingTeamID:   1000000777,
			TamerString:    "Beta",
			OwningPlayerID: 99,
			TargetingTeam:  1000000777,
		}),
		wildID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                       4001,
			ID2:                       4002,
			LastInAllyRangeSerialized: 10,
			TargetingTeam:             5,
		}),
		unknownTimeID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                          5001,
			ID2:                          5002,
			Tamed:                        true,
			DisableDefaultTamedTimestamp: true,
			TribeName:                    "Gamma",
			TamingTeamID:                 1000000888,
			TamerString:                  "Gamma",
			TargetingTeam:                1000000888,
		}),
		noTamedTimestampID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                           6001,
			ID2:                           6002,
			Tamed:                         false,
			LastInAllyRangeTimeSerialized: 1000,
			TribeName:                     "Delta",
			TamingTeamID:                  1000000999,
			TamerString:                   "Delta",
			OwningPlayerName:              "Dana",
			OwningPlayerID:                123,
			TargetingTeam:                 1000000999,
			TamedName:                     "NoTimestamp",
		}),
		aliasTimerID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                       7001,
			ID2:                       7002,
			LastInAllyRangeSerialized: 950,
			TribeName:                 "Epsilon",
			TamingTeamID:              1000001001,
			TamerString:               "Epsilon",
			TargetingTeam:             1000001001,
			TamedName:                 "AliasTimer",
		}),
		unclaimedID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                       8001,
			ID2:                       8002,
			LastInAllyRangeSerialized: 900,
			TargetingTeam:             2000000000,
			TamedName:                 "UnclaimedBaby",
		}),
		abandonedID: testfixtures.DinoGameObjectBytes(testfixtures.DinoGameObjectOptions{
			ID1:                       9001,
			ID2:                       9002,
			LastInAllyRangeSerialized: 850,
			TargetingTeam:             -2147483648,
			TamedName:                 "Abandoned",
		}),
	})
}
