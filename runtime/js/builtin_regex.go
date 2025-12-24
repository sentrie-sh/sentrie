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
	"regexp"
	"sync"

	"github.com/dop251/goja"
)

// regexCache caches compiled regex patterns for performance
var regexCache = struct {
	sync.RWMutex
	patterns map[string]*regexp.Regexp
}{
	patterns: make(map[string]*regexp.Regexp),
}

func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	regexCache.RLock()
	if re, ok := regexCache.patterns[pattern]; ok {
		regexCache.RUnlock()
		return re, nil
	}
	regexCache.RUnlock()

	regexCache.Lock()
	defer regexCache.Unlock()

	// Double-check after acquiring write lock
	if re, ok := regexCache.patterns[pattern]; ok {
		return re, nil
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	regexCache.patterns[pattern] = re
	return re, nil
}

var BuiltinRegexGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	_ = ex.Set("match", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("match requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return vm.NewGoError(err)
		}
		return vm.ToValue(re.MatchString(str))
	})

	_ = ex.Set("find", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("find requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return vm.NewGoError(err)
		}
		match := re.FindString(str)
		if match == "" {
			return goja.Null()
		}
		return vm.ToValue(match)
	})

	_ = ex.Set("findAll", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("findAll requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return vm.NewGoError(err)
		}
		matches := re.FindAllString(str, -1)
		return vm.ToValue(matches)
	})

	_ = ex.Set("replace", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return vm.NewGoError(errors.New("replace requires exactly 3 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()
		replacement := call.Argument(2).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return vm.NewGoError(err)
		}
		result := re.ReplaceAllString(str, replacement)
		return vm.ToValue(result)
	})

	_ = ex.Set("replaceAll", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return vm.NewGoError(errors.New("replaceAll requires exactly 3 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()
		replacement := call.Argument(2).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return vm.NewGoError(err)
		}
		result := re.ReplaceAllString(str, replacement)
		return vm.ToValue(result)
	})

	_ = ex.Set("split", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("split requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return vm.NewGoError(err)
		}
		parts := re.Split(str, -1)
		return vm.ToValue(parts)
	})

	return ex, nil
}

