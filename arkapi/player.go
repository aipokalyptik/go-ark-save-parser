package arkapi

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkprofile"
	"github.com/aipokalyptik/go-ark-save-parser/arktribute"
)

type PlayerAPI struct {
	profilePaths []string
	tribePaths   []string
	clusterPaths []string
	tributePaths []string
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
