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
	"strings"
	"unicode"

	"github.com/dop251/goja"
)

var BuiltinStringGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("trim", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(strings.TrimSpace(call.Argument(0).String()))
	})

	_ = ex.Set("trimLeft", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(strings.TrimLeftFunc(call.Argument(0).String(), unicode.IsSpace))
	})

	_ = ex.Set("trimRight", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(strings.TrimRightFunc(call.Argument(0).String(), unicode.IsSpace))
	})

	_ = ex.Set("toLowerCase", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(strings.ToLower(call.Argument(0).String()))
	})

	_ = ex.Set("toUpperCase", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(strings.ToUpper(call.Argument(0).String()))
	})

	_ = ex.Set("replace", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		oldStr := call.Argument(1).String()
		newStr := call.Argument(2).String()
		n := -1
		if len(call.Arguments) > 3 && call.Argument(3) != goja.Undefined() && call.Argument(3) != goja.Null() {
			n = int(call.Argument(3).ToInteger())
		}
		if n < 0 {
			return vm.ToValue(strings.ReplaceAll(s, oldStr, newStr))
		}
		return vm.ToValue(strings.Replace(s, oldStr, newStr, n))
	})

	_ = ex.Set("replaceAll", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		oldStr := call.Argument(1).String()
		newStr := call.Argument(2).String()
		return vm.ToValue(strings.ReplaceAll(s, oldStr, newStr))
	})

	_ = ex.Set("split", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		sep := call.Argument(1).String()
		var n int = -1
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			n = int(call.Argument(2).ToInteger())
		}
		if n < 0 {
			parts := strings.Split(s, sep)
			return vm.ToValue(parts)
		}
		parts := strings.SplitN(s, sep, n)
		return vm.ToValue(parts)
	})

	_ = ex.Set("substring", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		start := int(call.Argument(1).ToInteger())
		end := len(s)
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			end = int(call.Argument(2).ToInteger())
		}
		if start < 0 {
			start = 0
		}
		if end < 0 {
			end = 0
		}
		if start > len(s) {
			start = len(s)
		}
		if end > len(s) {
			end = len(s)
		}
		if start > end {
			start, end = end, start
		}
		return vm.ToValue(s[start:end])
	})

	_ = ex.Set("slice", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		start := int(call.Argument(1).ToInteger())
		end := len(s)
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			end = int(call.Argument(2).ToInteger())
		}
		if start < 0 {
			start = len(s) + start
		}
		if end < 0 {
			end = len(s) + end
		}
		if start < 0 {
			start = 0
		}
		if end < 0 {
			end = 0
		}
		if start > len(s) {
			start = len(s)
		}
		if end > len(s) {
			end = len(s)
		}
		if start > end {
			return vm.ToValue("")
		}
		return vm.ToValue(s[start:end])
	})

	_ = ex.Set("startsWith", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		prefix := call.Argument(1).String()
		pos := 0
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			pos = int(call.Argument(2).ToInteger())
		}
		if pos < 0 {
			pos = 0
		}
		if pos > len(s) {
			pos = len(s)
		}
		if pos+len(prefix) > len(s) {
			return vm.ToValue(false)
		}
		return vm.ToValue(strings.HasPrefix(s[pos:], prefix))
	})

	_ = ex.Set("endsWith", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		suffix := call.Argument(1).String()
		pos := len(s)
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			pos = int(call.Argument(2).ToInteger())
		}
		if pos < 0 {
			pos = 0
		}
		if pos > len(s) {
			pos = len(s)
		}
		if len(suffix) > pos {
			return vm.ToValue(false)
		}
		return vm.ToValue(strings.HasSuffix(s[:pos], suffix))
	})

	_ = ex.Set("indexOf", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		substr := call.Argument(1).String()
		fromIndex := 0
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			fromIndex = int(call.Argument(2).ToInteger())
		}
		if fromIndex < 0 {
			fromIndex = 0
		}
		if fromIndex >= len(s) {
			return vm.ToValue(-1)
		}
		idx := strings.Index(s[fromIndex:], substr)
		if idx == -1 {
			return vm.ToValue(-1)
		}
		return vm.ToValue(idx + fromIndex)
	})

	_ = ex.Set("lastIndexOf", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		substr := call.Argument(1).String()
		fromIndex := len(s)
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			fromIndex = int(call.Argument(2).ToInteger())
		}
		if fromIndex < 0 {
			fromIndex = 0
		}
		if fromIndex > len(s) {
			fromIndex = len(s)
		}
		idx := strings.LastIndex(s[:fromIndex], substr)
		return vm.ToValue(idx)
	})

	_ = ex.Set("padStart", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		targetLength := int(call.Argument(1).ToInteger())
		padString := " "
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			padString = call.Argument(2).String()
		}
		if len(padString) == 0 || targetLength <= len(s) {
			return vm.ToValue(s)
		}
		padCount := targetLength - len(s)
		padLen := len(padString)
		fullPads := padCount / padLen
		remainder := padCount % padLen
		padding := strings.Repeat(padString, fullPads) + padString[:remainder]
		return vm.ToValue(padding + s)
	})

	_ = ex.Set("padEnd", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		targetLength := int(call.Argument(1).ToInteger())
		padString := " "
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			padString = call.Argument(2).String()
		}
		if len(padString) == 0 || targetLength <= len(s) {
			return vm.ToValue(s)
		}
		padCount := targetLength - len(s)
		padLen := len(padString)
		fullPads := padCount / padLen
		remainder := padCount % padLen
		padding := strings.Repeat(padString, fullPads) + padString[:remainder]
		return vm.ToValue(s + padding)
	})

	_ = ex.Set("repeat", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		count := int(call.Argument(1).ToInteger())
		if count < 0 {
			return vm.NewGoError(errors.New("repeat count must be non-negative"))
		}
		return vm.ToValue(strings.Repeat(s, count))
	})

	_ = ex.Set("charAt", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		index := int(call.Argument(1).ToInteger())
		if index < 0 || index >= len(s) {
			return vm.ToValue("")
		}
		return vm.ToValue(string(s[index]))
	})

	_ = ex.Set("includes", func(call goja.FunctionCall) goja.Value {
		s := call.Argument(0).String()
		substr := call.Argument(1).String()
		fromIndex := 0
		if len(call.Arguments) > 2 && call.Argument(2) != goja.Undefined() && call.Argument(2) != goja.Null() {
			fromIndex = int(call.Argument(2).ToInteger())
		}
		if fromIndex < 0 {
			fromIndex = 0
		}
		if fromIndex >= len(s) {
			return vm.ToValue(false)
		}
		return vm.ToValue(strings.Contains(s[fromIndex:], substr))
	})

	_ = ex.Set("length", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(len(call.Argument(0).String()))
	})

	return ex, nil
}
