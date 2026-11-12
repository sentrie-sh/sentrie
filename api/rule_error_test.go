// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	runtimepkg "github.com/sentrie-sh/sentrie/runtime"
)

// decisionsWithErrors decodes a decision response into each rule's error,
// nil for a rule that succeeded, keyed by rule name.
func (s *APITestSuite) decisionsWithErrors(body []byte) map[string]*runtimepkg.RuleError {
	var resp struct {
		Decisions []struct {
			Rule  string                `json:"rule"`
			Error *runtimepkg.RuleError `json:"error"`
		} `json:"decisions"`
	}
	s.Require().NoError(json.Unmarshal(body, &resp), string(body))
	out := map[string]*runtimepkg.RuleError{}
	for _, d := range resp.Decisions {
		out[d.Rule] = d.Error
	}
	return out
}

// One failing rule makes the policy a 207 that still carries every rule's
// result (#143); the failing rule has the error, the others do not.
func (s *APITestSuite) TestDecisionPolicyWithAFailingRuleIsMultiStatus() {
	api := s.lambdaPackAPI()
	rec := s.postDecision(api, "/decision/lam/test/p", `{"facts": {}}`)
	s.Require().Equal(http.StatusMultiStatus, rec.Code, rec.Body.String())
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	decisions := s.decisionsWithErrors(rec.Body.Bytes())
	s.Require().Len(decisions, 3)
	s.Nil(decisions["nested"])
	s.Nil(decisions["filtered"])
	s.Require().NotNil(decisions["whole"])
	s.Equal("Evaluation Failed", decisions["whole"].Title)
	s.Equal(http.StatusInternalServerError, decisions["whole"].Status)
	s.Contains(decisions["whole"].Detail, "cannot be a callable")

	// a succeeding rule has no error field at all
	var raw struct {
		Decisions []map[string]json.RawMessage `json:"decisions"`
	}
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &raw))
	for _, d := range raw.Decisions {
		if string(d["rule"]) != `"whole"` {
			s.NotContains(d, "error")
		}
	}
}

// rulePackAPI serves one policy file, optionally with a TypeScript module.
func (s *APITestSuite) rulePackAPI(policy string, extra map[string]string, opts ...HTTPAPIOption) *HTTPAPI {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	files := map[string]string{
		loader.PackFileName: "[schema]\nversion = 1\n\n[pack]\nname = \"rule-error-pack\"\nversion = \"0.0.1\"\n",
		"policy.sentrie":    policy,
	}
	for name, content := range extra {
		files[name] = content
	}
	for name, content := range files {
		s.Require().NoError(os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644))
	}
	packFile, err := loader.LoadPack(ctx, dir)
	s.Require().NoError(err)
	programs, err := loader.LoadPrograms(ctx, packFile)
	s.Require().NoError(err)
	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, packFile))
	for _, program := range programs {
		s.Require().NoError(idx.AddProgram(ctx, program))
	}
	s.Require().NoError(idx.Validate(ctx))
	exec, err := runtimepkg.NewExecutor(idx, runtimepkg.WithEvalTimeout(0))
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = exec.Close() })
	return NewHTTPAPI(exec, opts...)
}

const ruleErrorAPIPolicy = `namespace re/test

policy p {
  rule ok = { yield true }
  rule bad1 = { yield 1 / 0 }
  rule bad2 = { yield "x" - 1 }
  export decision of ok
  export decision of bad1
  export decision of bad2
}

policy allbad {
  rule bad1 = { yield 1 / 0 }
  rule bad2 = { yield "x" - 1 }
  export decision of bad1
  export decision of bad2
}

policy clean {
  rule a = { yield true }
  rule b = { yield 2 }
  export decision of a
  export decision of b
}

policy typed {
  fact n: number
  rule ok = { yield n > 0 }
  export decision of ok
}

policy attachfail {
  rule ok = { yield true }
  rule fine = { yield true }
  export decision of ok attach good as 1 attach boom as 1 / 0
  export decision of fine attach note as "kept"
}

policy importer {
  rule fine = { yield true }
  rule viaImport = import decision ok from re/test/typed with n as "x"
  export decision of fine
  export decision of viaImport
}
`

// Every rule failing is still a 207 that lists each rule with its error.
func (s *APITestSuite) TestDecisionPolicyWithAllRulesFailingIsMultiStatus() {
	api := s.rulePackAPI(ruleErrorAPIPolicy, nil)
	rec := s.postDecision(api, "/decision/re/test/allbad", `{"facts": {}}`)
	s.Require().Equal(http.StatusMultiStatus, rec.Code, rec.Body.String())
	decisions := s.decisionsWithErrors(rec.Body.Bytes())
	s.Require().Len(decisions, 2)
	for rule, e := range decisions {
		s.Require().NotNil(e, rule)
		s.Equal("Evaluation Failed", e.Title, rule)
	}
	s.Contains(decisions["bad1"].Detail, "divide by zero")
	s.Contains(decisions["bad2"].Detail, "left operand is not a number")
}

// No failing rule is a plain 200 whose JSON has no error field anywhere.
func (s *APITestSuite) TestDecisionPolicyWithNoFailingRuleIsOKWithoutErrorField() {
	api := s.rulePackAPI(ruleErrorAPIPolicy, nil)
	rec := s.postDecision(api, "/decision/re/test/clean", `{"facts": {}}`)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.Len(s.decisionsWithErrors(rec.Body.Bytes()), 2)
	s.NotContains(rec.Body.String(), `"error"`)
}

