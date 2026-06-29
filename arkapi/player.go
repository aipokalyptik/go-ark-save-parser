package arkapi

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkprofile"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
	"github.com/google/uuid"
)

type PlayerAPI struct {
	save         *arksave.Save
	profilePaths []string
	tribePaths   []string
	clusterPaths []string
	tributePaths []string
}

type LocalFileSummary struct {
	Profiles int
	Tribes   int
	Clusters int
	Tributes int
}

type LocalProfileSummary struct {
	Files                LocalFileSummary
	HasParsedPlayers     bool
	ParsedPlayers        int
	HasParsedTribes      bool
	ParsedTribes         int
	HasTribePlayerLinks  bool
	TribePlayerLinks     int
	HasTotalDeaths       bool
	TotalDeaths          int32
	HasHighestLevel      bool
	HighestLevel         int32
	HasHighestExperience bool
	HighestExperience    float64
	HasAverageDeaths     bool
	AverageDeaths        float64
	HasAverageLevel      bool
	AverageLevel         float64
	HasAverageExperience bool
	AverageExperience    float64
	HasUnlockedEngrams   bool
	UnlockedEngrams      int
}

type LocalProfileFault struct {
	Operation string
	Err       error
}

type TribePlayerRelation struct {
	Tribe               arkobject.Tribe
	ActivePlayers       []arkobject.Player
	InactiveMemberIDs   []int32
	InactiveMemberNames []string
}

type TribePlayerRelationSummary struct {
	Players             int
	Tribes              int
	ActiveLinks         int
	InactiveMembers     int
	PlayersWithoutTribe int
	TribesWithInactive  int
	TribesWithoutActive int
}

type PlayerAndTribeDataSummary struct {
	Players             int               `json:"players"`
	Tribes              int               `json:"tribes"`
	PlayersWithNames    int               `json:"players_with_names"`
	TribesWithNames     int               `json:"tribes_with_names"`
	ActiveLinks         int               `json:"active_links"`
	InactiveMembers     int               `json:"inactive_members"`
	PlayersWithoutTribe int               `json:"players_without_tribe"`
	TribesWithInactive  int               `json:"tribes_with_inactive"`
	TribesWithoutActive int               `json:"tribes_without_active"`
	TribeRows           []TribeDataRow    `json:"tribe_rows"`
	PlayerRows          []PlayerDataRow   `json:"player_rows"`
	RelationRows        []RelationDataRow `json:"relation_rows"`
}

type PlayerDataRow struct {
	HasCharacterName bool  `json:"has_character_name"`
	HasPlayerName    bool  `json:"has_player_name"`
	Level            int32 `json:"level"`
	TribeID          int32 `json:"tribe_id"`
}

type TribeDataRow struct {
	HasName bool  `json:"has_name"`
	Members int   `json:"members"`
	Dinos   int32 `json:"dinos"`
}

type RelationDataRow struct {
	ActiveMembers   int `json:"active_members"`
	InactiveMembers int `json:"inactive_members"`
}

type PlayerInventorySummary struct {
	Players          int
	WithInventory    int
	WithoutInventory int
	TotalItems       int
	MaxItems         int
	MinItems         int
	AverageItems     float64
}

type PlayerInventoryLookup struct {
	PlayerDataID  uint64
	Found         bool
	InventoryUUID uuid.UUID
	Items         int
	HasLocation   bool
	Location      *arkobject.ActorTransform
}

type PlayerRosterSummary struct {
	Players      int
	WithNames    int
	HighestLevel int32
}

type PlayerAllSummary struct {
	Players         int
	Tribes          int
	HighestLevel    int32
	TotalDeaths     int32
	UnlockedEngrams int
}

type TribeRosterSummary struct {
	Tribes    int
	WithNames int
	Members   int
	Dinos     int32
}

type TribeDirectorySummary struct {
	Files           int
	Tribes          []arkobject.Tribe
	TotalDinos      int32
	HasAverageDinos bool
	AverageDinos    float64
}

func NewPlayer(save *arksave.Save) *PlayerAPI {
	return &PlayerAPI{save: save}
}

type PlayerPathFallback int

