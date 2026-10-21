// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/dop251/goja"
)

// writeModule writes a module under a fresh pack root and returns the registry and the pack root.
func (s *JSTestSuite) writeModule(rel, src string) (*Registry, string) {
	packRoot := s.T().TempDir()
	full := filepath.Join(packRoot, rel)
	s.Require().NoError(os.MkdirAll(filepath.Dir(full), 0o755))
	s.Require().NoError(os.WriteFile(full, []byte(src), 0o644))
	return NewRegistry(packRoot), packRoot
}

func (s *JSTestSuite) TestProgramFor_TranspileErrorNamesSourcePosition() {
	src := "export function ok(): number {\n  return 1;\n}\nexport const broken = ;\n"
	reg, packRoot := s.writeModule("lib/broken.ts", src)

	_, err := reg.PrepareUse("", []string{"local", "lib", "broken"}, packRoot)
	s.Require().Error(err)

	var se *SourceError
	s.Require().ErrorAs(err, &se)
	s.Equal("@local/lib/broken.ts", se.File)
	s.Equal(4, se.Line)
	s.Equal(23, se.Column)
	s.Contains(err.Error(), "@local/lib/broken.ts:4:23: ")
}

func (s *JSTestSuite) TestProgramFor_ReportsEveryTranspileError() {
	reg, packRoot := s.writeModule("multi.ts", "let a = 1;\nlet a = 2;\nlet b = 1;\nlet b = 2;\n")

	_, err := reg.PrepareUse("", []string{"local", "multi.ts"}, packRoot)
	s.Require().Error(err)
	s.Contains(err.Error(), "@local/multi.ts:2:")
	s.Contains(err.Error(), "@local/multi.ts:4:")
}

func (s *JSTestSuite) TestRequire_RuntimeThrowNamesSourceLine() {
	src := "type N = number;\n\nexport function boom(n: N): N {\n  if (n > 0) {\n    throw new Error(\"boom \" + n);\n  }\n  return n;\n}\n"
	reg, packRoot := s.writeModule("thrower.ts", src)

	ar := NewAliasRuntime(reg, packRoot)
	exports, err := ar.Require(s.T().Context(), packRoot, "@local/thrower")
	s.Require().NoError(err)
	fn, ok := goja.AssertFunction(exports.Get("boom"))
	s.Require().True(ok)

	_, err = fn(goja.Undefined(), ar.VM.ToValue(1))
	s.Require().Error(err)
	var ex *goja.Exception
	s.Require().ErrorAs(err, &ex)
	s.Contains(err.Error(), "boom 1")
	s.Contains(err.Error(), "@local/thrower.ts:5:")
	s.Contains(ex.String(), "at boom (@local/thrower.ts:5:")
}

func (s *JSTestSuite) TestProgramFor_RecompilesAfterFix() {
	reg, packRoot := s.writeModule("fixme.ts", "export const x = ;\n")
	modPath := filepath.Join(packRoot, "fixme.ts")

	_, err := reg.PrepareUse("", []string{"local", "fixme.ts"}, packRoot)
	s.Require().Error(err)

	// Unchanged source: the cached failure comes back without recompiling.
	_, again := reg.PrepareUse("", []string{"local", "fixme.ts"}, packRoot)
	s.Require().Error(again)
	s.Same(err, again)

	s.Require().NoError(os.WriteFile(modPath, []byte("export const x = 41 + 1;\n"), 0o644))
	mod, err := reg.PrepareUse("", []string{"local", "fixme.ts"}, packRoot)
	s.Require().NoError(err)
	s.Require().NotNil(mod.Program)

	// A success is kept: a later edit does not replace the compiled program.
	pgm := mod.Program
	s.Require().NoError(os.WriteFile(modPath, []byte("export const x = ;\n"), 0o644))
	got, err := reg.programFor(mod)
	s.Require().NoError(err)
	s.Same(pgm, got)
}

func (s *JSTestSuite) TestProgramFor_ReadErrorIsNotCached() {
	packRoot := s.T().TempDir()
	reg := NewRegistry(packRoot)
	modPath := filepath.Join(packRoot, "late.ts")
	mod := reg.getOrCreateModule("@local/late.ts", modPath, packRoot, false)
	s.Require().NotNil(mod)

	_, err := reg.programFor(mod)
	s.Require().Error(err)
	s.True(errors.Is(err, os.ErrNotExist))

	s.Require().NoError(os.WriteFile(modPath, []byte("export const x = 1;\n"), 0o644))
	pgm, err := reg.programFor(mod)
	s.Require().NoError(err)
	s.NotNil(pgm)
}

func (s *JSTestSuite) TestProgramFor_ConcurrentCallersShareOneProgram() {
	reg, packRoot := s.writeModule("shared.ts", "export const x = 1;\n")
	mod := reg.getOrCreateModule("@local/shared.ts", filepath.Join(packRoot, "shared.ts"), packRoot, false)
	s.Require().NotNil(mod)

	const callers = 16
	pgms := make([]*goja.Program, callers)
	errs := make([]error, callers)
	var wg sync.WaitGroup
	for i := range callers {
		wg.Go(func() { pgms[i], errs[i] = reg.programFor(mod) })
	}
	wg.Wait()
	for i := range callers {
		s.Require().NoError(errs[i])
		s.Same(pgms[0], pgms[i])
	}
}

func (s *JSTestSuite) TestCompileError_ParseErrorNamesModule() {
	_, err := goja.Compile("@local/x.ts", "var a b;", true)
	s.Require().Error(err)

	wrapped := compileError("@local/x.ts", err)
	var se *SourceError
	s.Require().ErrorAs(wrapped, &se)
	s.Equal("@local/x.ts", se.File)
	s.Contains(wrapped.Error(), "transpiled output")
	s.ErrorIs(wrapped, err)
}
