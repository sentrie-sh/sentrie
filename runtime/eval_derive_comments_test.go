// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

func (s *RuntimeTestSuite) TestExecRuleEvaluatesPolicyDeriveWithCommentsAndNestedBlock() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive onlyComment = (): number => {
  -- the body is a comment and a yield
  yield 10
}

policy pol {
  let _s = 0
  derive inner = (n: number): number => {
    -- policy-scoped derive
    let a = {
      -- nested block
      let b = n + onlyComment()
      -- before the nested yield
      yield b
    }
    -- before the yield
    yield a * 2
  }
  rule ok = { yield inner(1) }
  export decision of ok
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "comments2.sentrie", src: src})
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().NoError(err)
	s.Equal(int64(22), out.Decision.Value.Any())
}