const (
	PlayerPathFallbackNone PlayerPathFallback = iota
	PlayerPathFallbackPlayers
	PlayerPathFallbackTribes
)

type PlayerPathOptions struct {
	Fallback PlayerPathFallback
}

func NewPlayerFromPath(path string, opts PlayerPathOptions) (*PlayerAPI, func() error, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, noopPlayerClose, err
	}
	if info.IsDir() {
		api, err := NewPlayerFromDirectory(path)
		return api, noopPlayerClose, err
	}

	save, err := arksave.Open(path)
	if err != nil {
		return nil, noopPlayerClose, err
	}
	api := NewPlayer(save)
	useDirectory := false
	switch opts.Fallback {
	case PlayerPathFallbackPlayers:
		players, faults, err := api.PlayersWithFaults()
		if err != nil {
			_ = save.Close()
			return nil, noopPlayerClose, err
		}
		if len(faults) > 0 {
			_ = save.Close()
			return nil, noopPlayerClose, faults[0].Err
		}
		useDirectory = len(players) == 0
	case PlayerPathFallbackTribes:
		tribes, faults, err := api.TribeDetailsWithFaults()
		if err != nil {
			_ = save.Close()
			return nil, noopPlayerClose, err
		}
		if len(faults) > 0 {
			_ = save.Close()
			return nil, noopPlayerClose, faults[0].Err
		}
		useDirectory = len(tribes) == 0
	}
	if useDirectory {
		_ = save.Close()
		api, err := NewPlayerFromDirectory(filepath.Dir(path))
		return api, noopPlayerClose, err
	}
	return api, save.Close, nil
}

func noopPlayerClose() error {
	return nil
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

func (p *PlayerAPI) LocalFileSummary() LocalFileSummary {
	return LocalFileSummary{
		Profiles: len(p.profilePaths),
		Tribes:   len(p.tribePaths),
		Clusters: len(p.clusterPaths),
		Tributes: len(p.tributePaths),
	}
}

func LocalProfileSummaryFromPath(path string) (LocalProfileSummary, []LocalProfileFault, error) {
	api, err := NewPlayerFromDirectory(path)
	if err != nil {
		return LocalProfileSummary{}, nil, err
	}
	summary, faults := api.LocalProfileSummary()
	return summary, faults, nil
}

func (p *PlayerAPI) LocalProfileSummary() (LocalProfileSummary, []LocalProfileFault) {
	summary := LocalProfileSummary{Files: p.LocalFileSummary()}
	var faults []LocalProfileFault

	players, err := p.Players()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "players", Err: err})
	} else {
		summary.HasParsedPlayers = true
		summary.ParsedPlayers = len(players)
	}

	tribes, err := p.TribeSummaries()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "tribes", Err: err})
	} else {
		summary.HasParsedTribes = true
		summary.ParsedTribes = len(tribes)
	}

	links, err := p.TribePlayerLinkCount()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "tribe player map", Err: err})
	} else {
		summary.HasTribePlayerLinks = true
		summary.TribePlayerLinks = links
	}

	totalDeaths, err := p.TotalDeaths()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "total deaths", Err: err})
	} else {
		summary.HasTotalDeaths = true
		summary.TotalDeaths = totalDeaths
	}

	_, level, ok, err := p.PlayerWithHighestLevel()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "highest level", Err: err})
	} else if ok {
		summary.HasHighestLevel = true
		summary.HighestLevel = level
	}

	_, experience, ok, err := p.PlayerWithHighestExperience()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "highest experience", Err: err})
	} else if ok {
		summary.HasHighestExperience = true
		summary.HighestExperience = experience
	}

	averageDeaths, ok, err := p.AverageDeaths()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "average deaths", Err: err})
	} else if ok {
		summary.HasAverageDeaths = true
		summary.AverageDeaths = averageDeaths
	}

	averageLevel, ok, err := p.AverageLevel()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "average level", Err: err})
	} else if ok {
		summary.HasAverageLevel = true
		summary.AverageLevel = averageLevel
	}

	averageExperience, ok, err := p.AverageExperience()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "average experience", Err: err})
	} else if ok {
		summary.HasAverageExperience = true
		summary.AverageExperience = averageExperience
	}

	unlockedEngrams, err := p.UnlockedEngrams()
	if err != nil {
		faults = append(faults, LocalProfileFault{Operation: "unlocked engrams", Err: err})
	} else {
		summary.HasUnlockedEngrams = true
		summary.UnlockedEngrams = len(unlockedEngrams)
	}

	return summary, faults
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

