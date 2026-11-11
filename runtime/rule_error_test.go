// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"
	stdErr "errors"
	"net/http"
	"time"

	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

// ruleErrorPolicy has two healthy rules and three failing ones (#143), one of
// them a lambda-valued rule, which #141 makes that rule's error.
const ruleErrorPolicy = `namespace com/ex

policy pol {
  fact user: any
  let id = (x) => { yield x }
  rule a = { yield true }
  rule b = { yield 1 }
  rule badSub = { yield "x" - 1 }
  rule badDiv = { yield 1 / 0 }
  rule whole = { yield id }
  export decision of a
  export decision of b
  export decision of badSub
  export decision of badDiv
  export decision of whole
}
`

func (s *RuntimeTestSuite) execRuleErrorPolicy(facts map[string]any) (map[string]*ExecutorOutput, error) {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "rule_error.sentrie", src: ruleErrorPolicy})
	outputs, err := exec.ExecPolicy(ctx, "com/ex", "pol", facts)
	byName := map[string]*ExecutorOutput{}
	for _, o := range outputs {
		s.Require().NotContains(byName, o.RuleName, "one output per rule")
		byName[o.RuleName] = o
	}
	return byName, err
}

// One failing rule no longer fails the policy: every rule has an output, the
// failing ones carry their own error and report unknown.
func (s *RuntimeTestSuite) TestExecPolicyReportsEachRuleErrorAndKeepsTheOthers() {
	byName, err := s.execRuleErrorPolicy(map[string]any{"user": map[string]any{}})

	var ruleErrs RuleErrors
	s.Require().True(stdErr.As(err, &ruleErrs), "error is RuleErrors: %v", err)
	s.Len(ruleErrs, 3)
	s.Require().Len(byName, 5)

	s.Nil(byName["a"].Error)
	s.Equal(trinary.True, byName["a"].Decision.State)
	s.Nil(byName["b"].Error)

	for rule, detail := range map[string]string{
		"badSub": "left operand is not a number",
		"badDiv": "divide by zero",
		"whole":  "cannot be a callable",
	} {
		out := byName[rule]
		s.Require().NotNil(out.Error, rule)
		s.Equal("Evaluation Failed", out.Error.Title, rule)
		s.Equal(http.StatusInternalServerError, out.Error.Status, rule)
		s.Contains(out.Error.Detail, detail, rule)
		s.Equal(trinary.Unknown, out.Decision.State, rule)
	}
	s.ErrorIs(byName["whole"].Error, ErrCallableDecision)
	s.ErrorIs(err, ErrCallableDecision)
}

// A succeeding rule's JSON has no error field; a failing rule's has a
// problem-document style object next to its decision.
func (s *RuntimeTestSuite) TestRuleErrorJSONShape() {
	byName, _ := s.execRuleErrorPolicy(map[string]any{"user": map[string]any{}})

	var ok map[string]json.RawMessage
	raw, err := json.Marshal(byName["a"])
	s.Require().NoError(err)
	s.Require().NoError(json.Unmarshal(raw, &ok))
	s.NotContains(ok, "error")

	raw, err = json.Marshal(byName["badDiv"])
	s.Require().NoError(err)
	var failed struct {
		Decision map[string]any `json:"decision"`
		Error    map[string]any `json:"error"`
	}
	s.Require().NoError(json.Unmarshal(raw, &failed))
	s.Equal("Evaluation Failed", failed.Error["title"])
	s.EqualValues(http.StatusInternalServerError, failed.Error["status"])
	s.Contains(failed.Error["detail"], "divide by zero")
	s.Len(failed.Error, 3)
	s.NotNil(failed.Decision)
}

// A missing required fact is the request's error, not a failure of each rule
// (#143): ExecPolicy returns it alone, with no outputs.
func (s *RuntimeTestSuite) TestExecPolicyMissingFactIsARequestError() {
	byName, err := s.execRuleErrorPolicy(nil)

	s.Require().Error(err)
	var ruleErrs RuleErrors
	s.False(stdErr.As(err, &ruleErrs), "error is not RuleErrors: %v", err)
	var invErr xerr.InvalidInvocationError
	s.ErrorAs(err, &invErr)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "required fact not found: user")
	s.Empty(byName)
}

