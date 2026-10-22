// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

func (s *APITestSuite) TestParseTimeoutHeader() {
	cases := []struct {
		name    string
		values  []string
		want    time.Duration
		present bool
		errPart string
	}{
		{name: "absent", values: nil},
		{name: "milliseconds", values: []string{"250ms"}, want: 250 * time.Millisecond, present: true},
		{name: "seconds", values: []string{"2s"}, want: 2 * time.Second, present: true},
		{name: "compound", values: []string{"1m30s"}, want: 90 * time.Second, present: true},
		{name: "fraction", values: []string{"1.5s"}, want: 1500 * time.Millisecond, present: true},
		{name: "largest duration", values: []string{"2562047h47m16.854775807s"}, want: time.Duration(math.MaxInt64), present: true},
		{name: "empty", values: []string{""}, present: true, errPart: "is not a Go duration"},
		{name: "zero", values: []string{"0"}, present: true, errPart: "must be positive"},
		{name: "zero with unit", values: []string{"0s"}, present: true, errPart: "must be positive"},
		{name: "negative", values: []string{"-1s"}, present: true, errPart: "must be positive"},
		{name: "no unit", values: []string{"5"}, present: true, errPart: "is not a Go duration"},
		{name: "garbage", values: []string{"soon"}, present: true, errPart: "is not a Go duration"},
		{name: "RFC 3339 duration", values: []string{"PT2S"}, present: true, errPart: "is not a Go duration"},
		{name: "list in one value", values: []string{"1s, 2s"}, present: true, errPart: "is not a Go duration"},
		{name: "overflow", values: []string{"2562048h"}, present: true, errPart: "is not a Go duration"},
		{name: "overflow in nanoseconds", values: []string{"9223372036854775808ns"}, present: true, errPart: "is not a Go duration"},
		{name: "too long", values: []string{strings.Repeat("1", maxTimeoutHeaderLen) + "s"}, present: true, errPart: "longer than 32 bytes"},
		{name: "repeated", values: []string{"1s", "2s"}, present: true, errPart: "at most once"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			d, present, err := parseTimeoutHeader(tc.values)
			s.Equal(tc.present, present)
			if tc.errPart != "" {
				s.Require().ErrorContains(err, tc.errPart)
				s.Zero(d)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, d)
		})
	}
}

// TestTimeoutHeaderValueIsNotEchoedWhenTooLong: an over-long value is
// rejected before it is parsed or quoted back.
func (s *APITestSuite) TestTimeoutHeaderValueIsNotEchoedWhenTooLong() {
	long := strings.Repeat("9", 4096) + "s"
	_, _, err := parseTimeoutHeader([]string{long})
	s.Require().Error(err)
	s.NotContains(err.Error(), "9999")
}

// TestRequestDeadlineCombinations covers every row of #243's flag table, with
// and without the header.
func (s *APITestSuite) TestRequestDeadlineCombinations() {
	const d, m = 10 * time.Second, time.Minute
	cases := []struct {
		name     string
		def, max time.Duration
		header   string
		want     time.Duration
		errPart  string
	}{
		{name: "D, no ceiling, no header", def: d, want: d},
		{name: "D, no ceiling, shorter", def: d, header: "2s", want: 2 * time.Second},
		{name: "D, no ceiling, equal", def: d, header: "10s", want: d},
		{name: "D, no ceiling, longer is rejected", def: d, header: "11s", errPart: "above this server's maximum of 10s"},
		{name: "D, ceiling M, no header", def: d, max: m, want: d},
		{name: "D, ceiling M, shorter than D", def: d, max: m, header: "250ms", want: 250 * time.Millisecond},
		{name: "D, ceiling M, between D and M", def: d, max: m, header: "30s", want: 30 * time.Second},
		{name: "D, ceiling M, equal to M", def: d, max: m, header: "1m", want: m},
		{name: "D, ceiling M, above M is rejected", def: d, max: m, header: "1m0.001s", errPart: "above this server's maximum of 1m0s"},
		{name: "unbounded, no ceiling, no header", def: 0, want: 0},
		{name: "unbounded, no ceiling, any value", def: 0, header: "100h", want: 100 * time.Hour},
		{name: "unbounded, ceiling M, no header runs under M", def: 0, max: m, want: m},
		{name: "unbounded, ceiling M, shorter", def: 0, max: m, header: "5s", want: 5 * time.Second},
		{name: "unbounded, ceiling M, above M is rejected", def: 0, max: m, header: "2m", errPart: "above this server's maximum of 1m0s"},
		{name: "bad value is never a fallback", def: d, max: m, header: "soon", errPart: "is not a Go duration"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			api := NewHTTPAPI(nil, WithRequestTimeout(tc.def), WithMaxRequestTimeout(tc.max))
			h := http.Header{}
			if tc.header != "" {
				h.Set(TimeoutHeader, tc.header)
			}
			got, err := api.requestDeadline(h)
			if tc.errPart != "" {
				s.Require().ErrorContains(err, tc.errPart)
				return
			}
			s.Require().NoError(err)
			s.Equal(tc.want, got)
		})
	}
}