func (p *PlayerAPI) PlayersWithFaults() ([]arkobject.Player, []arksave.FaultyObjectInfo, error) {
	if p.save != nil {
		return p.savePlayersWithFaults()
	}
	out := make([]arkobject.Player, 0, len(p.profilePaths))
	var faults []arksave.FaultyObjectInfo
	for _, path := range p.profilePaths {
		profile, err := arkprofile.OpenPlayerProfile(path)
		if err != nil {
			faults = append(faults, localFileFault(path, err))
			continue
		}
		player, err := profile.Player()
		if err != nil {
			faults = append(faults, localFileFault(path, err))
			continue
		}
		if err := profile.PropertyError(); err != nil {
			faults = append(faults, localFileFault(path, err))
		}
		out = append(out, player)
	}
	return out, faults, nil
}

func (p *PlayerAPI) PlayerRosterSummary() (PlayerRosterSummary, error) {
	players, err := p.Players()
	if err != nil {
		return PlayerRosterSummary{}, err
	}
	return p.PlayerRosterSummaryForPlayers(players), nil
}

func (p *PlayerAPI) PlayerRosterSummaryWithFaults() (PlayerRosterSummary, []arksave.FaultyObjectInfo, error) {
	players, faults, err := p.PlayersWithFaults()
	if err != nil {
		return PlayerRosterSummary{}, nil, err
	}
	return p.PlayerRosterSummaryForPlayers(players), faults, nil
}

func PlayerRosterSummaryFromPath(path string) (PlayerRosterSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewPlayerFromPath(path, PlayerPathOptions{Fallback: PlayerPathFallbackPlayers})
	if err != nil {
		return PlayerRosterSummary{}, nil, err
	}
	summary, faults, err := api.PlayerRosterSummaryWithFaults()
	if closeErr := closeAPI(); err == nil && closeErr != nil {
		err = closeErr
	}
	return summary, faults, err
}

func (p *PlayerAPI) PlayerRosterSummaryForPlayers(players []arkobject.Player) PlayerRosterSummary {
	summary := PlayerRosterSummary{Players: len(players)}
	for _, player := range players {
		if player.HasName() {
			summary.WithNames++
		}
		if player.Level > summary.HighestLevel {
			summary.HighestLevel = player.Level
		}
	}
	return summary
}

func (p *PlayerAPI) PlayerAllSummary() (PlayerAllSummary, error) {
	players, err := p.Players()
	if err != nil {
		return PlayerAllSummary{}, err
	}
	tribes, err := p.TribeDetails()
	if err != nil {
		return PlayerAllSummary{}, err
	}
	return p.PlayerAllSummaryForData(players, tribes), nil
}

func (p *PlayerAPI) PlayerAllSummaryWithFaults() (PlayerAllSummary, []arksave.FaultyObjectInfo, error) {
	players, playerFaults, err := p.PlayersWithFaults()
	if err != nil {
		return PlayerAllSummary{}, nil, err
	}
	tribes, tribeFaults, err := p.TribeDetailsWithFaults()
	if err != nil {
		return PlayerAllSummary{}, nil, err
	}
	faults := append([]arksave.FaultyObjectInfo{}, playerFaults...)
	faults = append(faults, tribeFaults...)
	return p.PlayerAllSummaryForData(players, tribes), faults, nil
}

func PlayerAllSummaryFromPath(path string) (PlayerAllSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewPlayerFromPath(path, PlayerPathOptions{Fallback: PlayerPathFallbackPlayers})
	if err != nil {
		return PlayerAllSummary{}, nil, err
	}
	summary, faults, err := api.PlayerAllSummaryWithFaults()
	if closeErr := closeAPI(); err == nil && closeErr != nil {
		err = closeErr
	}
	return summary, faults, err
}

