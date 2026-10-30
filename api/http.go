// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/sentrie-sh/sentrie/api/middleware"
	"github.com/sentrie-sh/sentrie/runtime"
)

type ListenerServerPair struct {
	Listener net.Listener
	Server   *http.Server
}

func NewListenerServerPair(listener net.Listener, server *http.Server) *ListenerServerPair {
	return &ListenerServerPair{Listener: listener, Server: server}
}

func (p *ListenerServerPair) Close() error {
	err := p.Listener.Close()
	if err != nil {
		return err
	}
	err = p.Server.Close()
	if err != nil {
		return err
	}
	return nil
}

// Shutdown drains the server: it stops accepting connections and waits for
// in-flight requests until ctx is done. A ctx that ends first is returned as
// the error and the caller decides whether to fall back to Close. The listener
// is closed afterwards in case Serve never took ownership of it.
func (p *ListenerServerPair) Shutdown(ctx context.Context) error {
	err := p.Server.Shutdown(ctx)
	if cerr := p.Listener.Close(); cerr != nil && !errors.Is(cerr, net.ErrClosed) {
		err = errors.Join(err, cerr)
	}
	return err
}

// HTTPAPI provides HTTP endpoints for rule execution
type HTTPAPI struct {
	executor          runtime.Executor
	listeners         []*ListenerServerPair
	logger            *slog.Logger
	requestTimeout    time.Duration
	maxRequestTimeout time.Duration
	// readHeaderTimeout is serverReadHeaderTimeout; tests shorten it.
	readHeaderTimeout time.Duration
}

// HTTPAPIOption configures an HTTPAPI.
type HTTPAPIOption func(*HTTPAPI)

// WithRequestTimeout sets the default evaluation deadline of a decision
// request (#142; serve passes --eval-timeout). The executor enforces it: the
// API hands it over with runtime.ContextWithEvalTimeout, so there is one
// deadline per request. A value <= 0 leaves a request that does not send
// Sentrie-Timeout bounded only by the executor's own deadline and the client
// connection.
func WithRequestTimeout(d time.Duration) HTTPAPIOption {
	return func(api *HTTPAPI) {
		api.requestTimeout = d
	}
}

// WithMaxRequestTimeout sets the longest deadline a request may ask for with
// Sentrie-Timeout (#243; serve passes --eval-max-timeout). A longer request is
// rejected with 400, and a default that is unbounded or longer is capped to
// it. A value <= 0 makes the default the ceiling, so a request can only
// shorten its deadline; with no default either, any value is allowed.
func WithMaxRequestTimeout(d time.Duration) HTTPAPIOption {
	return func(api *HTTPAPI) {
		api.maxRequestTimeout = d
	}
}

// DefaultShutdownGracePeriod is how long serve lets in-flight requests finish
// after a shutdown signal before the connections are closed (#117). It is
// shorter than Kubernetes' default 30s terminationGracePeriodSeconds, so the
// process exits by itself before the kubelet sends SIGKILL.
const DefaultShutdownGracePeriod = 10 * time.Second

// responseWriteMargin is how long the server keeps the connection writable
// after the request deadline, to send the timeout response.
const responseWriteMargin = 5 * time.Second

// Connection limits for every listener (#247). A client that sends its
// headers slowly, or keeps an idle connection open, would otherwise hold a
// goroutine and a file descriptor for as long as serverReadTimeout allows.
const (
	serverReadTimeout       = 30 * time.Second
	serverReadHeaderTimeout = 5 * time.Second
	serverIdleTimeout       = 60 * time.Second
	serverMaxHeaderBytes    = 64 << 10 // 64 KiB; net/http's default is 1 MiB
)

// NewHTTPAPI creates a new HTTP API instance
func NewHTTPAPI(executor runtime.Executor, opts ...HTTPAPIOption) *HTTPAPI {
	api := &HTTPAPI{
		executor:          executor,
		logger:            slog.Default(),
		requestTimeout:    runtime.DefaultEvalTimeout,
		readHeaderTimeout: serverReadHeaderTimeout,
	}
	for _, opt := range opts {
		opt(api)
	}
	return api
}

// writeTimeout is the server write timeout: 30 seconds, or longer when the
// longest deadline a request may get needs it. With no ceiling it stays at 30
// seconds, which then caps how long a response can take.
func (api *HTTPAPI) writeTimeout() time.Duration {
	ceiling := api.maxTimeout()
	if ceiling > math.MaxInt64-responseWriteMargin {
		return time.Duration(math.MaxInt64)
	}
	return max(30*time.Second, ceiling+responseWriteMargin)
}

