// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto"
	"crypto/hmac"
	_ "crypto/sha256" // registers crypto.SHA256
	_ "crypto/sha512" // registers crypto.SHA384 and crypto.SHA512
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/dop251/goja"
)

// jwtHMACAlgorithms maps each supported HMAC JWS algorithm to its hash.
var jwtHMACAlgorithms = map[string]crypto.Hash{
	"HS256": crypto.SHA256,
	"HS384": crypto.SHA384,
	"HS512": crypto.SHA512,
}

var (
	errJWTSignature = errors.New("token signature is invalid")
	errJWTExpired   = errors.New("token is expired")
	errJWTNotYet    = errors.New("token is not valid yet")
)

// jwtToken is a JWT split into its decoded header and payload, the signing
// input, and the still-encoded signature.
type jwtToken struct {
	header       map[string]any
	payload      map[string]any
	signingInput string
	signature    string
}

func parseJWT(token string) (*jwtToken, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}
	header, err := decodeJWTSegment(parts[0], "header")
	if err != nil {
		return nil, err
	}
	payload, err := decodeJWTSegment(parts[1], "payload")
	if err != nil {
		return nil, err
	}
	return &jwtToken{
		header:       header,
		payload:      payload,
		signingInput: parts[0] + "." + parts[1],
		signature:    parts[2],
	}, nil
}

func decodeJWTSegment(segment, name string) (map[string]any, error) {
	raw, err := base64.RawURLEncoding.DecodeString(segment)
	if err != nil {
		return nil, fmt.Errorf("invalid token %s: %w", name, err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("invalid token %s: %w", name, err)
	}
	return out, nil
}

// checkJWTAlgorithm rejects "none" and any algorithm this module cannot verify.
func checkJWTAlgorithm(alg string) error {
	if alg == "none" {
		return errors.New(`algorithm "none" is not allowed`)
	}
	if _, ok := jwtHMACAlgorithms[alg]; !ok {
		return errors.New("unsupported algorithm: " + alg)
	}
	return nil
}

// hmacSecret validates the key argument for an HMAC algorithm.
func hmacSecret(key goja.Value) ([]byte, error) {
	secret, ok := key.Export().(string)
	if !ok {
		return nil, errors.New("secret must be a string")
	}
	if secret == "" {
		return nil, errors.New("secret must not be empty")
	}
	return []byte(secret), nil
}

// verifyJWTSignature checks tok's signature under alg with key. Argument
// problems (unsupported alg, bad key) come back as their own errors; a
// signature that does not match is errJWTSignature.
func verifyJWTSignature(tok *jwtToken, alg string, key goja.Value) error {
	if err := checkJWTAlgorithm(alg); err != nil {
		return err
	}
	secret, err := hmacSecret(key)
	if err != nil {
		return err
	}
	sig, err := base64.RawURLEncoding.DecodeString(tok.signature)
	if err != nil {
		return errJWTSignature
	}
	mac := hmac.New(jwtHMACAlgorithms[alg].New, secret)
	mac.Write([]byte(tok.signingInput))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return errJWTSignature
	}
	return nil
}

// jwtMaxLeeway caps options.leeway at one day, so a value given in the wrong
// unit (milliseconds) is an error rather than a silent disable of exp/nbf.
const jwtMaxLeeway = 24 * 60 * 60

// jwtOptions are the time-validation settings decode and verify accept.
type jwtOptions struct {
	// leeway is the clock skew, in seconds, tolerated on exp and nbf.
	leeway float64
	// checkTime validates exp and nbf; false is the inspect-only opt-out.
	checkTime bool
}

func parseJWTOptions(v goja.Value) (jwtOptions, error) {
	opts := jwtOptions{checkTime: true}
	if v == nil || goja.IsUndefined(v) || goja.IsNull(v) {
		return opts, nil
	}
	fields, ok := v.Export().(map[string]any)
	if !ok {
		return opts, errors.New("options must be an object")
	}
	for name, val := range fields {
		switch name {
		case "leeway":
			leeway, ok := jsNumber(val)
			if !ok || math.IsNaN(leeway) || leeway < 0 || leeway > jwtMaxLeeway {
				return opts, fmt.Errorf("options.leeway must be between 0 and %d seconds", int64(jwtMaxLeeway))
			}
			opts.leeway = leeway
		case "checkTime":
			checkTime, ok := val.(bool)
			if !ok {
				return opts, errors.New("options.checkTime must be a boolean")
			}
			opts.checkTime = checkTime
		default:
			return opts, fmt.Errorf("unknown option %q", name)
		}
	}
	return opts, nil
}

