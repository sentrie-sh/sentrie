// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
			return vm.NewGoError(errors.New("parse requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return vm.NewGoError(err)
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
			return vm.NewGoError(errors.New("join requires at least 1 argument"))
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
			return vm.NewGoError(err)
		}

		// Join remaining parts
		for i := 1; i < len(parts); i++ {
			relative, err := url.Parse(parts[i])
			if err != nil {
				return vm.NewGoError(err)
			}
			base = base.ResolveReference(relative)
		}

		return vm.ToValue(base.String())
	})

	_ = ex.Set("getHost", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("getHost requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		return vm.ToValue(parsed.Host)
	})

	_ = ex.Set("getPath", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("getPath requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		return vm.ToValue(parsed.Path)
	})

	_ = ex.Set("getQuery", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("getQuery requires exactly 1 argument"))
		}
		urlStr := call.Argument(0).String()

		parsed, err := url.Parse(urlStr)
		if err != nil {
			return vm.NewGoError(err)
		}

		return vm.ToValue(parsed.RawQuery)
	})

	_ = ex.Set("isValid", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("isValid requires exactly 1 argument"))
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

