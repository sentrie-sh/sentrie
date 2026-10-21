// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"
	"fmt"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
)

func (s *IndexTestSuite) indexFromSource(name, src string) *Index {
	ctx := s.T().Context()
	idx := CreateIndex()
	prog, err := parser.NewParserFromString(src, name).ParseProgram(ctx)
	s.Require().NoError(err)
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))
	return idx
}

func (s *IndexTestSuite) TestNondeterminismWarningsCoverEverySlot() {
	src := `namespace com/ex
policy pol {
  fact started?: number as started default now()
  use { v4, v7 } from @sentrie/uuid as uuid
  use { parse } from @sentrie/time as clock
  use { helper } from "./helper" as local
  let stamp = now()
  rule shadowed = {
    let now = 5
    yield now > 1
  }
  rule viaLambda = { yield any([1], (now) => { yield now() > 0 }) }
  rule viaImport = import decision ok from com/ex/auth with t as now()
  rule ok = default false when now() > 0 { yield count([1]) > 0 }
  export decision of ok attach at as now()
}
policy auth {
  fact t?: number as t
  rule ok = { yield true }
  export decision of ok
}`
	idx := s.indexFromSource("nondet.sentrie", src)

	lookup := func(moduleKey, export string) bool {
		return moduleKey == "@sentrie/uuid" && (export == "v4" || export == "v7")
	}
	warnings, err := idx.NondeterminismWarnings(s.T().Context(), lookup)
	s.Require().NoError(err)

	var got []string
	for _, w := range warnings {
		s.Equal("nondet.sentrie", w.At.File)
		got = append(got, fmt.Sprintf("%d %s", w.At.From.Line, w.Message))
	}
	nowMsg := `builtin "now" is nondeterministic: its result can differ between executions`
	s.Equal([]string{
		"2 " + nowMsg, // fact default (From.Line is 0-based)
		"3 @sentrie/uuid.v4 is nondeterministic: its result can differ between executions",
		"3 @sentrie/uuid.v7 is nondeterministic: its result can differ between executions",
		`5 local module "./helper" runs pack code, which can read Date.now() and Math.random() directly`,
		"6 " + nowMsg, // policy let
		// A lambda parameter named now does not shadow the builtin for calls:
		// the runtime's getTarget never resolves a call through a local.
		"11 " + nowMsg,
		"12 " + nowMsg, // import ... with t as now()
		"13 " + nowMsg, // rule when
		"14 " + nowMsg, // attachment
	}, got)

	// Without a module lookup, builtin calls and local modules are still reported.
	warnings, err = idx.NondeterminismWarnings(s.T().Context(), nil)
	s.Require().NoError(err)
	s.Len(warnings, 7)
}

func (s *IndexTestSuite) TestNondeterminismWarningsReportLocalModuleSources() {
	idx := s.indexFromSource("local.sentrie", `namespace com/ex
policy pol {
  use { helper } from @local/lib/helper as lib
  rule ok = { yield true }
  export decision of ok
}`)
	warnings, err := idx.NondeterminismWarnings(s.T().Context(), nil)
	s.Require().NoError(err)
	s.Require().Len(warnings, 1)
	s.Contains(warnings[0].Message, `local module "@local/lib/helper"`)
}

func (s *IndexTestSuite) TestNondeterminismWarningsEmptyForDeterministicPack() {
	idx := s.indexFromSource("det.sentrie", `namespace com/ex
policy pol {
  rule ok = { yield count([1]) > 0 }
  export decision of ok
}`)
	warnings, err := idx.NondeterminismWarnings(s.T().Context(), func(string, string) bool { return true })
	s.Require().NoError(err)
	s.Empty(warnings)
}

func (s *IndexTestSuite) TestNondeterminismWarningsCancelled() {
	idx := s.indexFromSource("det.sentrie", `namespace com/ex
policy pol {
  rule ok = { yield true }
  export decision of ok
}`)
	ctx, cancel := context.WithCancel(s.T().Context())
	cancel()
	_, err := idx.NondeterminismWarnings(ctx, nil)
	s.Require().Error(err)
}

func (s *IndexTestSuite) TestUseBuiltinModuleKey() {
	r := tokens.Range{}
	key, ok := UseBuiltinModuleKey(ast.NewUseStatement([]string{"v4"}, "", []string{"sentrie", "uuid"}, "uuid", r))
	s.True(ok)
	s.Equal("@sentrie/uuid", key)

	_, ok = UseBuiltinModuleKey(ast.NewUseStatement([]string{"f"}, "./mod", nil, "mod", r))
	s.False(ok)
	_, ok = UseBuiltinModuleKey(ast.NewUseStatement([]string{"f"}, "", []string{"local", "mod"}, "mod", r))
	s.False(ok)
	_, ok = UseBuiltinModuleKey(ast.NewUseStatement([]string{"f"}, "", []string{"sentrie"}, "sentrie", r))
	s.False(ok)
	_, ok = UseBuiltinModuleKey(nil)
	s.False(ok)
}

// TestNondeterminismWarningsCoverImportWithClauses checks that the walk shared
// with checkBuiltinCalls reaches the `with` expressions of an imported
// decision, which the runtime evaluates in the importing rule (#218).
func (s *IndexTestSuite) TestNondeterminismWarningsCoverImportWithClauses() {
	idx := s.indexFromSource("nondet.sentrie", `namespace com/ex
policy q {
  fact at: number as at
  rule ok = default true { yield at > 0 }
  export decision of ok
}
policy p {
  rule ok = import decision ok from com/ex/q with at as now()
  export decision of ok
}`)
	warnings, err := idx.NondeterminismWarnings(s.T().Context(), nil)
	s.Require().NoError(err)
	s.Require().Len(warnings, 1)
	s.Equal(7, warnings[0].At.From.Line)
	s.Contains(warnings[0].Message, `builtin "now" is nondeterministic`)
}
