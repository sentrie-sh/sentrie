// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"os"
	"path/filepath"
	"time"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/capabilities"
)

func (s *JSTestSuite) deterministicRuntime(on bool) *AliasRuntime {
	src := `{"version": 1}`
	if on {
		src = `{"version": 1, "deterministic_js": true}`
	}
	grant, err := capabilities.Parse([]byte(src))
	s.Require().NoError(err)
	reg := NewRegistry(s.T().TempDir())
	reg.RegisterTSBuiltin("js", string(BuiltinJSTS))
	reg.SetCapabilities(grant)
	ar := NewAliasRuntime(reg, reg.PackRoot)
	s.Require().NoError(ar.SetupStdLib(s.T().Context(), nil))
	return ar
}

func (s *JSTestSuite) run(vm *goja.Runtime, src string) goja.Value {
	v, err := vm.RunString(src)
	s.Require().NoError(err, src)
	return v
}

var deterministicStart = time.Date(2026, 3, 4, 5, 6, 7, 8_000_000, time.UTC)

func (s *JSTestSuite) TestBeginCallPinsTimeNowWithoutShims() {
	ar := s.deterministicRuntime(false)
	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@sentrie/js", "now"))
	pinned, ok := executionStartUnix(ar.VM)
	s.Require().True(ok)
	s.Equal(deterministicStart.Unix(), pinned)

	// The real globals are untouched.
	got := s.run(ar.VM, "Date.now()").ToInteger()
	s.InDelta(time.Now().UnixMilli(), got, float64(time.Minute.Milliseconds()))
	s.Nil(determinismFor(ar.VM))
}

func (s *JSTestSuite) TestDeterministicDate() {
	ar := s.deterministicRuntime(true)
	vm := ar.VM

	s.Equal(int64(0), s.run(vm, "Date.now()").ToInteger(), "before the first call the clock reads 0")

	s.Require().NoError(BeginCall(vm, deterministicStart, "@sentrie/js", "now"))
	want := deterministicStart.UnixMilli()
	s.Equal(want, s.run(vm, "Date.now()").ToInteger())
	s.Equal(want, s.run(vm, "new Date().getTime()").ToInteger())
	s.True(s.run(vm, "new Date() instanceof Date").ToBoolean())
	s.True(s.run(vm, "new Date().constructor === Date").ToBoolean())
	s.Equal(int64(0), s.run(vm, "new Date(0).getTime()").ToInteger())
	s.Equal(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC).UnixMilli(), s.run(vm, "new Date(Date.UTC(2020, 0, 2)).getTime()").ToInteger())
	s.Equal(time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC).UnixMilli(), s.run(vm, `Date.parse("2020-01-02T03:04:05Z")`).ToInteger())
	s.Equal(s.run(vm, "new Date(Date.now()).toString()").String(), s.run(vm, "Date()").String())

	later := deterministicStart.Add(time.Hour)
	s.Require().NoError(BeginCall(vm, later, "@sentrie/js", "now"))
	s.Equal(later.UnixMilli(), s.run(vm, "Date.now()").ToInteger())
}

func (s *JSTestSuite) randoms(vm *goja.Runtime, start time.Time, key, fn string) []float64 {
	s.Require().NoError(BeginCall(vm, start, key, fn))
	out := make([]float64, 0, 3)
	for range 3 {
		f := s.run(vm, "Math.random()").ToFloat()
		s.GreaterOrEqual(f, 0.0)
		s.Less(f, 1.0)
		out = append(out, f)
	}
	return out
}

func (s *JSTestSuite) TestDeterministicMathRandom() {
	a := s.deterministicRuntime(true)
	b := s.deterministicRuntime(true)

	first := s.randoms(a.VM, deterministicStart, "@sentrie/js", "random")
	s.NotEqual(first[0], first[1], "a call draws a stream, not a constant")
	s.Equal(first, s.randoms(a.VM, deterministicStart, "@sentrie/js", "random"), "same execution and function, same sequence")
	s.Equal(first, s.randoms(b.VM, deterministicStart, "@sentrie/js", "random"), "independent of the pooled VM")
	s.NotEqual(first, s.randoms(a.VM, deterministicStart, "@local/mod.ts", "random"), "seeded by the called function")
	s.NotEqual(first, s.randoms(a.VM, deterministicStart.Add(time.Nanosecond), "@sentrie/js", "random"), "seeded by the execution")
}

