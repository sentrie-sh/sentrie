// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package parser

import (
	"time"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *ParserTestSuite) TestPipelineExpressionLowering() {
	testCases := []struct {
		input    string
		expected string
	}{
		{"value |> len()", "len(value)"},
		{"value |> str.trim()", "str.trim(value)"},
		{"value |> str.replaceAll(\" \", \"-\")", "str.replaceAll(value, \" \", \"-\")"},
		{"value |> mod.sub.fn()", "mod.sub.fn(value)"},
		{"value |> len() |> math.abs()", "math.abs(len(value))"},
		{"value |> str.trim() |> len()", "len(str.trim(value))"},
		{"needle |> str.replace(haystack, #, \"$$\")", "str.replace(haystack, needle, \"$$\")"},
		{"x |> f(1, #)", "f(1, x)"},
		{"x |> f(#, 2)", "f(x, 2)"},
		{"x |> f(#, #)", "f(x, x)"},
		{"x |> f(1, #) |> g(#)", "g(f(1, x))"},
		{"x |> f(g(#))", "f(g(x))"},
		{"a + b |> len()", "len((a + b))"},
		{"a ? b : c |> len()", "len((a ? b : c))"},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "expression should parse: %s", tc.input)
		s.Nil(parser.err, "expression should not produce parser error: %s", tc.input)
		if expr == nil {
			continue
		}
		s.Equal(tc.expected, expr.String(), "unexpected lowering for: %s", tc.input)
	}
}

func (s *ParserTestSuite) TestPipelineExpressionMemoizationPreserved() {
	testCases := []struct {
		input          string
		expectedCall   string
		expectMemoized bool
		expectTTL      *time.Duration
	}{
		{
			input:          "x |> f()!",
			expectedCall:   "f(x)",
			expectMemoized: true,
			expectTTL:      nil,
		},
		{
			input:          "x |> f()!10",
			expectedCall:   "f(x)",
			expectMemoized: true,
			expectTTL:      durationPtr(10 * time.Second),
		},
		{
			input:          "x |> mod.f(1)!5",
			expectedCall:   "mod.f(x, 1)",
			expectMemoized: true,
			expectTTL:      durationPtr(5 * time.Second),
		},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.NotNil(expr, "expression should parse: %s", tc.input)
		s.Nil(parser.err, "expression should not produce parser error: %s", tc.input)

		call, ok := expr.(*ast.CallExpression)
		s.True(ok, "expected call expression for: %s", tc.input)
		s.Equal(tc.expectedCall, call.String(), "unexpected lowered call for: %s", tc.input)
		s.Equal(tc.expectMemoized, call.Memoized, "unexpected memoized flag for: %s", tc.input)

		if tc.expectTTL == nil {
			s.Nil(call.MemoizeTTL, "expected nil ttl for: %s", tc.input)
		} else {
			s.NotNil(call.MemoizeTTL, "expected ttl for: %s", tc.input)
			s.Equal(*tc.expectTTL, *call.MemoizeTTL, "unexpected ttl for: %s", tc.input)
		}
	}
}

func (s *ParserTestSuite) TestPipelineExpressionInvalidTargets() {
	testCases := []struct {
		input            string
		expectPipelineEr bool
	}{
		{"value |>", false},
		{"value |> (a + b)", true},
		{"value |> foo ? bar : baz", true},
		{"value |> len", true},
		{"value |> str.trim", true},
		{"value |> mod.sub.fn", true},
		{"value |> len |> math.abs", true},
		{"value |> str.trim |> len", true},
		{"x |> f!", true},
		{"x |> f!10", true},
		{"x |> mod.f!30", true},
		{"value |> [1, 2]", true},
		{"value |> {\"k\": 1}", true},
		{"value |> foo[0]", true},
		{"value |> foo().bar", true},
		{"value |> foo().bar()", true},
		{"value |> #", true},
		{"value |> (str.trim)", true},
	}

	for _, tc := range testCases {
		parser := NewParserFromString(tc.input, "test.sentra")
		expr := parser.parseExpression(s.T().Context(), LOWEST)
		s.Nil(expr, "expected parse failure for invalid pipeline target: %s", tc.input)
		s.NotNil(parser.err, "expected parser error for: %s", tc.input)
		if tc.expectPipelineEr {
			s.Contains(parser.err.Error(), "pipeline", "expected pipeline-specific error for: %s", tc.input)
		}
	}
}

