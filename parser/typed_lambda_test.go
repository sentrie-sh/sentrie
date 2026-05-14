// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package parser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseTypedLambdaReturnTypeAndEmptyParams(t *testing.T) {
	ctx := context.Background()
	src := `namespace n; policy p { export decision of r; rule r = { yield (): number => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl1.sentra").ParseProgram(ctx)
	require.NoError(t, err)
}

func TestParseTypedLambdaMalformedParamListErrors(t *testing.T) {
	ctx := context.Background()
	src := `namespace n; policy p { export decision of r; rule r = { yield (a number) => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl2.sentra").ParseProgram(ctx)
	require.Error(t, err)
}

func TestParseTypedLambdaMissingRParenErrors(t *testing.T) {
	ctx := context.Background()
	src := `namespace n; policy p { export decision of r; rule r = { yield (a: number => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl3.sentra").ParseProgram(ctx)
	require.Error(t, err)
}

func TestParseTypedLambdaOptionalBeforeRequiredRejected(t *testing.T) {
	ctx := context.Background()
	src := `namespace n; policy p { export decision of r; rule r = { yield (a?, b: number) => { yield 1 } } }`
	_, err := NewParserFromString(src, "tl4.sentra").ParseProgram(ctx)
	require.Error(t, err)
}
