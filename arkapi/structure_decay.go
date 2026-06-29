package arkapi

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

const secondsPerDay = 24 * 60 * 60

type StructureDemolishableOptions struct {
	MapName              string
	DecayMultiplier      float64
	GameUserSettingsPath string
	DecayPeriodsPath     string
	DecayPeriods         StructureDecayPeriods
	GroupBases           bool
}

type StructureDecayPeriods struct {
	Exact     map[string]float64 `json:"exact"`
	Substring map[string]float64 `json:"substring"`
}

type StructureDecayClass struct {
	Tier          string  `json:"tier"`
	PeriodSeconds float64 `json:"period_seconds"`
	Source        string  `json:"source"`
	Matched       string  `json:"matched,omitempty"`
}

type StructureDemolishableReport struct {
	Summary    StructureDemolishableSummary `json:"summary"`
	Structures []StructureDemolishableRow   `json:"structures"`
	Bases      []StructureDemolishableBase  `json:"bases,omitempty"`
}

type StructureDemolishableSummary struct {
	TotalStructures            int     `json:"total_structures"`
	EligibleStructures         int     `json:"eligible_structures"`
	UnknownTimestampStructures int     `json:"unknown_timestamp_structures"`
	DecayMultiplier            float64 `json:"decay_multiplier"`
	GameTime                   float64 `json:"game_time"`
	FaultCount                 int     `json:"fault_count"`
}

type StructureDemolishableOwner struct {
	SortKey          string `json:"sort_key"`
	TribeName        string `json:"tribe_name,omitempty"`
	TribeID          int32  `json:"tribe_id,omitempty"`
	PlayerName       string `json:"player_name,omitempty"`
	PlayerID         int32  `json:"player_id,omitempty"`
	OriginalPlacerID int32  `json:"original_placer_id,omitempty"`
}

type StructureDemolishableRow struct {
	UUID                  string                     `json:"uuid"`
	Blueprint             string                     `json:"blueprint"`
	ShortName             string                     `json:"short_name"`
	StructureID           int32                      `json:"structure_id"`
	Owner                 StructureDemolishableOwner `json:"owner"`
	Location              *arkobject.MapCoords       `json:"location,omitempty"`
	GameTime              float64                    `json:"game_time"`
	LastEnterStasisTime   float64                    `json:"last_enter_stasis_time,omitempty"`
	ElapsedSeconds        float64                    `json:"elapsed_seconds,omitempty"`
	DecayPeriodSeconds    float64                    `json:"decay_period_seconds"`
	AdjustedPeriodSeconds float64                    `json:"adjusted_period_seconds"`
	RemainingSeconds      float64                    `json:"remaining_seconds"`
	Eligible              bool                       `json:"eligible"`
	Tier                  string                     `json:"tier"`
	PeriodSource          string                     `json:"period_source"`
	OriginalCreationTime  float64                    `json:"original_creation_time,omitempty"`
	HasResetDecayTime     bool                       `json:"has_reset_decay_time,omitempty"`
	SavedWhenStasised     bool                       `json:"saved_when_stasised,omitempty"`
	WasPlacementSnapped   bool                       `json:"was_placement_snapped,omitempty"`
	LastInAllyRangeTime   float64                    `json:"last_in_ally_range_time_serialized,omitempty"`
	UnknownTimestamp      bool                       `json:"unknown_timestamp,omitempty"`
}

type StructureDemolishableBase struct {
	Owner              StructureDemolishableOwner `json:"owner"`
	AverageLocation    *arkobject.MapCoords       `json:"average_location,omitempty"`
	EligibleStructures int                        `json:"eligible_structures"`
	TotalStructures    int                        `json:"total_structures"`
	OldestElapsed      float64                    `json:"oldest_elapsed_seconds"`
	DominantTier       string                     `json:"dominant_tier"`
	StructureUUIDs     []string                   `json:"structure_uuids"`
}

func StructureDemolishableReportFromPath(savePath string, opts StructureDemolishableOptions) (StructureDemolishableReport, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewStructureFromPath(savePath)
	if err != nil {
		return StructureDemolishableReport{}, nil, err
	}
	defer closeAPI()
	return api.DemolishableReport(opts)
}

