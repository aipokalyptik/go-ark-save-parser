package arkapi

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkprofile"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
)

type PlayerAPI struct {
	save         *arksave.Save
	profilePaths []string
	tribePaths   []string
	clusterPaths []string
	tributePaths []string
}

func NewPlayer(save *arksave.Save) *PlayerAPI {
	return &PlayerAPI{save: save}
}

func NewPlayerFromDirectory(dir string) (*PlayerAPI, error) {
	api := &PlayerAPI{}
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		switch strings.ToLower(filepath.Ext(path)) {
		case ".arkprofile":
			api.profilePaths = append(api.profilePaths, path)
		case ".arktribe":
			api.tribePaths = append(api.tribePaths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	clusterFiles, err := arkcluster.Discover(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range clusterFiles {
		api.clusterPaths = append(api.clusterPaths, file.Path)
	}
	tributeFiles, err := arktribute.Discover(dir)
	if err != nil {
		return nil, err
	}
	for _, file := range tributeFiles {
		api.tributePaths = append(api.tributePaths, file.Path)
	}
	sort.Strings(api.profilePaths)
	sort.Strings(api.tribePaths)
	sort.Strings(api.clusterPaths)
	sort.Strings(api.tributePaths)
	return api, nil
}

func (p *PlayerAPI) ProfilePaths() []string {
	out := make([]string, len(p.profilePaths))
	copy(out, p.profilePaths)
	return out
}

func (p *PlayerAPI) TribePaths() []string {
	out := make([]string, len(p.tribePaths))
	copy(out, p.tribePaths)
	return out
}

func (p *PlayerAPI) ClusterPaths() []string {
	out := make([]string, len(p.clusterPaths))
	copy(out, p.clusterPaths)
	return out
}

func (p *PlayerAPI) TributePaths() []string {
	out := make([]string, len(p.tributePaths))
	copy(out, p.tributePaths)
	return out
}

func (p *PlayerAPI) Profiles() ([]*arkprofile.PlayerProfile, error) {
	out := make([]*arkprofile.PlayerProfile, 0, len(p.profilePaths))
	for _, path := range p.profilePaths {
		profile, err := arkprofile.OpenPlayerProfile(path)
		if err != nil {
			return nil, err
		}
		out = append(out, profile)
	}
	return out, nil
}

func (p *PlayerAPI) Players() ([]arkobject.Player, error) {
	if p.save != nil {
		return p.savePlayers()
	}
	profiles, err := p.Profiles()
	if err != nil {
		return nil, err
	}
	out := make([]arkobject.Player, 0, len(profiles))
	for _, profile := range profiles {
		player, err := profile.Player()
		if err != nil {
			return nil, err
		}
		out = append(out, player)
	}
	return out, nil
}

func (p *PlayerAPI) savePlayers() ([]arkobject.Player, error) {
	objects, err := p.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return isPlayerDataClass(info.ClassName)
	})
	if err != nil {
		return nil, err
	}
	out := make([]arkobject.Player, 0, len(objects))
	for _, info := range objects {
		player, err := arkobject.PlayerFromContainer(info.Object.Container())
		if err != nil {
			return nil, fmt.Errorf("parse player object %s: %w", info.UUID, err)
		}
		out = append(out, player)
	}
	return out, nil
}

func (p *PlayerAPI) PlayerByDataID(id uint64) (arkobject.Player, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, false, err
	}
	for _, player := range players {
		if player.PlayerDataID == id {
			return player, true, nil
		}
	}
	return arkobject.Player{}, false, nil
}

func (p *PlayerAPI) PlayerByUniqueID(id string) (arkobject.Player, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, false, err
	}
	for _, player := range players {
		if player.UniqueID == id {
			return player, true, nil
		}
	}
	return arkobject.Player{}, false, nil
}

func (p *PlayerAPI) PlayersByTribeID(tribeID int32) ([]arkobject.Player, error) {
	return p.filterPlayers(func(player arkobject.Player) bool {
		return player.TribeID == tribeID
	})
}

func (p *PlayerAPI) PlayersByCharacterName(name string) ([]arkobject.Player, error) {
	return p.filterPlayers(func(player arkobject.Player) bool {
		return player.CharacterName == name
	})
}

func (p *PlayerAPI) PlayersByPlayerName(name string) ([]arkobject.Player, error) {
	return p.filterPlayers(func(player arkobject.Player) bool {
		return player.PlayerName == name
	})
}

func (p *PlayerAPI) DeathsByPlayerID() (map[uint64]int32, error) {
	players, err := p.Players()
	if err != nil {
		return nil, err
	}
	out := map[uint64]int32{}
	for _, player := range players {
		out[player.PlayerDataID] = player.NumDeaths
	}
	return out, nil
}

