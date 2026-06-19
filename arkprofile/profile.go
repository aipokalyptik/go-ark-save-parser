package arkprofile

import (
	"os"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
)

type PlayerProfile struct {
	Path    string
	Archive *arkarchive.Archive
}

type TribeSave struct {
	Path    string
	Archive *arkarchive.Archive
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

func openArchive(path string) (*arkarchive.Archive, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return arkarchive.Parse(data, arkarchive.Options{FromStore: true})
}
