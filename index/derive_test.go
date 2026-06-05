// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"strings"

	"github.com/sentrie-sh/sentrie/parser"
	"github.com/stretchr/testify/require"
)

func (s *IndexTestSuite) TestDeriveAddProgramResolveAndSpan() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex

derive helper = () => { yield 1 }

policy pol {
  let _seed = 0
  derive inner = () => { yield helper() }
  rule allow = { yield inner() == 1 }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "one.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))

	d, err := idx.ResolveDerive("com/ex/pol/inner")
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), d.String())
	_ = d.Span()
}

func (s *IndexTestSuite) TestDeriveCycleViaSlashFQNDetected() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex

derive a = () => { yield com/ex/b() }
derive b = () => { yield com/ex/a() }

policy pol {
  let _seed = 0
  rule allow = { yield true }
  export decision of allow
}
`
	p := parser.NewParserFromString(src, "cyc.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.True(s.T(), strings.Contains(err.Error(), "cyclic derive") || strings.Contains(err.Error(), "derive dependency"))
}

func (s *IndexTestSuite) TestDeriveDuplicateFQNRejected() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src1 := `namespace com/ex
derive dup = () => { yield 1 }
policy p1 {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	src2 := `namespace com/ex
derive dup = () => { yield 2 }
policy p2 {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog1, err := parser.NewParserFromString(src1, "a.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog1))
	prog2, err := parser.NewParserFromString(src2, "b.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	err = idx.AddProgram(ctx, prog2)
	require.Error(s.T(), err)
}

func (s *IndexTestSuite) TestExportDeriveUnknownNameRejected() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
export derive missing
derive x = () => { yield 1 }
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "e.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	err = idx.AddProgram(ctx, prog)
	require.Error(s.T(), err)
}

func (s *IndexTestSuite) TestDeriveFatBodyWalkPassesValidation() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive leaf = () => { yield 1 }
derive fat = () => {
  let v = leaf()
  yield v + (true ? 1 : 0) + (null ?: 2) + (-1) + [1, 2][0] + (1 as number) + (!false) + ({"a":1}["a"])
}
policy pol {
  let _s = 0
  rule allow = { yield fat() > 0 }
  export decision of allow
}
`
	prog, err := parser.NewParserFromString(src, "fat.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestResolveDeriveNotFound() {
	idx := CreateIndex()
	_, err := idx.ResolveDerive("com/nope/never")
	require.Error(s.T(), err)
}

func (s *IndexTestSuite) TestAddProgramExportDeriveAndVerifyExported() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive published = () => { yield 1 }
export derive published
policy p {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "exp.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))

	ns := idx.Namespaces["com/ex"]
	require.NoError(s.T(), ns.VerifyDeriveExported("published"))
	require.Error(s.T(), ns.VerifyDeriveExported("unpublished"))
}

func (s *IndexTestSuite) TestAlphaSecretNotExportedInIndex() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/alpha
derive secret = () => { yield 1 }
policy pa {
  let _s = 0
  rule x = { yield true }
  export decision of x
}
`
	prog, err := parser.NewParserFromString(src, "a.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
	ns := idx.Namespaces["com/alpha"]
	require.Error(s.T(), ns.VerifyDeriveExported("secret"))
}

func (s *IndexTestSuite) TestDerivePurityRejectsFieldAccessModuleCall() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield mod.fn(1) }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "mod.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "TypeScript module calls are not permitted")
}

func (s *IndexTestSuite) TestDerivePurityRejectsFactIdentifier() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
policy pol {
  fact user:string
  let _s = 0
  derive d = () => { yield user }
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "fact.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "facts are not available inside a derive")
}

func (s *IndexTestSuite) TestDerivePurityRejectsYieldLambda() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield (x: number): number => { yield x } }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "lam.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "derive cannot yield a lambda value")
}

func (s *IndexTestSuite) TestDerivePurityLetArithmeticPasses() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = (x: number): number => {
  let doubled = x * 2
  yield doubled + 1
}
policy pol {
  let _s = 0
  rule r = { yield d(1) == 3 }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "let.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestDerivePurityAllowsLambdaInsidePureBuiltinCall() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield count(filter([1, 2], (a: number): trinary => { yield a == 1 })) }
policy pol {
  let _s = 0
  rule r = { yield d() == 1 }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "filt.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestDerivePurityAllowsDeriveAsCallbackToPureBuiltin() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive pred = (a: number): trinary => { yield a == 1 }
derive d = () => { yield count(filter([1, 2], pred)) }
policy pol {
  let _s = 0
  rule r = { yield d() == 1 }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "filt_derive_cb.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestDeriveDefineSiteOrderHelperAfterCallerFails() {
	ctx := s.T().Context()
	idx := CreateIndex()
	srcCaller := `namespace com/ex
derive caller = () => { yield helper() }
policy pol {
  let _s = 0
  rule r = { yield caller() == 1 }
  export decision of r
}
`
	srcHelper := `namespace com/ex
derive helper = () => { yield 1 }
policy pol2 {
  let _s = 0
  rule r2 = { yield true }
  export decision of r2
}
`
	p1, err := parser.NewParserFromString(srcCaller, "caller.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, p1))
	p2, err := parser.NewParserFromString(srcHelper, "helper.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, p2))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
}

func (s *IndexTestSuite) TestDeriveDefineSiteOrderHelperFirstPasses() {
	ctx := s.T().Context()
	idx := CreateIndex()
	srcHelper := `namespace com/ex
derive helper = () => { yield 1 }
policy pol2 {
  let _s = 0
  rule r2 = { yield true }
  export decision of r2
}
`
	srcCaller := `namespace com/ex
derive caller = () => { yield helper() }
policy pol {
  let _s = 0
  rule r = { yield caller() == 1 }
  export decision of r
}
`
	p1, err := parser.NewParserFromString(srcHelper, "h.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, p1))
	p2, err := parser.NewParserFromString(srcCaller, "c.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, p2))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestDerivePuritySlashCalleeInRuleYieldCompletes() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive ns = () => { yield 1 }
policy pol {
  let _seed = 0
  rule gate = { yield com/ex/ns() == 1 }
  export decision of gate
}
`
	p := parser.NewParserFromString(src, "slash.sentra")
	prog, err := p.ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}

