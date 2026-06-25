package safefile

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFileReadsRegularFileWithinLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(path, []byte("ark"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	data, err := ReadFile(path, 3)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "ark" {
		t.Fatalf("ReadFile() = %q, want ark", data)
	}
}

func TestReadFileRejectsFilesOverLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bin")
	if err := os.WriteFile(path, []byte("ark"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	_, err := ReadFile(path, 2)
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("ReadFile(over limit) error = %v, want ErrFileTooLarge", err)
	}
}

func TestReadFileRejectsDirectories(t *testing.T) {
	_, err := ReadFile(t.TempDir(), 0)
	if err == nil || !strings.Contains(err.Error(), "not a regular file") {
		t.Fatalf("ReadFile(directory) error = %v, want not a regular file", err)
	}
}
