// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) TestIsBuiltinAllowedInDerive() {
	s.True(isBuiltinAllowedInDerive("now"))
	s.False(isBuiltinAllowedInDerive("not_a_builtin"))
}

func (s *RuntimeTestSuite) TestExecRuleInvokesNamespaceDerive() {
	ctx := context.Background()
	src := `namespace com/ex

derive bump = (n: number): number => { yield n + 1 }

policy pol {
  let _seed = 0
  rule allow = {
    yield bump(1) == 2
  }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "drv.sentra")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func (s *RuntimeTestSuite) TestExecRuleDeriveTooManyArgsErrors() {
	ctx := context.Background()
	src := `namespace com/ex

derive id = (n: number): number => { yield n }

policy pol {
  let _seed = 0
  rule allow = { yield id(1, 2) == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "args.sentra")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "too many arguments")
}

func (s *RuntimeTestSuite) TestExecRuleDeriveReturnTypeMismatchErrors() {
	ctx := context.Background()
	src := `namespace com/ex

derive bad = (n: number): string => { yield n }

policy pol {
  let _seed = 0
  rule allow = { yield bad(1) == "1" }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "ret.sentra")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "derive return")
}
