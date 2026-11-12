// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/runtime"
)

// DecisionRequest represents the request body for rule execution
type DecisionRequest struct {
	Facts map[string]any `json:"facts"`
}

// DecisionResponse represents the response from rule execution
type DecisionResponse struct {
	Decisions []*runtime.ExecutorOutput `json:"decisions"`
	Error     string                    `json:"error,omitempty"`
}

// handleDecision handles POST /decision/{namespace...} requests
func (api *HTTPAPI) handleDecision(w http.ResponseWriter, r *http.Request) {
	// r.Context() adds request-scoped cancellation when the client goes away.
	// The evaluation deadline is the executor's, set per request below, so
	// there is one deadline and a timeout names it.
	ctx := r.Context()

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, "+TimeoutHeader)
	w.Header().Set("Access-Control-Expose-Headers", TimeoutAppliedHeader)

	path := r.PathValue("target")
	if path == "" {
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid Path", "The path parameter is required but was not provided")
		return
	}

	api.logger.InfoContext(ctx, "handleDecision", "path", path)

	// Handle preflight OPTIONS requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Only allow POST requests
	if r.Method != "POST" {
		api.writeErrorResponse(w, r, http.StatusMethodNotAllowed, "Method Not Allowed", "Only POST requests are supported for this endpoint")
		return
	}

	// Sentrie-Timeout chooses the deadline (#243); a bad value is the
	// caller's error, never a silent fallback to the default.
	deadline, err := api.requestDeadline(r.Header)
	if err != nil {
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid "+TimeoutHeader, err.Error())
		return
	}
	if deadline > 0 {
		ctx = runtime.ContextWithEvalTimeout(ctx, deadline)
		w.Header().Set(TimeoutAppliedHeader, deadline.String())
	}

	// Create span for path resolution
	namespace, policy, rule, err := api.executor.Index().ResolveSegments(strings.TrimPrefix(path, "/decision/"))
	if err != nil {
		api.writeErrorResponse(w, r, http.StatusNotFound, "Invalid Path", err.Error())
		return
	}

	// Parse query parameters for runconfig
	runConfig := make(map[string]string)
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			runConfig[key] = values[0]
		}
	}

	// Parse request body. limitBody caps what can be read (#115); a body
	// that declares or turns out to be over the limit is a 413, never a 400.
	if api.bodyOverLimit(r) {
		api.writeBodyTooLarge(w, r)
		return
	}
	var req DecisionRequest
	dec := json.NewDecoder(r.Body)
	dec.UseNumber() // keep integers beyond 2^53 exact (#166)
	if err := dec.Decode(&req); err != nil {
		if bodyTooLarge(err) {
			api.writeBodyTooLarge(w, r)
			return
		}
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid JSON", "The request body could not be parsed as valid JSON")
		return
	}
	// Only an over-limit drain matters: any other read error (such as a client
	// reset during trailing bytes) is ignored because the value is already decoded.
	if err := api.drainBody(r); bodyTooLarge(err) {
		api.writeBodyTooLarge(w, r)
		return
	}
	if err := box.CheckJSONNumbers(req.Facts); err != nil {
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid JSON", "The request body contains a number outside the supported range")
		return
	}

	// Execute policy/rule
	var outputs []*runtime.ExecutorOutput
	var runErr error
	if len(rule) == 0 {
		outputs, runErr = api.executor.ExecPolicy(ctx, namespace, policy, req.Facts)
	} else {
		output, e := api.executor.ExecRule(ctx, namespace, policy, rule, req.Facts)
		if output != nil {
			outputs = []*runtime.ExecutorOutput{output}
		}
		runErr = e
	}

	// A policy whose rules ran answers with every rule's result, and a rule
	// that failed carries its own error (#143): 207 Multi-Status when any rule
	// failed, even all of them. Any other failure, including a named rule's,
	// is a problem document.
	status := http.StatusOK
	var ruleErrs runtime.RuleErrors
	if len(rule) == 0 && errors.As(runErr, &ruleErrs) {
		status = http.StatusMultiStatus
	} else if runErr != nil {
		api.writeErrorResponse(w, r, runtime.ErrorStatus(runErr), "Evaluation Failed", runErr.Error())
		return
	}

	response := DecisionResponse{
		Decisions: outputs,
	}

	// Encode before writing the status: a value JSON cannot represent (a
	// non-finite number from arithmetic) is a 500, not a 200 with no body.
	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(response); err != nil {
		api.logger.ErrorContext(ctx, "Error encoding response", "error", err)
		api.writeErrorResponse(w, r, http.StatusInternalServerError, "Encoding Error", "The decision could not be encoded as JSON")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(body.Bytes()); err != nil {
		api.logger.ErrorContext(ctx, "Error writing response", "error", err)
	}
}