func (p *PlayerAPI) TotalDeaths() (int32, error) {
	players, err := p.Players()
	if err != nil {
		return 0, err
	}
	var total int32
	for _, player := range players {
		total += player.NumDeaths
	}
	return total, nil
}

func (p *PlayerAPI) AverageDeaths() (float64, bool, error) {
	players, err := p.Players()
	if err != nil {
		return 0, false, err
	}
	if len(players) == 0 {
		return 0, false, nil
	}
	var total int32
	for _, player := range players {
		total += player.NumDeaths
	}
	return float64(total) / float64(len(players)), true, nil
}

func (p *PlayerAPI) PlayerWithMostDeaths() (arkobject.Player, int32, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.NumDeaths > best.NumDeaths {
			best = player
		}
	}
	return best, best.NumDeaths, true, nil
}

func (p *PlayerAPI) PlayerWithFewestDeaths() (arkobject.Player, int32, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.NumDeaths < best.NumDeaths {
			best = player
		}
	}
	return best, best.NumDeaths, true, nil
}

func (p *PlayerAPI) LevelsByPlayerID() (map[uint64]int32, error) {
	players, err := p.Players()
	if err != nil {
		return nil, err
	}
	out := map[uint64]int32{}
	for _, player := range players {
		out[player.PlayerDataID] = player.Level
	}
	return out, nil
}

func (p *PlayerAPI) TotalLevel() (int32, error) {
	players, err := p.Players()
	if err != nil {
		return 0, err
	}
	var total int32
	for _, player := range players {
		total += player.Level
	}
	return total, nil
}

func (p *PlayerAPI) ExperienceByPlayerID() (map[uint64]float64, error) {
	players, err := p.Players()
	if err != nil {
		return nil, err
	}
	out := map[uint64]float64{}
	for _, player := range players {
		out[player.PlayerDataID] = player.Experience
	}
	return out, nil
}

func (p *PlayerAPI) TotalExperience() (float64, error) {
	players, err := p.Players()
	if err != nil {
		return 0, err
	}
	var total float64
	for _, player := range players {
		total += player.Experience
	}
	return total, nil
}

func (p *PlayerAPI) AverageExperience() (float64, bool, error) {
	players, err := p.Players()
	if err != nil {
		return 0, false, err
	}
	if len(players) == 0 {
		return 0, false, nil
	}
	var total float64
	for _, player := range players {
		total += player.Experience
	}
	return total / float64(len(players)), true, nil
}

func (p *PlayerAPI) AverageLevel() (float64, bool, error) {
	players, err := p.Players()
	if err != nil {
		return 0, false, err
	}
	if len(players) == 0 {
		return 0, false, nil
	}
	var total int32
	for _, player := range players {
		total += player.Level
	}
	return float64(total) / float64(len(players)), true, nil
}

func (p *PlayerAPI) EngramPointsByPlayerID() (map[uint64]int32, error) {
	players, err := p.Players()
	if err != nil {
		return nil, err
	}
	out := map[uint64]int32{}
	for _, player := range players {
		out[player.PlayerDataID] = player.EngramPoints
	}
	return out, nil
}

func (p *PlayerAPI) TotalEngramPoints() (int32, error) {
	players, err := p.Players()
	if err != nil {
		return 0, err
	}
	var total int32
	for _, player := range players {
		total += player.EngramPoints
	}
	return total, nil
}

func (p *PlayerAPI) UnlockedEngrams() ([]string, error) {
	players, err := p.Players()
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	for _, player := range players {
		for _, engram := range player.UnlockedEngrams {
			if engram != "" {
				seen[engram] = true
			}
		}
	}
	out := make([]string, 0, len(seen))
	for engram := range seen {
		out = append(out, engram)
	}
	sort.Strings(out)
	return out, nil
}

func (p *PlayerAPI) PlayerWithHighestLevel() (arkobject.Player, int32, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.Level > best.Level {
			best = player
		}
	}
	return best, best.Level, true, nil
}

func (p *PlayerAPI) PlayerWithLowestLevel() (arkobject.Player, int32, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.Level < best.Level {
			best = player
		}
	}
	return best, best.Level, true, nil
}

func (p *PlayerAPI) PlayerWithHighestExperience() (arkobject.Player, float64, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.Experience > best.Experience {
			best = player
		}
	}
	return best, best.Experience, true, nil
}

func (p *PlayerAPI) PlayerWithLowestExperience() (arkobject.Player, float64, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.Experience < best.Experience {
			best = player
		}
	}
	return best, best.Experience, true, nil
}

