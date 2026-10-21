// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"strings"

	"github.com/dop251/goja"
)

var BuiltinJwtGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("decode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 || len(call.Arguments) > 2 {
			return throwErr(vm, errors.New("decode requires 1 or 2 arguments"))
		}
		tokenStr := call.Argument(0).String()
		var secret string
		if len(call.Arguments) > 1 && call.Argument(1) != goja.Undefined() && call.Argument(1) != goja.Null() {
			secret = call.Argument(1).String()
		}

		parts := strings.Split(tokenStr, ".")
		if len(parts) != 3 {
			return throwErr(vm, errors.New("invalid token format"))
		}

		// Decode and parse header
		headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return throwErr(vm, err)
		}

		var header map[string]interface{}
		if err := json.Unmarshal(headerBytes, &header); err != nil {
			return throwErr(vm, err)
		}

		// Decode and parse payload
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return throwErr(vm, err)
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			return throwErr(vm, err)
		}

		// If secret provided, verify signature
		if secret != "" {
			alg, ok := header["alg"].(string)
			if !ok {
				return throwErr(vm, errors.New("algorithm not found in header"))
			}

			// Verify signature
			message := parts[0] + "." + parts[1]
			signature := parts[2]

			var h hash.Hash
			switch alg {
			case "HS256":
				h = hmac.New(sha256.New, []byte(secret))
			case "HS384":
				h = hmac.New(sha512.New384, []byte(secret))
			case "HS512":
				h = hmac.New(sha512.New, []byte(secret))
			default:
				return throwErr(vm, errors.New("unsupported algorithm: "+alg))
			}

			h.Write([]byte(message))
			expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

			if signature != expectedSig {
				return throwErr(vm, errors.New("token signature is invalid"))
			}
		}

		return vm.ToValue(payload)
	})

	_ = ex.Set("verify", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
			return throwErr(vm, errors.New("verify requires 2 or 3 arguments"))
		}
		tokenStr := call.Argument(0).String()
		secret := call.Argument(1).String()
		algorithm := "HS256"
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			algorithm = call.Argument(2).String()
		}

		parts := strings.Split(tokenStr, ".")
		if len(parts) != 3 {
			return vm.ToValue(false)
		}

		// Decode header to get algorithm
		headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return vm.ToValue(false)
		}

		var header map[string]interface{}
		if err := json.Unmarshal(headerBytes, &header); err != nil {
			return vm.ToValue(false)
		}

		alg, ok := header["alg"].(string)
		if !ok || alg != algorithm {
			return vm.ToValue(false)
		}

		// Verify signature
		message := parts[0] + "." + parts[1]
		signature := parts[2]

		var h hash.Hash
		switch algorithm {
		case "HS256":
			h = hmac.New(sha256.New, []byte(secret))
		case "HS384":
			h = hmac.New(sha512.New384, []byte(secret))
		case "HS512":
			h = hmac.New(sha512.New, []byte(secret))
		default:
			return vm.ToValue(false)
		}

		h.Write([]byte(message))
		expectedSig := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

		return vm.ToValue(signature == expectedSig)
	})

	_ = ex.Set("getPayload", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("getPayload requires exactly 1 argument"))
		}
		tokenStr := call.Argument(0).String()

		parts := strings.Split(tokenStr, ".")
		if len(parts) != 3 {
			return throwErr(vm, errors.New("invalid token format"))
		}

		// Decode payload (second part)
		decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return throwErr(vm, err)
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(decoded, &payload); err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(payload)
	})

	_ = ex.Set("getHeader", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("getHeader requires exactly 1 argument"))
		}
		tokenStr := call.Argument(0).String()

		parts := strings.Split(tokenStr, ".")
		if len(parts) != 3 {
			return throwErr(vm, errors.New("invalid token format"))
		}

		// Decode header (first part)
		decoded, err := base64.RawURLEncoding.DecodeString(parts[0])
		if err != nil {
			return throwErr(vm, err)
		}

		var header map[string]interface{}
		if err := json.Unmarshal(decoded, &header); err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(header)
	})

	return ex, nil
}
