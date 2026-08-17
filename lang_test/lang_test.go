// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package lang_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/stretchr/testify/require"
)

func TestFixturesParseIndexAndValidate(t *testing.T) {
	entries, err := os.ReadDir(".")
	require.NoError(t, err)

	ctx := t.Context()
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".sentrie") {
			continue
		}

		t.Run(ent.Name(), func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(".", ent.Name())
			data, err := os.ReadFile(path)
			require.NoError(t, err)

			src := string(data)
			idx := index.CreateIndex()
			prog, err := parser.NewParserFromString(src, ent.Name()).ParseProgram(ctx)
			require.NoError(t, err, "parse %s", path)
			require.NoError(t, idx.AddProgram(ctx, prog), "index %s", path)
			require.NoError(t, idx.Validate(ctx), "validate %s", path)
		})
	}
}
