// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"testing"

	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/stretchr/testify/require"
)

func TestBuiltinsRegistryImport(t *testing.T) {
	t.Parallel()
	require.Equal(t, len(builtins.Table), len(builtins.DeriveSafeNames()))
	for _, name := range builtins.DeriveSafeNames() {
		require.Contains(t, builtins.Table, name)
	}
}
