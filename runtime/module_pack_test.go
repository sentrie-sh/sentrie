// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
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
	ctx := s.T().Context()
	dir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, loader.PackFileName), []byte(testPackManifest), 0o644))
	for name, content := range files {
		path := filepath.Join(dir, name)
		s.Require().NoError(os.MkdirAll(filepath.Dir(path), 0o755))
		s.Require().NoError(os.WriteFile(path, []byte(content), 0o644))
	}

	packFile, err := loader.LoadPack(ctx, dir)
	s.Require().NoError(err)
	programs, err := loader.LoadPrograms(ctx, packFile)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, packFile))
	for _, program := range programs {
		s.Require().NoError(idx.AddProgram(ctx, program))
	}
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx, opts...)
	s.Require().NoError(err)
	return exec.(*executorImpl)
}
