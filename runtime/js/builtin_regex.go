// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"context"
	"errors"
	"regexp"

	"github.com/binaek/perch"
	"github.com/dop251/goja"
)

// regexCacheEntries bounds the process-global compiled-pattern cache. Patterns
// can come from request input through facts, so an unbounded cache grows by
// one entry per distinct input for the life of the process (#210). The least
// recently used pattern is evicted past this count and recompiled on its next
// use.
const regexCacheEntries = 1024

// regexCache is a bounded LRU of compiled patterns, shared by every VM.
var regexCache = perch.New[*regexp.Regexp](regexCacheEntries)

// getCompiledRegex returns the compiled form of pattern from the cache,
// compiling it on a miss. Compile errors are returned and not cached.
func getCompiledRegex(pattern string) (*regexp.Regexp, error) {
	// Compiling is pure and does not block, and goja callbacks carry no
	// context, so the load has nothing to cancel.
	re, _, err := regexCache.Get(context.Background(), pattern, 0, func(context.Context, string) (*regexp.Regexp, error) {
		return regexp.Compile(pattern)
	})
	return re, err
}

// replaceFirst replaces the leftmost match of re in str, expanding $1-style
// references in replacement exactly as ReplaceAllString does. The regexp
// package has no first-match replacement, so this expands a single match.
func replaceFirst(re *regexp.Regexp, str, replacement string) string {
	loc := re.FindStringSubmatchIndex(str)
	if loc == nil {
		return str
	}
	expanded := re.ExpandString(nil, replacement, str, loc)
	return str[:loc[0]] + string(expanded) + str[loc[1]:]
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
		return vm.ToValue(replaceFirst(re, str, replacement))
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
