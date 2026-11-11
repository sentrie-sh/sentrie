// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
)

const bodyLimitPath = "/decision/sh/sentrie/example/user_access"

// bodyLimitAPI is the example pack API with its own body limit.
func (s *APITestSuite) bodyLimitAPI(limit int64) *HTTPAPI {
	base := s.examplePackHTTPAPI()
	return NewHTTPAPI(base.executor, WithMaxBodyBytes(limit))
}

// paddedFacts is a valid decision body of at least n bytes.
func paddedFacts(n int) string {
	return `{"facts":{"user":{"role":"admin","status":"active"},"pad":"` + strings.Repeat("x", n) + `"}}`
}

func (s *APITestSuite) assertBodyTooLarge(rec *httptest.ResponseRecorder) {
	s.Equal(http.StatusRequestEntityTooLarge, rec.Code)
	s.Equal("application/problem+json", rec.Header().Get("Content-Type"))
	var p map[string]any
	s.Require().NoError(json.Unmarshal(rec.Body.Bytes(), &p))
	s.Equal(float64(http.StatusRequestEntityTooLarge), p["status"])
	s.Equal("Request Body Too Large", p["title"])
}

// TestBodyLimitDefaultIsOneMiB: an API built without the option caps bodies
// at 1 MiB (#115).
func (s *APITestSuite) TestBodyLimitDefaultIsOneMiB() {
	s.Equal(int64(1<<20), NewHTTPAPI(nil).maxBodyBytes)
}

// TestBodyLimitDeclaredLengthOver: a Content-Length above the limit is a 413.
func (s *APITestSuite) TestBodyLimitDeclaredLengthOver() {
	rec := s.postDecision(s.bodyLimitAPI(64), bodyLimitPath, paddedFacts(100))
	s.assertBodyTooLarge(rec)
}

// TestBodyLimitChunkedOver: a body with no Content-Length is cut off by the
// reader and is a 413, not a 400.
func (s *APITestSuite) TestBodyLimitChunkedOver() {
	req := httptest.NewRequest(http.MethodPost, bodyLimitPath, io.MultiReader(strings.NewReader(paddedFacts(100))))
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	s.decisionHandler(s.bodyLimitAPI(64)).ServeHTTP(rec, req)
	s.assertBodyTooLarge(rec)
}

// TestBodyLimitTrailingDataOver: a first JSON value under the limit followed
// by enough data to exceed it is still a 413.
func (s *APITestSuite) TestBodyLimitTrailingDataOver() {
	body := paddedFacts(0) + strings.Repeat(" ", 200)
	req := httptest.NewRequest(http.MethodPost, bodyLimitPath, io.MultiReader(strings.NewReader(body)))
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	s.decisionHandler(s.bodyLimitAPI(int64(len(paddedFacts(0))+10))).ServeHTTP(rec, req)
	s.assertBodyTooLarge(rec)
}

// TestBodyLimitAtLimitAccepted: a body exactly at the limit is evaluated.
func (s *APITestSuite) TestBodyLimitAtLimitAccepted() {
	body := paddedFacts(100)
	rec := s.postDecision(s.bodyLimitAPI(int64(len(body))), bodyLimitPath, body)
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
}

// TestBodyLimitZeroIsUnlimited: 0 disables the limit.
func (s *APITestSuite) TestBodyLimitZeroIsUnlimited() {
	rec := s.postDecision(s.bodyLimitAPI(0), bodyLimitPath, paddedFacts(2<<20))
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
}

// TestBodyLimitTrailingDataWithinLimitIgnored: PROVISIONAL (#115). Data after
// the first JSON value is still ignored when the whole body fits the limit,
// as it was before the limit existed.
func (s *APITestSuite) TestBodyLimitTrailingDataWithinLimitIgnored() {
	body := paddedFacts(0) + ` trailing garbage`
	rec := s.postDecision(s.bodyLimitAPI(int64(len(body))+100), bodyLimitPath, body)
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
}

// TestBodyLimitOneByteOverDeclared: the boundary is exact on a declared
// length: limit+1 is a 413 where limit is accepted.
func (s *APITestSuite) TestBodyLimitOneByteOverDeclared() {
	body := paddedFacts(100)
	rec := s.postDecision(s.bodyLimitAPI(int64(len(body))-1), bodyLimitPath, body)
	s.assertBodyTooLarge(rec)
}

// TestBodyLimitChunkedAtLimitAccepted: a chunked body exactly at the limit
// passes the reader and the drain.
func (s *APITestSuite) TestBodyLimitChunkedAtLimitAccepted() {
	body := paddedFacts(100)
	req := httptest.NewRequest(http.MethodPost, bodyLimitPath, io.MultiReader(strings.NewReader(body)))
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	s.decisionHandler(s.bodyLimitAPI(int64(len(body)))).ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code, rec.Body.String())
}

// TestBodyLimitNegativeOptionIsUnlimited: WithMaxBodyBytes clamps a negative
// value to 0 (no limit).
func (s *APITestSuite) TestBodyLimitNegativeOptionIsUnlimited() {
	s.Equal(int64(0), NewHTTPAPI(nil, WithMaxBodyBytes(-5)).maxBodyBytes)
}

// TestBodyLimitBodyTooLargeIsNot400: an oversized unparseable body is a 413,
// not the 400 Invalid JSON a truncated read used to give.
func (s *APITestSuite) TestBodyLimitBodyTooLargeIsNot400() {
	rec := s.postDecision(s.bodyLimitAPI(16), bodyLimitPath, `{"facts":`+strings.Repeat("[", 200))
	s.assertBodyTooLarge(rec)
}
