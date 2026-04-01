// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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

package constraints_test

import (
	"math"
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
)

func TestNumberMinMaxEqNeqGtLt(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		val     box.Value
		args    []box.Value
		wantErr bool
	}{
		{"min ok", "min", box.Number(5), []box.Value{box.Number(3)}, false},
		{"min boundary", "min", box.Number(3), []box.Value{box.Number(3)}, false},
		{"min fail", "min", box.Number(2), []box.Value{box.Number(3)}, true},
		{"min wrong arg count", "min", box.Number(1), nil, true},
		{"min non-number arg", "min", box.Number(1), []box.Value{box.String("a")}, true},
		{"min non-number val", "min", box.String("a"), []box.Value{box.Number(0)}, true},

		{"max ok", "max", box.Number(2), []box.Value{box.Number(5)}, false},
		{"max boundary", "max", box.Number(5), []box.Value{box.Number(5)}, false},
		{"max fail", "max", box.Number(9), []box.Value{box.Number(5)}, true},
		{"max wrong arg count", "max", box.Number(1), []box.Value{}, true},

		{"eq ok", "eq", box.Number(4), []box.Value{box.Number(4)}, false},
		{"eq fail", "eq", box.Number(4), []box.Value{box.Number(5)}, true},
		{"eq wrong args", "eq", box.Number(1), []box.Value{box.Number(1), box.Number(2)}, true},

		{"neq ok", "neq", box.Number(1), []box.Value{box.Number(2)}, false},
		{"neq fail", "neq", box.Number(2), []box.Value{box.Number(2)}, true},

		{"gt ok", "gt", box.Number(3), []box.Value{box.Number(2)}, false},
		{"gt fail equal", "gt", box.Number(2), []box.Value{box.Number(2)}, true},
		{"gt fail less", "gt", box.Number(1), []box.Value{box.Number(2)}, true},

		{"lt ok", "lt", box.Number(1), []box.Value{box.Number(2)}, false},
		{"lt fail equal", "lt", box.Number(2), []box.Value{box.Number(2)}, true},
		{"lt fail greater", "lt", box.Number(3), []box.Value{box.Number(2)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := constraints.NumberContraintCheckers[tt.key]
			runChecker(t, c, tt.val, tt.args, tt.wantErr)
		})
	}
}

func TestNumberInAndNotIn(t *testing.T) {
	t.Run("in scalar set", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["in"]
		runChecker(t, c, box.Number(7), []box.Value{box.Number(7)}, false)
		runChecker(t, c, box.Number(7), []box.Value{box.Number(8)}, true)
		runChecker(t, c, box.Number(7), nil, true)
		runChecker(t, c, box.String("x"), []box.Value{box.Number(7)}, true)
	})

	t.Run("in list arg", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["in"]
		set := box.List([]box.Value{box.Number(1), box.Number(2), box.Number(3)})
		runChecker(t, c, box.Number(2), []box.Value{set}, false)
		runChecker(t, c, box.Number(9), []box.Value{set}, true)
	})

	t.Run("not_in scalar", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["not_in"]
		runChecker(t, c, box.Number(4), []box.Value{box.Number(5)}, false)
		runChecker(t, c, box.Number(5), []box.Value{box.Number(5)}, true)
		runChecker(t, c, box.String("x"), []box.Value{box.Number(0)}, true)
	})

	t.Run("not_in list", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["not_in"]
		set := box.List([]box.Value{box.Number(1), box.Number(2)})
		runChecker(t, c, box.Number(3), []box.Value{set}, false)
		runChecker(t, c, box.Number(2), []box.Value{set}, true)
	})

	t.Run("not_in list rejects non-number entries", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["not_in"]
		set := box.List([]box.Value{box.String("bad")})
		runChecker(t, c, box.Number(7), []box.Value{set}, true)
	})

	t.Run("not_in rejects non-number scalar set", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["not_in"]
		runChecker(t, c, box.Number(7), []box.Value{box.String("bad")}, true)
	})
}

