// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	stdErr "errors"
	"net/http"

	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

// factBindingPolicy declares aliased facts so the supplied key (the declared
// name) and the local binding (the alias) differ (#147).
const factBindingPolicy = `namespace com/ex

policy pol {
  fact user: number as u
  fact level?: number as lvl default 5
  fact note?: string
  rule who = { yield u == 41 }
  rule lvlIs5 = { yield lvl == 5 }
  rule lvlIs9 = { yield lvl == 9 }
  rule noteGiven = { yield note is defined }
  export decision of who
  export decision of lvlIs5
  export decision of lvlIs9
  export decision of noteGiven
}
`

func (s *RuntimeTestSuite) factBindingExec(src string) *executorImpl {
	_, exec := s.mustBuildDeriveExecutor(s.T().Context(), deriveTestProgram{name: "fact_binding.sentrie", src: src})
	return exec
}

func (s *RuntimeTestSuite) factStates(exec *executorImpl, facts map[string]any) map[string]trinary.Value {
	outputs, err := exec.ExecPolicy(s.T().Context(), "com/ex", "pol", facts)
	s.Require().NoError(err)
	states := map[string]trinary.Value{}
	for _, o := range outputs {
		states[o.RuleName] = o.Decision.State
	}
	return states
}

// The declared name is the key a caller supplies; the policy reads the alias.
func (s *RuntimeTestSuite) TestFactSuppliedByDeclaredNameBindsTheAlias() {
	exec := s.factBindingExec(factBindingPolicy)

	states := s.factStates(exec, map[string]any{"user": 41, "level": 9})
	s.Equal(trinary.True, states["who"])
	s.Equal(trinary.True, states["lvlIs9"])
	s.Equal(trinary.False, states["lvlIs5"], "the supplied level replaces the default")
}

// Clean break: the alias alone no longer supplies a fact. A required fact
// given only by alias is missing (a request error, 400); an optional one is
// not bound, so its default applies.
func (s *RuntimeTestSuite) TestFactSuppliedByAliasAloneDoesNotBind() {
	exec := s.factBindingExec(factBindingPolicy)
	ctx := s.T().Context()

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "pol", map[string]any{"u": 41})
	s.Require().Error(err)
	s.Nil(outputs)
	var invErr xerr.InvalidInvocationError
	s.ErrorAs(err, &invErr)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "required fact not found: user")

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "who", map[string]any{"u": 41})
	s.Require().Error(err)
	s.Contains(err.Error(), "required fact not found: user")

	states := s.factStates(exec, map[string]any{"user": 41, "lvl": 9})
	s.Equal(trinary.True, states["lvlIs5"], "the alias key is ignored, the default stays")
	s.Equal(trinary.False, states["lvlIs9"])
}

// An omitted optional fact without a default reads as undefined; supplying it
// by name binds it.
func (s *RuntimeTestSuite) TestOmittedOptionalFactWithoutDefaultIsUndefined() {
	exec := s.factBindingExec(factBindingPolicy)

	s.Equal(trinary.False, s.factStates(exec, map[string]any{"user": 41})["noteGiven"])
	s.Equal(trinary.True, s.factStates(exec, map[string]any{"user": 41, "note": "hi"})["noteGiven"])
}

// A supplied fact of the wrong type is the caller's mistake: an invalid
// invocation (400) that names the declared name, for a policy and for a rule,
// not a failure of each rule.
func (s *RuntimeTestSuite) TestWronglyTypedSuppliedFactIsAnInvalidInvocation() {
	exec := s.factBindingExec(factBindingPolicy)
	ctx := s.T().Context()
	facts := map[string]any{"user": "x"}

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "pol", facts)
	s.Require().Error(err)
	s.Nil(outputs)
	var ruleErrs RuleErrors
	s.False(stdErr.As(err, &ruleErrs), "not per-rule errors: %v", err)
	var invErr xerr.InvalidInvocationError
	s.ErrorAs(err, &invErr)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "fact 'user': ")
	s.Contains(err.Error(), "is not a number")

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "who", facts)
	s.Require().Error(err)
	s.ErrorAs(err, &invErr)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "fact 'user': ")
}

// A default that is not a literal is only checked when it is bound, and it is
// the policy's mistake: a request-level 500 naming the default, not a 400.
// Supplying the fact avoids the bad default.
func (s *RuntimeTestSuite) TestWronglyTypedNonLiteralDefaultStaysARuntime500() {
	exec := s.factBindingExec(`namespace com/ex

policy pol {
  fact n?: number default ["x"][0]
  rule r = { yield n == 1 }
  export decision of r
}
`)
	ctx := s.T().Context()

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "pol", map[string]any{})
	s.Require().Error(err)
	s.Nil(outputs)
	var ruleErrs RuleErrors
	s.False(stdErr.As(err, &ruleErrs), "request-level, not per-rule: %v", err)
	var invErr xerr.InvalidInvocationError
	s.False(stdErr.As(err, &invErr), "not an invalid invocation: %v", err)
	s.Equal(http.StatusInternalServerError, ErrorStatus(err))
	s.Contains(err.Error(), "default value of fact 'n'")
	s.Contains(err.Error(), "is not a number")

	outputs, err = exec.ExecPolicy(ctx, "com/ex", "pol", map[string]any{"n": 1})
	s.Require().NoError(err)
	s.Require().Len(outputs, 1)
	s.Equal(trinary.True, outputs[0].Decision.State)
}

