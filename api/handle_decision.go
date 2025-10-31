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
	"time"

	"github.com/sentrie-sh/sentrie/runtime"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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
	ctx, span := api.tracer.Start(ctx, "decision.request")
	defer span.End()

	// Start timing
	start := time.Now()

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

	// Create span for path resolution
	namespace, policy, rule, err := api.executor.Index().ResolveSegments(strings.TrimPrefix(path, "/decision/"))
	if err != nil {
		api.writeErrorResponse(w, r, http.StatusNotFound, "Invalid Path", err.Error())
		return
	}

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
		span.RecordError(err)
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid JSON", "The request body could not be parsed as valid JSON")
		return
	}

	// Execute policy/rule
	var outputs []*runtime.ExecutorOutput
	var runErr error
	if api.metrics != nil {
		api.metrics.ActiveEvaluations.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("sentrie.namespace", namespace),
				attribute.String("sentrie.policy", policy),
				attribute.String("sentrie.rule", rule),
			),
		)
	}
	if len(rule) == 0 {
		outputs, runErr = api.executor.ExecPolicy(ctx, namespace, policy, req.Facts)
	} else {
		output, e := api.executor.ExecRule(ctx, namespace, policy, rule, req.Facts)
		outputs = []*runtime.ExecutorOutput{output}
		runErr = e
	}

	// Record execution metrics
	execDuration := float64(time.Since(start).Nanoseconds()) / 1e6 // Convert to milliseconds
	if api.metrics != nil {
		api.metrics.DecisionDuration.Record(ctx, execDuration)
	}

	// Determine outcome for metrics
	outcome := "unknown"
	if runErr == nil && len(outputs) > 0 {
		switch outputs[0].Decision.State.String() {
		case "true":
			outcome = "true"
		case "false":
			outcome = "false"
		default:
			outcome = "unknown"
		}
	} else if runErr != nil {
		outcome = "error"
	}

	// Record decision count
	if api.metrics != nil {
		api.metrics.DecisionCount.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("sentrie.namespace", namespace),
				attribute.String("sentrie.policy", policy),
				attribute.String("sentrie.rule", rule),
				attribute.String("sentrie.outcome", outcome),
			),
		)

		api.metrics.ActiveEvaluations.Add(ctx, -1,
			metric.WithAttributes(
				attribute.String("sentrie.namespace", namespace),
				attribute.String("sentrie.policy", policy),
				attribute.String("sentrie.rule", rule),
			),
		)
	}

	response := DecisionResponse{
		Decisions: outputs,
		Error:     runErr.Error(),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		span.RecordError(err)
		api.logger.ErrorContext(ctx, "Error encoding response", "error", err)
	}
}
