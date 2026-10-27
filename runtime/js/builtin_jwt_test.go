// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto"
	"crypto/hmac"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// jwtIOExample is the HS256 example token published on jwt.io, signed with
// the secret "your-256-bit-secret". It has no exp or nbf claim.
const jwtIOExample = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
	"eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ." +
	"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

const jwtTestSecret = "test-secret"

func b64url(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

// jwtUnsigned encodes header and claims as the first two token segments.
func (s *JSTestSuite) jwtUnsigned(header, claims map[string]any) string {
	h, err := json.Marshal(header)
	s.Require().NoError(err)
	c, err := json.Marshal(claims)
	s.Require().NoError(err)
	return b64url(h) + "." + b64url(c)
}

// jwtHS signs claims with an HMAC algorithm and secret.
func (s *JSTestSuite) jwtHS(alg, secret string, claims map[string]any) string {
	hashes := map[string]crypto.Hash{"HS256": crypto.SHA256, "HS384": crypto.SHA384, "HS512": crypto.SHA512}
	input := s.jwtUnsigned(map[string]any{"alg": alg, "typ": "JWT"}, claims)
	mac := hmac.New(hashes[alg].New, []byte(secret))
	mac.Write([]byte(input))
	return input + "." + b64url(mac.Sum(nil))
}

// jwtEval runs expr against @sentrie/jwt bound as `m`, with `token` set to
// tok, and returns its JSON form or "threw:<message>".
func (s *JSTestSuite) jwtEval(tok, expr string) string {
	vm := goja.New()
	ex, err := BuiltinJwtGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("m", ex))
	s.Require().NoError(vm.Set("token", tok))
	SetExecutionStart(vm, time.Unix(testExecutionStartUnix, 0))
	got, err := vm.RunString(`(function () {
		try { return JSON.stringify(` + expr + `); } catch (e) { return "threw:" + e.message; }
	})()`)
	s.Require().NoError(err)
	return got.String()
}

func (s *JSTestSuite) TestJwtKnownAnswer() {
	s.Equal("true", s.jwtEval(jwtIOExample, `m.verify(token, "your-256-bit-secret")`))
	s.Equal(`"John Doe"`, s.jwtEval(jwtIOExample, `m.decode(token, "your-256-bit-secret").name`))
	s.Equal("false", s.jwtEval(jwtIOExample, `m.verify(token, "not-the-secret")`))
	// getHeader returns a Go map, so its JSON key order varies from run to run.
	s.JSONEq(`{"alg":"HS256","typ":"JWT"}`, s.jwtEval(jwtIOExample, `m.getHeader(token)`))
}

func (s *JSTestSuite) TestJwtTimeClaims() {
	now := testExecutionStartUnix
	cases := []struct {
		name   string
		claims map[string]any
		expr   string
		want   string
	}{
		{"future exp verifies", map[string]any{"exp": now + 60}, `m.verify(token, s)`, "true"},
		{"expired does not verify", map[string]any{"exp": now - 10}, `m.verify(token, s)`, "false"},
		{"exp equal to now is expired", map[string]any{"exp": now}, `m.verify(token, s)`, "false"},
		{"expired decode throws", map[string]any{"exp": now - 10}, `m.decode(token, s)`, "threw:token is expired"},
		{"leeway covers skew", map[string]any{"exp": now - 10}, `m.verify(token, s, "HS256", { leeway: 30 })`, "true"},
		{"leeway too small", map[string]any{"exp": now - 10}, `m.verify(token, s, "HS256", { leeway: 5 })`, "false"},
		{"decode leeway", map[string]any{"exp": now - 10}, `m.decode(token, s, { leeway: 30 }).exp`, "1699999990"},
		{"not yet valid", map[string]any{"nbf": now + 10}, `m.verify(token, s)`, "false"},
		{"not yet valid decode throws", map[string]any{"nbf": now + 10}, `m.decode(token, s)`, "threw:token is not valid yet"},
		{"nbf equal to now is valid", map[string]any{"nbf": now}, `m.verify(token, s)`, "true"},
		{"nbf within leeway", map[string]any{"nbf": now + 10}, `m.verify(token, s, undefined, { leeway: 10 })`, "true"},
		{"inspect-only opt-out", map[string]any{"exp": now - 10}, `m.decode(token, s, { checkTime: false }).exp`, "1699999990"},
		{"verify opt-out", map[string]any{"exp": now - 10, "nbf": now + 10}, `m.verify(token, s, "HS256", { checkTime: false })`, "true"},
		{"non-numeric exp does not verify", map[string]any{"exp": "tomorrow"}, `m.verify(token, s)`, "false"},
		{"non-numeric exp decode throws", map[string]any{"exp": "tomorrow"}, `m.decode(token, s)`, "threw:exp claim must be a number"},
		{"fractional exp just ahead", map[string]any{"exp": float64(now) + 0.5}, `m.verify(token, s)`, "true"},
		{"fractional exp just behind", map[string]any{"exp": float64(now) - 0.5}, `m.verify(token, s)`, "false"},
		{"fractional leeway covers fractional skew", map[string]any{"exp": float64(now) - 0.5}, `m.verify(token, s, "HS256", { leeway: 0.75 })`, "true"},
		{"fractional leeway too small", map[string]any{"exp": float64(now) - 0.5}, `m.verify(token, s, "HS256", { leeway: 0.25 })`, "false"},
		{"leeway at the cap", map[string]any{"exp": now - 86_000}, `m.verify(token, s, "HS256", { leeway: 86400 })`, "true"},
		{"leeway over the cap throws", map[string]any{"exp": now - 10}, `m.verify(token, s, "HS256", { leeway: 86401 })`, "threw:options.leeway must be between 0 and 86400 seconds"},
		{"leeway in milliseconds throws", map[string]any{"exp": now - 10}, `m.decode(token, s, { leeway: 3600000 })`, "threw:options.leeway must be between 0 and 86400 seconds"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			tok := s.jwtHS("HS256", jwtTestSecret, tc.claims)
			s.Equal(tc.want, s.jwtEval(tok, `(function (s) { return `+tc.expr+`; })("`+jwtTestSecret+`")`))
		})
	}
}

