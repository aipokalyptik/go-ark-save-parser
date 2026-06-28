package arktribute

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/internal/safefile"
)

const DefaultMaxFileSize = 64 << 20

type Options struct {
	MaxFileSize int64
}

type File struct {
	ID   string
	Path string
}

type FileFault struct {
	Path string
	Err  error
}

type Data struct {
	ID            string
	Path          string
	PlayerDataIDs []uint64
	TribeDataIDs  []uint64
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
		if !isTributeFileName(name) {
			continue
		}
		files = append(files, File{ID: strings.TrimSuffix(name, filepath.Ext(name)), Path: filepath.Join(dir, name)})
	}
	sort.Slice(files, func(i int, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func Open(path string) (*Data, error) {
	return OpenWithOptions(path, Options{})
}

func OpenWithOptions(path string, opts Options) (*Data, error) {
	raw, err := safefile.ReadFile(path, maxFileSize(opts))
	if err != nil {
		return nil, err
	}
	playerIDs, tribeIDs, err := Parse(raw)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(path)
	return &Data{
		ID:            strings.TrimSuffix(name, filepath.Ext(name)),
		Path:          path,
		PlayerDataIDs: playerIDs,
		TribeDataIDs:  tribeIDs,
	}, nil
}

func OpenDirectory(dir string) ([]*Data, error) {
	return OpenDirectoryWithOptions(dir, Options{})
}

func OpenDirectoryWithOptions(dir string, opts Options) ([]*Data, error) {
	out, faults, err := OpenDirectoryWithFaultsOptions(dir, opts)
	if err != nil {
		return nil, err
	}
	if len(faults) > 0 {
		return nil, faults[0].Err
	}
	return out, nil
}

func OpenDirectoryWithFaults(dir string) ([]*Data, []FileFault, error) {
	return OpenDirectoryWithFaultsOptions(dir, Options{})
}

func OpenDirectoryWithFaultsOptions(dir string, opts Options) ([]*Data, []FileFault, error) {
	files, err := Discover(dir)
	if err != nil {
		return nil, nil, err
	}
	out := make([]*Data, 0, len(files))
	var faults []FileFault
	for _, file := range files {
		data, err := OpenWithOptions(file.Path, opts)
		if err != nil {
			faults = append(faults, FileFault{Path: file.Path, Err: err})
			continue
		}
		out = append(out, data)
	}
	return out, faults, nil
}

func Parse(raw []byte) ([]uint64, []uint64, error) {
	r := bytes.NewReader(raw)
	playerIDs, err := readIDList(r, "player data")
	if err != nil {
		return nil, nil, err
	}
	tribeIDs, err := readIDList(r, "tribe data")
	if err != nil {
		return nil, nil, err
	}
	if r.Len() != 0 {
		return nil, nil, fmt.Errorf("unexpected trailing tribute bytes: %d", r.Len())
	}
	return playerIDs, tribeIDs, nil
}

func readIDList(r *bytes.Reader, label string) ([]uint64, error) {
	var count int32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("missing %s count", label)
		}
		return nil, err
	}
	if count < 0 {
		return nil, fmt.Errorf("negative %s count %d", label, count)
	}
	if uint64(count) > uint64(r.Len()/8) {
		return nil, fmt.Errorf("%s count %d exceeds remaining data", label, count)
	}
	ids := make([]uint64, 0, count)
	for i := int32(0); i < count; i++ {
		var id uint64
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func isTributeFileName(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".arktributetribe" || ext == ".arktributetribetribe"
}

func maxFileSize(opts Options) int64 {
	if opts.MaxFileSize != 0 {
		return opts.MaxFileSize
	}
	return DefaultMaxFileSize
}
