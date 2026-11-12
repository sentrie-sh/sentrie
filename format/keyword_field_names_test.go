// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

// TestKeywordFieldNamesRoundTrip checks that keyword names after `.`, in a
// shape and after `attach` are printed unchanged and stably (#159).
func (s *FormatTestSuite) TestKeywordFieldNamesRoundTrip() {
	s.assertFormats(
		"namespace t\nshape S { version: string default?: number in: boolean }\npolicy p {\nfact d: S\nlet a = d.title\nlet b = d.default\nlet c = x.true.null\nlet e = 2 in [d.in]\nexport decision of r attach version as d.version attach default as 5\n}",
		"namespace t\n\nshape S {\n  version: string\n  default?: number\n  in: boolean\n}\n\npolicy p {\n  fact d: S\n  let a = d.title\n  let b = d.default\n  let c = x.true.null\n  let e = 2 in [d.in]\n  export decision of r attach version as d.version attach default as 5\n}\n",
	)
}
