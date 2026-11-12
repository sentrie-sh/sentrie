// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

const bodyLimitDecisionPath = "/decision/sh/sentrie/signal/access/allow_admin"

// postServe posts body to the decision endpoint of a started serve with a
// known Content-Length.
func (s *CmdTestSuite) postServe(sp *startedServe, body string) (*http.Response, []byte) {
	s.T().Helper()
	resp, err := http.Post("http://"+sp.addr+bodyLimitDecisionPath, "application/json", bytes.NewReader([]byte(body)))
	s.Require().NoError(err)
	defer func() { _ = resp.Body.Close() }()
	out, err := io.ReadAll(resp.Body)
	s.Require().NoError(err)
	return resp, out
}

func paddedServeBody(n int) string {
	return `{"facts":{"user":{"role":"admin"},"pad":"` + strings.Repeat("x", n) + `"}}`
}

// The real serve wiring: the default limit is 1 MiB, so a body just under it
// is decided and one over it is a 413 problem document (#115).
func (s *CmdTestSuite) TestServeDefaultBodyLimitIsOneMiB() {
	sp := s.startServe()

	resp, body := s.postServe(sp, paddedServeBody(1<<20-1000))
	s.Equal(http.StatusOK, resp.StatusCode, string(body))

	resp, body = s.postServe(sp, paddedServeBody(1<<20))
	s.Equal(http.StatusRequestEntityTooLarge, resp.StatusCode, string(body))
	s.Equal("application/problem+json", resp.Header.Get("Content-Type"))
	var p map[string]any
	s.Require().NoError(json.Unmarshal(body, &p), string(body))
	s.Equal("Request Body Too Large", p["title"])
}

// --http-max-body-bytes sets the ceiling, and 0 lifts it.
func (s *CmdTestSuite) TestServeMaxBodyBytesFlagSetsAndLiftsTheLimit() {
	small := s.startServe("--http-max-body-bytes", "100")
	resp, body := s.postServe(small, paddedServeBody(200))
	s.Equal(http.StatusRequestEntityTooLarge, resp.StatusCode, string(body))
	resp, body = s.postServe(small, `{"facts":{"user":{"role":"admin"}}}`)
	s.Equal(http.StatusOK, resp.StatusCode, string(body))

	unlimited := s.startServe("--http-max-body-bytes", "0")
	resp, body = s.postServe(unlimited, paddedServeBody(2<<20))
	s.Equal(http.StatusOK, resp.StatusCode, string(body))
}
