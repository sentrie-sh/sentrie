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
// returns, with the executor and the API configured as serve does.
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

	exec, err := runtimepkg.NewExecutor(idx, runtimepkg.WithEvalTimeout(requestTimeout))
	s.Require().NoError(err)
	s.T().Cleanup(func() { _ = exec.Close() })
	return NewHTTPAPI(exec, WithRequestTimeout(requestTimeout))
}

// TestDecisionRequestStopsAtItsDeadline: evaluation stops when the deadline
// passes, the client gets 503, and the message names the deadline.
func (s *APITestSuite) TestDecisionRequestStopsAtItsDeadline() {
	api := s.loopingPackAPI(100 * time.Millisecond)

	done := make(chan *httptest.ResponseRecorder, 1)
	go func() { done <- s.postDecision(api, "/decision/loop/test/p/r", `{"facts": {}}`) }()
	select {
	case rec := <-done:
		s.Equal(http.StatusServiceUnavailable, rec.Code)
		s.Contains(rec.Body.String(), "deadline exceeded")
		s.Contains(rec.Body.String(), "evaluation exceeded its 100ms deadline")
	case <-time.After(5 * time.Second):
		s.FailNow("decision request did not stop at its deadline")
	}
}

func (s *APITestSuite) TestWriteTimeoutCoversTheRequestDeadline() {
	s.Equal(30*time.Second, NewHTTPAPI(nil).writeTimeout())
	s.Equal(65*time.Second, NewHTTPAPI(nil, WithRequestTimeout(time.Minute)).writeTimeout())
	s.Equal(30*time.Second, NewHTTPAPI(nil, WithRequestTimeout(0)).writeTimeout(), "--eval-timeout 0 keeps the 30s write cap")
}
