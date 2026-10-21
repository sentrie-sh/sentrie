// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"net/url"

	"github.com/dop251/goja"
)

var BuiltinEncodingGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	// Base64 encoding
	_ = ex.Set("base64Encode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("base64Encode requires exactly 1 argument"))
		}
		return vm.ToValue(base64.StdEncoding.EncodeToString([]byte(call.Argument(0).String())))
	})

	_ = ex.Set("base64Decode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("base64Decode requires exactly 1 argument"))
		}
		decoded, err := base64.StdEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(string(decoded))
	})

	_ = ex.Set("base64UrlEncode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("base64UrlEncode requires exactly 1 argument"))
		}
		return vm.ToValue(base64.URLEncoding.EncodeToString([]byte(call.Argument(0).String())))
	})

	_ = ex.Set("base64UrlDecode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("base64UrlDecode requires exactly 1 argument"))
		}
		decoded, err := base64.URLEncoding.DecodeString(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(string(decoded))
	})

	// Hex encoding
	_ = ex.Set("hexEncode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("hexEncode requires exactly 1 argument"))
		}
		return vm.ToValue(hex.EncodeToString([]byte(call.Argument(0).String())))
	})

	_ = ex.Set("hexDecode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("hexDecode requires exactly 1 argument"))
		}
		decoded, err := hex.DecodeString(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(string(decoded))
	})

	// URL encoding
	_ = ex.Set("urlEncode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("urlEncode requires exactly 1 argument"))
		}
		return vm.ToValue(url.QueryEscape(call.Argument(0).String()))
	})

	_ = ex.Set("urlDecode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("urlDecode requires exactly 1 argument"))
		}
		decoded, err := url.QueryUnescape(call.Argument(0).String())
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(decoded)
	})

	return ex, nil
}
