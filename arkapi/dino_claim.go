package arkapi

import (
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

const defaultDinoClaimPeriodSeconds = 8 * secondsPerDay

type DinoClaimableOptions struct {
	MapName              string
	ClaimMultiplier      float64
	ClaimPeriodSeconds   float64
	GameUserSettingsPath string
}

type DinoClaimableReport struct {
	Summary DinoClaimableSummary `json:"summary"`
	Dinos   []DinoClaimableRow   `json:"dinos"`
}

type DinoClaimableSummary struct {
	TotalDinos            int     `json:"total_dinos"`
	OwnedDinos            int     `json:"owned_dinos"`
	ClaimableDinos        int     `json:"claimable_dinos"`
	UnknownTimestampDinos int     `json:"unknown_timestamp_dinos"`
	ClaimMultiplier       float64 `json:"claim_multiplier"`
	ClaimPeriodSeconds    float64 `json:"claim_period_seconds"`
	AdjustedPeriodSeconds float64 `json:"adjusted_period_seconds"`
	GameTime              float64 `json:"game_time"`
	FaultCount            int     `json:"fault_count"`
}

type DinoClaimableOwner struct {
	SortKey      string `json:"sort_key"`
	TribeName    string `json:"tribe_name,omitempty"`
	TamerTribeID int32  `json:"tamer_tribe_id,omitempty"`
	TamerString  string `json:"tamer_string,omitempty"`
	PlayerName   string `json:"player_name,omitempty"`
	PlayerID     int32  `json:"player_id,omitempty"`
	TargetTeam   int32  `json:"target_team,omitempty"`
}

type DinoClaimableRow struct {
	UUID                  string               `json:"uuid"`
	Blueprint             string               `json:"blueprint"`
	ShortName             string               `json:"short_name"`
	DinoID1               uint32               `json:"dino_id1,omitempty"`
	DinoID2               uint32               `json:"dino_id2,omitempty"`
	TamedName             string               `json:"tamed_name,omitempty"`
	Owner                 DinoClaimableOwner   `json:"owner"`
	Location              *arkobject.MapCoords `json:"location,omitempty"`
	GameTime              float64              `json:"game_time"`
	ClaimReferenceTime    float64              `json:"claim_reference_time,omitempty"`
	ClaimReferenceSource  string               `json:"claim_reference_source,omitempty"`
	TamedTimeStamp        float64              `json:"tamed_time_stamp,omitempty"`
	LastInAllyRangeTime   float64              `json:"last_in_ally_range_time_serialized,omitempty"`
	ElapsedSeconds        float64              `json:"elapsed_seconds,omitempty"`
	ClaimPeriodSeconds    float64              `json:"claim_period_seconds"`
	AdjustedPeriodSeconds float64              `json:"adjusted_period_seconds"`
	RemainingSeconds      float64              `json:"remaining_seconds"`
	Claimable             bool                 `json:"claimable"`
	UnknownTimestamp      bool                 `json:"unknown_timestamp,omitempty"`
}

func DinoClaimableReportFromPath(savePath string, opts DinoClaimableOptions) (DinoClaimableReport, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewDinoFromPath(savePath)
	if err != nil {
		return DinoClaimableReport{}, nil, err
	}
	defer closeAPI()
	return api.ClaimableReport(opts)
}

func (d *DinoAPI) ClaimableReport(opts DinoClaimableOptions) (DinoClaimableReport, []arksave.FaultyObjectInfo, error) {
	if d.save.Context == nil {
		return DinoClaimableReport{}, nil, fmt.Errorf("save context is nil")
	}
	if opts.MapName == "" {
		opts.MapName = d.save.Context.MapName
	}
	multiplier, err := resolveDinoClaimMultiplier(opts)
	if err != nil {
		return DinoClaimableReport{}, nil, err
	}
	period, err := resolveDinoClaimPeriod(opts)
	if err != nil {
		return DinoClaimableReport{}, nil, err
	}
	dinos, faults, err := d.AllWithFaults()
	if err != nil {
		return DinoClaimableReport{}, nil, err
	}
	adjusted := period * multiplier
	report := DinoClaimableReport{
		Summary: DinoClaimableSummary{
			TotalDinos:            len(dinos),
			ClaimMultiplier:       multiplier,
			ClaimPeriodSeconds:    period,
			AdjustedPeriodSeconds: adjusted,
			GameTime:              d.save.Context.GameTime,
			FaultCount:            len(faults),
		},
	}
	for _, id := range sortedUUIDKeys(dinos) {
		dino := dinos[id]
		if !isOwnedClaimableCandidate(dino) {
			continue
		}
		report.Summary.OwnedDinos++
		row := dinoClaimableRow(id, dino, opts.MapName, d.save.Context.GameTime, period, adjusted)
		if row.UnknownTimestamp {
			report.Summary.UnknownTimestampDinos++
		}
		if row.Claimable {
			report.Summary.ClaimableDinos++
			report.Dinos = append(report.Dinos, row)
		}
	}
	sortDinoClaimableRows(report.Dinos)
	return report, faults, nil
}

func resolveDinoClaimMultiplier(opts DinoClaimableOptions) (float64, error) {
	if opts.ClaimMultiplier != 0 {
		if opts.ClaimMultiplier <= 0 || math.IsNaN(opts.ClaimMultiplier) || math.IsInf(opts.ClaimMultiplier, 0) {
			return 0, fmt.Errorf("claim multiplier must be a positive finite number")
		}
		return opts.ClaimMultiplier, nil
	}
	if opts.GameUserSettingsPath != "" {
		value, ok, err := ParsePvEDinoDecayPeriodMultiplier(opts.GameUserSettingsPath)
		if err != nil {
			return 0, err
		}
		if ok {
			return value, nil
		}
	}
	return 1, nil
}

func resolveDinoClaimPeriod(opts DinoClaimableOptions) (float64, error) {
	if opts.ClaimPeriodSeconds == 0 {
		return defaultDinoClaimPeriodSeconds, nil
	}
	if opts.ClaimPeriodSeconds <= 0 || math.IsNaN(opts.ClaimPeriodSeconds) || math.IsInf(opts.ClaimPeriodSeconds, 0) {
		return 0, fmt.Errorf("claim period must be a positive finite number of seconds")
	}
	return opts.ClaimPeriodSeconds, nil
}

func ParsePvEDinoDecayPeriodMultiplier(path string) (float64, bool, error) {
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
		if !ok || strings.TrimSpace(key) != "PvEDinoDecayPeriodMultiplier" {
			continue
		}
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return 0, false, fmt.Errorf("parse PvEDinoDecayPeriodMultiplier: %w", err)
		}
		if parsed <= 0 || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
			return 0, false, fmt.Errorf("PvEDinoDecayPeriodMultiplier must be a positive finite number")
		}
		return parsed, true, nil
	}
	return 0, false, nil
}

