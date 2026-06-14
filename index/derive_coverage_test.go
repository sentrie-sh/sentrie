// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/require"
)

func deriveCovRng(line int) tokens.Range {
	return tokens.Range{
		File: "cov.sentra",
		From: tokens.Pos{Line: line, Column: 1, Offset: 0},
		To:   tokens.Pos{Line: line, Column: 2, Offset: 1},
	}
}

func (s *IndexTestSuite) TestDeriveVisibleFromPolicyAndDeriveCaller() {
	nsFQN := ast.NewFQN([]string{"com", "ex"}, deriveCovRng(1))
	polA := &Policy{Name: "polA", FQN: ast.CreateFQN(nsFQN, "polA"), Namespace: &Namespace{FQN: nsFQN}}
	polB := &Policy{Name: "polB", FQN: ast.CreateFQN(nsFQN, "polB"), Namespace: &Namespace{FQN: nsFQN}}

	nsDerive := &Derive{Name: "open", FQN: ast.CreateFQN(nsFQN, "open"), Namespace: &Namespace{FQN: nsFQN}}
	s.NoError(nsDerive.VisibleFromPolicy(nil))
	s.NoError(nsDerive.VisibleFromPolicy(polB))
	s.NoError(nsDerive.VisibleFromDeriveCaller(&Derive{Namespace: &Namespace{FQN: nsFQN}}))

	scoped := &Derive{
		Name:      "secret",
		FQN:       ast.CreateFQN(polA.FQN, "secret"),
		Namespace: &Namespace{FQN: nsFQN},
		Policy:    polA,
	}
	s.NoError(scoped.VisibleFromPolicy(polA))
	s.Error(scoped.VisibleFromPolicy(polB))
	s.Error(scoped.VisibleFromPolicy(nil))
	s.Error(scoped.VisibleFromDeriveCaller(&Derive{Namespace: &Namespace{FQN: nsFQN}}))
	s.NoError(scoped.VisibleFromDeriveCaller(&Derive{Namespace: &Namespace{FQN: nsFQN}, Policy: polA}))
}

func (s *IndexTestSuite) TestDeriveSpanUsesStatementWhenPresent() {
	stmt := ast.NewDeriveStatement("d", stubDeriveCovLambda(), deriveCovRng(2))
	lam := stubDeriveCovLambda()
	d := &Derive{Statement: stmt, Lambda: lam}
	s.Equal(stmt.Span(), d.Span())
}

func stubDeriveCovLambda() *ast.LambdaExpression {
	return ast.NewLambdaExpression(
		nil,
		ast.NewBlockExpression(nil, ast.NewIntegerLiteral(1, deriveCovRng(3)), deriveCovRng(3)),
		deriveCovRng(3),
	)
}

func (s *IndexTestSuite) TestDerivePurityRejectsUnknownSlashDerive() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield com/ex/missing() }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "unk_slash.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "unknown derive")
}

func (s *IndexTestSuite) TestDerivePurityRejectsTwoParamDeriveAsFilterCallback() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive isIdxOne = (item: number, idx: number): trinary => { yield idx == 1 }
derive bad = () => { yield count(filter([1, 2, 3], isIdxOne)) }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "two_param_cb.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.True(s.T(),
		strings.Contains(err.Error(), "must be called as isIdxOne(...)") ||
			strings.Contains(err.Error(), "call is not permitted inside a derive"),
	)
}

func (s *IndexTestSuite) TestDerivePurityAllowsLambdaUnderHigherOrderBuiltin() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive ok = () => { yield any([1], (x: number): trinary => { yield true }) }
policy pol {
  let _s = 0
  rule r = { yield ok() }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "hof_lambda.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestPolicyDeriveConflictWithSeenIdentifier() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
policy pol {
  let dup = 1
  derive dup = () => { yield 1 }
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "pol_dup.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	err = idx.AddProgram(ctx, prog)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "derive declaration")
}

func (s *IndexTestSuite) TestDeriveSpanUsesLambdaWhenStatementAbsent() {
	lam := stubDeriveCovLambda()
	d := &Derive{Lambda: lam}
	s.Equal(lam.Span(), d.Span())
}

func (s *IndexTestSuite) TestDeriveDuplicateExportConflict() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive x = () => { yield 1 }
export derive x
export derive x
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "dup_export.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	err = idx.AddProgram(ctx, prog)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "derive export")
}

func (s *IndexTestSuite) TestAddProgramDuplicateShapeNameRejected() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
shape Dup { value: number }
shape Dup { other: string }
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "dup_shape2.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	err = idx.AddProgram(ctx, prog)
	require.Error(s.T(), err)
}

func (s *IndexTestSuite) TestPolicyDeriveLateHeaderRejected() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
policy pol {
  fact n:number
  derive late = () => { yield 1 }
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "late_derive.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	err = idx.AddProgram(ctx, prog)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "derive")
}

func (s *IndexTestSuite) TestDeriveExprWalkCoversRichExpressionKinds() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive fat = () => {
  let items = [1, 2]
  let mapped = transform items with ".length"
  yield mapped + (true ? 1 : 0) + (null ?: 0) + (-items[0]) + (1 as number) + (!false) + ({"a":1}["a"]) + (items is defined) + ([] is empty)
}
policy pol {
  let _s = 0
  rule allow = { yield fat() > 0 }
  export decision of allow
}
`
	prog, err := parser.NewParserFromString(src, "fat_walk.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestDerivePurityAllowsReduceWithLambdaCallback() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive sumThree = () => { yield reduce([1, 2, 3], 0, (acc: number, item: number): number => { yield acc + item }) }
policy pol {
  let _s = 0
  rule r = { yield sumThree() == 6 }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "reduce_cb.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestAddProgramWithCommentShapeAndDeriveExport() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
shape Counter {
  value: number
}
derive published = () => { yield 1 }
export derive published
export shape Counter
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "mixed.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}