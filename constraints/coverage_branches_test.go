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
	"github.com/sentrie-sh/sentrie/trinary"
)

// Exercises remaining branches for 100% statement coverage (wrong arg counts after
// type checks, and wrong value kinds for zero-arg checkers).

func (s *ConstraintsTestSuite) TestNumberRemainingBranches() {
	s.Run("wrong arg count or arg type", func() {
		type row struct {
			key  string
			val  box.Value
			args []box.Value
		}
		rows := []row{
			{"max", box.Number(1), []box.Value{box.String("x")}},
			{"max", box.String("x"), []box.Value{box.Number(5)}},
			{"eq", box.Number(1), []box.Value{box.String("x")}},
			{"eq", box.String("x"), []box.Value{box.Number(1)}},
			{"eq", box.Number(1), []box.Value{box.Number(1), box.Number(2)}},
			{"neq", box.Number(1), []box.Value{box.String("x")}},
			{"neq", box.String("x"), []box.Value{box.Number(1)}},
			{"neq", box.Number(1), []box.Value{box.Number(1), box.Number(2)}},
			{"gt", box.Number(1), []box.Value{box.String("x")}},
			{"gt", box.String("x"), []box.Value{box.Number(1)}},
			{"gt", box.Number(1), []box.Value{box.Number(1), box.Number(2)}},
			{"lt", box.Number(1), []box.Value{box.String("x")}},
			{"lt", box.String("x"), []box.Value{box.Number(1)}},
			{"lt", box.Number(1), []box.Value{box.Number(1), box.Number(2)}},
			{"not_in", box.Number(1), []box.Value{box.Number(1), box.Number(2)}},
		}
		for _, r := range rows {
			s.Run(r.key, func() {
				c := constraints.NumberContraintCheckers[r.key]
				s.runChecker(c, r.val, r.args, true)
			})
		}
	})
	s.Run("zero-arg non-number val", func() {
		bad := box.String("x")
		for _, key := range []string{
			"even", "odd", "positive", "negative", "non_negative", "non_positive",
			"finite", "infinite", "nan",
		} {
			s.Run(key, func() {
				c := constraints.NumberContraintCheckers[key]
				s.runChecker(c, bad, nil, true)
			})
		}
	})
	s.Run("finite with inf and nan", func() {
		c := constraints.NumberContraintCheckers["finite"]
		s.runChecker(c, box.Number(math.Inf(1)), nil, true)
		s.runChecker(c, box.Number(math.NaN()), nil, true)
	})
	s.Run("infinite with non-inf", func() {
		c := constraints.NumberContraintCheckers["infinite"]
		s.runChecker(c, box.Number(math.NaN()), nil, true)
	})
	s.Run("nan with non-nan", func() {
		c := constraints.NumberContraintCheckers["nan"]
		s.runChecker(c, box.Number(math.Inf(1)), nil, true)
	})
}

func (s *ConstraintsTestSuite) TestStringRemainingBranches() {
	s.Run("length minlength maxlength arg type", func() {
		s.runChecker(constraints.StringContraintCheckers["minlength"],
			box.String("abc"), []box.Value{box.String("x")}, true)
		s.runChecker(constraints.StringContraintCheckers["maxlength"],
			box.Number(1), []box.Value{box.Number(3)}, true)
		s.runChecker(constraints.StringContraintCheckers["maxlength"],
			box.String("ab"), []box.Value{box.String("x")}, true)
	})
	s.Run("wrong arg count with string val", func() {
		str := box.String("ab")
		type row struct {
			key  string
			args []box.Value
		}
		rows := []row{
			{"minlength", nil},
			{"maxlength", []box.Value{box.Number(1), box.Number(2)}},
			{"regexp", nil},
			{"starts_with", []box.Value{}},
			{"ends_with", []box.Value{box.String("a"), box.String("b")}},
			{"has_substring", nil},
			{"not_has_substring", []box.Value{box.String("a"), box.String("b")}},
		}
		for _, r := range rows {
			s.Run(r.key, func() {
				c := constraints.StringContraintCheckers[r.key]
				s.runChecker(c, str, r.args, true)
			})
		}
	})
	s.Run("prefix suffix substring arg not string or val not string", func() {
		s.runChecker(constraints.StringContraintCheckers["starts_with"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
		s.runChecker(constraints.StringContraintCheckers["ends_with"],
			box.Number(1), []box.Value{box.String("x")}, true)
		s.runChecker(constraints.StringContraintCheckers["ends_with"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
		s.runChecker(constraints.StringContraintCheckers["has_substring"],
			box.Number(1), []box.Value{box.String("x")}, true)
		s.runChecker(constraints.StringContraintCheckers["has_substring"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
		s.runChecker(constraints.StringContraintCheckers["not_has_substring"],
			box.Number(1), []box.Value{box.String("x")}, true)
		s.runChecker(constraints.StringContraintCheckers["not_has_substring"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
	})
	s.Run("non-string val", func() {
		bad := box.Number(1)
		for _, key := range []string{
			"email", "url", "uuid", "alphanumeric", "alpha", "numeric",
			"lowercase", "uppercase", "trimmed", "not_empty",
		} {
			s.Run(key, func() {
				c := constraints.StringContraintCheckers[key]
				s.runChecker(c, bad, nil, true)
			})
		}
	})
	s.Run("not_one_of non-string val", func() {
		c := constraints.StringContraintCheckers["not_one_of"]
		s.runChecker(c, box.Number(1), []box.Value{box.String("a")}, true)
	})
}

func (s *ConstraintsTestSuite) TestTrinaryRemainingBranches() {
	s.Run("neq wrong arg count", func() {
		c := constraints.TrinaryConstraintCheckers["neq"]
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{
			box.Trinary(trinary.False), box.Trinary(trinary.True),
		}, true)
	})
	s.Run("neq non-trinary val", func() {
		c := constraints.TrinaryConstraintCheckers["neq"]
		s.runChecker(c, box.String("x"), []box.Value{box.Trinary(trinary.False)}, true)
	})
	s.Run("is_true non-trinary val", func() {
		c := constraints.TrinaryConstraintCheckers["is_true"]
		s.runChecker(c, box.String("x"), nil, true)
	})
	s.Run("is_false non-trinary val", func() {
		c := constraints.TrinaryConstraintCheckers["is_false"]
		s.runChecker(c, box.Number(1), nil, true)
	})
}
