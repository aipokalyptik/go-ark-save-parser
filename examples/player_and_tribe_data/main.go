package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/aipokalyptik/go-ark-save-parser/arkapi"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

type summary struct {
	Players             int           `json:"players"`
	Tribes              int           `json:"tribes"`
	PlayersWithNames    int           `json:"players_with_names"`
	TribesWithNames     int           `json:"tribes_with_names"`
	ActiveLinks         int           `json:"active_links"`
	InactiveMembers     int           `json:"inactive_members"`
	PlayersWithoutTribe int           `json:"players_without_tribe"`
	TribeRows           []tribeRow    `json:"tribe_rows"`
	PlayerRows          []playerRow   `json:"player_rows"`
	RelationRows        []relationRow `json:"relation_rows"`
}

type playerRow struct {
	HasCharacterName bool  `json:"has_character_name"`
	HasPlayerName    bool  `json:"has_player_name"`
	Level            int32 `json:"level"`
	TribeID          int32 `json:"tribe_id"`
}

type tribeRow struct {
	HasName bool  `json:"has_name"`
	Members int   `json:"members"`
	Dinos   int32 `json:"dinos"`
}

type relationRow struct {
	ActiveMembers   int `json:"active_members"`
	InactiveMembers int `json:"inactive_members"`
}

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <save.ark-or-save-directory>", os.Args[0])
	}
	api, closeSave, err := openPlayerAPI(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer closeSave()

	data, err := buildSummary(api)
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(data); err != nil {
		log.Fatal(err)
	}
}

func buildSummary(api *arkapi.PlayerAPI) (summary, error) {
	players, err := api.Players()
	if err != nil {
		return summary{}, err
	}
	tribes, err := api.TribeDetails()
	if err != nil {
		return summary{}, err
	}
	relations, err := api.TribePlayerRelations()
	if err != nil {
		return summary{}, err
	}

	out := summary{
		Players:      len(players),
		Tribes:       len(tribes),
		PlayerRows:   make([]playerRow, 0, len(players)),
		TribeRows:    make([]tribeRow, 0, len(tribes)),
		RelationRows: make([]relationRow, 0, len(relations)),
	}
	activePlayerIDs := map[uint64]struct{}{}
	for _, tribe := range tribes {
		row := tribeRow{HasName: tribe.Name != "", Members: len(tribe.MemberIDs), Dinos: tribe.NumDinos}
		if row.HasName {
			out.TribesWithNames++
		}
		out.TribeRows = append(out.TribeRows, row)
	}
	for _, player := range players {
		row := playerRow{
			HasCharacterName: player.CharacterName != "",
			HasPlayerName:    player.PlayerName != "",
			Level:            player.Level,
			TribeID:          player.TribeID,
		}
		if row.HasCharacterName || row.HasPlayerName {
			out.PlayersWithNames++
		}
		out.PlayerRows = append(out.PlayerRows, row)
	}
	for _, relation := range relations {
		row := relationRow{ActiveMembers: len(relation.ActivePlayers), InactiveMembers: len(relation.InactiveMemberIDs)}
		out.ActiveLinks += row.ActiveMembers
		out.InactiveMembers += row.InactiveMembers
		for _, player := range relation.ActivePlayers {
			activePlayerIDs[player.PlayerDataID] = struct{}{}
		}
		out.RelationRows = append(out.RelationRows, row)
	}
	out.PlayersWithoutTribe = len(players) - len(activePlayerIDs)

	sort.Slice(out.PlayerRows, func(i int, j int) bool {
		if out.PlayerRows[i].TribeID != out.PlayerRows[j].TribeID {
			return out.PlayerRows[i].TribeID < out.PlayerRows[j].TribeID
		}
		if out.PlayerRows[i].Level != out.PlayerRows[j].Level {
			return out.PlayerRows[i].Level < out.PlayerRows[j].Level
		}
		if out.PlayerRows[i].HasCharacterName != out.PlayerRows[j].HasCharacterName {
			return !out.PlayerRows[i].HasCharacterName
		}
		return !out.PlayerRows[i].HasPlayerName && out.PlayerRows[j].HasPlayerName
	})
	sort.Slice(out.TribeRows, func(i int, j int) bool {
		if out.TribeRows[i].Members != out.TribeRows[j].Members {
			return out.TribeRows[i].Members < out.TribeRows[j].Members
		}
		if out.TribeRows[i].Dinos != out.TribeRows[j].Dinos {
			return out.TribeRows[i].Dinos < out.TribeRows[j].Dinos
		}
		return !out.TribeRows[i].HasName && out.TribeRows[j].HasName
	})
	sort.Slice(out.RelationRows, func(i int, j int) bool {
		if out.RelationRows[i].ActiveMembers != out.RelationRows[j].ActiveMembers {
			return out.RelationRows[i].ActiveMembers < out.RelationRows[j].ActiveMembers
		}
		return out.RelationRows[i].InactiveMembers < out.RelationRows[j].InactiveMembers
	})
	return out, nil
}

func openPlayerAPI(path string) (*arkapi.PlayerAPI, func(), error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, func() {}, err
	}
	if info.IsDir() {
		api, err := arkapi.NewPlayerFromDirectory(path)
		return api, func() {}, err
	}
	save, err := arksave.Open(path)
	if err != nil {
		return nil, func() {}, err
	}
	api := arkapi.NewPlayer(save)
	players, _, err := api.PlayersWithFaults()
	if err != nil {
		_ = save.Close()
		return nil, func() {}, err
	}
	if len(players) == 0 {
		_ = save.Close()
		api, err := arkapi.NewPlayerFromDirectory(filepath.Dir(path))
		return api, func() {}, err
	}
	return api, func() { _ = save.Close() }, nil
}
