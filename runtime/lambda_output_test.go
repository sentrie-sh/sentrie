// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"

	"github.com/sentrie-sh/sentrie/trinary"
)

// lambdaOutputPolicy exercises a lambda at every position output can hold one
// (#141): nested in a list or dict, as an attachment, in the trace through a
// filter argument, and as a rule's whole value (from the body or from a
// default branch, or alongside a failing attachment). helper is an unreported rule whose value is a lambda, used
// as a callback by usesHelper.
const lambdaOutputPolicy = `namespace com/ex

policy pol {
  let id = (x) => { yield x }
  rule inList = { yield [1, id] }
  rule inDict = { yield { "f": id, "n": 1 } }
  rule deep = { yield { "xs": [{ "f": id }] } }
  rule filtered = { yield filter([1, 2, 3], (x) => { yield x > 1 }) }
  rule withAttach = { yield true }
  rule whole = { yield id }
  rule viaImport = import decision whole from com/ex/pol
  rule fromDefault = default id when false { yield true }
  rule helper = (x) => { yield x > 1 }
  rule usesHelper = { yield filter([1, 2, 3], helper) }
  rule plainBadAttach = { yield 1 }
  rule withBadAttach = { yield (x) => { yield x } }
  export decision of inList
  export decision of inDict
  export decision of deep
  export decision of filtered
  export decision of withAttach attach fn as id
  export decision of whole
  export decision of viaImport
  export decision of fromDefault
  export decision of usesHelper
  export decision of withBadAttach attach boom as 1 / 0
  export decision of plainBadAttach attach boom as 1 / 0
}
`

func (s *RuntimeTestSuite) execLambdaOutput(rule string) (*ExecutorOutput, error) {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "lambda_output.sentrie", src: lambdaOutputPolicy})
	return exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
}

// A lambda nested in a decision value encodes as the callable placeholder at any depth.
func (s *RuntimeTestSuite) TestLambdaNestedInDecisionValueEncodesAsPlaceholder() {
	cases := map[string]string{
		"inList": `[1,{"kind":"callable"}]`,
		"inDict": `{"f":{"kind":"callable"},"n":1}`,
		"deep":   `{"xs":[{"f":{"kind":"callable"}}]}`,
	}
	for rule, want := range cases {
		out, err := s.execLambdaOutput(rule)
		s.Require().NoError(err, rule)
		raw, err := json.Marshal(out.Decision.reported().Value)
		s.Require().NoError(err, rule)
		s.JSONEq(want, string(raw), rule)
	}
}

// A lambda attachment is nested output, not an error.
func (s *RuntimeTestSuite) TestLambdaAttachmentEncodesAsPlaceholder() {
	out, err := s.execLambdaOutput("withAttach")
	s.Require().NoError(err)
	raw, err := json.Marshal(out)
	s.Require().NoError(err)
	var got struct {
		Attachments map[string]json.RawMessage `json:"attachments"`
	}
	s.Require().NoError(json.Unmarshal(raw, &got))
	s.JSONEq(`{"kind":"callable"}`, string(got.Attachments["fn"]))
}

// A lambda only in the trace (a filter argument) no longer fails the encode of the whole output.
func (s *RuntimeTestSuite) TestLambdaInTraceEncodes() {
	out, err := s.execLambdaOutput("filtered")
	s.Require().NoError(err)
	raw, err := json.Marshal(out)
	s.Require().NoError(err)
	s.Contains(string(raw), `{"kind":"callable"}`)
}

// A rule whose whole value is a lambda is that rule's error, with no decision.
func (s *RuntimeTestSuite) TestRuleWholeValueLambdaIsRuleError() {
	for _, rule := range []string{"whole", "fromDefault"} {
		out, err := s.execLambdaOutput(rule)
		s.Require().Error(err, rule)
		s.ErrorIs(err, ErrCallableDecision, rule)
		s.ErrorContains(err, "com/ex/pol/"+rule, rule)
		s.Require().NotNil(out, rule)
		s.Equal(trinary.Unknown, out.Decision.State, rule)
		s.Empty(out.Attachments, rule)
		s.Require().NotNil(out.RuleNode, rule)
		s.Contains(out.RuleNode.Err, ErrCallableDecision.Error(), rule)
	}
}

// The callable error wins over an attachment error: the attachments are
// dropped anyway, so their failure must not mask why the rule failed.
func (s *RuntimeTestSuite) TestRuleWholeValueLambdaWinsOverAttachmentError() {
	out, err := s.execLambdaOutput("withBadAttach")
	s.Require().Error(err)
	s.ErrorIs(err, ErrCallableDecision)
	s.Equal("rule com/ex/pol/withBadAttach: "+ErrCallableDecision.Error(), err.Error())
	s.Require().NotNil(out)
	s.Equal(trinary.Unknown, out.Decision.State)
	s.False(out.Decision.reported().Value.IsCallable())
	s.NotContains(err.Error(), "divide by zero")
	s.Nil(out.Attachments)
	s.Require().NotNil(out.RuleNode)
	s.Contains(out.RuleNode.Err, ErrCallableDecision.Error())
}

// An attachment error on a rule whose value is not a callable still surfaces as the rule's error.
func (s *RuntimeTestSuite) TestAttachmentErrorOnPlainRuleStillSurfaces() {
	_, err := s.execLambdaOutput("plainBadAttach")
	s.Require().Error(err)
	s.NotErrorIs(err, ErrCallableDecision)
	s.ErrorContains(err, "divide by zero")
}

// An unreported rule whose value is a lambda is a helper: referenced from
// another rule it is a callback, not an error.
func (s *RuntimeTestSuite) TestLambdaHelperRuleIsCallableFromAnotherRule() {
	out, err := s.execLambdaOutput("usesHelper")
	s.Require().NoError(err)
	raw, err := json.Marshal(out.Decision.reported().Value)
	s.Require().NoError(err)
	s.JSONEq(`[2,3]`, string(raw))
}

// The same lambda-valued rule reached through import decision fails the importing rule.
func (s *RuntimeTestSuite) TestImportDecisionOfLambdaRuleFails() {
	_, err := s.execLambdaOutput("viaImport")
	s.Require().Error(err)
	s.ErrorIs(err, ErrCallableDecision)
}