// A fact that fails its declared type is a request error too, and so is an
// imported policy's fact failing its type (#147, user decision 11).
func (s *RuntimeTestSuite) TestExecPolicyFactTypeErrorIsARequestError() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "fact_type.sentrie", src: `namespace com/ex

policy typed {
  fact n: number
  rule ok = { yield n > 0 }
  export decision of ok
}

policy importer {
  rule fine = { yield true }
  rule viaImport = import decision ok from com/ex/typed with n as "x"
  export decision of fine
  export decision of viaImport
}
`})

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "typed", map[string]any{"n": "x"})
	s.Require().Error(err)
	var ruleErrs RuleErrors
	s.False(stdErr.As(err, &ruleErrs), "error is not RuleErrors: %v", err)
	s.Nil(outputs)
	s.Contains(err.Error(), "is not a number")

	outputs, err = exec.ExecPolicy(ctx, "com/ex", "importer", nil)
	s.Require().Error(err)
	s.False(stdErr.As(err, &ruleErrs), "error is not RuleErrors: %v", err)
	s.Nil(outputs)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "is not a number")
}

// The evaluation deadline fails only the rules it stops, each with its own
// "evaluation exceeded" error (status 503); a quick rule still reports.
func (s *RuntimeTestSuite) TestExecPolicyDeadlineIsPerRuleError() {
	exec := s.newTempPackExecutor(deadlinePack, WithEvalTimeout(100*time.Millisecond))
	defer func() { s.NoError(exec.Close()) }()
	facts := map[string]any{"xs": numberList(5000)}

	var outputs []*ExecutorOutput
	err := s.runBounded(10*time.Second, func() error {
		var err error
		outputs, err = exec.ExecPolicy(s.T().Context(), "deadline/test", "p", facts)
		return err
	})
	var ruleErrs RuleErrors
	s.Require().True(stdErr.As(err, &ruleErrs), "error is RuleErrors: %v", err)
	s.Require().Len(outputs, 3)
	for _, out := range outputs {
		if out.RuleName == "trivial" {
			s.Nil(out.Error)
			s.Equal(trinary.True, out.Decision.State)
			continue
		}
		s.Require().NotNil(out.Error, out.RuleName)
		s.Equal(http.StatusServiceUnavailable, out.Error.Status, out.RuleName)
		s.Contains(out.Error.Detail, "evaluation exceeded its 100ms deadline", out.RuleName)
	}
}

// ExecRule returns the failing rule's output with its error, next to the error.
func (s *RuntimeTestSuite) TestExecRuleFailureCarriesTheRuleError() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "rule_error.sentrie", src: ruleErrorPolicy})

	out, err := exec.ExecRule(ctx, "com/ex", "pol", "badDiv", map[string]any{"user": 1})
	s.Require().Error(err)
	s.Require().NotNil(out)
	s.Require().NotNil(out.Error)
	s.Equal(err.Error(), out.Error.Detail)

	out, err = exec.ExecRule(ctx, "com/ex", "pol", "a", nil)
	s.Require().Error(err)
	s.Require().NotNil(out, "a rule failing before evaluation still has an output")
	s.Require().NotNil(out.Error)
	s.Equal(http.StatusBadRequest, out.Error.Status)
	s.Equal("a", out.RuleName)
}

// Every rule failing still returns one output per rule, each with its error;
// no rule failing returns a nil error and no Error on any output.
func (s *RuntimeTestSuite) TestExecPolicyAllAndNoneFailing() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "all_none.sentrie", src: `namespace com/ex

policy allbad {
  rule x = { yield 1 / 0 }
  rule y = { yield "x" - 1 }
  export decision of x
  export decision of y
}

policy clean {
  rule x = { yield true }
  rule y = { yield 2 }
  export decision of x
  export decision of y
}
`})

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "allbad", nil)
	var ruleErrs RuleErrors
	s.Require().True(stdErr.As(err, &ruleErrs), "error is RuleErrors: %v", err)
	s.Len(ruleErrs, 2)
	s.Require().Len(outputs, 2)
	for _, o := range outputs {
		s.Require().NotNil(o.Error, o.RuleName)
		s.Equal(trinary.Unknown, o.Decision.State, o.RuleName)
	}

	outputs, err = exec.ExecPolicy(ctx, "com/ex", "clean", nil)
	s.Require().NoError(err)
	s.Require().Len(outputs, 2)
	for _, o := range outputs {
		s.Nil(o.Error, o.RuleName)
	}
}