func isOwnedClaimableCandidate(dino arkobject.Dino) bool {
	return dino.IsTamed && !dino.IsDead && !dino.IsCryopodded
}

func dinoClaimableRow(id uuid.UUID, dino arkobject.Dino, mapName string, gameTime float64, period float64, adjusted float64) DinoClaimableRow {
	row := DinoClaimableRow{
		UUID:                  id.String(),
		Blueprint:             dino.Blueprint,
		ShortName:             dino.ShortName(),
		DinoID1:               dino.ID1,
		DinoID2:               dino.ID2,
		TamedName:             dino.TamedName,
		Owner:                 dinoClaimableOwner(dino.Owner),
		GameTime:              gameTime,
		ClaimReferenceTime:    dinoClaimReferenceTime(dino),
		ClaimReferenceSource:  dinoClaimReferenceSource(dino),
		TamedTimeStamp:        dino.TamedTimeStamp,
		LastInAllyRangeTime:   dino.LastInAllyRangeTimeSerialized,
		ClaimPeriodSeconds:    period,
		AdjustedPeriodSeconds: adjusted,
		RemainingSeconds:      adjusted,
		UnknownTimestamp:      dinoClaimReferenceTime(dino) == 0,
	}
	if dino.Location != nil {
		coords := dino.Location.AsMapCoords(mapName)
		row.Location = &coords
	}
	if row.UnknownTimestamp {
		return row
	}
	row.ElapsedSeconds = gameTime - row.ClaimReferenceTime
	row.RemainingSeconds = math.Max(0, adjusted-row.ElapsedSeconds)
	row.Claimable = row.ElapsedSeconds >= adjusted
	return row
}

func dinoClaimReferenceTime(dino arkobject.Dino) float64 {
	if dino.LastInAllyRangeTimeSerialized != 0 {
		return dino.LastInAllyRangeTimeSerialized
	}
	return dino.TamedTimeStamp
}

func dinoClaimReferenceSource(dino arkobject.Dino) string {
	if dino.LastInAllyRangeTimeSerialized != 0 {
		return "last_in_ally_range_time_serialized"
	}
	if dino.TamedTimeStamp != 0 {
		return "tamed_time_stamp"
	}
	return ""
}

func dinoClaimableOwner(owner arkobject.DinoOwner) DinoClaimableOwner {
	out := DinoClaimableOwner{
		TribeName:    owner.TribeName,
		TamerTribeID: owner.TamerTribeID,
		TamerString:  owner.TamerString,
		PlayerName:   owner.PlayerName,
		PlayerID:     owner.PlayerID,
		TargetTeam:   owner.TargetTeam,
	}
	switch {
	case owner.TribeName != "":
		out.SortKey = owner.TribeName
	case owner.TamerString != "":
		out.SortKey = owner.TamerString
	case owner.TamerTribeID != 0:
		out.SortKey = fmt.Sprintf("%012d", owner.TamerTribeID)
	case owner.TargetTeam != 0:
		out.SortKey = fmt.Sprintf("%012d", owner.TargetTeam)
	case owner.PlayerName != "":
		out.SortKey = owner.PlayerName
	case owner.PlayerID != 0:
		out.SortKey = fmt.Sprintf("%012d", owner.PlayerID)
	default:
		out.SortKey = "unknown"
	}
	return out
}

func sortDinoClaimableRows(rows []DinoClaimableRow) {
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
