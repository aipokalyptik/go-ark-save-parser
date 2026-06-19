package arkprofile

import (
	"fmt"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
	"github.com/aipokalyptik/go-ark-save-parser/arkobject"
	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/internal/safefile"
)

const DefaultMaxArchiveFileSize = 512 << 20

type Options struct {
	MaxFileSize int64
}

type PlayerProfile struct {
	Path       string
	Archive    *arkarchive.Archive
	Properties arkproperty.Container
}

type TribeSave struct {
	Path       string
	Archive    *arkarchive.Archive
	Properties arkproperty.Container
}

type TribeSummary struct {
	Name     string
	OwnerID  uint32
	TribeID  int32
	Members  []string
	NumDinos int32
}

func OpenPlayerProfile(path string) (*PlayerProfile, error) {
	return OpenPlayerProfileWithOptions(path, Options{})
}

func OpenPlayerProfileWithOptions(path string, opts Options) (*PlayerProfile, error) {
	archive, err := openArchive(path, opts)
	if err != nil {
		return nil, err
	}
	profile := &PlayerProfile{Path: path, Archive: archive}
	if len(archive.Objects) > 0 {
		profile.Properties = arkproperty.Container{Properties: archive.Objects[0].Properties}
	}
	return profile, nil
}

func (p *PlayerProfile) Player() (arkobject.Player, error) {
	return arkobject.PlayerFromContainer(p.Properties)
}

func OpenTribeSave(path string) (*TribeSave, error) {
	return OpenTribeSaveWithOptions(path, Options{})
}

func OpenTribeSaveWithOptions(path string, opts Options) (*TribeSave, error) {
	archive, err := openArchive(path, opts)
	if err != nil {
		return nil, err
	}
	tribe := &TribeSave{Path: path, Archive: archive}
	if len(archive.Objects) > 0 {
		tribe.Properties = arkproperty.Container{Properties: archive.Objects[0].Properties}
	}
	return tribe, nil
}

func (t *TribeSave) Summary() (TribeSummary, error) {
	raw, ok := t.Properties.Value("TribeData")
	if !ok {
		return TribeSummary{}, fmt.Errorf("missing TribeData")
	}
	data, ok := raw.(arkproperty.Container)
	if !ok {
		return TribeSummary{}, fmt.Errorf("TribeData has type %T, want arkproperty.Container", raw)
	}
	var summary TribeSummary
	if raw, ok := data.Value("TribeName"); ok {
		value, _ := raw.(string)
		summary.Name = value
	}
	if raw, ok := data.Value("OwnerPlayerDataId"); ok {
		switch value := raw.(type) {
		case uint32:
			summary.OwnerID = value
		case int32:
			summary.OwnerID = uint32(value)
		}
	}
	if raw, ok := data.Value("TribeID"); ok {
		value, _ := raw.(int32)
		summary.TribeID = value
	}
	if raw, ok := data.Value("NumTribeDinos"); ok {
		value, _ := raw.(int32)
		summary.NumDinos = value
	}
	if raw, ok := data.Value("MembersPlayerName"); ok {
		for _, item := range stringArrayValues(raw) {
			summary.Members = append(summary.Members, item)
		}
	}
	return summary, nil
}

func stringArrayValues(value any) []string {
	switch values := value.(type) {
	case []any:
		out := make([]string, 0, len(values))
		for _, item := range values {
			if name, ok := item.(string); ok {
				out = append(out, name)
			}
		}
		return out
	case arkproperty.Array:
		out := make([]string, 0, len(values.Values))
		for _, item := range values.Values {
			if name, ok := item.(string); ok {
				out = append(out, name)
			}
		}
		return out
	default:
		return nil
	}
}

func openArchive(path string, opts Options) (*arkarchive.Archive, error) {
	data, err := safefile.ReadFile(path, maxFileSize(opts))
	if err != nil {
		return nil, err
	}
	return arkarchive.Parse(data, arkarchive.Options{FromStore: true, Format: arkarchive.FormatAuto})
}

func maxFileSize(opts Options) int64 {
	if opts.MaxFileSize != 0 {
		return opts.MaxFileSize
	}
	return DefaultMaxArchiveFileSize
}
