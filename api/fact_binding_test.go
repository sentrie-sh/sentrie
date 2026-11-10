// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	runtimepkg "github.com/sentrie-sh/sentrie/runtime"
)

const factBindingAPIPolicy = `namespace fb/test

policy p {
  fact user: number as u
  fact level?: number as lvl default 5
  rule who = { yield u == 41 }
  rule lvlIs5 = { yield lvl == 5 }
  export decision of who
  export decision of lvlIs5
}
`

func (s *APITestSuite) factPackAPI(src string) *HTTPAPI {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	files := map[string]string{
		loader.PackFileName: "[schema]\nversion = 1\n\n[pack]\nname = \"fb-pack\"\nversion = \"0.0.1\"\n",
		"policy.sentrie":    src,
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
	exec, err := runtimepkg.NewExecutor(idx)
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = exec.Close() })
	return NewHTTPAPI(exec)
}

func (s *APITestSuite) requireProblem(code int, body []byte, contentType string, detail ...string) {
	s.T().Helper()
	var problem ProblemDetails
	s.Require().NoError(json.Unmarshal(body, &problem), string(body))
	s.Equal(code, problem.Status, string(body))
	s.Equal("application/problem+json", contentType)
	for _, d := range detail {
		s.Contains(problem.Detail, d)
	}
}

// Callers send the declared name on serve; the alias alone is a missing fact.
func (s *APITestSuite) TestDecisionFactsAreSuppliedByDeclaredName() {
	api := s.factPackAPI(factBindingAPIPolicy)

	rec := s.postDecision(api, "/decision/fb/test/p", `{"facts": {"user": 41, "level": 9}}`)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	rec = s.postDecision(api, "/decision/fb/test/p/who", `{"facts": {"user": 41}}`)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())

	for _, path := range []string{"/decision/fb/test/p", "/decision/fb/test/p/who"} {
		rec = s.postDecision(api, path, `{"facts": {"u": 41}}`)
		s.Require().Equal(http.StatusBadRequest, rec.Code, path+" "+rec.Body.String())
		s.requireProblem(http.StatusBadRequest, rec.Body.Bytes(), rec.Header().Get("Content-Type"), "required fact not found: user")
	}
}

// A wrongly typed supplied fact is a request-level 400 problem document on a
// policy (never a 207 with per-rule errors) and on a single rule.
func (s *APITestSuite) TestDecisionWronglyTypedFactIs400() {
	api := s.factPackAPI(factBindingAPIPolicy)

	for _, path := range []string{"/decision/fb/test/p", "/decision/fb/test/p/who"} {
		rec := s.postDecision(api, path, `{"facts": {"user": "x"}}`)
		s.Require().Equal(http.StatusBadRequest, rec.Code, path+" "+rec.Body.String())
		s.requireProblem(http.StatusBadRequest, rec.Body.Bytes(), rec.Header().Get("Content-Type"), "fact 'user'", "is not a number")
		s.NotContains(rec.Body.String(), `"decisions"`)
	}

	// the optional fact is checked too
	rec := s.postDecision(api, "/decision/fb/test/p", `{"facts": {"user": 41, "level": "high"}}`)
	s.Require().Equal(http.StatusBadRequest, rec.Code, rec.Body.String())
	s.requireProblem(http.StatusBadRequest, rec.Body.Bytes(), rec.Header().Get("Content-Type"), "fact 'level'")
}

// PROVISIONAL (#147): a non-literal wrong default is the policy's fault, a
// 500 problem document, until a caller supplies the fact.
func (s *APITestSuite) TestDecisionWronglyTypedNonLiteralDefaultIs500() {
	api := s.factPackAPI(`namespace fb/test

policy p {
  fact n?: number default ["x"][0]
  rule r = { yield n == 1 }
  export decision of r
}
`)
	rec := s.postDecision(api, "/decision/fb/test/p", `{"facts": {}}`)
	s.Require().Equal(http.StatusInternalServerError, rec.Code, rec.Body.String())
	s.requireProblem(http.StatusInternalServerError, rec.Body.Bytes(), rec.Header().Get("Content-Type"), "default value of fact 'n'")

	rec = s.postDecision(api, "/decision/fb/test/p", `{"facts": {"n": 1}}`)
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
}

// PROVISIONAL (#147): an imported policy's wrongly typed `with` fact is the
// importing rule's own failure, so the policy is a 207 with that rule's 500.
func (s *APITestSuite) TestDecisionImportedWronglyTypedFactIsPerRule500() {
	api := s.factPackAPI(`namespace fb/test

policy target {
  fact n?: number
  rule ok = { yield true }
  export decision of ok
}

policy p {
  rule fine = { yield true }
  rule viaImport = import decision ok from fb/test/target with n as "x"
  export decision of fine
  export decision of viaImport
}
`)
	rec := s.postDecision(api, "/decision/fb/test/p", `{"facts": {}}`)
	s.Require().Equal(http.StatusMultiStatus, rec.Code, rec.Body.String())
	decisions := s.decisionsWithErrors(rec.Body.Bytes())
	s.Require().Len(decisions, 2)
	s.Nil(decisions["fine"])
	s.Require().NotNil(decisions["viaImport"])
	s.Equal(http.StatusInternalServerError, decisions["viaImport"].Status)
	s.Contains(decisions["viaImport"].Detail, "fact 'n'")
}
