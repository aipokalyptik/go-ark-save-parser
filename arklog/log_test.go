package arklog

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerFiltersByEnabledLevel(t *testing.T) {
	var out bytes.Buffer
	logger := New(&out)

	logger.SetLevel(API, true)
	logger.Info("info hidden")
	logger.API("api visible")
	logger.Save("save hidden")

	got := out.String()
	if strings.Contains(got, "info hidden") || strings.Contains(got, "save hidden") {
		t.Fatalf("disabled log output leaked: %q", got)
	}
	if !strings.Contains(got, "[api] api visible") {
		t.Fatalf("enabled API output missing from %q", got)
	}
}

func TestLoggerAllLevelEnablesEveryLevel(t *testing.T) {
	var out bytes.Buffer
	logger := New(&out)

	logger.SetLevel(All, true)
	logger.Info("info visible")
	logger.Error("error visible")
	logger.Parser("parser visible")

	got := out.String()
	for _, want := range []string{
		"[info] info visible",
		"[error] error visible",
		"[parser] parser visible",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("output %q missing %q", got, want)
		}
	}
}
