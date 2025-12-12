// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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

// HTTPAPI provides HTTP endpoints for rule execution
type HTTPAPI struct {
	executor  runtime.Executor
	listeners []*ListenerServerPair
	logger    *slog.Logger
}

// NewHTTPAPI creates a new HTTP API instance
func NewHTTPAPI(executor runtime.Executor) *HTTPAPI {
	return &HTTPAPI{
		executor: executor,
		logger:   slog.Default(),
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
		api.listeners = append(api.listeners, NewListenerServerPair(ln, &http.Server{
			Handler:      mux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			BaseContext: func(l net.Listener) context.Context {
				return ctx
			},
		}))
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

// StopServer gracefully stops the HTTP server
func (api *HTTPAPI) StopServer(ctx context.Context) error {
	if api.listeners != nil {
		for _, ln := range api.listeners {
			_ = ln.Close()
		}
		api.listeners = nil
	}

	return nil
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
