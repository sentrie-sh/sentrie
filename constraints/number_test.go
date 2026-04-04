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

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
)

func (s *ConstraintsTestSuite) TestNumberMinMaxEqNeqGtLt() {
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
		s.Run(tt.name, func() {
			c := constraints.NumberContraintCheckers[tt.key]
			s.runChecker(c, tt.val, tt.args, tt.wantErr)
		})
	}
}

func (s *ConstraintsTestSuite) TestNumberInAndNotIn() {
	s.Run("in scalar set", func() {
		c := constraints.NumberContraintCheckers["in"]
		s.runChecker(c, box.Number(7), []box.Value{box.Number(7)}, false)
		s.runChecker(c, box.Number(7), []box.Value{box.Number(8)}, true)
		s.runChecker(c, box.Number(7), nil, true)
		s.runChecker(c, box.String("x"), []box.Value{box.Number(7)}, true)
	})
	s.Run("in list arg", func() {
		c := constraints.NumberContraintCheckers["in"]
		set := box.List([]box.Value{box.Number(1), box.Number(2), box.Number(3)})
		s.runChecker(c, box.Number(2), []box.Value{set}, false)
		s.runChecker(c, box.Number(9), []box.Value{set}, true)
	})
	s.Run("not_in scalar", func() {
		c := constraints.NumberContraintCheckers["not_in"]
		s.runChecker(c, box.Number(4), []box.Value{box.Number(5)}, false)
		s.runChecker(c, box.Number(5), []box.Value{box.Number(5)}, true)
		s.runChecker(c, box.String("x"), []box.Value{box.Number(0)}, true)
	})
	s.Run("not_in list", func() {
		c := constraints.NumberContraintCheckers["not_in"]
		set := box.List([]box.Value{box.Number(1), box.Number(2)})
		s.runChecker(c, box.Number(3), []box.Value{set}, false)
		s.runChecker(c, box.Number(2), []box.Value{set}, true)
	})
	s.Run("not_in list rejects non-number entries", func() {
		c := constraints.NumberContraintCheckers["not_in"]
		set := box.List([]box.Value{box.String("bad")})
		s.runChecker(c, box.Number(7), []box.Value{set}, true)
	})
	s.Run("not_in rejects non-number scalar set", func() {
		c := constraints.NumberContraintCheckers["not_in"]
		s.runChecker(c, box.Number(7), []box.Value{box.String("bad")}, true)
	})
}

func (s *ConstraintsTestSuite) TestNumberRange() {
	c := constraints.NumberContraintCheckers["range"]
	s.runChecker(c, box.Number(5), []box.Value{box.Number(1), box.Number(10)}, false)
	s.runChecker(c, box.Number(1), []box.Value{box.Number(1), box.Number(10)}, false)
	s.runChecker(c, box.Number(10), []box.Value{box.Number(1), box.Number(10)}, false)
	s.runChecker(c, box.Number(0), []box.Value{box.Number(1), box.Number(10)}, true)
	s.runChecker(c, box.Number(11), []box.Value{box.Number(1), box.Number(10)}, true)
	s.runChecker(c, box.Number(5), []box.Value{box.Number(1)}, true)
	s.runChecker(c, box.String("x"), []box.Value{box.Number(1), box.Number(10)}, true)
	s.runChecker(c, box.Number(5), []box.Value{box.String("a"), box.Number(10)}, true)
	s.runChecker(c, box.Number(5), []box.Value{box.Number(1), box.String("b")}, true)
	s.Run("min greater than max", func() {
		s.runChecker(c, box.Number(5), []box.Value{box.Number(10), box.Number(1)}, true)
	})
}

func (s *ConstraintsTestSuite) TestNumberEvenOdd() {
	s.Run("even", func() {
		c := constraints.NumberContraintCheckers["even"]
		s.runChecker(c, box.Number(4), nil, false)
		s.runChecker(c, box.Number(3), nil, true)
		s.runChecker(c, box.String("x"), nil, true)
	})
	s.Run("odd", func() {
		c := constraints.NumberContraintCheckers["odd"]
		s.runChecker(c, box.Number(3), nil, false)
		s.runChecker(c, box.Number(4), nil, true)
	})
}

func (s *ConstraintsTestSuite) TestNumberMultipleOf() {
	c := constraints.NumberContraintCheckers["multiple_of"]
	s.runChecker(c, box.Number(12), []box.Value{box.Number(4)}, false)
	s.runChecker(c, box.Number(12), []box.Value{box.Number(5)}, true)
	s.runChecker(c, box.Number(12), []box.Value{box.Number(0)}, true)
	s.runChecker(c, box.Number(12), nil, true)
	s.runChecker(c, box.String("x"), []box.Value{box.Number(2)}, true)
	s.runChecker(c, box.Number(12), []box.Value{box.String("x")}, true)
	s.Run("negative divisor", func() {
		s.runChecker(c, box.Number(-6), []box.Value{box.Number(-3)}, false)
	})
}

func (s *ConstraintsTestSuite) TestNumberSignConstraints() {
	s.Run("positive", func() {
		c := constraints.NumberContraintCheckers["positive"]
		s.runChecker(c, box.Number(0.1), nil, false)
		s.runChecker(c, box.Number(0), nil, true)
		s.runChecker(c, box.String("x"), nil, true)
	})
	s.Run("negative", func() {
		c := constraints.NumberContraintCheckers["negative"]
		s.runChecker(c, box.Number(-1), nil, false)
		s.runChecker(c, box.Number(0), nil, true)
	})
	s.Run("non_negative", func() {
		c := constraints.NumberContraintCheckers["non_negative"]
		s.runChecker(c, box.Number(0), nil, false)
		s.runChecker(c, box.Number(-0.1), nil, true)
	})
	s.Run("non_positive", func() {
		c := constraints.NumberContraintCheckers["non_positive"]
		s.runChecker(c, box.Number(0), nil, false)
		s.runChecker(c, box.Number(0.1), nil, true)
	})
}

func (s *ConstraintsTestSuite) TestNumberFiniteInfiniteNaN() {
	s.Run("finite", func() {
		c := constraints.NumberContraintCheckers["finite"]
		s.runChecker(c, box.Number(1.5), nil, false)
		s.runChecker(c, box.Number(math.Inf(1)), nil, true)
		s.runChecker(c, box.Number(math.NaN()), nil, true)
		s.runChecker(c, box.String("x"), nil, true)
	})
	s.Run("infinite", func() {
		c := constraints.NumberContraintCheckers["infinite"]
		s.runChecker(c, box.Number(math.Inf(-1)), nil, false)
		s.runChecker(c, box.Number(1), nil, true)
	})
	s.Run("nan", func() {
		c := constraints.NumberContraintCheckers["nan"]
		s.runChecker(c, box.Number(math.NaN()), nil, false)
		s.runChecker(c, box.Number(1), nil, true)
	})
}
