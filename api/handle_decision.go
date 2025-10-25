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

// handleDecision handles POST /decision/{namespace...} requests
func (api *HTTPAPI) handleDecision(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tracer := api.tracer
	meter := api.meter

	// Create decision-specific metrics
	decisionCount, _ := meter.Int64Counter(
		"sentrie.decision.count",
		metric.WithDescription("Number of decisions made"),
	)
	decisionDuration, _ := meter.Float64Histogram(
		"sentrie.decision.duration",
		metric.WithDescription("Decision execution duration in milliseconds"),
		metric.WithUnit("ms"),
	)
	decisionFactsCount, _ := meter.Int64Histogram(
		"sentrie.decision.facts.count",
		metric.WithDescription("Number of facts per decision"),
	)

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
	ctx, pathSpan := tracer.Start(ctx, "decision.path_resolution")
	namespace, policy, rule, err := api.executor.Index().ResolveSegments(strings.TrimPrefix(path, "/decision/"))
	pathSpan.End()
	if err != nil {
		pathSpan.RecordError(err)
		api.writeErrorResponse(w, r, http.StatusNotFound, "Invalid Path", err.Error())
		return
	}

	// Add span attributes
	pathSpan.SetAttributes(
		attribute.String("sentrie.namespace", namespace),
		attribute.String("sentrie.policy", policy),
		attribute.String("sentrie.rule", rule),
	)

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
	ctx, parseSpan := tracer.Start(ctx, "decision.request_parsing")
	var req DecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		parseSpan.RecordError(err)
		api.writeErrorResponse(w, r, http.StatusBadRequest, "Invalid JSON", "The request body could not be parsed as valid JSON")
		return
	}
	parseSpan.End()

	// Record facts count
	factsCount := int64(len(req.Facts))
	decisionFactsCount.Record(ctx, factsCount)

	// Execute policy/rule
	ctx, execSpan := tracer.Start(ctx, "decision.execution")
	var outputs []*runtime.ExecutorOutput
	var runErr error
	if len(rule) == 0 {
		outputs, runErr = api.executor.ExecPolicy(ctx, namespace, policy, req.Facts)
	} else {
		output, e := api.executor.ExecRule(ctx, namespace, policy, rule, req.Facts)
		outputs = []*runtime.ExecutorOutput{output}
		runErr = e
	}
	execSpan.End()

	// Record execution metrics
	execDuration := float64(time.Since(start).Nanoseconds()) / 1e6 // Convert to milliseconds
	decisionDuration.Record(ctx, execDuration)

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
	decisionCount.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("sentrie.namespace", namespace),
			attribute.String("sentrie.policy", policy),
			attribute.String("sentrie.rule", rule),
			attribute.String("sentrie.outcome", outcome),
		),
	)

	response := DecisionResponse{
		Decisions: outputs,
		Error:     runErr.Error(),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	ctx, encodeSpan := tracer.Start(ctx, "decision.response_encoding")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		encodeSpan.RecordError(err)
		api.logger.DebugContext(ctx, "Error encoding response", "error", err)
	}
	encodeSpan.End()
}