func (p *PlayerAPI) PlayerWithMostEngramPoints() (arkobject.Player, int32, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.EngramPoints > best.EngramPoints {
			best = player
		}
	}
	return best, best.EngramPoints, true, nil
}

func (p *PlayerAPI) PlayerWithFewestEngramPoints() (arkobject.Player, int32, bool, error) {
	players, err := p.Players()
	if err != nil {
		return arkobject.Player{}, 0, false, err
	}
	if len(players) == 0 {
		return arkobject.Player{}, 0, false, nil
	}
	best := players[0]
	for _, player := range players[1:] {
		if player.EngramPoints < best.EngramPoints {
			best = player
		}
	}
	return best, best.EngramPoints, true, nil
}

func (p *PlayerAPI) filterPlayers(match func(arkobject.Player) bool) ([]arkobject.Player, error) {
	players, err := p.Players()
	if err != nil {
		return nil, err
	}
	out := make([]arkobject.Player, 0)
	for _, player := range players {
		if match(player) {
			out = append(out, player)
		}
	}
	return out, nil
}

func (p *PlayerAPI) Clusters() ([]*arkcluster.Data, error) {
	out := make([]*arkcluster.Data, 0, len(p.clusterPaths))
	for _, path := range p.clusterPaths {
		cluster, err := arkcluster.Open(path)
		if err != nil {
			return nil, err
		}
		out = append(out, cluster)
	}
	return out, nil
}

func (p *PlayerAPI) Tributes() ([]*arktribute.Data, error) {
	out := make([]*arktribute.Data, 0, len(p.tributePaths))
	for _, path := range p.tributePaths {
		tribute, err := arktribute.Open(path)
		if err != nil {
			return nil, err
		}
		out = append(out, tribute)
	}
	return out, nil
}

func (p *PlayerAPI) Tribes() ([]*arkprofile.TribeSave, error) {
	out := make([]*arkprofile.TribeSave, 0, len(p.tribePaths))
	for _, path := range p.tribePaths {
		tribe, err := arkprofile.OpenTribeSave(path)
		if err != nil {
			return nil, err
		}
		out = append(out, tribe)
	}
	return out, nil
}

func (p *PlayerAPI) TribeSummaries() ([]arkprofile.TribeSummary, error) {
	tribes, err := p.Tribes()
	if err != nil {
		return nil, err
	}
	out := make([]arkprofile.TribeSummary, 0, len(tribes))
	for _, tribe := range tribes {
		summary, err := tribe.Summary()
		if err != nil {
			return nil, err
		}
		out = append(out, summary)
	}
	return out, nil
}

func (p *PlayerAPI) TribeDetails() ([]arkobject.Tribe, error) {
	if p.save != nil {
		return p.saveTribeDetails()
	}
	tribes, err := p.Tribes()
	if err != nil {
		return nil, err
	}
	out := make([]arkobject.Tribe, 0, len(tribes))
	for _, tribe := range tribes {
		detail, err := tribe.Tribe()
		if err != nil {
			return nil, err
		}
		out = append(out, detail)
	}
	return out, nil
}

func (p *PlayerAPI) saveTribeDetails() ([]arkobject.Tribe, error) {
	objects, err := p.save.ParsedObjects(func(info arksave.ObjectClassInfo) bool {
		return isTribeDataClass(info.ClassName)
	})
	if err != nil {
		return nil, err
	}
	out := make([]arkobject.Tribe, 0, len(objects))
	for _, info := range objects {
		tribe, err := arkobject.TribeFromContainer(info.Object.Container())
		if err != nil {
			return nil, fmt.Errorf("parse tribe object %s: %w", info.UUID, err)
		}
		out = append(out, tribe)
	}
	return out, nil
}

func (p *PlayerAPI) TribeDinoCountsByID() (map[int32]int32, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return nil, err
	}
	out := map[int32]int32{}
	for _, tribe := range tribes {
		out[tribe.TribeID] = tribe.NumDinos
	}
	return out, nil
}

func (p *PlayerAPI) TotalTribeDinos() (int32, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return 0, err
	}
	var total int32
	for _, tribe := range tribes {
		total += tribe.NumDinos
	}
	return total, nil
}

func (p *PlayerAPI) AverageTribeDinos() (float64, bool, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return 0, false, err
	}
	if len(tribes) == 0 {
		return 0, false, nil
	}
	var total int32
	for _, tribe := range tribes {
		total += tribe.NumDinos
	}
	return float64(total) / float64(len(tribes)), true, nil
}

