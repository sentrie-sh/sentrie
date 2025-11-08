package trinary

import (
	"slices"
)

// Parse parses a string into a trinary value.
// It converts:
//   - 1, t, T, TRUE, true, True           => True
//   - 0, f, F, FALSE, false, False        => False
//   - -1, n, N, UNKNOWN, unknown, Unknown => Unknown
//
// Any other value returns Unknown.
func Parse(s string) Value {
	trueValues := []string{"1", "t", "T", "TRUE", "true", "True"}
	falseValues := []string{"0", "f", "F", "FALSE", "false", "False"}
	unknownValues := []string{"-1", "n", "N", "UNKNOWN", "unknown", "Unknown"}

	if slices.Contains(trueValues, s) {
		return True
	}
	if slices.Contains(falseValues, s) {
		return False
	}
	if slices.Contains(unknownValues, s) {
		return Unknown
	}
	return Unknown
}
