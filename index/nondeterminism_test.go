// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"context"

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
  rule viaLambda = { yield any([1], (now) => { yield now > 0 }) }
  rule ok = default false when now() > 0 { yield count([1]) > 0 }
  export decision of ok attach at as now()
}`
	idx := s.indexFromSource("nondet.sentrie", src)

	lookup := func(moduleKey, export string) bool {
		return moduleKey == "@sentrie/uuid" && (export == "v4" || export == "v7")
	}
	warnings, err := idx.NondeterminismWarnings(s.T().Context(), lookup)
	s.Require().NoError(err)

	var msgs []string
	for _, w := range warnings {
		s.Equal("nondet.sentrie", w.At.File)
		msgs = append(msgs, w.Message)
	}
	nowMsg := `builtin "now" is nondeterministic: its result can differ between executions`
	s.Equal([]string{
		nowMsg, // fact default (line 3)
		"@sentrie/uuid.v4 is nondeterministic: its result can differ between executions",
		"@sentrie/uuid.v7 is nondeterministic: its result can differ between executions",
		nowMsg, // let (line 7)
		nowMsg, // rule when (line 13)
		nowMsg, // attachment (line 14)
	}, msgs)

	// Sorted by position.
	for i := 1; i < len(warnings); i++ {
		s.LessOrEqual(warnings[i-1].At.From.Line, warnings[i].At.From.Line)
	}

	// Without a module lookup only builtin calls are reported.
	warnings, err = idx.NondeterminismWarnings(s.T().Context(), nil)
	s.Require().NoError(err)
	s.Len(warnings, 4)
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