func jsNumber(v any) (float64, bool) {
	switch n := v.(type) {
	case int64:
		return float64(n), true
	case float64:
		return n, true
	}
	return 0, false
}

// checkJWTTime validates the exp and nbf claims against now (Unix seconds).
// Per RFC 7519 a token is valid while now < exp and once now >= nbf; leeway
// widens both bounds.
func checkJWTTime(payload map[string]any, now int64, leeway float64) error {
	t := float64(now)
	exp, ok, err := numericClaim(payload, "exp")
	if err != nil {
		return err
	}
	if ok && t >= exp+leeway {
		return errJWTExpired
	}
	nbf, ok, err := numericClaim(payload, "nbf")
	if err != nil {
		return err
	}
	if ok && t+leeway < nbf {
		return errJWTNotYet
	}
	return nil
}

func numericClaim(payload map[string]any, name string) (float64, bool, error) {
	v, present := payload[name]
	if !present {
		return 0, false, nil
	}
	n, ok := v.(float64)
	if !ok || math.IsNaN(n) || math.IsInf(n, 0) {
		return 0, false, fmt.Errorf("%s claim must be a number", name)
	}
	return n, true, nil
}

func isAbsent(v goja.Value) bool {
	return v == nil || goja.IsUndefined(v) || goja.IsNull(v)
}

var BuiltinJwtGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	// clock returns the time exp and nbf are checked against: the execution
	// start, the same instant @sentrie/time.now() reports. It is read before
	// the token is looked at, so a VM without it fails on every call.
	clock := func(opts jwtOptions) (int64, error) {
		if !opts.checkTime {
			return 0, nil
		}
		return executionStartUnix(vm)
	}
	timeCheck := func(payload map[string]any, now int64, opts jwtOptions) error {
		if !opts.checkTime {
			return nil
		}
		return checkJWTTime(payload, now, opts.leeway)
	}

	_ = ex.Set("decode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
			return throwErr(vm, errors.New("decode requires 2 or 3 arguments; use getPayload or getHeader to read a token without verifying it"))
		}
		key := call.Argument(1)
		if isAbsent(key) {
			// An unset secret (a missing env var, an absent fact) arrives as
			// undefined. Skipping verification for it would hand back
			// attacker-chosen claims, so a missing key is an error.
			return throwErr(vm, errors.New("decode requires a key, got "+key.String()+"; use getPayload or getHeader to read a token without verifying it"))
		}
		opts, err := parseJWTOptions(call.Argument(2))
		if err != nil {
			return throwErr(vm, err)
		}
		now, err := clock(opts)
		if err != nil {
			return throwErr(vm, err)
		}

		tok, err := parseJWT(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}

		if !isAbsent(key) {
			alg, ok := tok.header["alg"].(string)
			if !ok {
				return throwErr(vm, errors.New("algorithm not found in header"))
			}
			if err := verifyJWTSignature(tok, alg, key); err != nil {
				return throwErr(vm, err)
			}
		}

		if err := timeCheck(tok.payload, now, opts); err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(tok.payload)
	})

	_ = ex.Set("verify", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 || len(call.Arguments) > 4 {
			return throwErr(vm, errors.New("verify requires 2 to 4 arguments"))
		}
		key := call.Argument(1)
		algorithm := "HS256"
		if !isAbsent(call.Argument(2)) {
			algorithm = call.Argument(2).String()
		}
		if err := checkJWTAlgorithm(algorithm); err != nil {
			return throwErr(vm, err)
		}
		if _, err := hmacSecret(key); err != nil {
			return throwErr(vm, err)
		}
		opts, err := parseJWTOptions(call.Argument(3))
		if err != nil {
			return throwErr(vm, err)
		}
		now, err := clock(opts)
		if err != nil {
			return throwErr(vm, err)
		}

		// From here on, anything wrong with the token is an answer, not an error.
		tok, err := parseJWT(call.Argument(0).String())
		if err != nil {
			return vm.ToValue(false)
		}
		if alg, ok := tok.header["alg"].(string); !ok || alg != algorithm {
			return vm.ToValue(false)
		}
		if err := verifyJWTSignature(tok, algorithm, key); err != nil {
			return vm.ToValue(false)
		}
		if err := timeCheck(tok.payload, now, opts); err != nil {
			return vm.ToValue(false)
		}
		return vm.ToValue(true)
	})

	_ = ex.Set("getPayload", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("getPayload requires exactly 1 argument"))
		}
		tok, err := parseJWT(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(tok.payload)
	})

	_ = ex.Set("getHeader", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("getHeader requires exactly 1 argument"))
		}
		tok, err := parseJWT(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(tok.header)
	})

	return ex, nil
}
