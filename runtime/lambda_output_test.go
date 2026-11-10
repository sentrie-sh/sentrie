// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import "encoding/json"

// lambdaOutputPolicy exercises a lambda at every position output can hold one
// (#141): nested in a list or dict, as an attachment, in the trace through a
// filter argument, and as a rule's whole value.
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
  export decision of inList
  export decision of inDict
  export decision of deep
  export decision of filtered
  export decision of withAttach attach fn as id
  export decision of whole
  export decision of viaImport
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
	_, err := s.execLambdaOutput("whole")
	s.Require().Error(err)
	s.ErrorIs(err, ErrCallableDecision)
	s.ErrorContains(err, "com/ex/pol/whole")
}

// The same lambda-valued rule reached through import decision fails the importing rule.
func (s *RuntimeTestSuite) TestImportDecisionOfLambdaRuleFails() {
	_, err := s.execLambdaOutput("viaImport")
	s.Require().Error(err)
	s.ErrorIs(err, ErrCallableDecision)
}