func (s *APITestSuite) TestWriteTimeoutCoversTheCeiling() {
	s.Equal(30*time.Second, NewHTTPAPI(nil, WithRequestTimeout(10*time.Second), WithMaxRequestTimeout(20*time.Second)).writeTimeout())
	s.Equal(125*time.Second, NewHTTPAPI(nil, WithRequestTimeout(10*time.Second), WithMaxRequestTimeout(2*time.Minute)).writeTimeout(), "sized from the ceiling, not the default")
	s.Equal(65*time.Second, NewHTTPAPI(nil, WithRequestTimeout(0), WithMaxRequestTimeout(time.Minute)).writeTimeout())
	s.Equal(30*time.Second, NewHTTPAPI(nil, WithRequestTimeout(0)).writeTimeout(), "no ceiling keeps the 30s write cap")
	s.Equal(time.Duration(math.MaxInt64), NewHTTPAPI(nil, WithMaxRequestTimeout(time.Duration(math.MaxInt64))).writeTimeout(), "the margin saturates instead of overflowing")
}

// postDecisionWithTimeout posts to the looping pack with Sentrie-Timeout set
// (unless timeout is empty) and fails the test if no response arrives in 5s.
func (s *APITestSuite) postDecisionWithTimeout(api *HTTPAPI, timeout string) *httptest.ResponseRecorder {
	done := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		req := httptest.NewRequest(http.MethodPost, "/decision/loop/test/p/r", strings.NewReader(`{"facts": {}}`))
		req.Header.Set("Content-Type", "application/json")
		if timeout != "" {
			req.Header.Set(TimeoutHeader, timeout)
		}
		rec := httptest.NewRecorder()
		s.decisionHandler(api).ServeHTTP(rec, req)
		done <- rec
	}()
	select {
	case rec := <-done:
		return rec
	case <-time.After(5 * time.Second):
		s.FailNow("decision request did not return")
		return nil
	}
}

func (s *APITestSuite) TestTimeoutHeaderShortensTheDeadline() {
	api := s.loopingPackAPIWith(WithRequestTimeout(5 * time.Second))
	start := time.Now()
	rec := s.postDecisionWithTimeout(api, "100ms")
	s.Less(time.Since(start), 2*time.Second)
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	s.Equal("100ms", rec.Header().Get(TimeoutAppliedHeader))
	s.Contains(rec.Body.String(), "evaluation exceeded its 100ms deadline")
}

func (s *APITestSuite) TestTimeoutHeaderLengthensTheDeadlineWithinTheCeiling() {
	api := s.loopingPackAPIWith(WithRequestTimeout(50*time.Millisecond), WithMaxRequestTimeout(time.Second))
	start := time.Now()
	rec := s.postDecisionWithTimeout(api, "300ms")
	s.GreaterOrEqual(time.Since(start), 300*time.Millisecond, "the 50ms default did not apply")
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	s.Equal("300ms", rec.Header().Get(TimeoutAppliedHeader))
	s.Contains(rec.Body.String(), "evaluation exceeded its 300ms deadline")
}

func (s *APITestSuite) TestTimeoutHeaderAboveTheCeilingIsRejected() {
	api := s.loopingPackAPIWith(WithRequestTimeout(50*time.Millisecond), WithMaxRequestTimeout(time.Second))
	rec := s.postDecisionWithTimeout(api, "2s")
	s.Equal(http.StatusBadRequest, rec.Code)
	s.Equal("application/problem+json", rec.Header().Get("Content-Type"))
	s.Empty(rec.Header().Get(TimeoutAppliedHeader), "nothing was evaluated")
	s.Contains(rec.Body.String(), "above this server's maximum of 1s")
}

func (s *APITestSuite) TestInvalidTimeoutHeaderIsRejected() {
	api := s.loopingPackAPIWith(WithRequestTimeout(50 * time.Millisecond))
	for _, v := range []string{"soon", "0", "-1s", "PT1S", strings.Repeat("1", 64)} {
		rec := s.postDecisionWithTimeout(api, v)
		s.Equal(http.StatusBadRequest, rec.Code, v)
		s.Equal("application/problem+json", rec.Header().Get("Content-Type"), v)
		s.Contains(rec.Body.String(), "Invalid "+TimeoutHeader, v)
		s.Empty(rec.Header().Get(TimeoutAppliedHeader), v)
	}
}

func (s *APITestSuite) TestAppliedHeaderEchoesTheDefault() {
	api := s.loopingPackAPIWith(WithRequestTimeout(100 * time.Millisecond))
	rec := s.postDecisionWithTimeout(api, "")
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	s.Equal("100ms", rec.Header().Get(TimeoutAppliedHeader))
}

// TestUnboundedDefaultRunsUnderTheCeiling: --eval-timeout 0 with
// --eval-max-timeout M gives a request without the header the deadline M.
func (s *APITestSuite) TestUnboundedDefaultRunsUnderTheCeiling() {
	api := s.loopingPackAPIWith(WithRequestTimeout(0), WithMaxRequestTimeout(100*time.Millisecond))
	rec := s.postDecisionWithTimeout(api, "")
	s.Equal(http.StatusServiceUnavailable, rec.Code)
	s.Equal("100ms", rec.Header().Get(TimeoutAppliedHeader))
	s.Contains(rec.Body.String(), "evaluation exceeded its 100ms deadline")
}

func (s *APITestSuite) TestTimeoutHeadersAreAllowedAndExposedForCORS() {
	api := s.loopingPackAPIWith(WithRequestTimeout(50 * time.Millisecond))
	rec := s.postDecisionWithTimeout(api, "")
	s.Contains(rec.Header().Get("Access-Control-Allow-Headers"), TimeoutHeader)
	s.Equal(TimeoutAppliedHeader, rec.Header().Get("Access-Control-Expose-Headers"))
}