func (p *PlayerAPI) PlayerAllSummaryForData(players []arkobject.Player, tribes []arkobject.Tribe) PlayerAllSummary {
	summary := PlayerAllSummary{
		Players: len(players),
		Tribes:  len(tribes),
	}
	engrams := map[string]struct{}{}
	for _, player := range players {
		if player.Level > summary.HighestLevel {
			summary.HighestLevel = player.Level
		}
		summary.TotalDeaths += player.NumDeaths
		for _, engram := range player.UnlockedEngrams {
			engrams[engram] = struct{}{}
		}
	}
	summary.UnlockedEngrams = len(engrams)
	return summary
}

func (p *PlayerAPI) savePlayers() ([]arkobject.Player, error) {
	players, faults, err := p.savePlayersWithFaults()
	if err != nil {
		return nil, err
	}
	if len(faults) > 0 {
		return nil, faults[0].Err
	}
	return players, nil
}

func (p *PlayerAPI) savePlayersWithFaults() ([]arkobject.Player, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := p.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isPlayerDataClass(info.ClassName)
	})
	if err != nil {
		return nil, nil, err
	}
	out := make([]arkobject.Player, 0, len(objects))
	for _, info := range objects {
		player, err := arkobject.PlayerFromContainer(info.Object.Container())
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: info.UUID, ClassName: info.ClassName, Err: fmt.Errorf("parse player object %s: %w", info.UUID, err)})
			continue
		}
		out = append(out, player)
	}
	embedded, embeddedFaults, err := p.saveEmbeddedPlayersWithFaults()
	if err != nil {
		return nil, nil, err
	}
	out = mergePlayersByDataID(out, embedded)
	faults = append(faults, embeddedFaults...)
	return out, faults, nil
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

func (p *PlayerAPI) PlayerPawnByDataID(id uint64) (*arkobject.GameObject, bool, error) {
	if p.save == nil {
		return nil, false, nil
	}
	objects, faults, err := p.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return strings.Contains(info.ClassName, "PlayerPawn")
	})
	if err != nil {
		return nil, false, err
	}
	for _, info := range objects {
		value, ok := info.Object.Value("LinkedPlayerDataID")
		if !ok {
			continue
		}
		if numericPropertyAsUint64(value) == id {
			return info.Object, true, nil
		}
	}
	if len(faults) > 0 {
		return nil, false, fmt.Errorf("parse player pawn candidate %s: %w", faults[0].UUID, faults[0].Err)
	}
	return nil, false, nil
}

func (p *PlayerAPI) PlayerInventoryByDataID(id uint64) (*arkobject.Inventory, bool, error) {
	if p.save == nil {
		return nil, false, nil
	}
	pawn, ok, err := p.PlayerPawnByDataID(id)
	if err != nil || !ok {
		return nil, ok, err
	}
	value, ok := pawn.Value("MyInventoryComponent")
	if !ok {
		return nil, false, nil
	}
	inventoryID, ok := objectReferencePropertyUUID(value)
	if !ok {
		return nil, false, nil
	}
	object, err := p.save.Object(inventoryID)
	if err != nil {
		return nil, false, err
	}
	inventory := arkobject.InventoryFromObject(object)
	return &inventory, true, nil
}

func (p *PlayerAPI) PlayerInventoriesWithFaults() (map[uint64]arkobject.Inventory, []arksave.FaultyObjectInfo, error) {
	out := map[uint64]arkobject.Inventory{}
	if p.save == nil {
		return out, nil, nil
	}
	pawns, _, err := p.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return strings.Contains(info.ClassName, "PlayerPawn")
	})
	if err != nil {
		return nil, nil, err
	}
	var faults []arksave.FaultyObjectInfo
	for _, pawn := range pawns {
		value, ok := pawn.Object.Value("LinkedPlayerDataID")
		if !ok {
			continue
		}
		dataID := numericPropertyAsUint64(value)
		if dataID == 0 {
			continue
		}
		if _, exists := out[dataID]; exists {
			continue
		}
		inventoryValue, ok := pawn.Object.Value("MyInventoryComponent")
		if !ok {
			continue
		}
		inventoryID, ok := objectReferencePropertyUUID(inventoryValue)
		if !ok {
			continue
		}
		object, err := p.save.Object(inventoryID)
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: inventoryID, Err: err})
			continue
		}
		out[dataID] = arkobject.InventoryFromObject(object)
	}
	return out, faults, nil
}

