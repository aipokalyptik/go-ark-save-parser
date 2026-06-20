package arkapi

import (
	"testing"

	"github.com/aipokalyptik/go-ark-save-parser/internal/testfixtures"
)

func createSyntheticArchive(t *testing.T, path string, className string) {
	t.Helper()
	testfixtures.WriteArchive(t, path, className)
}