func (p *PlayerAPI) TribeWithMostDinos() (arkobject.Tribe, int32, bool, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return arkobject.Tribe{}, 0, false, err
	}
	if len(tribes) == 0 {
		return arkobject.Tribe{}, 0, false, nil
	}
	best := tribes[0]
	for _, tribe := range tribes[1:] {
		if tribe.NumDinos > best.NumDinos {
			best = tribe
		}
	}
	return best, best.NumDinos, true, nil
}

func (p *PlayerAPI) TribeWithFewestDinos() (arkobject.Tribe, int32, bool, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return arkobject.Tribe{}, 0, false, err
	}
	if len(tribes) == 0 {
		return arkobject.Tribe{}, 0, false, nil
	}
	best := tribes[0]
	for _, tribe := range tribes[1:] {
		if tribe.NumDinos < best.NumDinos {
			best = tribe
		}
	}
	return best, best.NumDinos, true, nil
}

func (p *PlayerAPI) TribeByID(id int32) (arkobject.Tribe, bool, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return arkobject.Tribe{}, false, err
	}
	for _, tribe := range tribes {
		if tribe.TribeID == id {
			return tribe, true, nil
		}
	}
	return arkobject.Tribe{}, false, nil
}

func (p *PlayerAPI) TribePlayerMap() (map[int32][]arkobject.Player, error) {
	players, err := p.Players()
	if err != nil {
		return nil, err
	}
	tribes, err := p.TribeDetails()
	if err != nil {
		return nil, err
	}
	byID := map[uint64]arkobject.Player{}
	for _, player := range players {
		byID[player.PlayerDataID] = player
	}
	out := map[int32][]arkobject.Player{}
	for _, tribe := range tribes {
		out[tribe.TribeID] = []arkobject.Player{}
		for _, memberID := range tribe.MemberIDs {
			player, ok := byID[uint64(memberID)]
			if ok {
				out[tribe.TribeID] = append(out[tribe.TribeID], player)
			}
		}
	}
	return out, nil
}

func (p *PlayerAPI) TribeOfPlayer(player arkobject.Player) (arkobject.Tribe, bool, error) {
	return p.TribeByID(player.TribeID)
}

func (p *PlayerAPI) ObjectOwnerByPlayerDataID(id uint64) (arkobject.ObjectOwner, bool, error) {
	player, ok, err := p.PlayerByDataID(id)
	if err != nil || !ok {
		return arkobject.ObjectOwner{}, false, err
	}
	tribe, ok, err := p.TribeOfPlayer(player)
	if err != nil || !ok {
		return arkobject.ObjectOwner{}, false, err
	}
	return arkobject.ObjectOwnerFromProfile(player, tribe), true, nil
}

func (p *PlayerAPI) DinoOwnerByPlayerDataID(id uint64) (arkobject.DinoOwner, bool, error) {
	player, ok, err := p.PlayerByDataID(id)
	if err != nil || !ok {
		return arkobject.DinoOwner{}, false, err
	}
	tribe, ok, err := p.TribeOfPlayer(player)
	if err != nil || !ok {
		return arkobject.DinoOwner{}, false, err
	}
	return arkobject.DinoOwnerFromProfile(player, tribe), true, nil
}

func (p *PlayerAPI) TribesByName(name string) ([]arkobject.Tribe, error) {
	return p.filterTribes(func(tribe arkobject.Tribe) bool {
		return tribe.Name == name
	})
}

func (p *PlayerAPI) TribesByOwnerID(ownerID int32) ([]arkobject.Tribe, error) {
	return p.filterTribes(func(tribe arkobject.Tribe) bool {
		return tribe.OwnerID == ownerID
	})
}

func (p *PlayerAPI) TribesByMemberName(name string) ([]arkobject.Tribe, error) {
	return p.filterTribes(func(tribe arkobject.Tribe) bool {
		return containsString(tribe.Members, name)
	})
}

func (p *PlayerAPI) TribesByMemberID(id int32) ([]arkobject.Tribe, error) {
	return p.filterTribes(func(tribe arkobject.Tribe) bool {
		return containsInt32(tribe.MemberIDs, id)
	})
}

func (p *PlayerAPI) filterTribes(match func(arkobject.Tribe) bool) ([]arkobject.Tribe, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return nil, err
	}
	out := make([]arkobject.Tribe, 0)
	for _, tribe := range tribes {
		if match(tribe) {
			out = append(out, tribe)
		}
	}
	return out, nil
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func containsInt32(values []int32, want int32) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func isPlayerDataClass(className string) bool {
	return strings.Contains(className, "PrimalPlayerDataBP")
}

func isTribeDataClass(className string) bool {
	return strings.Contains(className, "PrimalTribeData")
}
