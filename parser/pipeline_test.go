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
		{"value |> len", "len(value)"},
		{"value |> str.trim", "str.trim(value)"},
		{"value |> str.replaceAll(\" \", \"-\")", "str.replaceAll(value, \" \", \"-\")"},
		{"value |> mod.sub.fn", "mod.sub.fn(value)"},
		{"value |> len |> math.abs", "math.abs(len(value))"},
		{"value |> str.trim |> len", "len(str.trim(value))"},
		{"a + b |> len", "len((a + b))"},
		{"a ? b : c |> len", "len((a ? b : c))"},
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
			input:          "x |> f!",
			expectedCall:   "f(x)",
			expectMemoized: true,
			expectTTL:      nil,
		},
		{
			input:          "x |> f!10",
			expectedCall:   "f(x)",
			expectMemoized: true,
			expectTTL:      durationPtr(10 * time.Second),
		},
		{
			input:          "x |> mod.f!30",
			expectedCall:   "mod.f(x)",
			expectMemoized: true,
			expectTTL:      durationPtr(30 * time.Second),
		},
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
		{"value |> [1, 2]", true},
		{"value |> {\"k\": 1}", true},
		{"value |> foo[0]", true},
		{"value |> foo().bar", true},
		{"value |> foo().bar()", true},
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

func durationPtr(d time.Duration) *time.Duration {
	return &d
}