func (s *JSTestSuite) TestJwtSignatureChecks() {
	valid := s.jwtHS("HS256", jwtTestSecret, map[string]any{"sub": "alice"})
	parts := strings.Split(valid, ".")

	// Flip one character of the signature.
	sig := []byte(parts[2])
	if sig[0] == 'A' {
		sig[0] = 'B'
	} else {
		sig[0] = 'A'
	}
	tamperedSig := parts[0] + "." + parts[1] + "." + string(sig)

	// Swap the payload, keeping the original signature.
	forged := strings.Split(s.jwtHS("HS256", jwtTestSecret, map[string]any{"sub": "mallory"}), ".")
	tamperedPayload := parts[0] + "." + forged[1] + "." + parts[2]

	none := s.jwtUnsigned(map[string]any{"alg": "none"}, map[string]any{"sub": "alice"}) + "."
	forgedSig := s.jwtUnsigned(map[string]any{"alg": "HS256"}, map[string]any{"sub": "admin"}) + ".AAAA"

	cases := []struct {
		name string
		tok  string
		expr string
		want string
	}{
		{"HS384 round trip", s.jwtHS("HS384", jwtTestSecret, map[string]any{}), `m.verify(token, "test-secret", "HS384")`, "true"},
		{"HS512 round trip", s.jwtHS("HS512", jwtTestSecret, map[string]any{}), `m.verify(token, "test-secret", "HS512")`, "true"},
		{"tampered signature", tamperedSig, `m.verify(token, "test-secret")`, "false"},
		{"tampered signature decode", tamperedSig, `m.decode(token, "test-secret")`, "threw:token signature is invalid"},
		{"tampered payload", tamperedPayload, `m.verify(token, "test-secret")`, "false"},
		{"wrong secret", valid, `m.verify(token, "other-secret")`, "false"},
		{"wrong alg", s.jwtHS("HS384", jwtTestSecret, map[string]any{}), `m.verify(token, "test-secret", "HS256")`, "false"},
		{"none token does not verify", none, `m.verify(token, "test-secret")`, "false"},
		{"none token decode with key throws", none, `m.decode(token, "test-secret")`, `threw:algorithm "none" is not allowed`},
		{"none as expected alg throws", valid, `m.verify(token, "test-secret", "none")`, `threw:algorithm "none" is not allowed`},
		{"malformed token does not verify", "a.b", `m.verify(token, "test-secret")`, "false"},
		{"empty secret throws", valid, `m.verify(token, "")`, "threw:secret must not be empty"},
		{"empty secret decode throws", valid, `m.decode(token, "")`, "threw:secret must not be empty"},
		{"non-string secret throws", valid, `m.verify(token, 42)`, "threw:secret must be a string"},
		{"unknown option throws", valid, `m.verify(token, "test-secret", "HS256", { checktime: false })`, `threw:unknown option "checktime"`},
		{"negative leeway throws", valid, `m.decode(token, "test-secret", { leeway: -1 })`, "threw:options.leeway must be between 0 and 86400 seconds"},
		{"decode without a key throws", valid, `m.decode(token)`, "threw:decode requires 2 or 3 arguments; use getPayload or getHeader to read a token without verifying it"},
		{"forged token with undefined key throws", forgedSig, `m.decode(token, undefined)`, "threw:decode requires a key, got undefined; use getPayload or getHeader to read a token without verifying it"},
		{"forged token with null key throws", forgedSig, `m.decode(token, null)`, "threw:decode requires a key, got null; use getPayload or getHeader to read a token without verifying it"},
		{"none token with undefined key throws", none, `m.decode(token, undefined, { checkTime: false })`, "threw:decode requires a key, got undefined; use getPayload or getHeader to read a token without verifying it"},
		{"getPayload still inspects", forgedSig, `m.getPayload(token).sub`, `"admin"`},
		{"non-object options throw", valid, `m.decode(token, "test-secret", "yes")`, "threw:options must be an object"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			s.Equal(tc.want, s.jwtEval(tc.tok, tc.expr))
		})
	}
}

// TestJwtRequiresExecutionStart checks that a VM without the pinned clock
// fails rather than falling back to the wall clock, and that the
// inspect-only opt-out does not need the clock at all.
func (s *JSTestSuite) TestJwtRequiresExecutionStart() {
	tok := s.jwtHS("HS256", jwtTestSecret, map[string]any{"sub": "alice"})
	vm := goja.New()
	ex, err := BuiltinJwtGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("m", ex))
	s.Require().NoError(vm.Set("token", tok))

	_, err = vm.RunString(`m.verify(token, "test-secret")`)
	s.Require().Error(err)
	s.Contains(err.Error(), "execution start time is not set")

	got, err := vm.RunString(`m.verify(token, "test-secret", "HS256", { checkTime: false })`)
	s.Require().NoError(err)
	s.True(got.ToBoolean())
}
