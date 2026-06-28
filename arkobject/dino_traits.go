package arkobject

import (
	"strconv"
	"strings"
)

func parseGeneTraits(values []string) []GeneTrait {
	if len(values) == 0 {
		return nil
	}
	out := make([]GeneTrait, 0, len(values))
	for _, raw := range values {
		out = append(out, parseGeneTrait(raw))
	}
	return out
}

func parseGeneTrait(raw string) GeneTrait {
	trait := GeneTrait{Raw: raw, Name: raw}
	open := strings.Index(raw, "[")
	if open < 0 || !strings.HasSuffix(raw, "]") {
		return trait
	}
	trait.Name = raw[:open]
	level, err := strconv.Atoi(raw[open+1 : len(raw)-1])
	if err == nil {
		trait.Level = level
	}
	return trait
}
