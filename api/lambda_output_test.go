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

func (s *APITestSuite) lambdaPackAPI() *HTTPAPI {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	files := map[string]string{
		loader.PackFileName: "[schema]\nversion = 1\n\n[pack]\nname = \"lam-pack\"\nversion = \"0.0.1\"\n",
		"policy.sentrie": `namespace lam/test

policy p {
  let id = (x) => { yield x }
  rule nested = { yield { "f": id, "xs": [1, id] } }
  rule filtered = { yield filter([1, 2, 3], (x) => { yield x > 1 }) }
  rule whole = { yield id }
  export decision of nested
  export decision of filtered
  export decision of whole
}
`,
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

// A lambda nested in the value, or only in the trace, is a 200 with the placeholder.
func (s *APITestSuite) TestDecisionEncodesLambdaAsPlaceholder() {
	api := s.lambdaPackAPI()
	for _, rule := range []string{"nested", "filtered"} {
		rec := s.postDecision(api, "/decision/lam/test/p/"+rule, `{"facts": {}}`)
		s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
		s.Contains(rec.Body.String(), `{"kind":"callable"}`, rule)
	}
}

// A rule whose whole value is a lambda is that rule's error: a 500 problem
// document whose detail names the rule.
func (s *APITestSuite) TestDecisionWholeValueLambdaIsRuleError() {
	api := s.lambdaPackAPI()
	rec := s.postDecision(api, "/decision/lam/test/p/whole", `{"facts": {}}`)
	s.Require().Equal(http.StatusInternalServerError, rec.Code, rec.Body.String())
	s.Equal("application/problem+json", rec.Header().Get("Content-Type"))
	var problem ProblemDetails
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &problem))
	s.Equal(http.StatusInternalServerError, problem.Status)
	s.Contains(problem.Detail, "lam/test/p/whole")
	s.Contains(problem.Detail, "cannot be a callable")
}
