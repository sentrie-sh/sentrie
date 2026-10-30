// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package tokens

import "testing"

// TestRangeStringPrintsBothLinesAlike checks that the single- and multi-line
// forms print lines and columns as stored, so a two-line range does not print
// its end one line early (#236).
func TestRangeStringPrintsBothLinesAlike(t *testing.T) {
	testCases := []struct {
		r    Range
		want string
	}{
		{NewRange("f", Pos{Line: 2, Column: 3}, Pos{Line: 2, Column: 7}), "f:2:3-7"},
		{NewRange("f", Pos{Line: 2, Column: 3}, Pos{Line: 4, Column: 2}), "f:2:3-4:2"},
		{NewRange("f", Pos{Line: 2, Column: 3}, Pos{Line: 2, Column: 3}), "f:2:3"},
		{NewRangeFromPos("f", Pos{Line: 1, Column: 1}), "f:1:1"},
		{Range{File: "f", From: Pos{Line: 5, Column: 2}}, "f:5:2"},
	}
	for _, tc := range testCases {
		if got := tc.r.String(); got != tc.want {
			t.Fatalf("%+v: got %q, want %q", tc.r, got, tc.want)
		}
	}
}
