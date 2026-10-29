// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package js

import (
	"math"

	"github.com/dop251/goja"
)

// mathModuleVM returns a VM with @sentrie/math bound to the global `m`.
func (s *JSTestSuite) mathModuleVM() *goja.Runtime {
	vm := goja.New()
	ex, err := BuiltinMathGo(vm)
	s.Require().NoError(err)
	s.Require().NoError(vm.Set("m", ex))
	return vm
}

// mathEval evaluates a script against @sentrie/math and returns the result as
// a float64.
func (s *JSTestSuite) mathEval(script string) float64 {
	v, err := s.mathModuleVM().RunString(script)
	s.Require().NoError(err, "script: %s", script)
	s.Require().True(goja.IsNumber(v), "script %s returned %v", script, v)
	return v.ToFloat()
}

// mathEvalThrows asserts that a script throws a JS exception whose message
// contains want.
func (s *JSTestSuite) mathEvalThrows(script, want string) {
	_, err := s.mathModuleVM().RunString(script)
	s.Require().Error(err, "script: %s", script)
	var ex *goja.Exception
	s.Require().ErrorAs(err, &ex, "script: %s", script)
	s.Contains(err.Error(), want, "script: %s", script)
}

func (s *JSTestSuite) TestMathAbs() {
	s.Equal(3.5, s.mathEval(`m.abs(-3.5)`))
	s.Equal(2.0, s.mathEval(`m.abs(2)`))
	s.Equal(0.0, s.mathEval(`m.abs(-0)`))
	s.mathEvalThrows(`m.abs()`, "abs: expected 1 argument(s), got 0")
	s.mathEvalThrows(`m.abs(1, 2)`, "abs: expected 1 argument(s), got 2")
}

func (s *JSTestSuite) TestMathFloorCeilRoundTrunc() {
	s.Equal(2.0, s.mathEval(`m.floor(2.7)`))
	s.Equal(-3.0, s.mathEval(`m.floor(-2.1)`))
	s.Equal(3.0, s.mathEval(`m.ceil(2.1)`))
	s.Equal(-2.0, s.mathEval(`m.ceil(-2.7)`))
	s.Equal(3.0, s.mathEval(`m.round(2.5)`))
	s.Equal(-3.0, s.mathEval(`m.round(-2.5)`), "round is half away from zero, unlike Math.round")
	s.Equal(2.0, s.mathEval(`m.round(2.4)`))
	s.Equal(2.0, s.mathEval(`m.trunc(2.9)`))
	s.Equal(-2.0, s.mathEval(`m.trunc(-2.9)`))
}

func (s *JSTestSuite) TestMathNegativeZeroIsNormalized() {
	// ceil(-0.5) and trunc(-0.5) are -0 in IEEE arithmetic; the module returns 0.
	for _, script := range []string{`m.ceil(-0.5)`, `m.trunc(-0.5)`, `m.round(-0.4)`, `m.max(-0, -0)`} {
		v := s.mathEval(script)
		s.Equal(0.0, v, script)
		s.False(math.Signbit(v), "%s returned negative zero", script)
	}
	s.Equal(true, s.mathEvalBool(`Object.is(m.ceil(-0.5), 0)`))
}

func (s *JSTestSuite) mathEvalBool(script string) bool {
	v, err := s.mathModuleVM().RunString(script)
	s.Require().NoError(err, "script: %s", script)
	b, ok := v.Export().(bool)
	s.Require().True(ok, "script %s returned %v", script, v)
	return b
}

func (s *JSTestSuite) TestMathMinMax() {
	s.Equal(1.0, s.mathEval(`m.min(3, 1, 2)`))
	s.Equal(3.0, s.mathEval(`m.max(3, 1, 2)`))
	s.Equal(-1.5, s.mathEval(`m.min(-1.5)`))
	s.Equal(1.0, s.mathEval(`m.min([3, 1, 2])`))
	s.Equal(3.0, s.mathEval(`m.max([3, 1, 2])`))
	s.mathEvalThrows(`m.min()`, "min: expected at least 1 argument, got 0")
	s.mathEvalThrows(`m.max([])`, "max: the array must not be empty")
	s.mathEvalThrows(`m.max([1], 2)`, "max: an array must be the only argument, got 2 arguments")
	s.mathEvalThrows(`m.min(1, "2")`, "min: argument 2 must be a number, got string")
	s.mathEvalThrows(`m.max([1, null])`, "max: array element 1 must be a number, got null")
	s.mathEvalThrows(`m.max(1, NaN)`, "max: argument 2 must be a finite number, got NaN")
}

