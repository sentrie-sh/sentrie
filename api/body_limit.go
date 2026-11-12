// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// DefaultMaxBodyBytes is the largest request body the API reads unless
// WithMaxBodyBytes says otherwise (#115): 1 MiB.
const DefaultMaxBodyBytes int64 = 1 << 20

// WithMaxBodyBytes sets the largest request body, in bytes, that any handler
// reads (#115; serve passes --http-max-body-bytes). A larger body is answered
// with 413. Zero means no limit; a negative value is treated as zero.
func WithMaxBodyBytes(n int64) HTTPAPIOption {
	return func(api *HTTPAPI) {
		api.maxBodyBytes = max(n, 0)
	}
}

// limitBody wraps every request body in http.MaxBytesReader, so no handler
// behind it can read more than maxBodyBytes. It sits in front of the whole
// mux, which covers every route, including ones added later. It only limits
// the reader; the handler that reads the body turns the error into a 413 with
// bodyTooLarge and writeBodyTooLarge.
func (api *HTTPAPI) limitBody(next http.Handler) http.Handler {
	limit := api.maxBodyBytes
	if limit <= 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, limit)
		next.ServeHTTP(w, r)
	})
}

// bodyOverLimit reports whether the request declares, before any of it is
// read, a body larger than the limit. A chunked body (ContentLength -1) is
// caught by the reader instead.
func (api *HTTPAPI) bodyOverLimit(r *http.Request) bool {
	return api.maxBodyBytes > 0 && r.ContentLength > api.maxBodyBytes
}

// bodyTooLarge reports whether err comes from reading past the body limit.
func bodyTooLarge(err error) bool {
	var maxErr *http.MaxBytesError
	return errors.As(err, &maxErr)
}

// drainBody reads what is left of a body after its JSON value was decoded, so
// a body whose first value fits but whose total size does not is still a 413.
// Data after the value is otherwise ignored, as before. With no limit nothing
// is read, since the rest could be unbounded.
func (api *HTTPAPI) drainBody(r *http.Request) error {
	if api.maxBodyBytes <= 0 {
		return nil
	}
	_, err := io.Copy(io.Discard, r.Body)
	return err
}

// writeBodyTooLarge answers 413 with a problem document naming the limit.
func (api *HTTPAPI) writeBodyTooLarge(w http.ResponseWriter, r *http.Request) {
	api.writeErrorResponse(w, r, http.StatusRequestEntityTooLarge, "Request Body Too Large",
		fmt.Sprintf("The request body is larger than this server's limit of %d bytes", api.maxBodyBytes))
}
