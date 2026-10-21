// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"errors"
	"net/url"
	"strings"

	"github.com/dop251/goja"
)

var BuiltinUrlGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("parse", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("parse requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return throwErr(vm, err)
		}

		result := vm.NewObject()
		_ = result.Set("scheme", parsed.Scheme)
		_ = result.Set("host", parsed.Host)
		_ = result.Set("path", parsed.Path)
		_ = result.Set("query", parsed.RawQuery)
		_ = result.Set("fragment", parsed.Fragment)
		_ = result.Set("user", parsed.User.String())

		return result
	})

	_ = ex.Set("join", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return throwErr(vm, errors.New("join requires at least 1 argument"))
		}

		// Collect all parts
		var parts []string
		for i := 0; i < len(call.Arguments); i++ {
			if call.Argument(i) != goja.Undefined() && call.Argument(i) != goja.Null() {
				parts = append(parts, call.Argument(i).String())
			}
		}

		if len(parts) == 0 {
			return vm.ToValue("")
		}

		// Start with first part as base
		base, err := url.Parse(parts[0])
		if err != nil {
			return throwErr(vm, err)
		}

		// Join remaining parts
		for i := 1; i < len(parts); i++ {
			relative, err := url.Parse(parts[i])
			if err != nil {
				return throwErr(vm, err)
			}
			base = base.ResolveReference(relative)
		}

		return vm.ToValue(base.String())
	})

	_ = ex.Set("getHost", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("getHost requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(parsed.Host)
	})

	_ = ex.Set("getPath", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("getPath requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(parsed.Path)
	})

	_ = ex.Set("getQuery", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("getQuery requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return throwErr(vm, err)
		}

		return vm.ToValue(parsed.RawQuery)
	})

	_ = ex.Set("isValid", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return throwErr(vm, errors.New("isValid requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return vm.ToValue(false)
		}

		// Basic validation: scheme and host should be present for most URLs
		// This is a simplified validation
		isValid := parsed.Scheme != "" && (parsed.Host != "" || strings.HasPrefix(urlStr, "/"))
		return vm.ToValue(isValid)
	})

	return ex, nil
}
