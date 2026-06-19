package arkcluster

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/aipokalyptik/go-ark-save-parser/arkarchive"
)

type File struct {
	ID   string
	Path string
}

type Data struct {
	ID      string
	Path    string
	Archive *arkarchive.Archive
}

func Discover(dir string) ([]File, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]File, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isClusterFileName(name) {
			continue
		}
		files = append(files, File{ID: name, Path: filepath.Join(dir, name)})
	}
	sort.Slice(files, func(i int, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func Open(path string) (*Data, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	archive, err := arkarchive.Parse(raw, arkarchive.Options{FromStore: false, Format: arkarchive.FormatAuto})
	if err != nil {
		return nil, err
	}
	return &Data{ID: filepath.Base(path), Path: path, Archive: archive}, nil
}

func OpenDirectory(dir string) ([]*Data, error) {
	files, err := Discover(dir)
	if err != nil {
		return nil, err
	}
	out := make([]*Data, 0, len(files))
	for _, file := range files {
		data, err := Open(file.Path)
		if err != nil {
			return nil, err
		}
		out = append(out, data)
	}
	return out, nil
}

func isClusterFileName(name string) bool {
	if name == "" || strings.HasPrefix(name, ".") {
		return false
	}
	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case "":
		return isExtensionlessClusterID(name)
	default:
		return false
	}
}

func isExtensionlessClusterID(name string) bool {
	if strings.HasPrefix(name, "EOS_") && len(name) > len("EOS_") {
		for _, r := range name[len("EOS_"):] {
			if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-') {
				return false
			}
		}
		return true
	}
	if len(name) != 32 {
		return false
	}
	for _, r := range name {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
