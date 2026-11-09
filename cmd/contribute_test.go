// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"errors"

	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *CmdTestSuite) TestValidateCmdReportsContributionErrors() {
	cases := map[string]struct {
		body string
		want string
	}{
		"unknown rule": {
			body: `rule a = true
  contribute deny { yield "x" }
  export decision of a`,
			want: "cannot contribute to unknown rule 'deny'",
		},
		"rule declared later": {
			body: `contribute deny { yield "x" }
  rule deny: list[string] = []
  export decision of deny`,
			want: "cannot contribute to unknown rule 'deny'",
		},
		"untyped rule": {
			body: `rule deny = []
  contribute deny { yield "x" }
  export decision of deny`,
			want: "the rule declares no type",
		},
		"non-list, non-boolean rule": {
			body: `rule deny: number = 1
  contribute deny { yield 2 }
  export decision of deny`,
			want: "the rule declares type number",
		},
		"literal of the wrong kind": {
			body: `rule deny: list[string] = []
  contribute deny { yield 42 }
  export decision of deny`,
			want: "yields a number, but the rule takes string",
		},
	}
	for name, c := range cases {
		s.Run(name, func() {
			dir := s.writeTestPack("namespace com/ex\npolicy pol {\n  " + c.body + "\n}\n")
			err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
			s.Require().Error(err)
			s.True(errors.Is(err, xerr.ErrIndex))
			s.Contains(err.Error(), c.want)
			s.Contains(err.Error(), "policy.sentrie:")
		})
	}
}

func (s *CmdTestSuite) TestValidateCmdReportsContributionToAnImportedRule() {
	dir := s.writeTestPack(`namespace com/ex
policy other {
  rule allow: boolean = true
  export decision of allow
}
policy pol {
  rule allow: boolean = import decision allow from com/ex/other
  contribute allow { yield true }
  export decision of allow
}
`)
	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), "imports its decision")
}

func (s *CmdTestSuite) TestValidateCmdAcceptsContributions() {
	dir := s.writeTestPack(contributeCmdPolicy)
	s.Require().NoError(runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir}))
}

const contributeCmdPolicy = `namespace com/ex
policy pol {
  fact user: document as user
  rule deny: list[string] = []
  contribute deny when user.disabled { yield "user disabled" }
  contribute deny when user.owner != user.id { yield "not owner" }
  contribute deny { yield user.reason }
  rule allow: boolean = false
  contribute allow when user.admin { yield true }
  export decision of deny
  export decision of allow
    attach reasons as deny
}
`

func (s *CmdTestSuite) TestExecCmdJSONOutputCarriesContributedValues() {
	dir := s.writeTestPack(contributeCmdPolicy)
	facts := `{"user": {"disabled": true, "admin": true, "owner": "x", "id": "y", "reason": "r"}}`

	run := func(rule string) map[string]any {
		var runErr error
		out := s.captureStdout(func() {
			runErr = runExecCLI(s.T().Context(), []string{rule, "--pack-location", dir, "--facts", facts, "--output", "json"})
		})
		s.Require().NoError(runErr)
		var decoded []map[string]any
		s.Require().NoError(json.Unmarshal([]byte(out), &decoded), out)
		s.Require().Len(decoded, 1)
		return decoded[0]
	}

	deny := run("com/ex/pol/deny")
	value := deny["decision"].(map[string]any)["value"]
	s.Equal([]any{"user disabled", "not owner", "r"}, value)

	allow := run("com/ex/pol/allow")
	s.Equal(true, allow["decision"].(map[string]any)["value"])
	s.Equal([]any{"user disabled", "not owner", "r"}, allow["attachments"].(map[string]any)["reasons"])
}

func (s *CmdTestSuite) TestValidateCmdChecksCapabilitiesAndWarnsInsideContributions() {
	dir := s.writeTestPack(`namespace com/ex
policy pol {
  rule deny: list[string] = []
  contribute deny when count([1]) > 0 { yield "x" }
  rule seen: list[number] = []
  contribute seen { yield now() }
  export decision of deny
  export decision of seen
}
`)
	caps := s.writeCapabilities(`{"version": 1, "builtins": ["now"], "modules": []}`)
	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir, "--capabilities", caps})
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))
	s.Contains(err.Error(), `builtin "count"`)
	s.Contains(err.Error(), "policy.sentrie:4:")

	var runErr error
	out := s.captureStderr(func() {
		runErr = runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	})
	s.Require().NoError(runErr)
	s.Contains(out, `builtin "now" is nondeterministic`)
	s.Contains(out, "policy.sentrie:6:")
}

func (s *CmdTestSuite) TestExecCmdFailsWhenAContributionFails() {
	dir := s.writeTestPack(`namespace com/ex
policy pol {
  fact user: document as user
  rule deny: list[string] = []
  contribute deny { yield user.id }
  export decision of deny
}
`)
	err := runExecCLI(s.T().Context(), []string{"com/ex/pol/deny", "--pack-location", dir, "--facts", `{"user": {"id": 7}}`})
	s.Require().Error(err)
}
