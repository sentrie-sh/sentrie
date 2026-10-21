// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
)

// TestRegexReplaceFirstAndAll pins #209: replace rewrites only the leftmost
// match, replaceAll rewrites every match, and both expand $n references.
func (s *JSTestSuite) TestRegexReplaceFirstAndAll() {
	cases := []struct {
		expr string
		want string
	}{
		{`m.replace("a", "aaa", "b")`, "baa"},
		{`m.replaceAll("a", "aaa", "b")`, "bbb"},
		{`m.replace("a", "xyz", "b")`, "xyz"},
		{`m.replaceAll("a", "xyz", "b")`, "xyz"},
		{`m.replace("o+", "foo boo", "0")`, "f0 boo"},
		{`m.replaceAll("o+", "foo boo", "0")`, "f0 b0"},
		{`m.replace("(\\w+)@(\\w+)", "a@b c@d", "$2@$1")`, "b@a c@d"},
		{`m.replaceAll("(\\w+)@(\\w+)", "a@b c@d", "$2@$1")`, "b@a d@c"},
		{`m.replace("^", "abc", ">")`, ">abc"},
		{`m.replace("x*", "abc", "-")`, "-abc"},
		{`m.replaceAll("x*", "abc", "-")`, "-a-b-c-"},
	}
	for _, tc := range cases {
		s.Run(tc.expr, func() {
			got, err := s.runBuiltinScript(BuiltinRegexGo, tc.expr)
			s.Require().NoError(err)
			s.Equal(tc.want, got.String())
		})
	}
}

// TestRegexReplaceThrowsOnBadPattern keeps both replace forms on the
// throwing convention from #111.
func (s *JSTestSuite) TestRegexReplaceThrowsOnBadPattern() {
	for _, expr := range []string{`m.replace("(", "x", "y")`, `m.replaceAll("(", "x", "y")`} {
		s.Run(expr, func() {
			_, err := s.runBuiltinScript(BuiltinRegexGo, expr)
			s.Require().Error(err)
			s.Contains(err.Error(), "missing closing )")
		})
	}
}

// TestRegexCacheStaysBounded pins #210: many distinct patterns, as a policy
// building patterns from request input would produce, never grow the cache
// past regexCacheEntries.
func (s *JSTestSuite) TestRegexCacheStaysBounded() {
	s.Equal(regexCacheEntries, regexCache.Cap())
	for i := range 3 * regexCacheEntries {
		re, err := getCompiledRegex(fmt.Sprintf("^bounded-%d$", i))
		s.Require().NoError(err)
		s.True(re.MatchString(fmt.Sprintf("bounded-%d", i)))
	}
	s.LessOrEqual(regexCache.Stats().Size, regexCacheEntries)

	// A hit hands back the same compiled value; a compile error is not cached.
	first, err := getCompiledRegex("^hot$")
	s.Require().NoError(err)
	second, err := getCompiledRegex("^hot$")
	s.Require().NoError(err)
	s.Same(first, second)
	_, err = getCompiledRegex("(")
	s.Require().Error(err)
	_, cached := regexCache.Peek("(")
	s.False(cached)
}

// TestRegexCacheConcurrentVMs drives the shared cache from many VMs at once,
// mixing a hot pattern with distinct ones, so `go test -race` covers it.
func (s *JSTestSuite) TestRegexCacheConcurrentVMs() {
	const workers, calls = 8, 200
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for w := range workers {
		wg.Go(func() {
			vm := goja.New()
			ex, err := BuiltinRegexGo(vm)
			if err != nil {
				errs <- err
				return
			}
			if err := vm.Set("m", ex); err != nil {
				errs <- err
				return
			}
			script := fmt.Sprintf(`(function () {
				for (let i = 0; i < %d; i++) {
					if (!m.match("^hot-[0-9]+$", "hot-" + i)) throw new Error("hot miss " + i);
					if (!m.match("^w%d-" + i + "$", "w%d-" + i)) throw new Error("distinct miss " + i);
				}
				return true;
			})()`, calls, w, w)
			if _, err := vm.RunString(script); err != nil {
				errs <- err
			}
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		s.NoError(err)
	}
	s.LessOrEqual(regexCache.Stats().Size, regexCacheEntries)
}