func (s *IndexTestSuite) TestDerivePurityRejectsBareDefineShortDerive() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive helper = () => { yield 1 }
derive bad = () => { yield helper }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "bare.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "must be called as helper(...)")
}

func (s *IndexTestSuite) TestDerivePurityRejectsRuleIdentifier() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
policy pol {
  let _s = 0
  rule gate = { yield true }
  derive d = () => { yield gate }
  export decision of gate
}
`
	prog, err := parser.NewParserFromString(src, "rule.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "rules cannot be referenced inside a derive")
}

func (s *IndexTestSuite) TestDerivePurityRejectsUnknownIdentifier() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => { yield mystery }
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "unkid.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "identifier \"mystery\" is not available")
}

func (s *IndexTestSuite) TestDerivePurityRejectsDisallowedCall() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive d = () => {
  let items = [1]
  yield items[0](1)
}
policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	prog, err := parser.NewParserFromString(src, "call.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "call is not permitted inside a derive")
}

func (s *IndexTestSuite) TestDerivePurityRejectsCrossPolicySlashFQN() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
policy polA {
  let _s = 0
  derive secret = () => { yield 1 }
  rule ra = { yield true }
  export decision of ra
}
policy polB {
  let _s = 0
  derive caller = () => { yield com/ex/polA/secret() }
  rule rb = { yield caller() == 1 }
  export decision of rb
}
`
	prog, err := parser.NewParserFromString(src, "crosspol.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "not visible from policy")
}

func (s *IndexTestSuite) TestDerivePurityRejectsNamespaceDeriveCallingPolicyDerive() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
derive nsCaller = () => { yield com/ex/polA/secret() }
policy polA {
  let _s = 0
  derive secret = () => { yield 1 }
  rule ra = { yield true }
  export decision of ra
}
`
	prog, err := parser.NewParserFromString(src, "nspol.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	err = idx.Validate(ctx)
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "policy-scoped and not visible outside its policy")
}

func (s *IndexTestSuite) TestDerivePurityAllowsSamePolicySlashFQN() {
	ctx := s.T().Context()
	idx := CreateIndex()
	src := `namespace com/ex
policy polA {
  let _s = 0
  derive secret = () => { yield 1 }
  derive caller = () => { yield com/ex/polA/secret() }
  rule gate = { yield caller() == 1 }
  export decision of gate
}
`
	prog, err := parser.NewParserFromString(src, "samepol.sentra").ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	require.NoError(s.T(), idx.Validate(ctx))
}
