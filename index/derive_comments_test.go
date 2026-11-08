// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/stretchr/testify/require"
)

// validateDeriveCommentSource parses src, adds it to a fresh index and returns
// the result of Validate.
func (s *IndexTestSuite) validateDeriveCommentSource(src, name string) error {
	ctx := s.T().Context()
	idx := CreateIndex()
	prog, err := parser.NewParserFromString(src, name).ParseProgram(ctx)
	require.NoError(s.T(), err)
	require.NoError(s.T(), idx.AddProgram(ctx, prog))
	return idx.Validate(ctx)
}

func (s *IndexTestSuite) TestDeriveCommentOnlyBeforeYieldPassesValidation() {
	// A comment is the only statement in the body, in a namespace derive, in a
	// policy derive and inside a lambda body.
	src := `namespace com/ex
derive nsOnly = () => {
  -- nothing but a comment
  yield 1
}
policy pol {
  let _s = 0
  derive polOnly = () => {
    -- nothing but a comment
    yield nsOnly()
  }
  rule allow = { yield polOnly() == 1 }
  export decision of allow
}
`
	require.NoError(s.T(), s.validateDeriveCommentSource(src, "only.sentra"))
}

func (s *IndexTestSuite) TestDeriveCommentInsideNestedBlockPassesValidation() {
	// A block expression nested in a let value and in a yield, each with comments.
	src := `namespace com/ex
derive nested = (n: number): number => {
  let a = {
    -- nested block, before its let
    let b = n + 1
    -- nested block, before its yield
    yield b
  }
  yield {
    -- nested block in the yield
    yield a
  }
}
policy pol {
  let _s = 0
  rule allow = { yield nested(1) == 2 }
  export decision of allow
}
`
	require.NoError(s.T(), s.validateDeriveCommentSource(src, "nested.sentra"))
}

func (s *IndexTestSuite) TestDeriveCommentInPolicyDerivePassesValidation() {
	src := `namespace com/ex
policy pol {
  let _s = 0
  derive inner = (n: number): number => {
    -- policy-scoped derive
    let m = n + 1
    -- between lets
    let k = m * 2
    -- before the yield
    yield k
  }
  rule allow = { yield inner(1) == 4 }
  export decision of allow
}
`
	require.NoError(s.T(), s.validateDeriveCommentSource(src, "poli.sentra"))
}

func (s *IndexTestSuite) TestDeriveCommentDoesNotHideRejectedConstructs() {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "impure builtin in a let after a comment",
			body: "-- ahead of an impure call\n  let t = now()\n  yield t",
			want: `builtin "now" is nondeterministic and is not permitted inside a derive`,
		},
		{
			name: "impure builtin in the yield after a comment",
			body: "-- ahead of an impure call\n  yield now()",
			want: `builtin "now" is nondeterministic and is not permitted inside a derive`,
		},
		{
			name: "impure builtin inside a nested block after a comment",
			body: "let t = {\n    -- ahead of an impure call\n    yield now()\n  }\n  yield t",
			want: `builtin "now" is nondeterministic and is not permitted inside a derive`,
		},
		{
			name: "unknown identifier after a comment",
			body: "-- ahead of a bad name\n  yield mystery",
			want: `identifier "mystery" is not available`,
		},
		{
			name: "module call after a comment",
			body: "-- ahead of a module call\n  yield mod.fn(1)",
			want: "TypeScript module calls are not permitted",
		},
		{
			name: "yielded lambda after a comment",
			body: "-- ahead of a lambda\n  yield (x: number) => { yield x }",
			want: "derive cannot yield a lambda value",
		},
		{
			name: "let used before it is declared",
			body: "-- a comment does not bring b into scope\n  let a = b\n  -- another\n  let b = 1\n  yield a",
			want: `identifier "b" is not available`,
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			src := "namespace com/ex\nderive d = (): number => {\n  " + tc.body + "\n}\n" +
				"policy pol {\n  let _s = 0\n  rule r = { yield true }\n  export decision of r\n}\n"
			err := s.validateDeriveCommentSource(src, "reject.sentra")
			require.Error(s.T(), err)
			require.Contains(s.T(), err.Error(), tc.want)
		})
	}
}

func (s *IndexTestSuite) TestDeriveCycleDetectedThroughCommentedPolicyDerives() {
	// The callee sits behind a comment in a policy-scoped derive, in a nested
	// block and after a comment in a lambda body.
	src := `namespace com/ex
policy pol {
  let _s = 0
  derive a = () => {
    -- ahead of the call
    let v = {
      -- nested block ahead of the call
      yield com/ex/pol/b()
    }
    yield v
  }
  derive b = () => {
    -- ahead of the call
    yield first([1], (x: number) => {
      -- inside a lambda body
      yield x == com/ex/pol/a()
    })
  }
  rule allow = { yield true }
  export decision of allow
}
`
	err := s.validateDeriveCommentSource(src, "cycpol.sentra")
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "cyclic derive dependency")
}
