package sandbox

import "strings"

func normalizeOutput(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, strings.TrimRight(line, " \t\r"))
	}
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return strings.Join(result, "\n")
}
