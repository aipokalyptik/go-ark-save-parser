package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/aipokalyptik/go-ark-save-parser/arkproperty"
	"github.com/aipokalyptik/go-ark-save-parser/arksave"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: class_lookup <save.ark> <class-substring> [class-substring...]")
		os.Exit(2)
	}

	save, err := arksave.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open save: %v\n", err)
		os.Exit(1)
	}
	defer save.Close()

	substrings := os.Args[2:]
	infos, _, err := save.SelectedObjectPropertiesWithFaults(func(info arksave.ObjectClassInfo) bool {
		for _, substr := range substrings {
			if strings.Contains(info.ClassName, substr) {
				return true
			}
		}
		return false
	}, []string{"StructureID", "bIsEngram"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "lookup class substring: %v\n", err)
		os.Exit(1)
	}

	objects := 0
	matchedClasses := map[string]struct{}{}
	for _, info := range infos {
		container := arkproperty.Container{Properties: info.Properties}
		if _, ok := container.Value("StructureID"); !ok {
			continue
		}
		if _, ok := container.Value("bIsEngram"); ok {
			continue
		}
		objects++
		matchedClasses[info.ClassName] = struct{}{}
	}

	fmt.Printf("objects=%d classes=%d\n", objects, len(matchedClasses))
}
