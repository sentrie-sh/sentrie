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
	"encoding/json"
	"net/http"
	"strings"

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
	ctx := r.Context()

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

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

	// Parse request body
	var req DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid JSON", "The request body could not be parsed as valid JSON")
		return
	}

	// Execute policy/rule
	var outputs []*runtime.ExecutorOutput
	var runErr error
	if len(rule) == 0 {
		outputs, runErr = api.executor.ExecPolicy(ctx, namespace, policy, req.Facts)
	} else {
		output, e := api.executor.ExecRule(ctx, namespace, policy, rule, req.Facts)
		outputs = []*runtime.ExecutorOutput{output}
		runErr = e
	}

	response := DecisionResponse{
		Decisions: outputs,
		Error:     runErr.Error(),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.logger.ErrorContext(ctx, "Error encoding response", "error", err)
	}
}
