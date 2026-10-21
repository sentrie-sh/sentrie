// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	runtimepkg "github.com/sentrie-sh/sentrie/runtime"
)

// loopingPackAPI serves a pack whose only rule calls a module that never
// returns. The executor's own deadline is disabled, so only the request
// deadline can stop the evaluation.
func (s *APITestSuite) loopingPackAPI(requestTimeout time.Duration) *HTTPAPI {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	files := map[string]string{
		loader.PackFileName: "[schema]\nversion = 1\n\n[pack]\nname = \"loop-pack\"\nversion = \"0.0.1\"\n",
		"spin.ts":           "export function spin(): number { while (true) {} }\n",
		"policy.sentrie": `namespace loop/test

policy p {
  use { spin } from "./spin" as m
  rule r = { yield m.spin() }
  export decision of r
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

	exec, err := runtimepkg.NewExecutor(idx, runtimepkg.WithEvalTimeout(0))
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = exec.Close() })
	return NewHTTPAPI(exec, WithRequestTimeout(requestTimeout))
}

// TestDecisionRequestStopsAtItsDeadline is #142's request-scoped cancellation:
// evaluation stops when the request deadline passes and the client gets 503.
func (s *APITestSuite) TestDecisionRequestStopsAtItsDeadline() {
	api := s.loopingPackAPI(100 * time.Millisecond)

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- s.postDecision(api, "/decision/loop/test/p/r", `{"facts": {}}`) }()
	select {
	case rec := <-done:
		s.Equal(http.StatusServiceUnavailable, rec.Code)
		s.Contains(rec.Body.String(), "deadline exceeded")
	case <-time.After(5 * time.Second):
		s.FailNow("decision request did not stop at its deadline")
	}
}

func (s *APITestSuite) TestWriteTimeoutCoversTheRequestDeadline() {
	s.Equal(30*time.Second, NewHTTPAPI(nil).writeTimeout())
	s.Equal(65*time.Second, NewHTTPAPI(nil, WithRequestTimeout(time.Minute)).writeTimeout())
}
