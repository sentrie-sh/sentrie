// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
)

const contributeSrc = `namespace com/ex

policy pol {
  fact user: document as user

  -- list rule: base list, then contributions in declaration order
  rule deny: list[string] = ["base"]
  contribute deny when user.disabled { yield "user disabled" }
  contribute deny when user.owner != user.id { yield "not owner" }
  contribute deny when false { yield "never" }
  contribute deny { yield user.missing }
  contribute deny { yield "always" }

  -- list rule whose own guard does not hold starts empty
  rule guarded: list[string] = when user.never { yield ["base"] }
  contribute guarded { yield "only" }

  -- boolean rule: OR of its own value and its contributions
  rule allow: boolean = false
  contribute allow when user.admin { yield true }
  contribute allow { yield user.owner == user.id }

  -- a rule that reads the contributed value
  rule reasons = count(deny)

  -- a contribution with a value of the wrong element type
  rule wrong: list[string] = []
  contribute wrong { yield user.id }

  -- a typed rule whose own body has the wrong type
  rule badbase: list[string] = "nope"

  export decision of deny
  export decision of guarded
  export decision of allow
    attach reasons as deny
  export decision of reasons
  export decision of wrong
  export decision of badbase
}
`

func (s *RuntimeTestSuite) contributeExec() *executorImpl {
	_, exec := s.mustBuildDeriveExecutor(s.T().Context(), deriveTestProgram{name: "contribute.sentrie", src: contributeSrc})
	return exec
}

func contributeUser(disabled, admin bool, owner string) map[string]any {
	return map[string]any{"user": map[string]any{
		"disabled": disabled, "admin": admin, "owner": owner, "id": "u1", "never": false,
	}}
}

func (s *RuntimeTestSuite) TestContributeListInDeclarationOrder() {
	exec := s.contributeExec()
	ctx := s.T().Context()

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "deny", contributeUser(true, false, "someone"))
	s.Require().NoError(err)
	s.Equal([]any{"base", "user disabled", "not owner", "always"}, box.ToBoundaryAny(out.Decision.Value))
	s.Equal(trinary.True, out.Decision.State)

	out, err = exec.ExecRule(ctx, "com/ex", "pol", "deny", contributeUser(false, false, "u1"))
	s.Require().NoError(err)
	s.Equal([]any{"base", "always"}, box.ToBoundaryAny(out.Decision.Value))

	// another rule referring to deny reads the combined list
	out, err = exec.ExecRule(ctx, "com/ex", "pol", "reasons", contributeUser(true, false, "someone"))
	s.Require().NoError(err)
	s.Equal(float64(4), out.Decision.Value.Any())

	// an attachment on another export reads the combined list
	out, err = exec.ExecRule(ctx, "com/ex", "pol", "allow", contributeUser(true, false, "someone"))
	s.Require().NoError(err)
	s.Equal([]any{"base", "user disabled", "not owner", "always"}, box.ToBoundaryAny(out.Attachments["reasons"]))
}

func (s *RuntimeTestSuite) TestContributeListGuardedBaseStartsEmpty() {
	out, err := s.contributeExec().ExecRule(s.T().Context(), "com/ex", "pol", "guarded", contributeUser(false, false, "u1"))
	s.Require().NoError(err)
	s.Equal([]any{"only"}, box.ToBoundaryAny(out.Decision.Value))
}

func (s *RuntimeTestSuite) TestContributeBooleanIsOr() {
	exec := s.contributeExec()
	ctx := s.T().Context()
	for name, c := range map[string]struct {
		admin bool
		owner string
		want  trinary.Value
	}{
		"no contribution holds":   {false, "someone", trinary.False},
		"guarded contribution":    {true, "someone", trinary.True},
		"unguarded contribution":  {false, "u1", trinary.True},
		"both contributions hold": {true, "u1", trinary.True},
	} {
		s.Run(name, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", "allow", contributeUser(false, c.admin, c.owner))
			s.Require().NoError(err)
			s.Equal(c.want, out.Decision.State)
			s.Equal(box.Trinary(c.want), out.Decision.Value)
		})
	}
}

func (s *RuntimeTestSuite) TestContributeTypeMismatchFails() {
	exec := s.contributeExec()
	ctx := s.T().Context()

	user := contributeUser(false, false, "u1")
	user["user"].(map[string]any)["id"] = 7
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "wrong", user)
	s.Require().Error(err)
	s.Equal(trinary.Unknown, out.Decision.State)
	s.Equal(box.Trinary(trinary.Unknown), out.Decision.Value)

	out, err = exec.ExecRule(ctx, "com/ex", "pol", "badbase", contributeUser(false, false, "u1"))
	s.Require().Error(err)
	s.Equal(trinary.Unknown, out.Decision.State)
	s.Equal(box.Trinary(trinary.Unknown), out.Decision.Value)
}

func (s *RuntimeTestSuite) TestContributeTraceRecordsEachContribution() {
	out, err := s.contributeExec().ExecRule(s.T().Context(), "com/ex", "pol", "deny", contributeUser(false, false, "u1"))
	s.Require().NoError(err)

	var contributions []*trace.Node
	var walk func(n *trace.Node)
	walk = func(n *trace.Node) {
		if n == nil {
			return
		}
		if n.Op == "rule-contribution" {
			contributions = append(contributions, n)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(out.RuleNode)

	s.Require().Len(contributions, 5)
	skipped := make([]any, len(contributions))
	for i, c := range contributions {
		s.Equal(i, c.Meta["index"])
		skipped[i] = c.Meta["skipped"]
	}
	s.Equal([]any{"when", "when", "when", "undefined", nil}, skipped)
}
