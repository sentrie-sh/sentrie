// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

const (
	// TimeoutHeader lets a decision request choose its own evaluation deadline
	// (#243), as a Go duration such as 250ms, 2s or 1m30s. A value above the
	// API's ceiling (WithMaxRequestTimeout) is rejected with 400.
	TimeoutHeader = "Sentrie-Timeout"

	// TimeoutAppliedHeader reports the deadline a decision request was
	// evaluated under, as time.Duration.String(). It is omitted when no
	// deadline applied.
	TimeoutAppliedHeader = "Sentrie-Timeout-Applied"

	// maxTimeoutHeaderLen bounds a Sentrie-Timeout value before it is parsed.
	// The longest useful Go duration ("2562047h47m16.854775807s") is 24 bytes.
	maxTimeoutHeaderLen = 32
)

var errTimeoutHeaderRepeated = errors.New(TimeoutHeader + " must be sent at most once")

// parseTimeoutHeader parses the Sentrie-Timeout values of a request. ok is
// false when the header is absent. The value is untrusted: it is length-checked
// before parsing, time.ParseDuration rejects anything that would overflow a
// time.Duration, and only a positive duration is accepted. Error messages quote
// the value only once it is known to be short.
func parseTimeoutHeader(values []string) (d time.Duration, ok bool, err error) {
	switch len(values) {
	case 0:
		return 0, false, nil
	case 1:
	default:
		return 0, true, errTimeoutHeaderRepeated
	}
	v := values[0]
	if len(v) > maxTimeoutHeaderLen {
		return 0, true, fmt.Errorf("%s is longer than %d bytes", TimeoutHeader, maxTimeoutHeaderLen)
	}
	d, err = time.ParseDuration(v)
	if err != nil {
		return 0, true, fmt.Errorf("%s %q is not a Go duration such as 250ms, 2s or 1m30s", TimeoutHeader, v)
	}
	if d <= 0 {
		return 0, true, fmt.Errorf("%s %q must be positive", TimeoutHeader, v)
	}
	return d, true, nil
}

// maxTimeout is the longest deadline a request may ask for: the ceiling when
// one is set, otherwise the default. Zero means unbounded.
func (api *HTTPAPI) maxTimeout() time.Duration {
	if api.maxRequestTimeout > 0 {
		return api.maxRequestTimeout
	}
	return api.requestTimeout
}

// defaultTimeout is the deadline of a request that does not send
// Sentrie-Timeout: the default, capped to the ceiling when one is set. An
// unbounded default with a ceiling therefore runs under the ceiling. Zero
// means no deadline.
func (api *HTTPAPI) defaultTimeout() time.Duration {
	d := api.requestTimeout
	if ceiling := api.maxRequestTimeout; ceiling > 0 && (d <= 0 || d > ceiling) {
		return ceiling
	}
	return d
}

// requestDeadline picks the evaluation deadline for one decision request: the
// default, or the Sentrie-Timeout value when it is within maxTimeout. A value
// above it is an error, never clamped. Zero means no deadline. The header
// chooses the deadline and nothing else.
func (api *HTTPAPI) requestDeadline(h http.Header) (time.Duration, error) {
	requested, ok, err := parseTimeoutHeader(h.Values(TimeoutHeader))
	if err != nil {
		return 0, err
	}
	if !ok {
		return api.defaultTimeout(), nil
	}
	if ceiling := api.maxTimeout(); ceiling > 0 && requested > ceiling {
		return 0, fmt.Errorf("%s %s is above this server's maximum of %s", TimeoutHeader, requested, ceiling)
	}
	return requested, nil
}