// newServer builds the server every listener uses, with its read, header,
// write and idle limits. BaseContext derives each request's context from ctx
// without its cancellation: cancelling ctx is what starts a shutdown, and the
// requests in flight must be left to finish (#117). Closing a connection
// still cancels its request.
func (api *HTTPAPI) newServer(ctx context.Context, handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		ReadTimeout:       serverReadTimeout,
		ReadHeaderTimeout: api.readHeaderTimeout,
		WriteTimeout:      api.writeTimeout(),
		IdleTimeout:       serverIdleTimeout,
		MaxHeaderBytes:    serverMaxHeaderBytes,
		BaseContext: func(net.Listener) context.Context {
			return context.WithoutCancel(ctx)
		},
	}
}

func (api *HTTPAPI) Setup(ctx context.Context, port int, listen []string) error {
	mux := http.NewServeMux()

	// Register the decision endpoint using Go 1.24 syntax
	mux.Handle("POST /decision/{target...}",
		middleware.RequestIDMiddleware(
			http.HandlerFunc(api.handleDecision),
		),
	)

	// Health check endpoint
	mux.Handle("GET /health", http.HandlerFunc(api.handleHealth))

	bindings, err := resolveBindings(port, listen)
	if err != nil {
		return err
	}

	// Start listeners on all addresses
	api.listeners = make([]*ListenerServerPair, 0, len(bindings))
	for _, binding := range bindings {
		ln, err := net.Listen("tcp", binding)
		if err != nil {
			// Close any already opened listeners
			for _, l := range api.listeners {
				_ = l.Close()
			}
			api.listeners = nil
			return fmt.Errorf("failed to listen on %s: %w", binding, err)
		}
		api.listeners = append(api.listeners, NewListenerServerPair(ln, api.newServer(ctx, mux)))
		api.logger.DebugContext(ctx, "Listening on server", "binding", binding)
	}
	return nil
}

// StartServer starts the HTTP server on the specified addresses
func (api *HTTPAPI) StartServer(ctx context.Context, port int, listen []string) {
	// Start serving on all listeners
	var wg sync.WaitGroup
	errChan := make(chan error, len(api.listeners))

	for _, ln := range api.listeners {
		server := ln.Server
		wg.Go(func() {
			api.logger.DebugContext(ctx,
				"Decision endpoint available",
				"method", "POST",
				"address", ln.Listener.Addr().String(),
				"url", fmt.Sprintf("http://%s/decision/{namespace}/{policy}/{rule}", ln.Listener.Addr().String()))

			api.logger.DebugContext(ctx,
				"Health check endpoint available",
				slog.String("method", "GET"),
				slog.String("address", ln.Listener.Addr().String()),
				slog.String("url", fmt.Sprintf("http://%s/health", ln.Listener.Addr().String())))
			if err := server.Serve(ln.Listener); err != nil && err != http.ErrServerClosed {
				errChan <- err
			}
		})
	}

	defer func() {
		wg.Wait()
		close(errChan)
	}()

}

// StopServer shuts every listener's server down, letting in-flight requests
// finish until ctx is done. Pass a context that is not already cancelled (serve
// derives one with context.WithoutCancel and a timeout of
// DefaultShutdownGracePeriod). A server still busy when ctx ends is closed, so
// the caller can exit, and the error says so. Errors from the listeners are
// joined.
func (api *HTTPAPI) StopServer(ctx context.Context) error {
	// Only read the slice: serve runs StartServer in a goroutine that reads it
	// too, and a signal can arrive right after Setup. A second StopServer is
	// harmless because Shutdown is idempotent and a closed listener is ignored.
	listeners := api.listeners

	errs := make([]error, len(listeners))
	var wg sync.WaitGroup
	for i, ln := range listeners {
		wg.Go(func() {
			err := ln.Shutdown(ctx)
			if err == nil {
				return
			}
			// The grace period ran out with requests still in flight: drop them.
			api.logger.WarnContext(ctx, "Shutdown grace period ended, closing connections",
				"address", ln.Listener.Addr().String(), "error", err)
			errs[i] = errors.Join(
				fmt.Errorf("graceful shutdown of %s: %w", ln.Listener.Addr(), err),
				ln.Server.Close(),
			)
		})
	}
	wg.Wait()

	return errors.Join(errs...)
}

// handleHealth handles GET /health requests
func (api *HTTPAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"status": "healthy",
		"time":   time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.logger.DebugContext(r.Context(), "Error encoding health response", "error", err)
	}
}

// writeErrorResponse writes a Problem Details error response in JSON format
func (api *HTTPAPI) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, title, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(statusCode)

	response := ProblemDetails{
		Type:     fmt.Sprintf("https://sentrie.sh/problems/%d", statusCode),
		Title:    title,
		Status:   statusCode,
		Detail:   detail,
		Instance: middleware.GetRequestIDFromRequest(r),
		Ext: map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.logger.DebugContext(r.Context(), "Error encoding problem details response", "error", err)
	}
}