// Healthy and failing rules are told apart in one 207, whatever their order.
func (s *APITestSuite) TestDecisionPolicyMixedRulesAreComparedAsASet() {
	api := s.rulePackAPI(ruleErrorAPIPolicy, nil)
	rec := s.postDecision(api, "/decision/re/test/p", `{"facts": {}}`)
	s.Require().Equal(http.StatusMultiStatus, rec.Code, rec.Body.String())
	decisions := s.decisionsWithErrors(rec.Body.Bytes())
	s.Len(decisions, 3)
	s.Nil(decisions["ok"])
	s.NotNil(decisions["bad1"])
	s.NotNil(decisions["bad2"])
}

// A named rule that fails is a problem document, not a 207.
func (s *APITestSuite) TestDecisionNamedRuleFailureIsAProblemDocument() {
	api := s.rulePackAPI(ruleErrorAPIPolicy, nil)
	rec := s.postDecision(api, "/decision/re/test/p/bad1", `{"facts": {}}`)
	s.Require().Equal(http.StatusInternalServerError, rec.Code, rec.Body.String())
	var problem struct {
		Title  string `json:"title"`
		Status int    `json:"status"`
		Detail string `json:"detail"`
	}
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &problem))
	s.Equal("Evaluation Failed", problem.Title)
	s.Contains(problem.Detail, "divide by zero")
	s.NotContains(rec.Body.String(), `"decisions"`)
}

// A missing required fact is a 400 problem document for a policy request,
// not a per-rule error, and so is an imported policy's wrongly typed with
// fact: one 400 for the whole request, with no decisions.
func (s *APITestSuite) TestDecisionFactErrorsAreRequestLevel() {
	api := s.rulePackAPI(ruleErrorAPIPolicy, nil)

	rec := s.postDecision(api, "/decision/re/test/typed", `{"facts": {}}`)
	s.Require().Equal(http.StatusBadRequest, rec.Code, rec.Body.String())
	s.Contains(rec.Body.String(), "required fact not found: n")
	s.NotContains(rec.Body.String(), `"decisions"`)

	rec = s.postDecision(api, "/decision/re/test/importer", `{"facts": {}}`)
	s.Require().Equal(http.StatusBadRequest, rec.Code, rec.Body.String())
	s.Contains(rec.Body.String(), "imported policy re/test/typed")
	s.Contains(rec.Body.String(), "is not a number")
	s.NotContains(rec.Body.String(), `"decisions"`)
}

// A deadline that stops one rule is that rule's 503 inside a 207.
func (s *APITestSuite) TestDecisionDeadlineIsAPerRuleErrorInMultiStatus() {
	api := s.rulePackAPI(`namespace dl/test

policy p {
  use { spin } from "./spin" as m
  rule quick = { yield true }
  rule slow = { yield m.spin() }
  export decision of quick
  export decision of slow
}
`, map[string]string{"spin.ts": "export function spin(): number { while (true) {} }\n"},
		WithRequestTimeout(150*time.Millisecond))

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- s.postDecision(api, "/decision/dl/test/p", `{"facts": {}}`) }()
	select {
	case rec := <-done:
		s.Require().Equal(http.StatusMultiStatus, rec.Code, rec.Body.String())
		decisions := s.decisionsWithErrors(rec.Body.Bytes())
		s.Require().Len(decisions, 2)
		s.Nil(decisions["quick"])
		s.Require().NotNil(decisions["slow"])
		s.Equal(http.StatusServiceUnavailable, decisions["slow"].Status)
	case <-time.After(10 * time.Second):
		s.FailNow("evaluation did not stop at its deadline")
	}
}

// A rule whose attachment fails reports unknown with its error and no
// attachments in the 207, never the true decision computed before the
// attachment failed (#143 review); its sibling is untouched. Asked for by
// name it is a problem document.
func (s *APITestSuite) TestDecisionAttachmentFailureIsUnknown() {
	api := s.rulePackAPI(ruleErrorAPIPolicy, nil)
	rec := s.postDecision(api, "/decision/re/test/attachfail", `{"facts": {}}`)
	s.Require().Equal(http.StatusMultiStatus, rec.Code, rec.Body.String())

	var resp struct {
		Decisions []struct {
			Rule        string                `json:"rule"`
			Decision    map[string]any        `json:"decision"`
			Attachments map[string]any        `json:"attachments"`
			Error       *runtimepkg.RuleError `json:"error"`
		} `json:"decisions"`
	}
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &resp), rec.Body.String())
	s.Require().Len(resp.Decisions, 2)
	for _, d := range resp.Decisions {
		switch d.Rule {
		case "ok":
			s.Require().NotNil(d.Error)
			s.Contains(d.Error.Detail, "divide by zero")
			s.Equal("unknown", d.Decision["state"], rec.Body.String())
			s.NotEqual(true, d.Decision["value"], rec.Body.String())
			s.Empty(d.Attachments, rec.Body.String())
		case "fine":
			s.Nil(d.Error)
			s.Equal("true", d.Decision["state"], rec.Body.String())
			s.Equal("kept", d.Attachments["note"])
		default:
			s.Failf("unexpected rule", "%s", d.Rule)
		}
	}

	rec = s.postDecision(api, "/decision/re/test/attachfail/ok", `{"facts": {}}`)
	s.Require().Equal(http.StatusInternalServerError, rec.Code, rec.Body.String())
	s.Contains(rec.Body.String(), "divide by zero")
	s.NotContains(rec.Body.String(), `"decisions"`)
	s.NotContains(rec.Body.String(), `"good"`)
}
