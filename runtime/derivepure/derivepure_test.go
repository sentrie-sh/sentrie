// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package derivepure

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsPureBuiltin(t *testing.T) {
	require.True(t, IsPureBuiltin("now"))
	require.True(t, IsPureBuiltin("count"))
	require.False(t, IsPureBuiltin("not_a_builtin"))
	require.False(t, IsPureBuiltin(""))
}

func TestPureBuiltinNamesSortedAndComplete(t *testing.T) {
	names := PureBuiltinNames()
	require.Len(t, names, len(pureBuiltinNames))
	require.True(t, slices.IsSorted(names))
	for _, name := range names {
		require.True(t, IsPureBuiltin(name))
	}
}
