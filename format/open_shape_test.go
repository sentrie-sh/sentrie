// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
)

// TestOpenShapeMarkerRoundTrip checks that the `...` marker is printed on its
// own line after the fields and comments, and that formatting is stable (#145).
func (s *FormatTestSuite) TestOpenShapeMarkerRoundTrip() {
	s.assertFormats(
		"namespace t\nshape S { a: string b?: number ... }\nshape E { ... }\nshape C {\n a: string -- the a\n ... -- open\n}",
		"namespace t\n\nshape S {\n  a: string\n  b?: number\n  ...\n}\n\nshape E {\n  ...\n}\n\nshape C {\n  a: string -- the a\n  ... -- open\n}\n",
	)
}

// TestOpenShapeMarkerCountsInEqual checks that Equal tells an open shape from
// a closed one, and ignores where the marker sits in the source.
func (s *FormatTestSuite) TestOpenShapeMarkerCountsInEqual() {
	ctx := s.T().Context()
	parse := func(src string) *ast.Program {
		prog, err := parser.NewParserFromString(src, "case.sentrie").ParseProgram(ctx)
		s.Require().NoError(err)
		return prog
	}
	open := parse("namespace t\nshape S { a: string ... }\n")
	closed := parse("namespace t\nshape S { a: string }\n")
	spaced := parse("namespace t\nshape S {\n  a: string\n\n  ...\n}\n")

	commentBefore := parse("namespace t\nshape S {\n  a: string\n  -- c\n  ...\n}\n")
	commentAfter := parse("namespace t\nshape S {\n  a: string\n  ...\n  -- c\n}\n")

	s.Require().Error(Equal(open, closed))
	s.Require().Error(Equal(commentBefore, commentAfter), "a comment moved across the marker")
	s.Require().NoError(Equal(open, spaced))
}
