// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/stretchr/testify/require"
)

// TestBooleanTypeRefStaticKind: boolean rejects unknown so its static kind is
// ValueBool, while trinary stays ValueTrinary (#144).
func TestBooleanTypeRefStaticKind(t *testing.T) {
	t.Parallel()
	rng := testBuiltinKindRange()

	boolean := ast.NewTrinaryTypeRef(rng)
	boolean.Boolean = true
	kind, ok := typeRefKind(nil, nil, boolean)
	require.True(t, ok)
	require.Equal(t, box.ValueBool, kind)

	kind, ok = typeRefKind(nil, nil, ast.NewTrinaryTypeRef(rng))
	require.True(t, ok)
	require.Equal(t, box.ValueTrinary, kind)
}
