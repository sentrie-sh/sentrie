// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sentrie-sh/sentrie/api/middleware"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	runtimepkg "github.com/sentrie-sh/sentrie/runtime"
)

func examplePackDir() string {
	_, current, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(current), "..", "example_pack")
}

func (s *APITestSuite) examplePackHTTPAPI() *HTTPAPI {
	if s.examplePackAPI != nil {
		return s.examplePackAPI
	}

	ctx := s.T().Context()

	packFile, err := loader.LoadPack(ctx, examplePackDir())
	s.Require().NoError(err)

	programs, err := loader.LoadPrograms(ctx, packFile)
	s.Require().NoError(err)
	s.Require().NotEmpty(programs)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, packFile))
	for _, program := range programs {
		s.Require().NoError(idx.AddProgram(ctx, program))
	}
	s.Require().NoError(idx.Validate(ctx))

	exec, err := runtimepkg.NewExecutor(idx)
	s.Require().NoError(err)

	s.examplePackAPI = NewHTTPAPI(exec)
	return s.examplePackAPI
}

func (s *APITestSuite) decisionHandler(api *HTTPAPI) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /decision/{target...}",
		middleware.RequestIDMiddleware(http.HandlerFunc(api.handleDecision)),
	)
	return mux
}

func (s *APITestSuite) postDecision(api *HTTPAPI, path string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.decisionHandler(api).ServeHTTP(rec, req)
	return rec
}

func (s *APITestSuite) TestHandleDecisionPolicySuccess() {
	api := s.examplePackHTTPAPI()

	rec := s.postDecision(api, "/decision/sh/sentrie/example/user_access", `{
		"facts": {
			"user": {"role": "admin", "status": "active"}
		}
	}`)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	var payload map[string]json.RawMessage
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &payload))
	s.Contains(payload, "decisions")
	s.NotContains(payload, "error")

	var decisions []json.RawMessage
	s.Require().NoError(json.Unmarshal(payload["decisions"], &decisions))
	s.NotEmpty(decisions)
}

func (s *APITestSuite) TestHandleDecisionRuleSuccess() {
	api := s.examplePackHTTPAPI()

	rec := s.postDecision(api, "/decision/sh/sentrie/example/user_access/allow_admin", `{
		"facts": {
			"user": {"role": "admin", "status": "active"}
		}
	}`)

	s.Equal(http.StatusOK, rec.Code)

	var payload map[string]json.RawMessage
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &payload))
	s.Contains(payload, "decisions")
	s.NotContains(payload, "error")

	var decisions []json.RawMessage
	s.Require().NoError(json.Unmarshal(payload["decisions"], &decisions))
	s.Len(decisions, 1)
}

func (s *APITestSuite) TestHandleDecisionMissingRequiredFact() {
	api := s.examplePackHTTPAPI()

	rec := s.postDecision(api, "/decision/sh/sentrie/example/user_access", `{"facts": {}}`)

	// Every rule fails on the missing fact, each with its own 400 error (#143).
	s.Equal(http.StatusMultiStatus, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	decisions := s.decisionsWithErrors(rec.Body.Bytes())
	s.Require().NotEmpty(decisions)
	for rule, ruleErr := range decisions {
		s.Require().NotNil(ruleErr, rule)
		s.Equal(http.StatusBadRequest, ruleErr.Status, rule)
		s.Equal("Evaluation Failed", ruleErr.Title, rule)
		s.Contains(ruleErr.Detail, "required fact not found: user", rule)
	}
}

func (s *APITestSuite) TestHandleDecisionEvaluationInternalError() {
	api := s.examplePackHTTPAPI()

	rec := s.postDecision(api, "/decision/sh/sentrie/example/shapes/example", `{"facts": {}}`)

	s.Equal(http.StatusMultiStatus, rec.Code)
	s.Equal("application/json", rec.Header().Get("Content-Type"))

	found := false
	for _, ruleErr := range s.decisionsWithErrors(rec.Body.Bytes()) {
		if ruleErr != nil && strings.Contains(ruleErr.Detail, "invalid value for let declaration user") {
			found = true
			s.Equal(http.StatusInternalServerError, ruleErr.Status)
			s.Equal("Evaluation Failed", ruleErr.Title)
		}
	}
	s.True(found, rec.Body.String())
}

func (s *APITestSuite) TestHandleDecisionInvalidPath() {
	api := s.examplePackHTTPAPI()

	rec := s.postDecision(api, "/decision/does/not/exist", `{"facts": {}}`)

	s.Equal(http.StatusNotFound, rec.Code)
	s.Equal("application/problem+json", rec.Header().Get("Content-Type"))
}

func (s *APITestSuite) TestHandleDecisionInvalidJSON() {
	api := s.examplePackHTTPAPI()

	req := httptest.NewRequest(http.MethodPost, "/decision/sh/sentrie/example/user_access", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.decisionHandler(api).ServeHTTP(rec, req)

	s.Equal(http.StatusBadRequest, rec.Code)
	s.Equal("application/problem+json", rec.Header().Get("Content-Type"))
}
