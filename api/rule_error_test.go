// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"

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