func (p *PlayerAPI) InventoryItemCount(inventory arkobject.Inventory) int {
	return inventory.NumberOfItems()
}

func (p *PlayerAPI) PlayerInventorySummaryForPlayers(players []arkobject.Player) (PlayerInventorySummary, []arksave.FaultyObjectInfo, error) {
	if p.save == nil {
		return PlayerInventorySummary{}, nil, fmt.Errorf("player inventory summary requires a save-backed PlayerAPI")
	}
	inventories, faults, err := p.PlayerInventoriesWithFaults()
	if err != nil {
		return PlayerInventorySummary{}, nil, err
	}
	out := PlayerInventorySummary{Players: len(players)}
	for i, player := range players {
		items := 0
		inventory, ok := inventories[player.PlayerDataID]
		if ok {
			out.WithInventory++
			items = p.InventoryItemCount(inventory)
		} else {
			out.WithoutInventory++
		}
		if i == 0 || items < out.MinItems {
			out.MinItems = items
		}
		if items > out.MaxItems {
			out.MaxItems = items
		}
		out.TotalItems += items
	}
	if len(players) > 0 {
		out.AverageItems = float64(out.TotalItems) / float64(len(players))
	}
	return out, faults, nil
}

func PlayerInventorySummaryFromPath(path string) (PlayerInventorySummary, []arksave.FaultyObjectInfo, error) {
	save, err := arksave.Open(path)
	if err != nil {
		return PlayerInventorySummary{}, nil, err
	}
	defer save.Close()

	api := NewPlayer(save)
	players, _, err := api.PlayersWithFaults()
	if err != nil {
		return PlayerInventorySummary{}, nil, err
	}
	if len(players) == 0 {
		directoryAPI, err := NewPlayerFromDirectory(filepath.Dir(path))
		if err != nil {
			return PlayerInventorySummary{}, nil, err
		}
		players, err = directoryAPI.Players()
		if err != nil {
			return PlayerInventorySummary{}, nil, err
		}
	}
	return api.PlayerInventorySummaryForPlayers(players)
}

func PlayerInventoryLookupFromPath(path string, playerDataID uint64) (PlayerInventoryLookup, error) {
	save, err := arksave.Open(path)
	if err != nil {
		return PlayerInventoryLookup{}, err
	}
	defer save.Close()

	api := NewPlayer(save)
	lookup := PlayerInventoryLookup{PlayerDataID: playerDataID}
	inventory, ok, err := api.PlayerInventoryByDataID(playerDataID)
	if err != nil || !ok {
		return lookup, err
	}
	lookup.Found = true
	lookup.InventoryUUID = inventory.UUID
	lookup.Items = inventory.NumberOfItems()
	location, hasLocation, err := api.PlayerLocationByDataID(playerDataID)
	if err != nil {
		return PlayerInventoryLookup{}, err
	}
	lookup.HasLocation = hasLocation
	lookup.Location = location
	return lookup, nil
}

