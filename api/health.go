// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/sentrie-sh/sentrie/index"
)

// Health states reported by /health and /health/deep. There is no "degraded":
// nothing the process can observe sits between working and broken.
const (
	statusHealthy   = "healthy"
	statusUnhealthy = "unhealthy"
)

// Fixed messages for index failures. The index's own error text carries the
// pack's absolute path, which an unauthenticated probe must not return; the
// full error goes to the debug log instead.
const (
	msgIndexValidationFailed = "index validation failed"
	msgIndexCommitFailed     = "index commit failed"
)

// registerHealthRoutes adds the two probe routes to mux.
func (api *HTTPAPI) registerHealthRoutes(mux *http.ServeMux) {
	mux.Handle("GET /health", http.HandlerFunc(api.handleHealth))
	mux.Handle("GET /health/deep", http.HandlerFunc(api.handleDeepHealth))
}

// healthResponse is the body of GET /health.
type healthResponse struct {
	// CheckID is a fresh random nonce per response, so a caller can tell a live
	// answer from a cached or replayed one.
	CheckID string `json:"check_id"`
	Status  string `json:"status"`
	Time    string `json:"time"`
}

// deepHealthResponse is the body of GET /health/deep. It reports only what the
// process can observe about itself. The issue's executor caches, JS instance
// pools, pack location and operational probes are not here because the runtime
// exposes nothing to report them from, and a made-up value would be worse than
// an absent one.
type deepHealthResponse struct {
	Status     string               `json:"status"`
	Time       string               `json:"time"`
	Components deepHealthComponents `json:"components"`
	Errors     []deepHealthError    `json:"errors"`
}

type deepHealthComponents struct {
	Index     indexHealth     `json:"index"`
	Executor  executorHealth  `json:"executor"`
	JSRuntime jsRuntimeHealth `json:"js_runtime"`
	Pack      packHealth      `json:"pack"`
}

type deepHealthError struct {
	Component string `json:"component"`
	Message   string `json:"message"`
}

type indexHealth struct {
	Status          string          `json:"status"`
	Validated       bool            `json:"validated"`
	Committed       bool            `json:"committed"`
	ValidationError *string         `json:"validation_error"`
	CommitError     *string         `json:"commit_error"`
	Statistics      indexStatistics `json:"statistics"`
}

type indexStatistics struct {
	Namespaces int `json:"namespaces"`
	Policies   int `json:"policies"`
	Programs   int `json:"programs"`
	Rules      int `json:"rules"`
	Shapes     int `json:"shapes"`
}

type executorHealth struct {
	Status      string `json:"status"`
	Initialized bool   `json:"initialized"`
}

type jsRuntimeHealth struct {
	Status              string         `json:"status"`
	RegistryInitialized bool           `json:"registry_initialized"`
	BuiltinModules      builtinModules `json:"builtin_modules"`
}

type builtinModules struct {
	Registered int      `json:"registered"`
	Modules    []string `json:"modules"`
}

// packHealth deliberately omits the pack's filesystem location: it is whatever
// path the operator gave `serve`, often absolute, and a health endpoint is
// reachable by anyone who can reach the port.
type packHealth struct {
	Status     string  `json:"status"`
	Loaded     bool    `json:"loaded"`
	Name       string  `json:"name,omitempty"`
	Version    string  `json:"version,omitempty"`
	Accessible *bool   `json:"accessible,omitempty"`
	Error      *string `json:"error,omitempty"`
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// handleHealth handles GET /health. It is a liveness probe: it answers 200 for
// any process that can serve a request. Use /health/deep for readiness.
func (api *HTTPAPI) handleHealth(w http.ResponseWriter, r *http.Request) {
	api.writeHealthJSON(w, r, http.StatusOK, healthResponse{
		CheckID: rand.Text(),
		Status:  statusHealthy,
		Time:    nowRFC3339(),
	})
}

// handleDeepHealth handles GET /health/deep. It answers 200 when every
// component is healthy and 503 when any is not, so a load balancer can act on
// the status code alone.
func (api *HTTPAPI) handleDeepHealth(w http.ResponseWriter, r *http.Request) {
	resp := api.deepHealth(r.Context())
	code := http.StatusOK
	if resp.Status != statusHealthy {
		code = http.StatusServiceUnavailable
	}
	api.writeHealthJSON(w, r, code, resp)
}

func (api *HTTPAPI) writeHealthJSON(w http.ResponseWriter, r *http.Request, code int, body any) {
	w.Header().Set("Content-Type", "application/json")
	// A health answer is only worth anything if it is fresh.
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		api.logger.DebugContext(r.Context(), "Error encoding health response", "error", err)
	}
}

