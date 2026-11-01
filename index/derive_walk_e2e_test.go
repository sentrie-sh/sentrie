// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"fmt"
	"strings"

	sast "github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/tokens"
)

// deriveWalkWrappers wraps a probe expression (%s) in one expression kind each,
// so the derive checks must reach the probe through that kind's children. The
// kinds are reached through forEachDeriveExprChild (and so forEachExprChild); a
// kind whose children the walker dropped would hide the probe from the check.
var deriveWalkWrappers = []struct{ name, tmpl string }{
	{"call argument", "count([%s])"},
	{"infix left", "%s + 1"},
	{"infix right", "1 + %s"},
	{"unary", "-(%s)"},
	{"ternary condition", "(%s) ? 1 : 2"},
	{"ternary then", "true ? %s : 2"},
	{"ternary else", "true ? 1 : %s"},
	{"elvis else", "null ?: %s"},
	{"index base", "[%s][0]"},
	{"index key", "[1][%s]"},
	{"field access base", "(%s).x"},
	{"list element", "[1, %s]"},
	{"map value", `{"k": %s}`},
	{"cast", "%s as number"},
	{"transform argument", `transform [%s] with ".length"`},
	{"is defined", "%s is defined"},
	{"is empty", "[%s] is empty"},
	{"is value side", "1 is (%s)"},
}

// deriveWalkSource builds a policy whose derive d yields body. The extra
// derives are appended to the namespace.
func deriveWalkSource(body string, extra string) string {
	return fmt.Sprintf(`namespace com/ex
%s
policy pol {
  fact user:string
  let _s = 0
  derive d = () => { yield %s }
  rule r = { yield true }
  export decision of r
}
`, extra, body)
}

func (s *IndexTestSuite) validateDeriveWalkSource(src string) error {
	ctx := s.T().Context()
	idx := CreateIndex()
	prog, err := parser.NewParserFromString(src, "walk.sentrie").ParseProgram(ctx)
	s.Require().NoError(err, src)
	s.Require().NoError(idx.AddProgram(ctx, prog))
	return idx.Validate(ctx)
}

// A derive body is walked by forEachDeriveExprChild; an impure construct nested
// in any expression kind must still be refused with the same message as when it
// is the whole body.
func (s *IndexTestSuite) TestDerivePurityRejectsImpureCallNestedInEveryExpressionKind() {
	probes := []struct{ name, expr, msg string }{
		{"nondeterministic builtin", "now()", `builtin "now" is nondeterministic and is not permitted inside a derive`},
		{"typescript module call", "mod.fn(1)", "TypeScript module calls are not permitted inside a derive"},
		{"fact reference", "user", `facts are not available inside a derive ("user")`},
	}
	for _, p := range probes {
		for _, w := range deriveWalkWrappers {
			// `1 is user` is refused earlier, when the index resolves the `is`
			// operand, with a different message; the walk is not reached.
			if w.name == "is value side" && p.name == "fact reference" {
				continue
			}
			s.Run(p.name+"/"+w.name, func() {
				err := s.validateDeriveWalkSource(deriveWalkSource(fmt.Sprintf(w.tmpl, p.expr), ""))
				s.Require().Error(err)
				s.Contains(err.Error(), "derive com/ex/pol/d: ")
				s.Contains(err.Error(), p.msg)
			})
		}
	}
}

// The control for the test above: the same wrappers around a pure probe pass,
// so the rejections are caused by the probe and not by the wrapper.
func (s *IndexTestSuite) TestDerivePurityAcceptsPureProbeInEveryExpressionKind() {
	for _, w := range deriveWalkWrappers {
		s.Run(w.name, func() {
			s.NoError(s.validateDeriveWalkSource(deriveWalkSource(fmt.Sprintf(w.tmpl, "1"), "")))
		})
	}
}

// Derive cycle detection finds a call to another derive through every
// expression kind, not only a call at the top of the body.
func (s *IndexTestSuite) TestDeriveCycleDetectedThroughEveryExpressionKind() {
	for _, w := range deriveWalkWrappers {
		s.Run(w.name, func() {
			src := fmt.Sprintf(`namespace com/ex
derive a = () => { yield %s }
derive b = () => { yield com/ex/a() }

policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`, fmt.Sprintf(w.tmpl, "com/ex/b()"))
			err := s.validateDeriveWalkSource(src)
			s.Require().Error(err)
			s.Contains(err.Error(), "cyclic derive dependency")
			s.Contains(err.Error(), "com/ex/a")
			s.Contains(err.Error(), "com/ex/b")
		})
	}
}

func (s *IndexTestSuite) TestDeriveCycleDetectedThroughLambdaAndLet() {
	cases := map[string]string{
		"lambda body yield": `any([1], (x: number): trinary => { yield com/ex/b() == 1 })`,
		"lambda body let":   `any([1], (x: number): trinary => { let t = com/ex/b() yield t == 1 })`,
	}
	for name, body := range cases {
		s.Run(name, func() {
			src := fmt.Sprintf(`namespace com/ex
derive a = () => { yield %s }
derive b = () => { yield com/ex/a() }

policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`, body)
			err := s.validateDeriveWalkSource(src)
			s.Require().Error(err)
			s.Contains(err.Error(), "cyclic derive dependency")
		})
	}

	s.Run("body let", func() {
		src := `namespace com/ex
derive a = () => {
  let t = com/ex/b()
  yield t
}
derive b = () => { yield com/ex/a() }

policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
		err := s.validateDeriveWalkSource(src)
		s.Require().Error(err)
		s.Contains(err.Error(), "cyclic derive dependency")
	})
}

func (s *IndexTestSuite) TestDeriveCycleDetectedThroughPipeline() {
	src := `namespace com/ex
derive b = (n: number): number => { yield com/ex/a(n) }
derive a = (n: number): number => { yield n |> b() }

policy pol {
  let _s = 0
  rule r = { yield true }
  export decision of r
}
`
	err := s.validateDeriveWalkSource(src)
	s.Require().Error(err)
	s.Contains(err.Error(), "cyclic derive dependency")
}

// The control for the cycle tests: a chain of derives that never returns to its
// start is accepted, however deeply the calls are nested.
func (s *IndexTestSuite) TestDeriveChainWithoutCycleAccepted() {
	src := strings.Join([]string{
		`namespace com/ex`,
		`derive c = () => { yield 1 }`,
		`derive b = () => { yield any([1], (x: number): trinary => { let t = com/ex/c() yield t == 1 }) }`,
		`derive a = (n: number): number => { yield [n |> c(), com/ex/c()][0] }`,
		`policy pol {`,
		`  let _s = 0`,
		`  rule r = { yield true }`,
		`  export decision of r`,
		`}`,
	}, "\n")
	s.NoError(s.validateDeriveWalkSource(src))
}

// The derive walker yields only real children: an absent operand (a call
// argument, the value side of an `is`, a block's yield) is skipped rather than
// handed to the caller as a nil expression.
func (s *IndexTestSuite) TestForEachDeriveExprChildSkipsNilChildren() {
	r := tokens.Range{}
	id := sast.NewIdentifier("x", r)
	nodes := map[string]sast.Expression{
		"call with nil argument": sast.NewCallExpression(id, []sast.Expression{nil, id}, false, nil, r),
		"is with no value side":  sast.NewIsExpression(id, nil, nil, r),
		"block without yield":    sast.NewBlockExpression(nil, nil, r),
	}
	for name, n := range nodes {
		s.Run(name, func() {
			s.NoError(forEachDeriveExprChild(n, func(c sast.Expression) error {
				s.NotNil(c, "a nil child was yielded")
				return nil
			}))
		})
	}
}
