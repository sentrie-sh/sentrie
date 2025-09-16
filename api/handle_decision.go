package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/xerr"
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

	namespace, policy, rule, err := resolveSegmentsFromPath(strings.TrimPrefix(path, "/decision/"), api.executor.Index())
	if err != nil {
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid Path", fmt.Sprintf("The provided path could not be parsed: %s", err.Error()))
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

	// Execute the rule
	d, attachments, traceNode, err := api.executor.ExecRule(r.Context(), namespace, policy, rule, req.Facts)

	if err != nil {
		// Use the error message directly
		api.writeErrorResponse(w, r, http.StatusInternalServerError, "Rule Execution Failed", fmt.Sprintf("An error occurred while executing the rule: %s", err.Error()))
		return
	}

	// Prepare response
	response := DecisionResponse{
		Decision:    d,
		Attachments: attachments,
		Trace:       traceNode,
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		slog.DebugContext(r.Context(), "Error encoding response", "error", err)
	}
}

func resolveSegmentsFromPath(path string, idx *index.Index) (ns, policy, rule string, err error) {
	// split by .
	parts := strings.Split(path, "/")
	// start joining the parts, until we have a namespace, or we run out of parts

	nsName := ""
	for {
		nextPart := parts[0]
		parts = parts[1:]

		if nextPart == "" {
			continue
		}
		if len(nsName) == 0 {
			nsName = nextPart
		} else {
			nsName = strings.Join([]string{nsName, nextPart}, ast.FQNSeparator)
		}
		n, err := idx.ResolveNamespace(nsName)
		if errors.Is(err, xerr.NotFoundError{}) {
			continue
		}

		// if we have an error, and it's not a namespace not found error, return the error
		if err != nil {
			return "", "", "", err
		}

		if n != nil {
			nsName = n.FQN.String()
			break
		}
		if len(parts) == 0 {
			return "", "", "", xerr.ErrNamespaceNotFound(path)
		}
	}

	// if we do not have at least 1 part left, return an error - it's a problem - we MUST have a policy name
	if len(parts) == 0 {
		return "", "", "", xerr.ErrPolicyNotFound(path)
	}

	// we have a namespace, the next segment is the policy name
	policyName, parts := parts[0], parts[1:]
	_, err = idx.ResolvePolicy(nsName, policyName)
	if err != nil {
		return "", "", "", err
	}

	// we have a policy, the next segment is the rule name
	ruleName := ""

	if len(parts) > 0 {
		ruleName = parts[0]
	}

	return nsName, policyName, ruleName, nil
}
