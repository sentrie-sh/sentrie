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
	"math"
	"math/rand"
	"time"

	"github.com/dop251/goja"
)

var BuiltinMathGo = func(vm *goja.Runtime) (*goja.Object, error) {
	ex := vm.NewObject()

	// Constants
	_ = ex.Set("E", math.E)
	_ = ex.Set("PI", math.Pi)
	_ = ex.Set("LN2", math.Ln2)
	_ = ex.Set("LN10", math.Ln10)
	_ = ex.Set("LOG2E", math.Log2E)
	_ = ex.Set("LOG10E", math.Log10E)
	_ = ex.Set("SQRT2", math.Sqrt2)
	_ = ex.Set("SQRT1_2", math.Sqrt2/2)
	_ = ex.Set("MAX_VALUE", math.MaxFloat64)
	_ = ex.Set("MIN_VALUE", math.SmallestNonzeroFloat64)

	// Basic operations
	_ = ex.Set("abs", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("abs requires exactly 1 argument"))
		}
		return vm.ToValue(math.Abs(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("ceil", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("ceil requires exactly 1 argument"))
		}
		return vm.ToValue(math.Ceil(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("floor", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("floor requires exactly 1 argument"))
		}
		return vm.ToValue(math.Floor(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("round", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("round requires exactly 1 argument"))
		}
		return vm.ToValue(math.Round(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("max", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.NewGoError(errors.New("max requires at least 1 argument"))
		}
		maxVal := call.Argument(0).ToFloat()
		for i := 1; i < len(call.Arguments); i++ {
			val := call.Argument(i).ToFloat()
			if val > maxVal {
				maxVal = val
			}
		}
		return vm.ToValue(maxVal)
	})

	_ = ex.Set("min", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.NewGoError(errors.New("min requires at least 1 argument"))
		}
		minVal := call.Argument(0).ToFloat()
		for i := 1; i < len(call.Arguments); i++ {
			val := call.Argument(i).ToFloat()
			if val < minVal {
				minVal = val
			}
		}
		return vm.ToValue(minVal)
	})

	_ = ex.Set("sqrt", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("sqrt requires exactly 1 argument"))
		}
		val := call.Argument(0).ToFloat()
		if val < 0 {
			return vm.NewGoError(errors.New("sqrt: square root of negative number"))
		}
		return vm.ToValue(math.Sqrt(val))
	})

	_ = ex.Set("pow", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("pow requires exactly 2 arguments"))
		}
		base := call.Argument(0).ToFloat()
		exp := call.Argument(1).ToFloat()
		return vm.ToValue(math.Pow(base, exp))
	})

	_ = ex.Set("exp", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("exp requires exactly 1 argument"))
		}
		return vm.ToValue(math.Exp(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("log", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("log requires exactly 1 argument"))
		}
		val := call.Argument(0).ToFloat()
		if val <= 0 {
			return vm.NewGoError(errors.New("log: logarithm of non-positive number"))
		}
		return vm.ToValue(math.Log(val))
	})

	_ = ex.Set("log10", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("log10 requires exactly 1 argument"))
		}
		val := call.Argument(0).ToFloat()
		if val <= 0 {
			return vm.NewGoError(errors.New("log10: logarithm of non-positive number"))
		}
		return vm.ToValue(math.Log10(val))
	})

	_ = ex.Set("log2", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("log2 requires exactly 1 argument"))
		}
		val := call.Argument(0).ToFloat()
		if val <= 0 {
			return vm.NewGoError(errors.New("log2: logarithm of non-positive number"))
		}
		return vm.ToValue(math.Log2(val))
	})

	// Trigonometric functions
	_ = ex.Set("sin", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("sin requires exactly 1 argument"))
		}
		return vm.ToValue(math.Sin(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("cos", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("cos requires exactly 1 argument"))
		}
		return vm.ToValue(math.Cos(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("tan", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("tan requires exactly 1 argument"))
		}
		return vm.ToValue(math.Tan(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("asin", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("asin requires exactly 1 argument"))
		}
		val := call.Argument(0).ToFloat()
		if val < -1 || val > 1 {
			return vm.NewGoError(errors.New("asin: value must be between -1 and 1"))
		}
		return vm.ToValue(math.Asin(val))
	})

	_ = ex.Set("acos", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("acos requires exactly 1 argument"))
		}
		val := call.Argument(0).ToFloat()
		if val < -1 || val > 1 {
			return vm.NewGoError(errors.New("acos: value must be between -1 and 1"))
		}
		return vm.ToValue(math.Acos(val))
	})

	_ = ex.Set("atan", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("atan requires exactly 1 argument"))
		}
		return vm.ToValue(math.Atan(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("atan2", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 2 {
			return vm.NewGoError(errors.New("atan2 requires exactly 2 arguments"))
		}
		y := call.Argument(0).ToFloat()
		x := call.Argument(1).ToFloat()
		return vm.ToValue(math.Atan2(y, x))
	})

	// Hyperbolic functions
	_ = ex.Set("sinh", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("sinh requires exactly 1 argument"))
		}
		return vm.ToValue(math.Sinh(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("cosh", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("cosh requires exactly 1 argument"))
		}
		return vm.ToValue(math.Cosh(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("tanh", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) != 1 {
			return vm.NewGoError(errors.New("tanh requires exactly 1 argument"))
		}
		return vm.ToValue(math.Tanh(call.Argument(0).ToFloat()))
	})

	_ = ex.Set("random", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			return vm.NewGoError(errors.New("random requires exactly 0 arguments"))
		}
		// Returns a random number in [0, 1) like JavaScript Math.random()
		// Create a new RNG with nanosecond timestamp seed for each call
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		return vm.ToValue(rng.Float64())
	})

	return ex, nil
}