func (s *ParserTestSuite) TestHasIdentifierRoot() {
	rng := tokens.BadRange("test.sentra")
	s.True(hasIdentifierRoot(ast.NewIdentifier("f", rng)))
	s.True(
		hasIdentifierRoot(
			ast.NewFieldAccessExpression(
				ast.NewIdentifier("mod", rng),
				"fn",
				rng,
			),
		),
	)
	s.False(
		hasIdentifierRoot(
			ast.NewFieldAccessExpression(
				ast.NewCallExpression(
					ast.NewIdentifier("factory", rng),
					nil,
					false,
					nil,
					rng,
				),
				"fn",
				rng,
			),
		),
	)
}

func (s *ParserTestSuite) TestPipelineHoleHelpers() {
	rng := tokens.BadRange("test.sentra")
	replacement := ast.NewIdentifier("x", rng)

	exprs := []struct {
		name     string
		input    ast.Expression
		expected string
	}{
		{
			name:     "call",
			input:    ast.NewCallExpression(ast.NewIdentifier("f", rng), []ast.Expression{ast.NewPipelineHoleExpression(rng)}, false, nil, rng),
			expected: "f(x)",
		},
		{
			name: "field_access_left",
			input: ast.NewFieldAccessExpression(
				ast.NewPipelineHoleExpression(rng),
				"trim",
				rng,
			),
			expected: "x.trim",
		},
		{
			name: "index_access",
			input: ast.NewIndexAccessExpression(
				ast.NewIdentifier("arr", rng),
				ast.NewPipelineHoleExpression(rng),
				rng,
			),
			expected: "arr[x]",
		},
		{
			name:     "list",
			input:    ast.NewListLiteral([]ast.Expression{ast.NewPipelineHoleExpression(rng)}, rng),
			expected: "[x]",
		},
		{
			name: "map",
			input: ast.NewMapLiteral([]ast.MapEntry{{
				Key:   ast.NewStringLiteral("k", rng),
				Value: ast.NewPipelineHoleExpression(rng),
			}}, rng),
			expected: "{k: x}",
		},
		{
			name:     "infix",
			input:    ast.NewInfixExpression(ast.NewPipelineHoleExpression(rng), ast.NewIntegerLiteral(1, rng), "+", rng),
			expected: "(x + 1)",
		},
		{
			name:     "unary",
			input:    ast.NewUnaryExpression("-", ast.NewPipelineHoleExpression(rng), rng),
			expected: "-x",
		},
		{
			name: "ternary",
			input: ast.NewTernaryExpression(
				ast.NewPipelineHoleExpression(rng),
				ast.NewIntegerLiteral(1, rng),
				ast.NewIntegerLiteral(0, rng),
				rng,
			),
			expected: "(x ? 1 : 0)",
		},
		{
			name:     "cast",
			input:    ast.NewCastExpression(ast.NewPipelineHoleExpression(rng), ast.NewNumberTypeRef(rng), rng),
			expected: "cast x as number",
		},
		{
			name:     "is_defined",
			input:    ast.NewIsDefinedExpression(ast.NewPipelineHoleExpression(rng), rng),
			expected: "is defined x",
		},
		{
			name:     "is_empty",
			input:    ast.NewIsEmptyExpression(ast.NewPipelineHoleExpression(rng), rng),
			expected: "is empty x",
		},
		{
			name:     "transform",
			input:    ast.NewTransformExpression(ast.NewPipelineHoleExpression(rng), "to_number", rng),
			expected: "transform  x to_number",
		},
		{
			name:     "preceding_comment",
			input:    ast.NewPrecedingCommentExpression("c", ast.NewPipelineHoleExpression(rng), rng),
			expected: "c -- x",
		},
		{
			name:     "trailing_comment",
			input:    ast.NewTrailingCommentExpression("c", ast.NewPipelineHoleExpression(rng), rng),
			expected: "x -- c",
		},
	}

	for _, tc := range exprs {
		s.True(containsPipelineHole(tc.input), tc.name)
		got := substitutePipelineHoles(tc.input, replacement)
		s.Equal(tc.expected, got.String(), tc.name)
		s.False(containsPipelineHole(got), tc.name)
	}

	s.False(containsPipelineHole(ast.NewIdentifier("plain", rng)))
	s.False(containsPipelineHoleInExprs([]ast.Expression{ast.NewIdentifier("plain", rng)}))
}

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
