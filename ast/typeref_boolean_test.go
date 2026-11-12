// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package ast

import (
	"testing"

	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

// TestTrinaryTypeRefStringAndHints: boolean and trinary render under their own
// names, and the not-a-type table holds exactly bool, int and float (#144).
func TestTrinaryTypeRefStringAndHints(t *testing.T) {
	t.Parallel()
	tri := NewTrinaryTypeRef(tokens.Range{})
	require.Equal(t, "trinary", tri.String())
	tri.Boolean = true
	require.Equal(t, "boolean", tri.String())

	for name, want := range map[string]string{"bool": "boolean", "int": "number", "float": "number"} {
		hint, ok := NotATypeHint(name)
		require.True(t, ok, name)
		require.Equal(t, want, hint)
	}
	for _, name := range []string{"boolean", "number", "trinary", "integer", "Bool", ""} {
		_, ok := NotATypeHint(name)
		require.False(t, ok, name)
	}
}
