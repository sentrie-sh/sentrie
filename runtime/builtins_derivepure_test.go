// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"testing"

	"github.com/sentrie-sh/sentrie/internal/derivepure"
	"github.com/stretchr/testify/require"
)

func TestDerivePureBuiltinNamesAreRegistered(t *testing.T) {
	for _, name := range derivepure.PureBuiltinNames() {
		_, ok := Builtins[name]
		require.True(t, ok, "pure builtin %q must exist in Builtins", name)
	}
}
