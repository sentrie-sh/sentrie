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

package middleware

import (
	"net/http"

	"github.com/sentrie-sh/sentrie/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"go.opentelemetry.io/otel/trace"
)

func OTelMiddleware(cfg *otel.OTelConfig, tracer trace.Tracer, meter metric.Meter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !cfg.Enabled {
			next.ServeHTTP(w, r)
			return
		}

		ctx := r.Context()
		ctx, span := tracer.Start(ctx,
			"http.request",
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.NetworkProtocolName(r.URL.Scheme)))
		defer span.End()

		r = ensureRequestIDInRequest(r)

		requestID := GetRequestIDFromRequest(r)
		span.SetAttributes(attribute.String("http.request.id", requestID))

		rww := &rWW{ResponseWriter: w}
		rqq := &rQQ{Request: *r}

		next.ServeHTTP(rww, rqq.WithContext(ctx)) // Pass the span context to the next handler

		span.SetAttributes(semconv.HTTPRequestBodySize(int(rqq.bytesRead)))
		span.SetAttributes(semconv.HTTPResponseBodySize(int(rww.bytesWritten)))
		span.SetAttributes(semconv.HTTPResponseStatusCode(rww.statusCode))
	})
}

type rQQ struct {
	http.Request
	bytesRead int64
}

func (r *rQQ) Read(p []byte) (n int, err error) {
	n, err = r.Request.Body.Read(p)
	r.bytesRead += int64(n)
	return n, err
}

type rWW struct {
	http.ResponseWriter
	bytesWritten int64
	statusCode   int
}

func (w *rWW) Write(p []byte) (n int, err error) {
	n, err = w.ResponseWriter.Write(p)
	w.bytesWritten += int64(n)
	return n, err
}

func (w *rWW) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
