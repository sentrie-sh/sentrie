// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
			return throwErr(vm, errors.New("match requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return throwErr(vm, err)
		}
		return vm.ToValue(re.MatchString(str))
	})

	_ = ex.Set("find", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("find requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return throwErr(vm, err)
		}
		match := re.FindString(str)
		if match == "" {
			return goja.Null()
		}
		return vm.ToValue(match)
	})

	_ = ex.Set("findAll", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("findAll requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return throwErr(vm, err)
		}
		matches := re.FindAllString(str, -1)
		return vm.ToValue(matches)
	})

	_ = ex.Set("replace", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return throwErr(vm, errors.New("replace requires exactly 3 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()
		replacement := call.Argument(2).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return throwErr(vm, err)
		}
		result := re.ReplaceAllString(str, replacement)
		return vm.ToValue(result)
	})

	_ = ex.Set("replaceAll", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 3 {
			return throwErr(vm, errors.New("replaceAll requires exactly 3 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()
		replacement := call.Argument(2).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return throwErr(vm, err)
		}
		result := re.ReplaceAllString(str, replacement)
		return vm.ToValue(result)
	})

	_ = ex.Set("split", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return throwErr(vm, errors.New("split requires exactly 2 arguments"))
		}
		pattern := call.Argument(0).String()
		str := call.Argument(1).String()

		re, err := getCompiledRegex(pattern)
		if err != nil {
			return throwErr(vm, err)
		}
		parts := re.Split(str, -1)
		return vm.ToValue(parts)
	})

	return ex, nil
}
