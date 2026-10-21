// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rsa"
	_ "crypto/sha256" // registers crypto.SHA256
	_ "crypto/sha512" // registers crypto.SHA384 and crypto.SHA512
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"

	"github.com/dop251/goja"
)

// jwtFamily is the signature scheme behind a JWS algorithm.
type jwtFamily int

const (
	jwtHMAC jwtFamily = iota
	jwtRSA
	jwtECDSA
)

// jwtAlgorithm describes one supported JWS algorithm (RFC 7518 section 3).
type jwtAlgorithm struct {
	family jwtFamily
	hash   crypto.Hash
	// curve is the curve an ECDSA algorithm requires; nil otherwise.
	curve elliptic.Curve
}

var jwtAlgorithms = map[string]jwtAlgorithm{
	"HS256": {family: jwtHMAC, hash: crypto.SHA256},
	"HS384": {family: jwtHMAC, hash: crypto.SHA384},
	"HS512": {family: jwtHMAC, hash: crypto.SHA512},
	"RS256": {family: jwtRSA, hash: crypto.SHA256},
	"RS384": {family: jwtRSA, hash: crypto.SHA384},
	"RS512": {family: jwtRSA, hash: crypto.SHA512},
	"ES256": {family: jwtECDSA, hash: crypto.SHA256, curve: elliptic.P256()},
	"ES384": {family: jwtECDSA, hash: crypto.SHA384, curve: elliptic.P384()},
	"ES512": {family: jwtECDSA, hash: crypto.SHA512, curve: elliptic.P521()},
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
	if _, ok := jwtAlgorithms[alg]; !ok {
		return errors.New("unsupported algorithm: " + alg)
	}
	return nil
}

// verifyJWTSignature checks tok's signature under alg with key, a value
// parseJWTKey returned for the same alg. A signature that does not match is
// errJWTSignature.
func verifyJWTSignature(tok *jwtToken, alg string, key any) error {
	if _, present := tok.header["crit"]; present {
		// RFC 7515 4.1.11: reject critical extensions this verifier does not
		// understand, which is all of them.
		return errors.New("unsupported crit header")
	}
	a := jwtAlgorithms[alg]
	sig, err := base64.RawURLEncoding.DecodeString(tok.signature)
	if err != nil {
		return errJWTSignature
	}
	if a.family == jwtHMAC {
		mac := hmac.New(a.hash.New, key.([]byte))
		mac.Write([]byte(tok.signingInput))
		if !hmac.Equal(sig, mac.Sum(nil)) {
			return errJWTSignature
		}
		return nil
	}

	h := a.hash.New()
	h.Write([]byte(tok.signingInput))
	digest := h.Sum(nil)
	switch a.family {
	case jwtRSA:
		if rsa.VerifyPKCS1v15(key.(*rsa.PublicKey), a.hash, digest, sig) != nil {
			return errJWTSignature
		}
	case jwtECDSA:
		// JWS carries an ECDSA signature as fixed-width R || S (RFC 7518 3.4).
		size := (a.curve.Params().BitSize + 7) / 8
		if len(sig) != 2*size {
			return errJWTSignature
		}
		r := new(big.Int).SetBytes(sig[:size])
		sv := new(big.Int).SetBytes(sig[size:])
		if !ecdsa.Verify(key.(*ecdsa.PublicKey), digest, r, sv) {
			return errJWTSignature
		}
	}
	return nil
}

// jwtOptions are the time-validation settings decode and verify accept.
type jwtOptions struct {
	// leeway is the clock skew, in seconds, tolerated on exp and nbf.
	leeway float64
	// checkTime validates exp and nbf; false is the inspect-only opt-out.
	checkTime bool
	// algorithms pins the header algorithms decode accepts; nil means the
	// algorithms of the key's family (jwtKeyAlgorithms).
	algorithms []string
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
			if !ok || math.IsNaN(leeway) || math.IsInf(leeway, 0) || leeway < 0 {
				return opts, errors.New("options.leeway must be a non-negative number of seconds")
			}
			opts.leeway = leeway
		case "checkTime":
			checkTime, ok := val.(bool)
			if !ok {
				return opts, errors.New("options.checkTime must be a boolean")
			}
			opts.checkTime = checkTime
		case "algorithms":
			list, ok := val.([]any)
			if !ok || len(list) == 0 {
				return opts, errors.New("options.algorithms must be a non-empty array of algorithm names")
			}
			opts.algorithms = make([]string, 0, len(list))
			for _, item := range list {
				alg, ok := item.(string)
				if !ok {
					return opts, errors.New("options.algorithms must be a non-empty array of algorithm names")
				}
				if err := checkJWTAlgorithm(alg); err != nil {
					return opts, err
				}
				opts.algorithms = append(opts.algorithms, alg)
			}
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
		if len(call.Arguments) < 1 || len(call.Arguments) > 3 {
			return throwErr(vm, errors.New("decode requires 1 to 3 arguments"))
		}
		key := call.Argument(1)
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
			if err := checkJWTAlgorithm(alg); err != nil {
				return throwErr(vm, err)
			}
			// The header is attacker-controlled, so it never picks the
			// algorithm family on its own: it must be one the caller allowed,
			// or, by default, one the key's own family supports.
			allowed := opts.algorithms
			if allowed == nil {
				allowed, err = jwtKeyAlgorithms(key)
				if err != nil {
					return throwErr(vm, err)
				}
			}
			if !slices.Contains(allowed, alg) {
				return throwErr(vm, fmt.Errorf("token algorithm %s is not allowed for this key (allowed: %s); pass options.algorithms to choose", alg, strings.Join(allowed, ", ")))
			}
			parsed, err := parseJWTKey(alg, key)
			if err != nil {
				return throwErr(vm, err)
			}
			if err := verifyJWTSignature(tok, alg, parsed); err != nil {
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
		parsed, err := parseJWTKey(algorithm, key)
		if err != nil {
			return throwErr(vm, err)
		}
		opts, err := parseJWTOptions(call.Argument(3))
		if err != nil {
			return throwErr(vm, err)
		}
		if opts.algorithms != nil {
			return throwErr(vm, errors.New("options.algorithms applies to decode; verify takes the algorithm as its third argument"))
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
		if err := verifyJWTSignature(tok, algorithm, parsed); err != nil {
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
