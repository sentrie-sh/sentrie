// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"os"
	"path/filepath"
)

// entryForms are the ways a pack can name a local module. Each returns the
// error from resolving name (a specifier without directory or extension)
// through that form, so a confinement hole in one form cannot hide behind
// the others.
var entryForms = map[string]func(reg *Registry, root, name string) error{
	"require relative": func(reg *Registry, root, name string) error {
		_, err := reg.LoadRequire(root, "./"+name)
		return err
	},
	"require @local": func(reg *Registry, root, name string) error {
		_, err := reg.LoadRequire(root, "@local/"+name)
		return err
	},
	"use local": func(reg *Registry, root, name string) error {
		_, err := reg.PrepareUse("", []string{"local", name}, root)
		return err
	},
	"use relative": func(reg *Registry, root, name string) error {
		_, err := reg.PrepareUse("./"+name, nil, root)
		return err
	},
}

// A pack-local file named <name>.<ext> that links outside the pack must be
// refused through every entry form when the specifier omits the extension
// (the probed file is the link), whichever extension is probed, and however
// many links lead out.
func (s *JSTestSuite) TestProbedExtensionLinkOutsideRejectedForEveryEntryForm() {
	outside := s.T().TempDir()
	s.writePackFile(outside, "secret.ts", "export const v = 1;")
	s.writePackFile(outside, "secret.js", "exports.v = 1;")

	cases := map[string]func(root string){
		"ts link out": func(root string) {
			s.Require().NoError(os.Symlink(filepath.Join(outside, "secret.ts"), filepath.Join(root, "leak.ts")))
		},
		"js link out": func(root string) {
			s.Require().NoError(os.Symlink(filepath.Join(outside, "secret.js"), filepath.Join(root, "leak.js")))
		},
		// leak.ts -> hop1 -> hop2 -> outside: a chain, not a single link.
		"ts chain out": func(root string) {
			hop2 := filepath.Join(root, "hop2")
			s.Require().NoError(os.Symlink(filepath.Join(outside, "secret.ts"), hop2))
			hop1 := filepath.Join(root, "hop1")
			s.Require().NoError(os.Symlink(hop2, hop1))
			s.Require().NoError(os.Symlink(hop1, filepath.Join(root, "leak.ts")))
		},
		// The probed .ts wins over an in-pack .js, so it must fail closed
		// rather than fall back to the safe file.
		"ts link out shadows in-pack js": func(root string) {
			s.Require().NoError(os.Symlink(filepath.Join(outside, "secret.ts"), filepath.Join(root, "leak.ts")))
			s.writePackFile(root, "leak.js", "exports.v = 2;")
		},
	}
	for name, setup := range cases {
		root := s.T().TempDir()
		setup(root)
		for form, resolve := range entryForms {
			// A fresh registry per form, so one form cannot mask a leak in
			// another by having warmed the module cache.
			err := resolve(NewRegistry(root), root, "leak")
			s.Require().Error(err, "%s via %s", name, form)
			s.Contains(err.Error(), "outside the packroot", "%s via %s", name, form)
		}
	}
}

// The same probe bypass from a nested module: a relative require resolved
// from a module's own Dir. The root is resolved up front so the only thing
// that can refuse the require is confinement of the probed file, not the
// symlinked-root handling.
func (s *JSTestSuite) TestProbedExtensionLinkOutsideRejectedFromNestedModule() {
	root, err := filepath.EvalSymlinks(s.T().TempDir())
	s.Require().NoError(err)
	outside := s.T().TempDir()
	s.writePackFile(outside, "secret.ts", "export const v = 1;")
	s.writePackFile(root, "main.ts", "const a = require('./lib/a');\nexport const v = a.v;")
	s.writePackFile(root, "lib/a.ts", "const l = require('../leak');\nexport const v = l.v;")
	s.Require().NoError(os.Symlink(filepath.Join(outside, "secret.ts"), filepath.Join(root, "leak.ts")))

	ar := NewAliasRuntime(NewRegistry(root), root)
	_, err = ar.Require(s.T().Context(), root, "./main")
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}

// A dangling .ts link is not a module: probing falls through to .js.
func (s *JSTestSuite) TestProbedExtensionDanglingLinkFallsThroughToJS() {
	root := s.T().TempDir()
	s.Require().NoError(os.Symlink(filepath.Join(root, "nowhere.ts"), filepath.Join(root, "mod.ts")))
	s.writePackFile(root, "mod.js", "exports.v = 9;")
	ar := NewAliasRuntime(NewRegistry(root), root)

	ex, err := ar.Require(s.T().Context(), root, "./mod")
	s.Require().NoError(err)
	s.Equal(int64(9), ex.Get("v").ToInteger())
}

