package arkapi

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkcluster"
	"github.com/aipokalyptik/go-ark-save-parser/arkprofile"
)

type PlayerAPI struct {
	profilePaths []string
	tribePaths   []string
	clusterPaths []string
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
	sort.Strings(api.profilePaths)
	sort.Strings(api.tribePaths)
	sort.Strings(api.clusterPaths)
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