// deepHealth inspects the executor and its index without changing either.
func (api *HTTPAPI) deepHealth(ctx context.Context) deepHealthResponse {
	resp := deepHealthResponse{Time: nowRFC3339(), Errors: []deepHealthError{}}
	fail := func(component, message string) {
		resp.Errors = append(resp.Errors, deepHealthError{Component: component, Message: message})
	}

	var idx *index.Index
	if api.executor != nil {
		idx = api.executor.Index()
	}

	resp.Components.Executor = executorHealth{Status: statusHealthy, Initialized: api.executor != nil && idx != nil}
	if !resp.Components.Executor.Initialized {
		resp.Components.Executor.Status = statusUnhealthy
		fail("executor", "executor is not initialized")
	}

	resp.Components.Index = indexHealth{Status: statusUnhealthy}
	if idx != nil {
		resp.Components.Index = api.indexHealth(ctx, idx, fail)
	} else {
		fail("index", "index is not available")
	}

	resp.Components.JSRuntime = jsRuntimeHealth{Status: statusUnhealthy, BuiltinModules: builtinModules{Modules: []string{}}}
	if api.executor != nil {
		if mods := api.executor.BuiltinModules(); len(mods) > 0 {
			resp.Components.JSRuntime = jsRuntimeHealth{
				Status:              statusHealthy,
				RegistryInitialized: true,
				BuiltinModules:      builtinModules{Registered: len(mods), Modules: mods},
			}
		}
	}
	if resp.Components.JSRuntime.Status != statusHealthy {
		fail("js_runtime", "no built-in modules are registered")
	}

	resp.Components.Pack = packHealthOf(idx, fail)

	resp.Status = statusHealthy
	for _, status := range []string{
		resp.Components.Index.Status,
		resp.Components.Executor.Status,
		resp.Components.JSRuntime.Status,
		resp.Components.Pack.Status,
	} {
		if status != statusHealthy {
			resp.Status = statusUnhealthy
		}
	}
	return resp
}

func (api *HTTPAPI) indexHealth(ctx context.Context, idx *index.Index, fail func(component, message string)) indexHealth {
	st := idx.Status()

	// A failed commit during Validate is recorded as the validation error as
	// well. Report it once, under commit_error.
	valErr := st.ValidationError
	if valErr != nil && valErr == st.CommitError {
		valErr = nil
	}

	h := indexHealth{
		Status:    statusHealthy,
		Validated: st.Validated && valErr == nil,
		Committed: st.Committed && st.CommitError == nil,
		Statistics: indexStatistics{
			Namespaces: st.Namespaces,
			Policies:   st.Policies,
			Programs:   st.Programs,
			Rules:      st.Rules,
			Shapes:     st.Shapes,
		},
	}
	switch {
	case valErr != nil:
		// The error text embeds source spans that start with the pack's absolute
		// path, so it stays in the log; the unauthenticated response says only
		// that it failed.
		api.logger.DebugContext(ctx, "Deep health: index validation failed", "error", valErr)
		msg := msgIndexValidationFailed
		h.ValidationError = &msg
		fail("index", msg)
	case !st.Validated:
		fail("index", "index has not been validated")
	}
	if st.CommitError != nil {
		api.logger.DebugContext(ctx, "Deep health: index commit failed", "error", st.CommitError)
		msg := msgIndexCommitFailed
		h.CommitError = &msg
		fail("index", msg)
	} else if st.Validated && valErr == nil && !st.Committed {
		fail("index", "index has not been committed")
	}
	if !h.Validated || !h.Committed {
		h.Status = statusUnhealthy
	}
	return h
}

func packHealthOf(idx *index.Index, fail func(component, message string)) packHealth {
	if idx == nil || idx.Pack == nil {
		fail("pack", "pack is not loaded")
		return packHealth{Status: statusUnhealthy}
	}
	p := packHealth{Status: statusHealthy, Loaded: true}
	if idx.Pack.Pack != nil {
		p.Name = idx.Pack.Pack.Name
		if idx.Pack.Pack.Version != nil {
			p.Version = idx.Pack.Pack.Version.String()
		}
	}
	// The pack is read into memory at startup, but modules are read from its
	// directory on first use, so a pack directory that has gone away matters.
	info, err := os.Stat(idx.Pack.Location)
	accessible := err == nil && info.IsDir()
	p.Accessible = &accessible
	if !accessible {
		msg := "pack directory is not accessible"
		p.Status = statusUnhealthy
		p.Error = &msg
		fail("pack", msg)
	}
	return p
}
