// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"net"
	"net/http"
	"os"
	"time"
)

// TestServerSetsConnectionLimits: every listener's server bounds header reads,
// idle keep-alive connections and header size, not just the whole read (#247).
func (s *APITestSuite) TestServerSetsConnectionLimits() {
	srv := NewHTTPAPI(nil).newServer(s.T().Context(), http.NewServeMux())
	s.Equal(serverReadTimeout, srv.ReadTimeout)
	s.Equal(serverReadHeaderTimeout, srv.ReadHeaderTimeout)
	s.Equal(serverIdleTimeout, srv.IdleTimeout)
	s.Equal(serverMaxHeaderBytes, srv.MaxHeaderBytes)
	s.Positive(srv.WriteTimeout)
	s.Less(srv.ReadHeaderTimeout, srv.ReadTimeout, "the header deadline must be tighter than the whole read")
}

// TestServerClosesASlowHeaderConnection: a client that starts a request and
// never finishes its headers is disconnected at the header deadline instead of
// holding the connection for the full read timeout.
func (s *APITestSuite) TestServerClosesASlowHeaderConnection() {
	api := NewHTTPAPI(nil)
	api.readHeaderTimeout = 100 * time.Millisecond
	srv := api.newServer(s.T().Context(), http.NewServeMux())

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	go func() { _ = srv.Serve(ln) }()
	s.T().Cleanup(func() { _ = srv.Close() })

	conn, err := net.Dial("tcp", ln.Addr().String())
	s.Require().NoError(err)
	defer func() { _ = conn.Close() }()
	_, err = conn.Write([]byte("POST /decision/x HTTP/1.1\r\nHost: localhost\r\n"))
	s.Require().NoError(err)

	start := time.Now()
	s.Require().NoError(conn.SetReadDeadline(start.Add(5 * time.Second)))
	buf := make([]byte, 512)
	for {
		if _, err = conn.Read(buf); err != nil {
			break
		}
	}
	elapsed := time.Since(start)

	// The server closes the connection (io.EOF or a reset); the client's own
	// 5s read deadline must not be what ended the read.
	s.Require().Error(err)
	s.NotErrorIs(err, os.ErrDeadlineExceeded, "the server should close the connection first")
	s.GreaterOrEqual(elapsed, 50*time.Millisecond)
	s.Less(elapsed, 2*time.Second)
}
