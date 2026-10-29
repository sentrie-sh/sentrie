// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

func (s *JSTestSuite) TestBuiltinJSTS_Embedded() {
	s.NotEmpty(BuiltinJSTS, "BuiltinJSTS should not be empty")
	s.Greater(len(BuiltinJSTS), 100, "BuiltinJSTS should contain substantial content")
	source := string(BuiltinJSTS)
	s.Contains(source, "export const round", "Should export Math.round")
	s.Contains(source, "export const length", "Should export String.length")
	s.Contains(source, "export const parse", "Should export JSON.parse")
	s.Contains(source, "export const stringify", "Should export JSON.stringify")
}

// TestBuiltinJSTS_NoNowExport guards #207: @sentrie/js must not re-export
// Date.now, an unpinned wall clock in milliseconds. @sentrie/time.now() is the
// pinned clock.
func (s *JSTestSuite) TestBuiltinJSTS_NoNowExport() {
	source := string(BuiltinJSTS)
	s.NotContains(source, "export const now", "@sentrie/js must not export now")
	s.NotContains(source, "Date.now", "@sentrie/js must not expose Date.now")
}
