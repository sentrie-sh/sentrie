// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"os"
	"path/filepath"
)

// symlinkedPackRoot returns a pack root reached through a symlink
// (<tmp>/current -> <tmp>/releases/1) and the real directory behind it.
func (s *JSTestSuite) symlinkedPackRoot() (linkRoot, realRoot string) {
	base := s.T().TempDir()
	realRoot = filepath.Join(base, "releases", "1")
	s.Require().NoError(os.MkdirAll(realRoot, 0o755))
	linkRoot = filepath.Join(base, "current")
	s.Require().NoError(os.Symlink(realRoot, linkRoot))
	return linkRoot, realRoot
}

func (s *JSTestSuite) writePackFile(root, rel, src string) {
	p := filepath.Join(root, rel)
	s.Require().NoError(os.MkdirAll(filepath.Dir(p), 0o755))
	s.Require().NoError(os.WriteFile(p, []byte(src), 0o644))
}

// writeNestedChain writes main.ts -> ./a -> ./lib/b -> ../c, so each
// require after the first is resolved from a module's own (resolved) Dir.
func (s *JSTestSuite) writeNestedChain(root string) {
	s.writePackFile(root, "main.ts", "const a = require('./a');\nexport const v = a.v + 1;")
	s.writePackFile(root, "a.ts", "const b = require('./lib/b');\nexport const v = b.v + 1;")
	s.writePackFile(root, "lib/b.ts", "const c = require('../c');\nexport const v = c.v + 1;")
	s.writePackFile(root, "c.ts", "export const v = 1;")
}

func (s *JSTestSuite) requireNestedChain(root, fromDir string) int64 {
	reg := NewRegistry(root)
	ar := NewAliasRuntime(reg, fromDir)
	ex, err := ar.Require(s.T().Context(), fromDir, "./main")
	s.Require().NoError(err)
	return ex.Get("v").ToInteger()
}

func (s *JSTestSuite) TestNestedRelativeRequires_PackRootBehindSymlink() {
	linkRoot, _ := s.symlinkedPackRoot()
	s.writeNestedChain(linkRoot)
	s.Equal(int64(4), s.requireNestedChain(linkRoot, linkRoot))
}

// The first require may come from the resolved side (a module's Dir) while
// the registry holds the symlinked root, or the other way round.
func (s *JSTestSuite) TestNestedRelativeRequires_MixedRootSpellings() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	s.writeNestedChain(realRoot)
	s.Equal(int64(4), s.requireNestedChain(linkRoot, realRoot))
	s.Equal(int64(4), s.requireNestedChain(realRoot, linkRoot))
}

// os.TempDir on macOS is under /var, itself a symlink to /private/var, so an
// unresolved temp dir is the natural form of the bug there. Elsewhere this
// runs against a plain directory and still must pass.
func (s *JSTestSuite) TestNestedRelativeRequires_UnresolvedTempDir() {
	root := s.T().TempDir()
	s.writeNestedChain(root)
	s.Equal(int64(4), s.requireNestedChain(root, root))
}

func (s *JSTestSuite) TestResolveUse_LocalPackRootBehindSymlink() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	s.writePackFile(linkRoot, "lib/helpers.ts", "export const v = 1;")
	reg := NewRegistry(linkRoot)

	key, path, dir, builtin, err := reg.resolveUse("", []string{"local", "lib", "helpers.ts"}, linkRoot)
	s.Require().NoError(err)
	s.False(builtin)
	s.Equal("@local/lib/helpers.ts", key)
	resolvedReal, err := filepath.EvalSymlinks(realRoot)
	s.Require().NoError(err)
	s.Equal(filepath.Join(resolvedReal, "lib", "helpers.ts"), path)
	s.Equal(filepath.Join(resolvedReal, "lib"), dir)
}

// A symlinked pack root must not widen confinement: a relative require from
// a resolved module Dir still cannot climb out of the real root with "..".
func (s *JSTestSuite) TestResolveRequire_SymlinkedRootRelativeEscapeRejected() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	s.writePackFile(filepath.Dir(realRoot), "secret.ts", "export const v = 1;")
	s.writePackFile(realRoot, "lib/a.ts", "export const v = 1;")
	resolvedReal, err := filepath.EvalSymlinks(realRoot)
	s.Require().NoError(err)
	reg := NewRegistry(linkRoot)

	for _, spec := range []string{"../../secret", "../../secret.ts", "./../../secret"} {
		_, _, _, _, err := reg.resolveRequire(filepath.Join(resolvedReal, "lib"), spec)
		s.Require().Error(err, spec)
		s.Contains(err.Error(), "outside the packroot", spec)

		_, err = reg.LoadRequire(filepath.Join(linkRoot, "lib"), spec)
		s.Require().Error(err, spec)
		s.Contains(err.Error(), "outside the packroot", spec)
	}

	_, _, _, _, err = reg.resolveRequire(linkRoot, "@local/../secret.ts")
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}

// A symlink inside a symlinked pack root that points outside the real root
// is still refused, both for a directory link and a file link.
func (s *JSTestSuite) TestResolveRequire_SymlinkedRootInnerLinkOutsideRejected() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	outside := s.T().TempDir()
	s.writePackFile(outside, "secret.ts", "export const v = 1;")
	s.Require().NoError(os.Symlink(outside, filepath.Join(realRoot, "escape")))
	s.Require().NoError(os.Symlink(filepath.Join(outside, "secret.ts"), filepath.Join(realRoot, "leak.ts")))
	reg := NewRegistry(linkRoot)

	for _, spec := range []string{"./escape/secret.ts", "./escape/secret", "./leak.ts", "@local/escape/secret.ts", "@local/leak.ts"} {
		_, err := reg.LoadRequire(linkRoot, spec)
		s.Require().Error(err, spec)
		s.Contains(err.Error(), "outside the packroot", spec)
	}
}

// A link inside the pack that stays inside it is still followed.
func (s *JSTestSuite) TestLoadRequire_InnerLinkInsidePackRootAllowed() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	s.writePackFile(realRoot, "impl/real.ts", "export const v = 7;")
	s.Require().NoError(os.Symlink(filepath.Join(realRoot, "impl", "real.ts"), filepath.Join(realRoot, "alias.ts")))
	s.Require().NoError(os.Symlink(filepath.Join(realRoot, "impl"), filepath.Join(realRoot, "implink")))
	reg := NewRegistry(linkRoot)
	ar := NewAliasRuntime(reg, linkRoot)

	for _, spec := range []string{"./alias", "./alias.ts", "./implink/real"} {
		ex, err := ar.Require(s.T().Context(), linkRoot, spec)
		s.Require().NoError(err, spec)
		s.Equal(int64(7), ex.Get("v").ToInteger(), spec)
	}
}
