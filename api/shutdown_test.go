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

// servePair serves handler on a fresh loopback listener and registers the
// pair with api, so api.StopServer shuts it down. It returns the base URL.
func (s *APITestSuite) servePair(api *HTTPAPI, baseCtx context.Context, handler http.Handler) string {
	srv := api.newServer(baseCtx, handler)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	api.listeners = append(api.listeners, NewListenerServerPair(ln, srv))
	go func() { _ = srv.Serve(ln) }()
	s.T().Cleanup(func() { _ = srv.Close() })
	return "http://" + ln.Addr().String()
}

// refuses reports whether nothing accepts connections at addr any more.
func refuses(addr string) bool {
	c, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		return true
	}
	_ = c.Close()
	return false
}

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
		s.Require().Failf("StopServer returned before the request finished", "err: %v", err)
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

// blockUntil is a handler that signals started and then waits for release,
// ignoring the request context: the kind of handler only a force-close ends.
func blockUntil(started chan<- struct{}, release <-chan struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started <- struct{}{}
		<-release
		_, _ = w.Write([]byte("done"))
	})
}

// TestStopServerDrainsEveryListenerConcurrently: all listeners stop accepting
// as soon as StopServer is called, while each still finishes its own request.
// A sequential drain would leave the later listeners open until the earlier
// ones had finished.
func (s *APITestSuite) TestStopServerDrainsEveryListenerConcurrently() {
	const n = 3
	api := NewHTTPAPI(nil)
	started := make(chan struct{}, n)
	release := make(chan struct{})
	urls := make([]string, n)
	for i := range urls {
		urls[i] = s.servePair(api, s.T().Context(), blockUntil(started, release))
	}

	bodies := make(chan string, n)
	for _, u := range urls {
		go func() {
			resp, err := http.Get(u)
			if err != nil {
				bodies <- "error: " + err.Error()
				return
			}
			defer func() { _ = resp.Body.Close() }()
			b, _ := io.ReadAll(resp.Body)
			bodies <- string(b)
		}()
	}
	for range n {
		<-started
	}

	ctx, cancel := context.WithTimeout(s.T().Context(), 30*time.Second)
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- api.StopServer(ctx) }()

	deadline := time.Now().Add(10 * time.Second)
	for _, u := range urls {
		addr := u[len("http://"):]
		for !refuses(addr) {
			s.Require().True(time.Now().Before(deadline), "listener %s still accepting while the others drain", addr)
			time.Sleep(5 * time.Millisecond)
		}
	}
	select {
	case err := <-stopped:
		s.Require().Failf("StopServer returned with requests in flight", "err: %v", err)
	default:
	}

	close(release)
	s.NoError(<-stopped)
	for range n {
		s.Equal("done", <-bodies)
	}
}

// TestStopServerNamesTheListenersThatMissedTheDeadline: with one drained and
// two stuck listeners, the error wraps context.DeadlineExceeded and names
// exactly the stuck ones; the stuck handlers ignore cancellation, so only the
// force-close ends their clients' requests.
func (s *APITestSuite) TestStopServerNamesTheListenersThatMissedTheDeadline() {
	api := NewHTTPAPI(nil)
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	defer close(release)

	idle := s.servePair(api, s.T().Context(), http.NotFoundHandler())
	stuckA := s.servePair(api, s.T().Context(), blockUntil(started, release))
	stuckB := s.servePair(api, s.T().Context(), blockUntil(started, release))

	clientErrs := make(chan error, 2)
	for _, u := range []string{stuckA, stuckB} {
		go func() {
			resp, err := http.Get(u)
			if err == nil {
				_ = resp.Body.Close()
			}
			clientErrs <- err
		}()
	}
	<-started
	<-started

	ctx, cancel := context.WithTimeout(s.T().Context(), 100*time.Millisecond)
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- api.StopServer(ctx) }()

	var err error
	select {
	case err = <-stopped:
	case <-time.After(10 * time.Second):
		s.FailNow("StopServer did not return after the grace period")
	}
	s.Require().ErrorIs(err, context.DeadlineExceeded)
	s.Contains(err.Error(), stuckA[len("http://"):])
	s.Contains(err.Error(), stuckB[len("http://"):])
	s.NotContains(err.Error(), idle[len("http://"):])

	for range 2 {
		s.Error(<-clientErrs, "the force-close should have dropped the stuck request")
	}
}

// TestShutdownClosesAListenerNobodyServes: a pair whose Serve never ran (the
// signal arrived between Setup and StartServer) still releases its port.
func (s *APITestSuite) TestShutdownClosesAListenerNobodyServes() {
	api := NewHTTPAPI(nil)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	pair := NewListenerServerPair(ln, api.newServer(s.T().Context(), http.NotFoundHandler()))
	addr := ln.Addr().String()

	s.NoError(pair.Shutdown(s.T().Context()))
	s.True(refuses(addr))
}

// TestStopServerDoesNotRaceWithStartServer: serve runs StartServer in a
// goroutine and a signal can arrive at any point after Setup, so StopServer
// must not write the listener slice StartServer reads. Run with -race.
func (s *APITestSuite) TestStopServerDoesNotRaceWithStartServer() {
	api := NewHTTPAPI(nil)
	s.Require().NoError(api.Setup(s.T().Context(), 0, []string{"127.0.0.1"}))

	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		close(started)
		api.StartServer(s.T().Context(), 0, []string{"127.0.0.1"})
	}()
	<-started

	ctx, cancel := context.WithTimeout(s.T().Context(), 5*time.Second)
	defer cancel()
	s.NoError(api.StopServer(ctx))
	// A second stop is harmless.
	s.NoError(api.StopServer(ctx))

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		s.Fail("StartServer did not return after StopServer")
	}
}
