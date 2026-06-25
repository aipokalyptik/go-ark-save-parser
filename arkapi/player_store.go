package arkapi

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/google/uuid"
)

var (
	embeddedTribeDataName  = []byte("/Script/ShooterGame.PrimalTribeData\x00")
	embeddedPlayerDataName = []byte("Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C\x00")
	embeddedNonePattern    = []byte("None")
)

func (p *PlayerAPI) saveEmbeddedPlayersWithFaults() ([]arkobject.Player, []arksave.FaultyObjectInfo, error) {
	raw, ok, err := p.gameModeCustomBytes()
	if err != nil || !ok {
		return nil, nil, err
	}
	archives := embeddedPlayerArchives(raw)
	out := make([]arkobject.Player, 0, len(archives))
	var faults []arksave.FaultyObjectInfo
	for _, rawArchive := range archives {
		archive, err := arkarchive.Parse(rawArchive.data, arkarchive.Options{FromStore: true, Format: arkarchive.FormatAuto})
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: rawArchive.id, ClassName: "/Game/PrimalEarth/CoreBlueprints/PrimalPlayerDataBP.PrimalPlayerDataBP_C", Err: err})
			continue
		}
		if len(archive.Objects) == 0 {
			continue
		}
		object := archive.Objects[0]
		player, err := arkobject.PlayerFromContainer(arkproperty.Container{Properties: object.Properties})
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: object.UUID, ClassName: object.ClassName, Err: fmt.Errorf("parse embedded player object %s: %w", object.UUID, err)})
			continue
		}
		if object.PropertyError != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: object.UUID, ClassName: object.ClassName, Err: object.PropertyError})
		}
		out = append(out, player)
	}
	return out, faults, nil
}

func (p *PlayerAPI) saveEmbeddedTribeDetailsWithFaults() ([]arkobject.Tribe, []arksave.FaultyObjectInfo, error) {
	raw, ok, err := p.gameModeCustomBytes()
	if err != nil || !ok {
		return nil, nil, err
	}
	archives := embeddedTribeArchives(raw)
	out := make([]arkobject.Tribe, 0, len(archives))
	var faults []arksave.FaultyObjectInfo
	for _, rawArchive := range archives {
		archive, err := arkarchive.Parse(rawArchive.data, arkarchive.Options{FromStore: true, Format: arkarchive.FormatAuto})
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: rawArchive.id, ClassName: "/Script/ShooterGame.PrimalTribeData", Err: err})
			continue
		}
		if len(archive.Objects) == 0 {
			continue
		}
		object := archive.Objects[0]
		tribe, err := arkobject.TribeFromContainer(arkproperty.Container{Properties: object.Properties})
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: object.UUID, ClassName: object.ClassName, Err: fmt.Errorf("parse embedded tribe object %s: %w", object.UUID, err)})
			continue
		}
		if object.PropertyError != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: object.UUID, ClassName: object.ClassName, Err: object.PropertyError})
		}
		out = append(out, tribe)
	}
	return out, faults, nil
}

func (p *PlayerAPI) gameModeCustomBytes() ([]byte, bool, error) {
	if p.save == nil {
		return nil, false, nil
	}
	raw, err := p.save.CustomValue("GameModeCustomBytes")
	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if len(raw) < 30 {
		return nil, false, nil
	}
	return raw, true, nil
}

type embeddedArchive struct {
	id   uuid.UUID
	data []byte
}

func embeddedPlayerArchives(data []byte) []embeddedArchive {
	positions := findAdjustedAll(data, embeddedPlayerDataName, -1)
	out := make([]embeddedArchive, 0, len(positions))
	for i, pos := range positions {
		if pos-20 < 0 || pos-4 > len(data) {
			continue
		}
		idBytes := append([]byte(nil), data[pos-20:pos-4]...)
		id, err := uuid.FromBytes(idBytes)
		if err != nil {
			continue
		}
		offset := pos - 36
		if offset < 0 || i+1 >= len(positions) {
			continue
		}
		lastNone := findAdjustedLastBefore(data, embeddedNonePattern, positions[i+1], -1)
		if lastNone < 0 {
			continue
		}
		size := lastNone + 4 - offset
		end := offset + size + 1
		if size < 0 || end > len(data) {
			continue
		}
		raw := append([]byte(nil), data[offset:end]...)
		raw = append(raw, 0x00, 0x01, 0x00, 0x00, 0x00)
		raw = append(raw, idBytes...)
		out = append(out, embeddedArchive{id: id, data: raw})
	}
	return out
}

func embeddedTribeArchives(data []byte) []embeddedArchive {
	positions := findAdjustedAll(data, embeddedTribeDataName, -1)
	out := make([]embeddedArchive, 0, len(positions))
	for _, pos := range positions {
		if pos-20 < 0 || pos-4 > len(data) {
			continue
		}
		idBytes := append([]byte(nil), data[pos-20:pos-4]...)
		id, err := uuid.FromBytes(idBytes)
		if err != nil {
			continue
		}
		uuidPositions := findAdjustedAll(data, idBytes, -1)
		if len(uuidPositions) < 2 {
			continue
		}
		offset := pos - 36
		start := offset + 1
		size := uuidPositions[1] - offset
		end := start + size
		if start < 0 || size < 0 || end > len(data) {
			continue
		}
		out = append(out, embeddedArchive{id: id, data: append([]byte(nil), data[start:end]...)})
	}
	return out
}

func findAdjustedAll(data []byte, pattern []byte, adjust int) []int {
	var out []int
	for pos := 0; pos >= 0 && pos < len(data); {
		idx := bytes.Index(data[pos:], pattern)
		if idx < 0 {
			break
		}
		actual := pos + idx
		if actual > 0 {
			out = append(out, actual+adjust)
		}
		pos = actual + len(pattern)
	}
	return out
}

func findAdjustedLastBefore(data []byte, pattern []byte, before int, adjust int) int {
	if before > len(data) {
		before = len(data)
	}
	if before < 0 {
		return -1
	}
	idx := bytes.LastIndex(data[:before], pattern)
	if idx < 0 {
		return -1
	}
	return idx + adjust
}

func mergePlayersByDataID(base []arkobject.Player, extra []arkobject.Player) []arkobject.Player {
	if len(extra) == 0 {
		return base
	}
	seen := make(map[uint64]int, len(base)+len(extra))
	for i, player := range base {
		seen[player.PlayerDataID] = i
	}
	out := append([]arkobject.Player(nil), base...)
	for _, player := range extra {
		if idx, ok := seen[player.PlayerDataID]; ok {
			out[idx] = player
			continue
		}
		seen[player.PlayerDataID] = len(out)
		out = append(out, player)
	}
	return out
}

func mergeTribesByID(base []arkobject.Tribe, extra []arkobject.Tribe) []arkobject.Tribe {
	if len(extra) == 0 {
		return base
	}
	seen := make(map[int32]int, len(base)+len(extra))
	for i, tribe := range base {
		seen[tribe.TribeID] = i
	}
	out := append([]arkobject.Tribe(nil), base...)
	for _, tribe := range extra {
		if idx, ok := seen[tribe.TribeID]; ok {
			out[idx] = tribe
			continue
		}
		seen[tribe.TribeID] = len(out)
		out = append(out, tribe)
	}
	return out
}
