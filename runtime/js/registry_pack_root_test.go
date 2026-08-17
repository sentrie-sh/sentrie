// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"os"
	"path/filepath"
)

func (s *JSTestSuite) TestResolveRequire_LocalAtLocalEscapesPackRoot() {
	packRoot := s.T().TempDir()
	reg := NewRegistry(packRoot)

	_, _, _, _, err := reg.resolveRequire(packRoot, "@local/../../etc/passwd")
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}

func (s *JSTestSuite) TestResolveRequire_RelativeEscapesPackRoot() {
	packRoot := s.T().TempDir()
	subDir := filepath.Join(packRoot, "lib")
	s.Require().NoError(os.Mkdir(subDir, 0o755))

	reg := NewRegistry(packRoot)

	_, _, _, _, err := reg.resolveRequire(subDir, "../../etc/passwd")
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}

func (s *JSTestSuite) TestResolveRequire_LocalSymlinkOutsidePackRootRejected() {
	packRoot := s.T().TempDir()
	outsideDir := s.T().TempDir()
	outsideFile := filepath.Join(outsideDir, "secret.js")
	s.Require().NoError(os.WriteFile(outsideFile, []byte("export const x = 1;"), 0o644))

	linkPath := filepath.Join(packRoot, "escape")
	s.Require().NoError(os.Symlink(outsideDir, linkPath))

	reg := NewRegistry(packRoot)

	_, _, _, _, err := reg.resolveRequire(packRoot, "@local/escape/secret.js")
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}

func (s *JSTestSuite) TestResolveRequire_LocalModuleInsidePackRoot() {
	packRoot := s.T().TempDir()
	modulePath := filepath.Join(packRoot, "foo.js")
	s.Require().NoError(os.WriteFile(modulePath, []byte("export const x = 1;"), 0o644))

	reg := NewRegistry(packRoot)

	key, path, dir, builtin, err := reg.resolveRequire(packRoot, "@local/foo.js")
	s.Require().NoError(err)
	s.Equal("@local/foo.js", key)
	resolvedModule, err := evalSymlinksExisting(modulePath)
	s.Require().NoError(err)
	s.Equal(resolvedModule, path)
	resolvedPackRoot, err := evalSymlinksExisting(packRoot)
	s.Require().NoError(err)
	s.Equal(resolvedPackRoot, dir)
	s.False(builtin)
}

func (s *JSTestSuite) TestResolveUse_LocalEscapesPackRoot() {
	packRoot := s.T().TempDir()
	reg := NewRegistry(packRoot)

	_, _, _, _, err := reg.resolveUse("", []string{"local", "..", "..", "etc", "passwd"}, packRoot)
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}