func (s *JSTestSuite) TestMathMinMaxBoundsArrayLength() {
	// A sparse array with a huge length costs nothing in JS; min and max must
	// refuse it before allocating, not size a buffer from the length.
	s.mathEvalThrows(`var a = []; a.length = 2**32 - 1; m.min(a)`,
		"min: the array has 4294967295 elements, more than the limit of 1048576")
	// Within the limit, a missing element throws at that element.
	s.mathEvalThrows(`var a = []; a.length = 1048576; m.max(a)`,
		"max: array element 0 must be a number, got undefined")
}

func (s *JSTestSuite) TestMathMinMaxAcceptsGoSlice() {
	// Lists arriving from policy code are Go slices wrapped by goja, not JS arrays.
	vm := s.mathModuleVM()
	s.Require().NoError(vm.Set("list", []any{4.0, int64(9), 2.5}))
	v, err := vm.RunString(`m.max(list) - m.min(list)`)
	s.Require().NoError(err)
	s.Equal(6.5, v.ToFloat())
}

func (s *JSTestSuite) TestMathPowSqrt() {
	s.Equal(1024.0, s.mathEval(`m.pow(2, 10)`))
	s.Equal(0.5, s.mathEval(`m.pow(2, -1)`))
	s.Equal(1.0, s.mathEval(`m.pow(0, 0)`))
	s.Equal(3.0, s.mathEval(`m.sqrt(9)`))
	s.Equal(0.0, s.mathEval(`m.sqrt(0)`))
	s.mathEvalThrows(`m.pow(10, 400)`, "pow: result is not a finite number")
	s.mathEvalThrows(`m.pow(-8, 1/3)`, "pow: result is not a finite number")
	s.mathEvalThrows(`m.pow(0, -1)`, "pow: result is not a finite number")
	s.mathEvalThrows(`m.pow(2)`, "pow: expected 2 argument(s), got 1")
	s.mathEvalThrows(`m.sqrt(-1)`, "sqrt: argument must be non-negative, got -1")
}

func (s *JSTestSuite) TestMathClamp() {
	s.Equal(5.0, s.mathEval(`m.clamp(5, 0, 10)`))
	s.Equal(0.0, s.mathEval(`m.clamp(-5, 0, 10)`))
	s.Equal(10.0, s.mathEval(`m.clamp(50, 0, 10)`))
	s.Equal(3.0, s.mathEval(`m.clamp(7, 3, 3)`))
	s.mathEvalThrows(`m.clamp(1, 10, 0)`, "clamp: lower bound 10 is greater than upper bound 0")
	s.mathEvalThrows(`m.clamp(1, 0)`, "clamp: expected 3 argument(s), got 2")
}

func (s *JSTestSuite) TestMathRejectsNonFiniteAndNonNumbers() {
	cases := []struct{ script, want string }{
		{`m.abs(NaN)`, "abs: argument 1 must be a finite number, got NaN"},
		{`m.floor(Infinity)`, "floor: argument 1 must be a finite number, got Infinity"},
		{`m.ceil(-Infinity)`, "ceil: argument 1 must be a finite number, got -Infinity"},
		{`m.round("2.5")`, "round: argument 1 must be a number, got string"},
		{`m.trunc(null)`, "trunc: argument 1 must be a number, got null"},
		{`m.sqrt(undefined)`, "sqrt: argument 1 must be a number, got undefined"},
		{`m.pow(true, 2)`, "pow: argument 1 must be a number, got boolean"},
		{`m.clamp(1, 0, Infinity)`, "clamp: argument 3 must be a finite number, got Infinity"},
		{`m.abs({})`, "abs: argument 1 must be a number, got object"},
	}
	for _, tc := range cases {
		s.mathEvalThrows(tc.script, tc.want)
	}
}

func (s *JSTestSuite) TestMathErrorsAreCatchable() {
	// Errors are thrown, never returned, so a try/catch sees them.
	s.Equal(true, s.mathEvalBool(`(() => { try { m.sqrt(-1); return false } catch (e) { return true } })()`))
}
