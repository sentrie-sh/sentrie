// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) convertJS(src string) (any, error) {
	rt := goja.New()
	v, err := rt.RunString(src)
	s.Require().NoError(err)
	return jsToSentrie(s.T().Context(), rt, v)
}

// TestJSToSentrieConversionTable covers #154's table: Date and Map convert
// faithfully, and unsupported values are errors rather than a silent {}.
func (s *RuntimeTestSuite) TestJSToSentrieConversionTable() {
	ok := map[string]any{
		`new Date(0)`:                        int64(0),
		`new Date(1700000000999)`:            int64(1700000000),
		`new Map([["a", 1], ["b", "x"]])`:    map[string]any{"a": int64(1), "b": "x"},
		`({ when: new Date(5000), n: 1.5 })`: map[string]any{"when": int64(5), "n": 1.5},
		`[new Date(2000), null, true]`:       []any{int64(2), nil, true},
		`new String("s")`:                    "s",
		`({ m: new Map([["k", [1]]]) })`:     map[string]any{"m": map[string]any{"k": []any{int64(1)}}},
	}
	for src, want := range ok {
		got, err := s.convertJS(src)
		s.Require().NoError(err, src)
		s.Equal(want, got, src)
	}

	nested, err := s.convertJS(`({ a: undefined })`)
	s.Require().NoError(err)
	s.True(box.FromBoundaryAny(nested.(map[string]any)["a"]).IsUndefined())

	bad := map[string]string{
		`new Map([[1, "x"]])`:    "is not a string",
		`new Set([1])`:           "Set",
		`(function() {})`:        "functions cannot cross",
		`({ f: function() {} })`: "return value.f",
		`Symbol("x")`:            "symbol",
		`10n`:                    "big.Int",
		`new Date(NaN)`:          "invalid Date",
		`/x/`:                    "RegExp",
		`(() => { const o = {}; o.self = o; return o; })()`: "cyclic",
		`({ get x() { throw new Error("boom"); } })`:        "boom",
		`new Proxy({ a: 1 }, {})`:                           "Proxy",
		`new Array(2 ** 32 - 1)`:                            "exceeds the limit",
		`(() => { const a = [1]; Object.defineProperty(a, 0, { get() { throw new Error("idx"); } }); return a; })()`: "idx",
		`new Map([["k", { get v() { throw new Error("deep"); } }]])`:                                                 "deep",
	}
	for src, want := range bad {
		_, err := s.convertJS(src)
		s.Require().Error(err, src)
		s.Contains(err.Error(), want, src)
	}
}

// TestInternalGlobalsAreNotReachableFromModules covers #154 part 2: neither
// __require nor the execution start time is a global a module can read or
// overwrite, and overwriting the old name does not move time.now().
func (s *RuntimeTestSuite) TestInternalGlobalsAreNotReachableFromModules() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"probe.ts": `
import { now } from "@sentrie/time";
export function internals(): string {
  return typeof (globalThis as any).__require + "," + typeof (globalThis as any).__executionStartTimeUnix;
}
export function tamperedNow(): number {
  (globalThis as any).__executionStartTimeUnix = 5;
  return now();
}
`,
		"policy.sentrie": `namespace internals/test

policy p {
  use { internals, tamperedNow } from "./probe" as m
  use { now } from @sentrie/time as tm

  rule hidden = { yield m.internals() == "undefined,undefined" }
  rule pinned = { yield m.tamperedNow() == tm.now() }

  export decision of hidden
  export decision of pinned
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	for _, rule := range []string{"hidden", "pinned"} {
		out, err := exec.ExecRule(ctx, "internals/test", "p", rule, map[string]any{})
		s.Require().NoError(err, rule)
		s.Equal(trinary.True, out.Decision.State, rule)
	}
}

// TestConversionStopsAtTheDeadline covers the #215 review finding: a getter
// that loops while the result is converted is interrupted by the call's
// deadline and reported as an error, not a panic or a hang.
func (s *RuntimeTestSuite) TestConversionStopsAtTheDeadline() {
	exec := s.newTempPackExecutor(map[string]string{
		"m.ts": `export function f(): any { return { get x() { while (true) {} } }; }`,
		"policy.sentrie": `namespace conv/test

policy p {
  use { f } from "./m" as m
  rule r = { yield m.f() }
  export decision of r
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	ctx, cancel := context.WithTimeout(s.T().Context(), 100*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panicked: %v", r)
			}
		}()
		_, err := exec.ExecRule(ctx, "conv/test", "p", "r", map[string]any{})
		done <- err
	}()
	select {
	case err := <-done:
		s.Require().Error(err)
		s.NotContains(err.Error(), "panicked")
		s.ErrorIs(err, context.DeadlineExceeded)
	case <-time.After(5 * time.Second):
		s.FailNow("conversion did not stop at the deadline")
	}
}