// A directory link inside the pack that leaves it, used as an intermediate
// segment of a nested relative require, is refused; so is a chain of
// directory links whose last hop leaves the pack.
func (s *JSTestSuite) TestNestedRequire_DirectoryLinkSegmentOutsideRejected() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	outside := s.T().TempDir()
	s.writePackFile(outside, "secret.ts", "export const v = 1;")
	s.writePackFile(realRoot, "lib/a.ts", "export const v = 1;")
	s.Require().NoError(os.Symlink(outside, filepath.Join(realRoot, "lib", "esc")))
	// lib/chain1 -> lib/chain2 -> outside
	s.Require().NoError(os.Symlink(outside, filepath.Join(realRoot, "lib", "chain2")))
	s.Require().NoError(os.Symlink(filepath.Join(realRoot, "lib", "chain2"), filepath.Join(realRoot, "lib", "chain1")))
	reg := NewRegistry(linkRoot)

	for _, spec := range []string{"./esc/secret", "./esc/secret.ts", "./chain1/secret", "./chain1/secret.ts", "./esc/../esc/secret"} {
		for _, from := range []string{filepath.Join(realRoot, "lib"), filepath.Join(linkRoot, "lib")} {
			_, err := reg.LoadRequire(from, spec)
			s.Require().Error(err, "%s from %s", spec, from)
			s.Contains(err.Error(), "outside the packroot", "%s from %s", spec, from)
		}
	}
}

// Executing nested module code that requires through an escaping directory
// link fails the whole require chain.
func (s *JSTestSuite) TestNestedRequire_DirectoryLinkOutsideRejectedWhenExecuted() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	outside := s.T().TempDir()
	s.writePackFile(outside, "secret.ts", "export const v = 1;")
	s.writePackFile(realRoot, "main.ts", "const a = require('./lib/a');\nexport const v = a.v;")
	s.writePackFile(realRoot, "lib/a.ts", "const s = require('./esc/secret');\nexport const v = s.v;")
	s.Require().NoError(os.Symlink(outside, filepath.Join(realRoot, "lib", "esc")))

	ar := NewAliasRuntime(NewRegistry(linkRoot), linkRoot)
	_, err := ar.Require(s.T().Context(), linkRoot, "./main")
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}

// ".." after a directory link is cleaned lexically before anything is
// resolved, so lib/esc/../x reads lib/x and never the outside directory the
// link points at (as a shell cd would).
func (s *JSTestSuite) TestNestedRequire_DotDotAfterDirectoryLinkReadsLexicalParent() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	outside := s.T().TempDir()
	s.writePackFile(outside, "x.ts", "export const v = 111;")
	s.writePackFile(realRoot, "lib/x.ts", "export const v = 7;")
	s.Require().NoError(os.Symlink(outside, filepath.Join(realRoot, "lib", "esc")))
	ar := NewAliasRuntime(NewRegistry(linkRoot), linkRoot)

	ex, err := ar.Require(s.T().Context(), filepath.Join(linkRoot, "lib"), "./esc/../x")
	s.Require().NoError(err)
	s.Equal(int64(7), ex.Get("v").ToInteger())
}

// Link chains that stay inside the pack are followed, from a nested module
// and through an extension-less probe, with the root behind a symlink.
func (s *JSTestSuite) TestNestedRequire_InPackLinkChainAllowed() {
	linkRoot, realRoot := s.symlinkedPackRoot()
	s.writePackFile(realRoot, "impl/real.ts", "export const v = 5;")
	// lib/alias.ts -> lib/hop.ts -> impl/real.ts
	s.writePackFile(realRoot, "lib/.keep", "")
	s.Require().NoError(os.Symlink(filepath.Join(realRoot, "impl", "real.ts"), filepath.Join(realRoot, "lib", "hop.ts")))
	s.Require().NoError(os.Symlink(filepath.Join(realRoot, "lib", "hop.ts"), filepath.Join(realRoot, "lib", "alias.ts")))
	s.writePackFile(realRoot, "main.ts", "const a = require('./lib/alias');\nexport const v = a.v + 1;")
	ar := NewAliasRuntime(NewRegistry(linkRoot), linkRoot)

	ex, err := ar.Require(s.T().Context(), linkRoot, "./main")
	s.Require().NoError(err)
	s.Equal(int64(6), ex.Get("v").ToInteger())
}

// A specifier naming the pack root itself has no extension, so probing tries
// <packroot>.ts, a file next to the root rather than inside it. It must be
// refused, not read.
func (s *JSTestSuite) TestRootItselfSpecifierDoesNotProbeSiblingOfRoot() {
	base := s.T().TempDir()
	root := filepath.Join(base, "pack")
	s.Require().NoError(os.Mkdir(root, 0o755))
	s.writePackFile(base, "pack.ts", "export const v = 1;")
	reg := NewRegistry(root)

	for _, spec := range []string{".", "./", "@local/", "@local/.", "../pack"} {
		_, err := reg.LoadRequire(root, spec)
		s.Require().Error(err, spec)
		s.Contains(err.Error(), "outside the packroot", spec)
	}
	_, err := reg.PrepareUse("", []string{"local"}, root)
	s.Require().Error(err)
	s.Contains(err.Error(), "outside the packroot")
}
