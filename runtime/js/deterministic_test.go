// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"time"

	"github.com/dop251/goja"
	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/constants"
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
	s.Equal(deterministicStart.Unix(), ar.VM.Get(constants.ExecutionStartTimeUnixKey).ToInteger())

	// The real globals are untouched.
	got := s.run(ar.VM, "Date.now()").ToInteger()
	s.InDelta(time.Now().UnixMilli(), got, float64(time.Minute.Milliseconds()))
	s.Nil(ar.VM.Get(determinismStateKey))
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
	now, ok := goja.AssertFunction(exports.Get("now"))
	s.Require().True(ok)
	random, ok := goja.AssertFunction(exports.Get("random"))
	s.Require().True(ok)

	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@sentrie/js", "now"))
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

func (s *JSTestSuite) TestDeterministicStateCannotBeReplacedFromJS() {
	ar := s.deterministicRuntime(true)
	s.run(ar.VM, `try { globalThis.__sentrie_determinism = 1 } catch (e) {}; try { delete globalThis.__sentrie_determinism } catch (e) {}`)
	s.Require().NoError(BeginCall(ar.VM, deterministicStart, "@sentrie/js", "now"))
	s.Equal(deterministicStart.UnixMilli(), s.run(ar.VM, "Date.now()").ToInteger())
	s.False(s.run(ar.VM, `Object.keys(globalThis).includes("__sentrie_determinism")`).ToBoolean())
}