const factImportPolicy = `namespace com/ex

policy target {
  fact subject?: string as who
  fact n?: number
  rule ok = { yield who == "x" }
  export decision of ok
}

policy byName {
  rule viaName = import decision ok from com/ex/target with subject as "x"
  export decision of viaName
}

policy badType {
  rule fine = { yield true }
  rule viaImport = import decision ok from com/ex/target with n as "x"
  export decision of fine
  export decision of viaImport
}
`

// An import's `with` names the target's fact by its declared name, binds it
// under the alias, and rejects the alias; the importing rule fails, not the
// request. The alias rejection is the runtime check, reached when the index
// check did not run.
func (s *RuntimeTestSuite) TestImportWithBindsByDeclaredNameAtRuntime() {
	ctx := s.T().Context()
	// not validated: the index check rejects the alias before the runtime can
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "fact_import.sentrie", src: factImportPolicy + `
policy byAlias {
  rule viaAlias = import decision ok from com/ex/target with who as "x"
  export decision of viaAlias
}
`})
	rawExec, err := NewExecutor(idx)
	s.Require().NoError(err)
	exec := rawExec.(*executorImpl)

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "byName", nil)
	s.Require().NoError(err)
	s.Require().Len(outputs, 1)
	s.Equal(trinary.True, outputs[0].Decision.State)

	outputs, err = exec.ExecPolicy(ctx, "com/ex", "byAlias", nil)
	var ruleErrs RuleErrors
	s.Require().True(stdErr.As(err, &ruleErrs), "error is RuleErrors: %v", err)
	s.Require().Len(outputs, 1)
	s.Contains(ruleErrs[0].Detail, `unknown fact "who"`)
	s.Contains(ruleErrs[0].Detail, "accepts facts: n, subject")
}

// An imported policy's wrongly typed `with` fact fails the whole request
// with a 400, like a caller's wrongly typed fact (#147, user decision 11):
// for the whole policy and for the importing rule alone.
func (s *RuntimeTestSuite) TestImportedWronglyTypedFactIsARequestError() {
	ctx := s.T().Context()
	exec := s.factBindingExec(factImportPolicy)

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "badType", nil)
	s.Require().Error(err)
	s.Nil(outputs)
	var ruleErrs RuleErrors
	s.False(stdErr.As(err, &ruleErrs), "not per-rule errors: %v", err)
	var invErr xerr.InvalidInvocationError
	s.ErrorAs(err, &invErr)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "imported policy com/ex/target: fact 'n': ")
	s.Contains(err.Error(), "is not a number")

	_, err = exec.ExecRule(ctx, "com/ex", "badType", "viaImport", nil)
	s.Require().Error(err)
	s.ErrorAs(err, &invErr)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "imported policy com/ex/target: fact 'n': ")
}

// importedFactErrorsPolicy has importing rules that leave out a required
// `with` fact or hit different `with` fact errors.
const importedFactErrorsPolicy = `namespace com/ex

policy target {
  fact req: number
  rule ok = { yield req == 1 }
  export decision of ok
}

policy other {
  fact o: number
  rule ok = { yield o == 1 }
  export decision of ok
}

policy missingWith {
  rule fine = { yield true }
  rule viaImport = import decision ok from com/ex/target
  export decision of fine
  export decision of viaImport
}

policy two {
  rule second = import decision ok from com/ex/target with req as "y"
  rule first = import decision ok from com/ex/other with o as "x"
  export decision of first
  export decision of second
}
`

// A missing required `with` fact is a request error and a 400, like a
// caller's missing required fact (#147, PROVISIONAL extension of user
// decisions 2 and 11), for the whole policy and for the importing rule.
func (s *RuntimeTestSuite) TestImportedMissingRequiredFactIsARequestError() {
	ctx := s.T().Context()
	exec := s.factBindingExec(importedFactErrorsPolicy)

	outputs, err := exec.ExecPolicy(ctx, "com/ex", "missingWith", nil)
	s.Require().Error(err)
	s.Nil(outputs)
	var ruleErrs RuleErrors
	s.False(stdErr.As(err, &ruleErrs), "not per-rule errors: %v", err)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "imported policy com/ex/target: required fact not found: req")

	_, err = exec.ExecRule(ctx, "com/ex", "missingWith", "viaImport", nil)
	s.Require().Error(err)
	s.Equal(http.StatusBadRequest, ErrorStatus(err))
	s.Contains(err.Error(), "imported policy com/ex/target: required fact not found: req")
}

// When importing rules hit different `with` fact errors, ExecPolicy returns
// the earliest declared rule's every time, not the first rule to finish.
func (s *RuntimeTestSuite) TestImportedFactErrorChoiceIsDeterministic() {
	ctx := s.T().Context()
	exec := s.factBindingExec(importedFactErrorsPolicy)
	for range 50 {
		_, err := exec.ExecPolicy(ctx, "com/ex", "two", nil)
		s.Require().Error(err)
		s.Contains(err.Error(), "imported policy com/ex/target: fact 'req'")
	}
}
