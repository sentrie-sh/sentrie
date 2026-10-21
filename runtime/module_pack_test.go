// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/stretchr/testify/require"
)

const testPackManifest = `[schema]
version = 1

[pack]
name = "module-test-pack"
version = "0.0.1"
`

// newTempPackExecutor writes files (relative path -> content) into a fresh pack
// directory, runs the load, index and validate pipeline over it, and returns an
// executor for the pack.
func (s *RuntimeTestSuite) newTempPackExecutor(files map[string]string, opts ...NewExecutorOption) *executorImpl {
	return buildTempPackExecutor(s.T(), files, opts...)
}

// buildTempPackExecutor is newTempPackExecutor for any testing.TB, so
// benchmarks can build the same packs.
func buildTempPackExecutor(tb testing.TB, files map[string]string, opts ...NewExecutorOption) *executorImpl {
	r := require.New(tb)
	ctx := tb.Context()
	dir := tb.TempDir()
	r.NoError(os.WriteFile(filepath.Join(dir, loader.PackFileName), []byte(testPackManifest), 0o644))
	for name, content := range files {
		path := filepath.Join(dir, name)
		r.NoError(os.MkdirAll(filepath.Dir(path), 0o755))
		r.NoError(os.WriteFile(path, []byte(content), 0o644))
	}

	packFile, err := loader.LoadPack(ctx, dir)
	r.NoError(err)
	programs, err := loader.LoadPrograms(ctx, packFile)
	r.NoError(err)

	idx := index.CreateIndex()
	r.NoError(idx.SetPack(ctx, packFile))
	for _, program := range programs {
		r.NoError(idx.AddProgram(ctx, program))
	}
	r.NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx, opts...)
	r.NoError(err)
	return exec.(*executorImpl)
}
