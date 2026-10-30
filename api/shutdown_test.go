// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"io"
	"net"
	"net/http"
	"time"
)

// serveSlow serves a handler that signals started, then waits for release, on
// a listener whose pair is registered with a fresh HTTPAPI. The server's base
// context comes from baseCtx, as Setup passes the serve context.
func (s *APITestSuite) serveSlow(baseCtx context.Context, started chan<- struct{}, release <-chan struct{}) (*HTTPAPI, string) {
	api := NewHTTPAPI(nil)
	srv := api.newServer(baseCtx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		select {
		case <-release:
			_, _ = w.Write([]byte("done"))
		case <-r.Context().Done():
		}
	}))
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	api.listeners = []*ListenerServerPair{NewListenerServerPair(ln, srv)}
	go func() { _ = srv.Serve(ln) }()
	s.T().Cleanup(func() { _ = srv.Close() })
	return api, "http://" + ln.Addr().String()
}

// TestStopServerLetsAnInFlightRequestFinish: cancelling the serve context (the
// shutdown signal) does not cancel a request in flight, and StopServer waits
// for it and returns nil (#117).
func (s *APITestSuite) TestStopServerLetsAnInFlightRequestFinish() {
	baseCtx, cancel := context.WithCancel(s.T().Context())
	started, release := make(chan struct{}, 1), make(chan struct{})
	api, url := s.serveSlow(baseCtx, started, release)

	type result struct {
		body string
		err  error
	}
	got := make(chan result, 1)
	go func() {
		resp, err := http.Get(url)
		if err != nil {
			got <- result{err: err}
			return
		}
		defer func() { _ = resp.Body.Close() }()
		b, err := io.ReadAll(resp.Body)
		got <- result{body: string(b), err: err}
	}()
	<-started
	cancel()

	stopped := make(chan error, 1)
	shutdownCtx, stopCancel := context.WithTimeout(context.WithoutCancel(baseCtx), 5*time.Second)
	defer stopCancel()
	go func() { stopped <- api.StopServer(shutdownCtx) }()

	select {
	case err := <-stopped:
		s.Failf("StopServer returned before the request finished", "err: %v", err)
	case <-time.After(200 * time.Millisecond):
	}
	close(release)

	s.NoError(<-stopped)
	r := <-got
	s.Require().NoError(r.err)
	s.Equal("done", r.body)
}

// TestStopServerClosesAfterTheGracePeriod: a request that outlives the grace
// period is dropped and StopServer reports it instead of hanging.
func (s *APITestSuite) TestStopServerClosesAfterTheGracePeriod() {
	started, release := make(chan struct{}, 1), make(chan struct{})
	defer close(release)
	api, url := s.serveSlow(s.T().Context(), started, release)

	go func() {
		resp, err := http.Get(url)
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	<-started

	ctx, cancel := context.WithTimeout(s.T().Context(), 100*time.Millisecond)
	defer cancel()
	err := api.StopServer(ctx)
	s.ErrorIs(err, context.DeadlineExceeded)
}
