package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/binaek/sentra/runtime"
)

// handleDecision handles POST /decision/{namespace...} requests
func (api *HTTPAPI) handleDecision(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	path := r.PathValue("target")
	if path == "" {
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid Path", "The path parameter is required but was not provided")
		return
	}

	slog.InfoContext(r.Context(), "handleDecision", "path", path)

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
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid JSON", "The request body could not be parsed as valid JSON")
		return
	}

	var outputs []*runtime.ExecutorOutput
	var runErr error
	if len(rule) == 0 {
		outputs, runErr = api.executor.ExecPolicy(r.Context(), namespace, policy, req.Facts)
	} else {
		output, e := api.executor.ExecRule(r.Context(), namespace, policy, rule, req.Facts)
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
		slog.DebugContext(r.Context(), "Error encoding response", "error", err)
	}
}