func (s *StructureAPI) DemolishableReport(opts StructureDemolishableOptions) (StructureDemolishableReport, []arksave.FaultyObjectInfo, error) {
	if s.save.Context == nil {
		return StructureDemolishableReport{}, nil, fmt.Errorf("save context is nil")
	}
	if opts.MapName == "" {
		opts.MapName = s.save.Context.MapName
	}
	multiplier, err := resolveStructureDecayMultiplier(opts)
	if err != nil {
		return StructureDemolishableReport{}, nil, err
	}
	periods, err := resolveStructureDecayPeriods(opts)
	if err != nil {
		return StructureDemolishableReport{}, nil, err
	}
	structures, faults, err := s.selectedStructureIndexWithFaults()
	if err != nil {
		return StructureDemolishableReport{}, nil, err
	}
	report := StructureDemolishableReport{
		Summary: StructureDemolishableSummary{
			TotalStructures: len(structures),
			DecayMultiplier: multiplier,
			GameTime:        s.save.Context.GameTime,
			FaultCount:      len(faults),
		},
	}
	for _, id := range sortedUUIDKeys(structures) {
		row := structureDemolishableRow(id, structures[id], opts.MapName, s.save.Context.GameTime, multiplier, periods)
		if row.UnknownTimestamp {
			report.Summary.UnknownTimestampStructures++
		}
		if row.Eligible {
			report.Summary.EligibleStructures++
			report.Structures = append(report.Structures, row)
		}
	}
	sortStructureDemolishableRows(report.Structures)
	if opts.GroupBases {
		report.Bases = structureDemolishableBases(structures, report.Structures, opts.MapName)
	}
	return report, faults, nil
}

func resolveStructureDecayMultiplier(opts StructureDemolishableOptions) (float64, error) {
	if opts.DecayMultiplier != 0 {
		if opts.DecayMultiplier <= 0 || math.IsNaN(opts.DecayMultiplier) || math.IsInf(opts.DecayMultiplier, 0) {
			return 0, fmt.Errorf("decay multiplier must be a positive finite number")
		}
		return opts.DecayMultiplier, nil
	}
	if opts.GameUserSettingsPath != "" {
		value, ok, err := ParsePvEStructureDecayPeriodMultiplier(opts.GameUserSettingsPath)
		if err != nil {
			return 0, err
		}
		if ok {
			return value, nil
		}
	}
	return 1, nil
}

func ParsePvEStructureDecayPeriodMultiplier(path string) (float64, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false, fmt.Errorf("read game user settings: %w", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "[") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok || strings.TrimSpace(key) != "PvEStructureDecayPeriodMultiplier" {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return 0, false, fmt.Errorf("parse PvEStructureDecayPeriodMultiplier: %w", err)
		}
		if parsed <= 0 || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return 0, false, fmt.Errorf("PvEStructureDecayPeriodMultiplier must be a positive finite number")
		}
		return parsed, true, nil
	}
	return 0, false, nil
}

func resolveStructureDecayPeriods(opts StructureDemolishableOptions) (StructureDecayPeriods, error) {
	periods := opts.DecayPeriods
	if opts.DecayPeriodsPath == "" {
		return periods, nil
	}
	raw, err := os.ReadFile(opts.DecayPeriodsPath)
	if err != nil {
		return StructureDecayPeriods{}, fmt.Errorf("read decay periods: %w", err)
	}
	if err := json.Unmarshal(raw, &periods); err != nil {
		return StructureDecayPeriods{}, fmt.Errorf("parse decay periods: %w", err)
	}
	return periods, nil
}

