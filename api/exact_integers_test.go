// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	runtimepkg "github.com/sentrie-sh/sentrie/runtime"
)

// TestDecisionKeepsLargeIntegerFactsExact covers #166 at the network boundary:
// an integer fact beyond 2^53 is decoded exactly, compared exactly and echoed
// back without rounding or exponent notation.
func (s *APITestSuite) TestDecisionKeepsLargeIntegerFactsExact() {
	api := s.numbersPackAPI()

	body := `{"facts": {"id": 9007199254740993, "other": 9007199254740992}}`

	rec := s.postDecision(api, "/decision/ints/test/p/echo", body)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.Contains(rec.Body.String(), `"value":9007199254740993`)

	rec = s.postDecision(api, "/decision/ints/test/p/same", body)
	s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
	s.Contains(rec.Body.String(), `"value":false`)
}

// TestDecisionRejectsOutOfRangeNumberFacts: a JSON number float64 cannot hold
// is bad JSON (400), as it was before UseNumber, not an infinity fact.
func (s *APITestSuite) TestDecisionRejectsOutOfRangeNumberFacts() {
	api := s.numbersPackAPI()
	for _, body := range []string{
		`{"facts": {"id": 1e400, "other": 1}}`,
		`{"facts": {"id": 1, "other": 1, "extra": {"deep": [-1e500]}}}`,
	} {
		rec := s.postDecision(api, "/decision/ints/test/p/echo", body)
		s.Equal(http.StatusBadRequest, rec.Code, body)
		s.Contains(rec.Body.String(), "Invalid JSON", body)
	}
}

// TestDecisionEncodingFailureIsAnError: a non-finite result from arithmetic
// cannot be encoded; the client gets a 500 problem response, not a 200 with an
// empty body.
func (s *APITestSuite) TestDecisionEncodingFailureIsAnError() {
	api := s.numbersPackAPI()
	rec := s.postDecision(api, "/decision/ints/test/p/overflow", `{"facts": {"id": 1e308, "other": 1}}`)
	s.Equal(http.StatusInternalServerError, rec.Code)
	s.NotEmpty(rec.Body.String())
}

// numbersPackAPI serves a small pack that echoes and compares number facts.
func (s *APITestSuite) numbersPackAPI() *HTTPAPI {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	files := map[string]string{
		loader.PackFileName: "[schema]\nversion = 1\n\n[pack]\nname = \"ints-pack\"\nversion = \"0.0.1\"\n",
		"policy.sentrie": `namespace ints/test

policy p {
  fact id: number
  fact other: number
  rule echo = { yield id }
  rule same = { yield id == other }
  rule overflow = { yield id * 10 }
  export decision of echo
  export decision of same
  export decision of overflow
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
