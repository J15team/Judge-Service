package comparator

import "strings"

// StrictComparator compares output strictly (with trailing whitespace normalization)
type StrictComparator struct{}

// NewStrictComparator creates a new StrictComparator
func NewStrictComparator() *StrictComparator {
	return &StrictComparator{}
}

// Compare compares expected and actual output
func (c *StrictComparator) Compare(expected, actual string) bool {
	// Normalize trailing whitespace on each line and trailing newlines
	expected = normalizeOutput(expected)
	actual = normalizeOutput(actual)
	return expected == actual
}

func normalizeOutput(s string) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		result = append(result, strings.TrimRight(line, " \t\r"))
	}
	// Remove trailing empty lines
	for len(result) > 0 && result[len(result)-1] == "" {
		result = result[:len(result)-1]
	}
	return strings.Join(result, "\n")
}
