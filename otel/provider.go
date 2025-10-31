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

package otel

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"runtime/metrics"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// OTelConfig holds configuration for OpenTelemetry initialization
type OTelConfig struct {
	Enabled        bool
	Endpoint       string
	Protocol       string
	ServiceName    string
	ServiceVersion string
	PackName       string
	TraceExecution bool
}

type ShutdownFn func(context.Context) error

// InitProvider initializes OpenTelemetry providers and returns a cleanup function
func InitProvider(ctx context.Context, config OTelConfig) (ShutdownFn, error) {
	if !config.Enabled {
		return func(context.Context) error { return nil }, nil
	}

	// Parse endpoint URL
	endpointURL, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// Create resource with service name and pack name
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(config.ServiceName),
		semconv.ServiceVersionKey.String(config.ServiceVersion),
	}
	if config.PackName != "" {
		attrs = append(attrs,
			semconv.ServiceNamespaceKey.String(config.PackName),
		)
	}
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attrs...,
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	var cleanupFuncs []func(context.Context) error

	// Initialize trace provider
	traceExporter, traceCleanup, err := createTraceExporter(ctx, config.Protocol, endpointURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, traceCleanup)

	tracerProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithResource(res),
	)
	cleanupFuncs = append(cleanupFuncs, tracerProvider.Shutdown)

	// Initialize metric provider
	metricExporter, metricCleanup, err := createMetricExporter(ctx, config.Protocol, endpointURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, metricCleanup)

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExporter)),
		sdkmetric.WithResource(res),
	)
	cleanupFuncs = append(cleanupFuncs, meterProvider.Shutdown)

	// Initialize log provider
	_, logCleanup, err := createLogExporter(ctx, config.Protocol, endpointURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create log exporter: %w", err)
	}
	cleanupFuncs = append(cleanupFuncs, logCleanup)

	loggerProvider := log.NewLoggerProvider()
	cleanupFuncs = append(cleanupFuncs, loggerProvider.Shutdown)

	// Set global providers
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	global.SetLoggerProvider(loggerProvider)

	// Set up propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Set up slog bridge
	logger := otelslog.NewLogger("sentrie")
	slog.SetDefault(logger)

	// Set up runtime metrics collector
	meter := meterProvider.Meter("sentrie/runtime")
	if err := setupRuntimeMetrics(ctx, meter); err != nil {
		return nil, fmt.Errorf("failed to setup runtime metrics: %w", err)
	}

	// Return combined cleanup function
	return func(ctx context.Context) error {
		var allErr error
		for _, cleanup := range cleanupFuncs {
			if err := cleanup(ctx); err != nil {
				allErr = errors.Join(allErr, err)
			}
		}
		return allErr
	}, nil
}

// createTraceExporter creates a trace exporter based on protocol
func createTraceExporter(ctx context.Context, protocol string, endpointURL *url.URL) (trace.SpanExporter, func(context.Context) error, error) {
	switch protocol {
	case "grpc":
		exporter, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpointURL.Host),
			otlptracegrpc.WithInsecure(),
		)
		return exporter, exporter.Shutdown, err
	case "http":
		// For HTTP, use WithEndpointURL to specify the full URL
		endpoint := fmt.Sprintf("%s://%s", endpointURL.Scheme, endpointURL.Host)
		exporter, err := otlptracehttp.New(ctx,
			otlptracehttp.WithEndpointURL(endpoint),
		)
		return exporter, exporter.Shutdown, err
	default:
		return nil, nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// createMetricExporter creates a metric exporter based on protocol
func createMetricExporter(ctx context.Context, protocol string, endpointURL *url.URL) (sdkmetric.Exporter, func(context.Context) error, error) {
	switch protocol {
	case "grpc":
		exporter, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(endpointURL.Host),
			otlpmetricgrpc.WithInsecure(),
		)
		return exporter, exporter.Shutdown, err
	case "http":
		// For HTTP, use WithEndpointURL to specify the full URL
		endpoint := fmt.Sprintf("%s://%s", endpointURL.Scheme, endpointURL.Host)
		exporter, err := otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpointURL(endpoint),
		)
		return exporter, exporter.Shutdown, err
	default:
		return nil, nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// createLogExporter creates a log exporter based on protocol
func createLogExporter(ctx context.Context, protocol string, endpointURL *url.URL) (log.Exporter, func(context.Context) error, error) {
	switch protocol {
	case "grpc":
		exporter, err := otlploggrpc.New(ctx,
			otlploggrpc.WithEndpoint(endpointURL.Host),
			otlploggrpc.WithInsecure(),
		)
		return exporter, exporter.Shutdown, err
	case "http":
		// For HTTP, use WithEndpointURL to specify the full URL
		endpoint := fmt.Sprintf("%s://%s", endpointURL.Scheme, endpointURL.Host)
		exporter, err := otlploghttp.New(ctx,
			otlploghttp.WithEndpointURL(endpoint),
		)
		return exporter, exporter.Shutdown, err
	default:
		return nil, nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

// setupRuntimeMetrics sets up runtime metrics collection
func setupRuntimeMetrics(ctx context.Context, meter metric.Meter) error {
	// Define metric names to collect
	metricNames := []string{
		"memory_classes_heap_objects_bytes",
		"memory_classes_total_bytes",
		"gc_cycles_total_gc_cycles",
		"gc_heap_goal_bytes",
		"gc_pauses_seconds",
		"sched_goroutines_goroutines",
		"cpu_classes_total_cpu_seconds",
	}

	// Map OpenTelemetry metric names to runtime metric names
	runtimeMetricMap := map[string]string{
		"memory_classes_heap_objects_bytes": "/memory/classes/heap/objects:bytes",
		"memory_classes_total_bytes":        "/memory/classes/total:bytes",
		"gc_cycles_total_gc_cycles":         "/gc/cycles/total:gc-cycles",
		"gc_heap_goal_bytes":                "/gc/heap/goal:bytes",
		"gc_pauses_seconds":                 "/gc/pauses:seconds",
		"sched_goroutines_goroutines":       "/sched/goroutines:goroutines",
		"cpu_classes_total_cpu_seconds":     "/cpu/classes/total:cpu-seconds",
	}

	// Create gauges for each metric
	gauges := make(map[string]metric.Int64Gauge)
	for _, name := range metricNames {
		gauge, err := meter.Int64Gauge(name)
		if err != nil {
			return fmt.Errorf("failed to create gauge for %s: %w", name, err)
		}
		gauges[name] = gauge
	}

	// Start a goroutine to periodically collect and report metrics
	// this approach is better than the built in Observer pattern in the OTel lib,
	// since we get to record multiple values in one cycle
	// and stay idle for the rest of the time
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Read all metrics
				descriptions := metrics.All()
				samples := make([]metrics.Sample, len(descriptions))
				for i, desc := range descriptions {
					samples[i].Name = desc.Name
				}
				metrics.Read(samples)

				// Update gauges for metrics we care about
				for _, sample := range samples {
					// Find the OpenTelemetry metric name for this runtime metric
					for otelName, runtimeName := range runtimeMetricMap {
						if sample.Name == runtimeName {
							if gauge, exists := gauges[otelName]; exists {
								switch sample.Value.Kind() {
								case metrics.KindUint64:
									gauge.Record(ctx, int64(sample.Value.Uint64()))
								case metrics.KindFloat64:
									gauge.Record(ctx, int64(sample.Value.Float64()))
								}
							}
							break
						}
					}
				}
			}
		}
	}()

	return nil
}
