package arkprofile

import (
	"fmt"
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
)

type PlayerProfile struct {
	Path    string
	Archive *arkarchive.Archive
}

type TribeSave struct {
	Path       string
	Archive    *arkarchive.Archive
	Properties map[string]any
}

type TribeSummary struct {
	Name     string
	OwnerID  uint32
	TribeID  int32
	Members  []string
	NumDinos int32
}

func OpenPlayerProfile(path string) (*PlayerProfile, error) {
	archive, err := openArchive(path)
	if err != nil {
		return nil, err
	}
	return &PlayerProfile{Path: path, Archive: archive}, nil
}

func OpenTribeSave(path string) (*TribeSave, error) {
	archive, err := openArchive(path)
	if err != nil {
		return nil, err
	}
	return &TribeSave{Path: path, Archive: archive}, nil
}

func (t *TribeSave) Summary() (TribeSummary, error) {
	raw, ok := t.Properties["TribeData"]
	if !ok {
		return TribeSummary{}, fmt.Errorf("missing TribeData")
	}
	data, ok := raw.(map[string]any)
	if !ok {
		return TribeSummary{}, fmt.Errorf("TribeData has type %T, want map[string]any", raw)
	}
	var summary TribeSummary
	if value, ok := data["TribeName"].(string); ok {
		summary.Name = value
	}
	if value, ok := data["OwnerPlayerDataId"].(uint32); ok {
		summary.OwnerID = value
	}
	if value, ok := data["TribeID"].(int32); ok {
		summary.TribeID = value
	}
	if value, ok := data["NumTribeDinos"].(int32); ok {
		summary.NumDinos = value
	}
	if values, ok := data["MembersPlayerName"].([]any); ok {
		for _, item := range values {
			if name, ok := item.(string); ok {
				summary.Members = append(summary.Members, name)
			}
		}
	}
	return summary, nil
}

func openArchive(path string) (*arkarchive.Archive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return arkarchive.Parse(data, arkarchive.Options{FromStore: true})
}