func TestNumberRange(t *testing.T) {
	c := constraints.NumberContraintCheckers["range"]
	runChecker(t, c, box.Number(5), []box.Value{box.Number(1), box.Number(10)}, false)
	runChecker(t, c, box.Number(1), []box.Value{box.Number(1), box.Number(10)}, false)
	runChecker(t, c, box.Number(10), []box.Value{box.Number(1), box.Number(10)}, false)
	runChecker(t, c, box.Number(0), []box.Value{box.Number(1), box.Number(10)}, true)
	runChecker(t, c, box.Number(11), []box.Value{box.Number(1), box.Number(10)}, true)
	runChecker(t, c, box.Number(5), []box.Value{box.Number(1)}, true)
	runChecker(t, c, box.String("x"), []box.Value{box.Number(1), box.Number(10)}, true)
	runChecker(t, c, box.Number(5), []box.Value{box.String("a"), box.Number(10)}, true)
	runChecker(t, c, box.Number(5), []box.Value{box.Number(1), box.String("b")}, true)
	t.Run("min greater than max", func(t *testing.T) {
		runChecker(t, c, box.Number(5), []box.Value{box.Number(10), box.Number(1)}, true)
	})
}

func TestNumberEvenOdd(t *testing.T) {
	t.Run("even", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["even"]
		runChecker(t, c, box.Number(4), nil, false)
		runChecker(t, c, box.Number(3), nil, true)
		runChecker(t, c, box.String("x"), nil, true)
	})
	t.Run("odd", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["odd"]
		runChecker(t, c, box.Number(3), nil, false)
		runChecker(t, c, box.Number(4), nil, true)
	})
}

func TestNumberMultipleOf(t *testing.T) {
	c := constraints.NumberContraintCheckers["multiple_of"]
	runChecker(t, c, box.Number(12), []box.Value{box.Number(4)}, false)
	runChecker(t, c, box.Number(12), []box.Value{box.Number(5)}, true)
	runChecker(t, c, box.Number(12), []box.Value{box.Number(0)}, true)
	runChecker(t, c, box.Number(12), nil, true)
	runChecker(t, c, box.String("x"), []box.Value{box.Number(2)}, true)
	runChecker(t, c, box.Number(12), []box.Value{box.String("x")}, true)
	t.Run("negative divisor", func(t *testing.T) {
		runChecker(t, c, box.Number(-6), []box.Value{box.Number(-3)}, false)
	})
}

func TestNumberSignConstraints(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["positive"]
		runChecker(t, c, box.Number(0.1), nil, false)
		runChecker(t, c, box.Number(0), nil, true)
		runChecker(t, c, box.String("x"), nil, true)
	})
	t.Run("negative", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["negative"]
		runChecker(t, c, box.Number(-1), nil, false)
		runChecker(t, c, box.Number(0), nil, true)
	})
	t.Run("non_negative", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["non_negative"]
		runChecker(t, c, box.Number(0), nil, false)
		runChecker(t, c, box.Number(-0.1), nil, true)
	})
	t.Run("non_positive", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["non_positive"]
		runChecker(t, c, box.Number(0), nil, false)
		runChecker(t, c, box.Number(0.1), nil, true)
	})
}

func TestNumberFiniteInfiniteNaN(t *testing.T) {
	t.Run("finite", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["finite"]
		runChecker(t, c, box.Number(1.5), nil, false)
		runChecker(t, c, box.Number(math.Inf(1)), nil, true)
		runChecker(t, c, box.Number(math.NaN()), nil, true)
		runChecker(t, c, box.String("x"), nil, true)
	})
	t.Run("infinite", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["infinite"]
		runChecker(t, c, box.Number(math.Inf(-1)), nil, false)
		runChecker(t, c, box.Number(1), nil, true)
	})
	t.Run("nan", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["nan"]
		runChecker(t, c, box.Number(math.NaN()), nil, false)
		runChecker(t, c, box.Number(1), nil, true)
	})
}