func ClassifyStructureDecay(blueprint string, overrides StructureDecayPeriods) StructureDecayClass {
	if overrides.Exact != nil {
		if period, ok := overrides.Exact[blueprint]; ok {
			return StructureDecayClass{Tier: "custom", PeriodSeconds: period, Source: "custom_exact", Matched: blueprint}
		}
	}
	if overrides.Substring != nil {
		for _, key := range sortedStringKeys(overrides.Substring) {
			if strings.Contains(blueprint, key) {
				return StructureDecayClass{Tier: "custom", PeriodSeconds: overrides.Substring[key], Source: "custom_substring", Matched: key}
			}
		}
	}
	lower := strings.ToLower(blueprint)
	switch {
	case strings.Contains(lower, "cryofridge") || (strings.Contains(lower, "cryo") && strings.Contains(lower, "fridge")):
		return StructureDecayClass{Tier: "tek_utility", PeriodSeconds: 40 * secondsPerDay, Source: "built_in", Matched: "cryofridge"}
	case strings.Contains(lower, "/tek/") || strings.Contains(lower, "_tek") || strings.Contains(lower, "tek"):
		return StructureDecayClass{Tier: "tek", PeriodSeconds: 20 * secondsPerDay, Source: "built_in", Matched: "tek"}
	case strings.Contains(lower, "storagebox_huge") || strings.Contains(lower, "vault"):
		return StructureDecayClass{Tier: "vault", PeriodSeconds: 16 * secondsPerDay, Source: "built_in", Matched: "vault"}
	case strings.Contains(lower, "greenhouse"):
		return StructureDecayClass{Tier: "greenhouse", PeriodSeconds: 16 * secondsPerDay, Source: "built_in", Matched: "greenhouse"}
	case strings.Contains(lower, "/metal/") || strings.Contains(lower, "_metal"):
		return StructureDecayClass{Tier: "metal", PeriodSeconds: 16 * secondsPerDay, Source: "built_in", Matched: "metal"}
	case strings.Contains(lower, "/stone/") || strings.Contains(lower, "_stone"):
		return StructureDecayClass{Tier: "stone", PeriodSeconds: 12 * secondsPerDay, Source: "built_in", Matched: "stone"}
	case strings.Contains(lower, "/wood/") || strings.Contains(lower, "_wood"):
		return StructureDecayClass{Tier: "wood", PeriodSeconds: 8 * secondsPerDay, Source: "built_in", Matched: "wood"}
	case strings.Contains(lower, "adobe"):
		return StructureDecayClass{Tier: "adobe", PeriodSeconds: 8 * secondsPerDay, Source: "built_in", Matched: "adobe"}
	case strings.Contains(lower, "/thatch/") || strings.Contains(lower, "_thatch"):
		return StructureDecayClass{Tier: "thatch", PeriodSeconds: 4 * secondsPerDay, Source: "built_in", Matched: "thatch"}
	default:
		return StructureDecayClass{Tier: "unknown", PeriodSeconds: 40 * secondsPerDay, Source: "unknown_fallback"}
	}
}

func structureDemolishableRow(id uuid.UUID, structure arkobject.Structure, mapName string, gameTime float64, multiplier float64, periods StructureDecayPeriods) StructureDemolishableRow {
	class := ClassifyStructureDecay(structure.Blueprint, periods)
	adjusted := class.PeriodSeconds * multiplier
	row := StructureDemolishableRow{
		UUID:                  id.String(),
		Blueprint:             structure.Blueprint,
		ShortName:             arkobject.ShortNameFromBlueprint(structure.Blueprint),
		StructureID:           structure.ID,
		Owner:                 structureDemolishableOwner(structure.Owner),
		GameTime:              gameTime,
		LastEnterStasisTime:   structure.LastEnterStasisTime,
		DecayPeriodSeconds:    class.PeriodSeconds,
		AdjustedPeriodSeconds: adjusted,
		RemainingSeconds:      adjusted,
		Tier:                  class.Tier,
		PeriodSource:          class.Source,
		OriginalCreationTime:  structure.OriginalCreationTime,
		HasResetDecayTime:     structure.HasResetDecayTime,
		SavedWhenStasised:     structure.SavedWhenStasised,
		WasPlacementSnapped:   structure.WasPlacementSnapped,
		LastInAllyRangeTime:   structure.LastInAllyRangeTimeSerialized,
		UnknownTimestamp:      structure.LastEnterStasisTime == 0,
	}
	if structure.Location != nil {
		coords := structure.Location.AsMapCoords(mapName)
		row.Location = &coords
	}
	if row.UnknownTimestamp {
		return row
	}
	row.ElapsedSeconds = gameTime - structure.LastEnterStasisTime
	row.RemainingSeconds = math.Max(0, adjusted-row.ElapsedSeconds)
	row.Eligible = row.ElapsedSeconds >= adjusted
	return row
}

func structureDemolishableOwner(owner arkobject.ObjectOwner) StructureDemolishableOwner {
	out := StructureDemolishableOwner{
		TribeName:        owner.TribeName,
		TribeID:          owner.TribeID,
		PlayerName:       owner.PlayerName,
		PlayerID:         owner.PlayerID,
		OriginalPlacerID: owner.OriginalPlacerID,
	}
	switch {
	case owner.TribeName != "":
		out.SortKey = owner.TribeName
	case owner.TribeID != 0:
		out.SortKey = fmt.Sprintf("%012d", owner.TribeID)
	case owner.PlayerName != "":
		out.SortKey = owner.PlayerName
	case owner.PlayerID != 0:
		out.SortKey = fmt.Sprintf("%012d", owner.PlayerID)
	default:
		out.SortKey = "unknown"
	}
	return out
}

