package arkobject

import "strings"

func clusterDinoComponentClass(className string) bool {
	return strings.Contains(className, "CharacterStatus") ||
		strings.Contains(className, "DinoCharacterStatus") ||
		strings.Contains(className, "AIController") ||
		strings.Contains(className, "InventoryComponent")
}

func clusterClassNamesContaining(classNames []string, token string) []string {
	out := make([]string, 0)
	for _, className := range classNames {
		if strings.Contains(className, token) {
			out = append(out, className)
		}
	}
	return out
}

func clusterClassNamesContainingAny(classNames []string, tokens ...string) []string {
	out := make([]string, 0)
	for _, className := range classNames {
		for _, token := range tokens {
			if strings.Contains(className, token) {
				out = append(out, className)
				break
			}
		}
	}
	return out
}
