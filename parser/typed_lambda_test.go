// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTypedLambdaReturnTypeAndEmptyParams(t *testing.T) {
	ctx := t.Context()
	src := `namespace n; policy p { export decision of r; rule r = { yield (): number => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl1.sentra").ParseProgram(ctx)
	require.NoError(t, err)
}

func TestParseTypedLambdaMalformedParamListErrors(t *testing.T) {
	ctx := t.Context()
	src := `namespace n; policy p { export decision of r; rule r = { yield (a number) => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl2.sentra").ParseProgram(ctx)
	require.Error(t, err)
}

func TestParseTypedLambdaMissingRParenErrors(t *testing.T) {
	ctx := t.Context()
	src := `namespace n; policy p { export decision of r; rule r = { yield (a: number => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl3.sentra").ParseProgram(ctx)
	require.Error(t, err)
}

func TestParseTypedLambdaOptionalBeforeRequiredRejected(t *testing.T) {
	ctx := t.Context()
	src := `namespace n; policy p { export decision of r; rule r = { yield (a?, b: number) => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl4.sentra").ParseProgram(ctx)
	require.Error(t, err)
}

func TestParseTypedLambdaDuplicateParamsRejected(t *testing.T) {
	ctx := t.Context()
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "adjacent typed duplicate",
			src:  `namespace n; policy p { export decision of r; rule r = { yield (a: number, a: number): number => { yield a } } }`,
		},
		{
			name: "non-adjacent typed duplicate",
			src:  `namespace n; policy p { export decision of r; rule r = { yield (a: number, b: number, a: number): number => { yield a } } }`,
		},
		{
			name: "duplicate optional params",
			src:  `namespace n; policy p { export decision of r; rule r = { yield (a?, a?: number): number => { yield a } } }`,
		},
		{
			name: "mixed typed and untyped duplicate",
			src:  `namespace n; policy p { export decision of r; rule r = { yield (a: number, a): number => { yield a } } }`,
		},
		{
			name: "duplicate in derive typed lambda",
			src: `namespace n
derive bad = (x: number, x: number): number => { yield x }
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewParserFromString(tc.src, tc.name+".sentra").ParseProgram(ctx)
			require.Error(t, err)
			require.Contains(t, err.Error(), "duplicate lambda parameter")
		})
	}
}