func (p *PlayerAPI) PlayerLocationByDataID(id uint64) (*arkobject.ActorTransform, bool, error) {
	pawn, ok, err := p.PlayerPawnByDataID(id)
	if err != nil || !ok {
		return nil, ok, err
	}
	value, ok := pawn.Value("SavedBaseWorldLocation")
	if !ok {
		return nil, false, nil
	}
	switch v := value.(type) {
	case arkproperty.Vector:
		return &arkobject.ActorTransform{X: v.X, Y: v.Y, Z: v.Z}, true, nil
	case arkobject.ActorTransform:
		return &v, true, nil
	case *arkobject.ActorTransform:
		if v == nil {
			return nil, false, nil
		}
		return v, true, nil
	default:
		return nil, false, nil
	}
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

func PlayerUnlockedEngramsFromPath(path string) ([]string, error) {
	api, closeAPI, err := NewPlayerFromPath(path, PlayerPathOptions{Fallback: PlayerPathFallbackPlayers})
	if err != nil {
		return nil, err
	}
	engrams, err := api.UnlockedEngrams()
	if closeErr := closeAPI(); err == nil && closeErr != nil {
		err = closeErr
	}
	return engrams, err
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

func numericPropertyAsUint64(value any) uint64 {
	switch v := value.(type) {
	case uint64:
		return v
	case uint32:
		return uint64(v)
	case int64:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int32:
		if v < 0 {
			return 0
		}
		return uint64(v)
	case int:
		if v < 0 {
			return 0
		}
		return uint64(v)
	default:
		return 0
	}
}

func objectReferencePropertyUUID(value any) (uuid.UUID, bool) {
	ref, ok := value.(arkproperty.ObjectReference)
	if !ok {
		return uuid.Nil, false
	}
	switch v := ref.Value.(type) {
	case uuid.UUID:
		return v, true
	case string:
		id, err := uuid.Parse(v)
		return id, err == nil
	default:
		return uuid.Nil, false
	}
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

func (p *PlayerAPI) TribeDetailsWithFaults() ([]arkobject.Tribe, []arksave.FaultyObjectInfo, error) {
	if p.save != nil {
		return p.saveTribeDetailsWithFaults()
	}
	out := make([]arkobject.Tribe, 0, len(p.tribePaths))
	var faults []arksave.FaultyObjectInfo
	for _, path := range p.tribePaths {
		tribeSave, err := arkprofile.OpenTribeSave(path)
		if err != nil {
			faults = append(faults, localFileFault(path, err))
			continue
		}
		tribe, err := tribeSave.Tribe()
		if err != nil {
			faults = append(faults, localFileFault(path, err))
			continue
		}
		if err := tribeSave.PropertyError(); err != nil {
			faults = append(faults, localFileFault(path, err))
		}
		out = append(out, tribe)
	}
	return out, faults, nil
}

func localFileFault(path string, err error) arksave.FaultyObjectInfo {
	return arksave.FaultyObjectInfo{ClassName: path, Err: err}
}

func (p *PlayerAPI) TribeRosterSummary() (TribeRosterSummary, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return TribeRosterSummary{}, err
	}
	return p.TribeRosterSummaryForTribes(tribes), nil
}

func (p *PlayerAPI) TribeRosterSummaryWithFaults() (TribeRosterSummary, []arksave.FaultyObjectInfo, error) {
	tribes, faults, err := p.TribeDetailsWithFaults()
	if err != nil {
		return TribeRosterSummary{}, nil, err
	}
	return p.TribeRosterSummaryForTribes(tribes), faults, nil
}

func TribeRosterSummaryFromPath(path string) (TribeRosterSummary, []arksave.FaultyObjectInfo, error) {
	api, closeAPI, err := NewPlayerFromPath(path, PlayerPathOptions{Fallback: PlayerPathFallbackTribes})
	if err != nil {
		return TribeRosterSummary{}, nil, err
	}
	summary, faults, err := api.TribeRosterSummaryWithFaults()
	if closeErr := closeAPI(); err == nil && closeErr != nil {
		err = closeErr
	}
	return summary, faults, err
}

func TribeDirectorySummaryFromPath(path string) (TribeDirectorySummary, error) {
	api, err := NewPlayerFromDirectory(path)
	if err != nil {
		return TribeDirectorySummary{}, err
	}
	return api.TribeDirectorySummary()
}

func (p *PlayerAPI) TribeDirectorySummary() (TribeDirectorySummary, error) {
	tribes, err := p.TribeDetails()
	if err != nil {
		return TribeDirectorySummary{}, err
	}
	totalDinos := int32(0)
	for _, tribe := range tribes {
		totalDinos += tribe.NumDinos
	}
	summary := TribeDirectorySummary{
		Files:      len(p.TribePaths()),
		Tribes:     tribes,
		TotalDinos: totalDinos,
	}
	if len(tribes) > 0 {
		summary.HasAverageDinos = true
		summary.AverageDinos = float64(totalDinos) / float64(len(tribes))
	}
	return summary, nil
}

func (p *PlayerAPI) TribeRosterSummaryForTribes(tribes []arkobject.Tribe) TribeRosterSummary {
	summary := TribeRosterSummary{Tribes: len(tribes)}
	for _, tribe := range tribes {
		if tribe.HasName() {
			summary.WithNames++
		}
		summary.Members += tribe.MemberCount()
		summary.Dinos += tribe.NumDinos
	}
	return summary
}

func (p *PlayerAPI) saveTribeDetails() ([]arkobject.Tribe, error) {
	tribes, faults, err := p.saveTribeDetailsWithFaults()
	if err != nil {
		return nil, err
	}
	if len(faults) > 0 {
		return nil, faults[0].Err
	}
	return tribes, nil
}

func (p *PlayerAPI) saveTribeDetailsWithFaults() ([]arkobject.Tribe, []arksave.FaultyObjectInfo, error) {
	objects, faults, err := p.save.ParsedObjectsWithFaults(func(info arksave.ObjectClassInfo) bool {
		return isTribeDataClass(info.ClassName)
	})
	if err != nil {
		return nil, nil, err
	}
	out := make([]arkobject.Tribe, 0, len(objects))
	for _, info := range objects {
		tribe, err := arkobject.TribeFromContainer(info.Object.Container())
		if err != nil {
			faults = append(faults, arksave.FaultyObjectInfo{UUID: info.UUID, ClassName: info.ClassName, Err: fmt.Errorf("parse tribe object %s: %w", info.UUID, err)})
			continue
		}
		out = append(out, tribe)
	}
	embedded, embeddedFaults, err := p.saveEmbeddedTribeDetailsWithFaults()
	if err != nil {
		return nil, nil, err
	}
	out = mergeTribesByID(out, embedded)
	faults = append(faults, embeddedFaults...)
	return out, faults, nil
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

func (p *PlayerAPI) TribePlayerLinkCount() (int, error) {
	tribePlayers, err := p.TribePlayerMap()
	if err != nil {
		return 0, err
	}
	links := 0
	for _, players := range tribePlayers {
		links += len(players)
	}
	return links, nil
}

func (p *PlayerAPI) TribePlayerRelations() ([]TribePlayerRelation, error) {
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
	out := make([]TribePlayerRelation, 0, len(tribes))
	for _, tribe := range tribes {
		relation := TribePlayerRelation{Tribe: tribe}
		for idx, memberID := range tribe.MemberIDs {
			player, ok := byID[uint64(memberID)]
			if ok {
				relation.ActivePlayers = append(relation.ActivePlayers, player)
				continue
			}
			relation.InactiveMemberIDs = append(relation.InactiveMemberIDs, memberID)
			if idx < len(tribe.Members) {
				relation.InactiveMemberNames = append(relation.InactiveMemberNames, tribe.Members[idx])
			}
		}
		out = append(out, relation)
	}
	return out, nil
}

func (p *PlayerAPI) TribePlayerRelationSummary() (TribePlayerRelationSummary, error) {
	players, err := p.Players()
	if err != nil {
		return TribePlayerRelationSummary{}, err
	}
	tribes, err := p.TribeDetails()
	if err != nil {
		return TribePlayerRelationSummary{}, err
	}
	relations, err := p.TribePlayerRelations()
	if err != nil {
		return TribePlayerRelationSummary{}, err
	}
	return p.TribePlayerRelationSummaryForData(players, tribes, relations), nil
}

func TribePlayerRelationSummaryFromPath(path string) (TribePlayerRelationSummary, error) {
	api, closeAPI, err := NewPlayerFromPath(path, PlayerPathOptions{Fallback: PlayerPathFallbackPlayers})
	if err != nil {
		return TribePlayerRelationSummary{}, err
	}
	summary, err := api.TribePlayerRelationSummary()
	if closeErr := closeAPI(); err == nil && closeErr != nil {
		err = closeErr
	}
	return summary, err
}

func (p *PlayerAPI) TribePlayerRelationSummaryForData(players []arkobject.Player, tribes []arkobject.Tribe, relations []TribePlayerRelation) TribePlayerRelationSummary {
	summary := TribePlayerRelationSummary{
		Players: len(players),
		Tribes:  len(tribes),
	}
	tribeIDs := map[int32]struct{}{}
	for _, tribe := range tribes {
		tribeIDs[tribe.TribeID] = struct{}{}
	}
	for _, player := range players {
		if _, ok := tribeIDs[player.TribeID]; !ok {
			summary.PlayersWithoutTribe++
		}
	}
	for _, relation := range relations {
		summary.ActiveLinks += len(relation.ActivePlayers)
		summary.InactiveMembers += len(relation.InactiveMemberIDs)
		if len(relation.InactiveMemberIDs) > 0 {
			summary.TribesWithInactive++
		}
		if len(relation.ActivePlayers) == 0 {
			summary.TribesWithoutActive++
		}
	}
	return summary
}

func (p *PlayerAPI) PlayerAndTribeDataSummary() (PlayerAndTribeDataSummary, error) {
	players, err := p.Players()
	if err != nil {
		return PlayerAndTribeDataSummary{}, err
	}
	tribes, err := p.TribeDetails()
	if err != nil {
		return PlayerAndTribeDataSummary{}, err
	}
	relations, err := p.TribePlayerRelations()
	if err != nil {
		return PlayerAndTribeDataSummary{}, err
	}
	return p.PlayerAndTribeDataSummaryForData(players, tribes, relations), nil
}

func PlayerAndTribeDataSummaryFromPath(path string) (PlayerAndTribeDataSummary, error) {
	api, closeAPI, err := NewPlayerFromPath(path, PlayerPathOptions{Fallback: PlayerPathFallbackPlayers})
	if err != nil {
		return PlayerAndTribeDataSummary{}, err
	}
	summary, err := api.PlayerAndTribeDataSummary()
	if closeErr := closeAPI(); err == nil && closeErr != nil {
		err = closeErr
	}
	return summary, err
}

func (p *PlayerAPI) PlayerAndTribeDataSummaryForData(players []arkobject.Player, tribes []arkobject.Tribe, relations []TribePlayerRelation) PlayerAndTribeDataSummary {
	out := PlayerAndTribeDataSummary{
		Players:      len(players),
		Tribes:       len(tribes),
		PlayerRows:   make([]PlayerDataRow, 0, len(players)),
		TribeRows:    make([]TribeDataRow, 0, len(tribes)),
		RelationRows: make([]RelationDataRow, 0, len(relations)),
	}
	tribeIDs := map[int32]struct{}{}
	for _, tribe := range tribes {
		tribeIDs[tribe.TribeID] = struct{}{}
		row := TribeDataRow{HasName: tribe.HasName(), Members: tribe.MemberCount(), Dinos: tribe.NumDinos}
		if row.HasName {
			out.TribesWithNames++
		}
		out.TribeRows = append(out.TribeRows, row)
	}
	for _, player := range players {
		row := PlayerDataRow{
			HasCharacterName: player.CharacterName != "",
			HasPlayerName:    player.PlayerName != "",
			Level:            player.Level,
			TribeID:          player.TribeID,
		}
		if player.HasName() {
			out.PlayersWithNames++
		}
		if !player.InTribe() {
			out.PlayersWithoutTribe++
		} else if _, ok := tribeIDs[player.TribeID]; !ok {
			out.PlayersWithoutTribe++
		}
		out.PlayerRows = append(out.PlayerRows, row)
	}
	for _, relation := range relations {
		row := RelationDataRow{ActiveMembers: len(relation.ActivePlayers), InactiveMembers: len(relation.InactiveMemberIDs)}
		out.ActiveLinks += row.ActiveMembers
		out.InactiveMembers += row.InactiveMembers
		if row.InactiveMembers > 0 {
			out.TribesWithInactive++
		}
		if row.ActiveMembers == 0 {
			out.TribesWithoutActive++
		}
		out.RelationRows = append(out.RelationRows, row)
	}

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
	return out
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