func (s *JSTestSuite) TestDeterministicShimsReachModuleExports() {
	ar := s.deterministicRuntime(true)
	exports, err := ar.Require(s.T().Context(), ar.BaseDir, "@sentrie/js")
	s.Require().NoError(err)
	random, ok := goja.AssertFunction(exports.Get("random"))
	s.Require().True(ok)

	// @sentrie/js has no now since #207; module code reaches the clock
	// through the global Date, which the shims replace.
	now, ok := goja.AssertFunction(s.run(ar.VM, "(function () { return Date.now(); })"))
	s.Require().True(ok)
	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@local/mod.ts", "now"))
	v, err := now(goja.Undefined())
	s.Require().NoError(err)
	s.Equal(deterministicStart.UnixMilli(), v.ToInteger())

	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@sentrie/js", "random"))
	r1, err := random(goja.Undefined())
	s.Require().NoError(err)
	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@sentrie/js", "random"))
	r2, err := random(goja.Undefined())
	s.Require().NoError(err)
	s.Equal(r1.ToFloat(), r2.ToFloat())
}

// The shim state lives on the Go side: nothing is added to the global scope.
func (s *JSTestSuite) TestDeterministicShimsAddNoGlobals() {
	plain := s.deterministicRuntime(false)
	shimmed := s.deterministicRuntime(true)
	keys := `Object.getOwnPropertyNames(globalThis).concat(Object.getOwnPropertySymbols(globalThis).map(String)).sort().join(",")`
	s.Equal(s.run(plain.VM, keys).String(), s.run(shimmed.VM, keys).String())
	s.NotNil(determinismFor(shimmed.VM))
}

// Pack code that replaces shared intrinsics before constructing a Date must
// not be able to reach the real Date (and with it the wall clock).
func (s *JSTestSuite) TestDeterministicDateResistsIntrinsicTampering() {
	ar := s.deterministicRuntime(true)
	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@local/leak.ts", "probe"))
	got := s.run(ar.VM, `(function () {
  var leaked = [];
  var bind = Function.prototype.bind, apply = Function.prototype.apply;
  var concat = Array.prototype.concat, slice = Array.prototype.slice;
  Function.prototype.bind = function () { leaked.push(this); return bind.apply(this, arguments); };
  Function.prototype.apply = function (t, a) { leaked.push(this); return apply.call(this, t, a); };
  Array.prototype.concat = function () { leaked.push(this); return concat.apply(this, arguments); };
  Array.prototype.slice = function () { leaked.push(this); return slice.apply(this, arguments); };
  var Real = Reflect.construct;
  Reflect.construct = function (target) { leaked.push(target); return Real.apply(null, arguments); };
  try {
    new Date(2020, 0, 1);
    new Date();
    Date();
    new Date(0);
  } finally {
    Function.prototype.bind = bind; Function.prototype.apply = apply;
    Array.prototype.concat = concat; Array.prototype.slice = slice;
    Reflect.construct = Real;
  }
  var dates = leaked.filter(function (x) { return typeof x === "function" && x !== Date && /Date/.test(String(x.name)); });
  return dates.length + ":" + (Object.getPrototypeOf(new Date()).constructor === Date);
})()`).String()
	s.Equal("0:true", got)
	s.Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local).UnixMilli(), s.run(ar.VM, "new Date(2020, 0, 1).getTime()").ToInteger())
}

func (s *JSTestSuite) TestDeterministicDateKeepsSubclassesAndIdentity() {
	ar := s.deterministicRuntime(true)
	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@sentrie/js", "now"))
	s.True(s.run(ar.VM, "class X extends Date {}; var x = new X(); x instanceof X && x instanceof Date").ToBoolean())
	s.Equal(deterministicStart.UnixMilli(), s.run(ar.VM, "class Y extends Date {}; new Y().getTime()").ToInteger())
	s.Equal(int64(5), s.run(ar.VM, "class Z extends Date {}; new Z(5).getTime()").ToInteger())
	s.Equal("Date", s.run(ar.VM, "Date.name").String())
	s.Equal(int64(7), s.run(ar.VM, "Date.length").ToInteger())
}

// Module top-level code runs when the module first loads, before any call:
// there the shimmed clock reads 0, never the wall clock.
func (s *JSTestSuite) TestDeterministicClockReadsZeroAtModuleLoad() {
	ar := s.deterministicRuntime(true)
	s.Require().NoError(os.WriteFile(filepath.Join(ar.BaseDir, "early.ts"),
		[]byte("export const loadedAt = Date.now();\nexport const builtAt = new Date().getTime();"), 0o644))
	exports, err := ar.Require(s.T().Context(), ar.BaseDir, "./early")
	s.Require().NoError(err)
	s.Equal(int64(0), exports.Get("loadedAt").ToInteger())
	s.Equal(int64(0), exports.Get("builtAt").ToInteger())
}
