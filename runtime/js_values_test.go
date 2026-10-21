// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *RuntimeTestSuite) convertJS(src string) (any, error) {
	rt := goja.New()
	v, err := rt.RunString(src)
	s.Require().NoError(err)
	return jsToSentrie(rt, v)
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
	}
	for src, want := range bad {
		_, err := s.convertJS(src)
		s.Require().Error(err, src)
		s.Contains(err.Error(), want, src)
	}
}

// TestJSToSentrieRejectsOpaqueBuiltins covers #224: goja reports WeakMap,
// WeakSet, DataView, ArrayBuffer and typed arrays as class "Object", so they
// used to convert to {} (or pass through as raw Go slices). Each is now an
// error naming the value's path.
func (s *RuntimeTestSuite) TestJSToSentrieRejectsOpaqueBuiltins() {
	bad := map[string]string{
		`new WeakMap()`:                   "unexpected return type WeakMap at return value",
		`new WeakSet()`:                   "unexpected return type WeakSet at return value",
		`({ w: new WeakMap() })`:          "WeakMap at return value.w",
		`[1, new WeakSet()]`:              "WeakSet at return value[1]",
		`new Map([["k", new WeakMap()]])`: `WeakMap at return value.get("k")`,
		`(() => { class W extends WeakMap {}; return new W(); })()`: "WeakMap",
		`Object.setPrototypeOf(new WeakSet(), Object.prototype)`:    "WeakSet",
		`new ArrayBuffer(4)`:                                            "ArrayBuffer at return value",
		`new Uint8Array([1, 2])`:                                        "Uint8Array at return value",
		`new Float64Array(1)`:                                           "Float64Array",
		`new BigInt64Array(1)`:                                          "BigInt64Array",
		`new DataView(new ArrayBuffer(4))`:                              "DataView",
		`({ bytes: new Uint8Array(1) })`:                                "Uint8Array at return value.bytes",
		`(() => { class B extends Uint8Array {}; return new B(1); })()`: "Uint8Array",
	}
	for src, want := range bad {
		_, err := s.convertJS(src)
		s.Require().Error(err, src)
		s.Contains(err.Error(), want, src)
	}

	ok := map[string]any{
		// No internal slot, so it is an ordinary object with no own keys.
		`Object.create(WeakMap.prototype)`: map[string]any{},
		// The explicit route for binary data.
		`Array.from(new Uint8Array([1, 2]))`: []any{int64(1), int64(2)},
		// goja has no WeakRef or FinalizationRegistry, so a module cannot
		// return one. If goja adds them, extend opaqueBuiltin and this test.
		`typeof WeakRef + "," + typeof FinalizationRegistry`: "undefined,undefined",
	}
	for src, want := range ok {
		got, err := s.convertJS(src)
		s.Require().NoError(err, src)
		s.Equal(want, got, src)
	}

	// A Go byte slice a builtin returns (for example a digest) is not a typed
	// array and still takes the wrapped-Go path.
	rt := goja.New()
	got, err := jsToSentrie(rt, rt.ToValue([]byte{1, 2}))
	s.Require().NoError(err)
	s.Equal([]any{int64(1), int64(2)}, got)

	// Overwritten globals disable the check but must not break conversion.
	s.NotPanics(func() {
		_, err := s.convertJS(`(() => { const w = new WeakMap(); delete globalThis.WeakMap; delete globalThis.ArrayBuffer; return { w }; })()`)
		s.NoError(err)
	})
}

// TestModuleReturningWeakMapIsAnError is #224's reproduction end to end: a rule
// calling an export that returns a WeakMap fails instead of yielding {}.
func (s *RuntimeTestSuite) TestModuleReturningWeakMapIsAnError() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"weak.ts": `export function f() { return new WeakMap(); }
`,
		"policy.sentrie": `namespace weak/test

policy p {
  use { f } from "./weak" as m

  rule r = { yield m.f() }

  export decision of r
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	_, err := exec.ExecRule(ctx, "weak/test", "p", "r", map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), "unexpected return type WeakMap at return value")
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
