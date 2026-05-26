// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/sentrie-sh/sentrie/trinary"
)

func elvisTestRange() tokens.Range {
	return tokens.Range{File: "elvis_test.sentra", From: tokens.Pos{}, To: tokens.Pos{}}
}

func (s *RuntimeTestSuite) TestEvalElvisFiveCases() {
	ctx := s.T().Context()
	p := &index.Policy{}
	ec := NewExecutionContext(p, &executorImpl{})
	defer ec.Dispose()
	exec := &executorImpl{}
	ec.SetLocal("undefLocal", box.Undefined(), true)

	rng := elvisTestRange()
	s.Run("null", func() {
		el := ast.NewTernaryElvis(ast.NewNullLiteral(rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		requireNoErrorNum(s, err, v, 99)
	})
	s.Run("undefined", func() {
		el := ast.NewTernaryElvis(ast.NewIdentifier("undefLocal", rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		requireNoErrorNum(s, err, v, 99)
	})
	s.Run("zero", func() {
		el := ast.NewTernaryElvis(ast.NewIntegerLiteral(0, rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		requireNoErrorNum(s, err, v, 0)
	})
	s.Run("empty_string", func() {
		el := ast.NewTernaryElvis(ast.NewStringLiteral("", rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		s.NoError(err)
		str, ok := v.StringValue()
		s.True(ok)
		s.Equal("", str)
	})
	s.Run("false", func() {
		el := ast.NewTernaryElvis(ast.NewTrinaryLiteral(trinary.False, rng), ast.NewIntegerLiteral(99, rng), rng)
		v, _, err := eval(ctx, ec, exec, p, el)
		s.NoError(err)
		s.True(box.EqualValues(v, box.Trinary(trinary.False)))
	})
}

func (s *RuntimeTestSuite) TestExecRuleElvisParsedShortCircuitsDivideByZero() {
	ctx := s.T().Context()
	src := `namespace n
policy p {
  let _seed = 0
  rule r = { yield _seed ?: (1 / 0) == 0 }
  export decision of r
}
`
	p := parser.NewParserFromString(src, "elvis_exec.sentra")
	prog, err := p.ParseProgram(ctx)
	s.Require().NoError(err)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: "."}))
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	out, err := exec.ExecRule(ctx, "n", "p", "r", nil)
	s.NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

func requireNoErrorNum(s *RuntimeTestSuite, err error, v box.Value, want float64) {
	s.T().Helper()
	s.Require().NoError(err)
	n, ok := v.NumberValue()
	s.True(ok)
	s.Equal(want, n)
}
