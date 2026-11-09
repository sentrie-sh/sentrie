// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"slices"
	"time"
)

// packLocationOf returns the --pack-location a started serve was given.
func (s *CmdTestSuite) packLocationOf(sp *startedServe) string {
	s.T().Helper()
	i := slices.Index(sp.cmd.Args, "--pack-location")
	s.Require().GreaterOrEqual(i, 0)
	return sp.cmd.Args[i+1]
}

func (s *CmdTestSuite) getFromServe(sp *startedServe, path string) (*http.Response, []byte) {
	s.T().Helper()
	resp, err := http.Get("http://" + sp.addr + path)
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return resp, body
}

// The real `serve` wiring answers both probes: liveness with a nonce, and the
// deep check healthy for a pack that validated at startup.
func (s *CmdTestSuite) TestServeAnswersHealthAndDeepHealth() {
	sp := s.startServe()

	resp, body := s.getFromServe(sp, "/health")
	s.Equal(http.StatusOK, resp.StatusCode)
	var live struct {
		CheckID string `json:"check_id"`
		Status  string `json:"status"`
		Time    string `json:"time"`
	}
	s.Require().NoError(json.Unmarshal(body, &live), string(body))
	s.NotEmpty(live.CheckID)
	s.Equal("healthy", live.Status)
	_, err := time.Parse(time.RFC3339, live.Time)
	s.Require().NoError(err)
	_, second := s.getFromServe(sp, "/health")
	s.NotEqual(string(body), string(second), "each response carries a fresh check_id")

	resp, body = s.getFromServe(sp, "/health/deep")
	s.Equal(http.StatusOK, resp.StatusCode, string(body))
	var deep struct {
		Status     string                    `json:"status"`
		Components map[string]map[string]any `json:"components"`
		Errors     []any                     `json:"errors"`
	}
	s.Require().NoError(json.Unmarshal(body, &deep), string(body))
	s.Equal("healthy", deep.Status)
	s.NotNil(deep.Errors)
	s.Empty(deep.Errors)
	for _, name := range []string{"index", "executor", "js_runtime", "pack"} {
		s.Equal("healthy", deep.Components[name]["status"], name)
	}
	s.Equal(true, deep.Components["index"]["validated"])
	s.Equal(true, deep.Components["index"]["committed"])
	s.Equal("signal-pack", deep.Components["pack"]["name"])
	s.NotContains(string(body), s.packLocationOf(sp), "the pack path stays off the wire")
}