func sortStructureDemolishableRows(rows []StructureDemolishableRow) {
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Owner.SortKey != rows[j].Owner.SortKey {
			return rows[i].Owner.SortKey < rows[j].Owner.SortKey
		}
		latI, longI := demolishableLocationSort(rows[i].Location)
		latJ, longJ := demolishableLocationSort(rows[j].Location)
		if latI != latJ {
			return latI < latJ
		}
		if longI != longJ {
			return longI < longJ
		}
		if rows[i].ShortName != rows[j].ShortName {
			return rows[i].ShortName < rows[j].ShortName
		}
		return rows[i].UUID < rows[j].UUID
	})
}

func demolishableLocationSort(location *arkobject.MapCoords) (float64, float64) {
	if location == nil || location.InCryopod {
		return math.Inf(1), math.Inf(1)
	}
	return location.Lat, location.Long
}

func structureDemolishableBases(structures map[uuid.UUID]arkobject.Structure, eligible []StructureDemolishableRow, mapName string) []StructureDemolishableBase {
	eligibleByUUID := map[uuid.UUID]StructureDemolishableRow{}
	for _, row := range eligible {
		id, err := uuid.Parse(row.UUID)
		if err == nil {
			eligibleByUUID[id] = row
		}
	}
	components := structureComponents(structures)
	bases := make([]StructureDemolishableBase, 0, len(components))
	for _, component := range components {
		base := StructureDemolishableBase{TotalStructures: len(component)}
		tierCounts := map[string]int{}
		var latTotal, longTotal float64
		var locationCount int
		for _, id := range component {
			row, ok := eligibleByUUID[id]
			if !ok {
				continue
			}
			base.EligibleStructures++
			base.StructureUUIDs = append(base.StructureUUIDs, row.UUID)
			if base.Owner.SortKey == "" {
				base.Owner = row.Owner
			}
			if row.ElapsedSeconds > base.OldestElapsed {
				base.OldestElapsed = row.ElapsedSeconds
			}
			tierCounts[row.Tier]++
			if row.Location != nil && !row.Location.InCryopod {
				latTotal += row.Location.Lat
				longTotal += row.Location.Long
				locationCount++
			}
		}
		if base.EligibleStructures == 0 {
			continue
		}
		sort.Strings(base.StructureUUIDs)
		base.DominantTier = dominantTier(tierCounts)
		if locationCount != 0 {
			base.AverageLocation = &arkobject.MapCoords{Lat: latTotal / float64(locationCount), Long: longTotal / float64(locationCount)}
		}
		bases = append(bases, base)
	}
	sort.Slice(bases, func(i, j int) bool {
		if bases[i].Owner.SortKey != bases[j].Owner.SortKey {
			return bases[i].Owner.SortKey < bases[j].Owner.SortKey
		}
		latI, longI := demolishableLocationSort(bases[i].AverageLocation)
		latJ, longJ := demolishableLocationSort(bases[j].AverageLocation)
		if latI != latJ {
			return latI < latJ
		}
		if longI != longJ {
			return longI < longJ
		}
		return bases[i].StructureUUIDs[0] < bases[j].StructureUUIDs[0]
	})
	return bases
}

func structureComponents(structures map[uuid.UUID]arkobject.Structure) [][]uuid.UUID {
	adjacent := map[uuid.UUID][]uuid.UUID{}
	for id, structure := range structures {
		for _, linkedID := range structure.LinkedStructureUUIDs {
			if _, ok := structures[linkedID]; !ok {
				continue
			}
			adjacent[id] = append(adjacent[id], linkedID)
			adjacent[linkedID] = append(adjacent[linkedID], id)
		}
	}
	visited := map[uuid.UUID]bool{}
	components := [][]uuid.UUID{}
	for _, start := range sortedUUIDKeys(structures) {
		if visited[start] {
			continue
		}
		queue := []uuid.UUID{start}
		visited[start] = true
		component := []uuid.UUID{}
		for len(queue) != 0 {
			id := queue[0]
			queue = queue[1:]
			component = append(component, id)
			for _, next := range adjacent[id] {
				if visited[next] {
					continue
				}
				visited[next] = true
				queue = append(queue, next)
			}
		}
		sort.Slice(component, func(i, j int) bool { return component[i].String() < component[j].String() })
		components = append(components, component)
	}
	return components
}

func dominantTier(counts map[string]int) string {
	bestTier := ""
	bestCount := -1
	for _, tier := range sortedStringKeysFromCounts(counts) {
		if counts[tier] > bestCount {
			bestTier = tier
			bestCount = counts[tier]
		}
	}
	return bestTier
}

func sortedStringKeys(values map[string]float64) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func sortedStringKeysFromCounts(values map[string]int) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
