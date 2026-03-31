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
	"github.com/sentrie-sh/sentrie/trinary"
)

// Exercises remaining branches for 100% statement coverage (wrong arg counts after
// type checks, and wrong value kinds for zero-arg checkers).

func TestNumberRemainingBranches(t *testing.T) {
	t.Run("wrong arg count or arg type", func(t *testing.T) {
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
			t.Run(r.key, func(t *testing.T) {
				c := constraints.NumberContraintCheckers[r.key]
				runChecker(t, c, r.val, r.args, true)
			})
		}
	})

	t.Run("zero-arg non-number val", func(t *testing.T) {
		bad := box.String("x")
		for _, key := range []string{
			"even", "odd", "positive", "negative", "non_negative", "non_positive",
			"finite", "infinite", "nan",
		} {
			t.Run(key, func(t *testing.T) {
				c := constraints.NumberContraintCheckers[key]
				runChecker(t, c, bad, nil, true)
			})
		}
	})

	t.Run("finite with inf and nan", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["finite"]
		runChecker(t, c, box.Number(math.Inf(1)), nil, true)
		runChecker(t, c, box.Number(math.NaN()), nil, true)
	})

	t.Run("infinite with non-inf", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["infinite"]
		runChecker(t, c, box.Number(math.NaN()), nil, true)
	})

	t.Run("nan with non-nan", func(t *testing.T) {
		c := constraints.NumberContraintCheckers["nan"]
		runChecker(t, c, box.Number(math.Inf(1)), nil, true)
	})
}

func TestStringRemainingBranches(t *testing.T) {
	t.Run("length minlength maxlength arg type", func(t *testing.T) {
		runChecker(t, constraints.StringContraintCheckers["minlength"],
			box.String("abc"), []box.Value{box.String("x")}, true)
		runChecker(t, constraints.StringContraintCheckers["maxlength"],
			box.Number(1), []box.Value{box.Number(3)}, true)
		runChecker(t, constraints.StringContraintCheckers["maxlength"],
			box.String("ab"), []box.Value{box.String("x")}, true)
	})

	t.Run("wrong arg count with string val", func(t *testing.T) {
		s := box.String("ab")
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
			t.Run(r.key, func(t *testing.T) {
				c := constraints.StringContraintCheckers[r.key]
				runChecker(t, c, s, r.args, true)
			})
		}
	})

	t.Run("prefix suffix substring arg not string or val not string", func(t *testing.T) {
		runChecker(t, constraints.StringContraintCheckers["starts_with"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
		runChecker(t, constraints.StringContraintCheckers["ends_with"],
			box.Number(1), []box.Value{box.String("x")}, true)
		runChecker(t, constraints.StringContraintCheckers["ends_with"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
		runChecker(t, constraints.StringContraintCheckers["has_substring"],
			box.Number(1), []box.Value{box.String("x")}, true)
		runChecker(t, constraints.StringContraintCheckers["has_substring"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
		runChecker(t, constraints.StringContraintCheckers["not_has_substring"],
			box.Number(1), []box.Value{box.String("x")}, true)
		runChecker(t, constraints.StringContraintCheckers["not_has_substring"],
			box.String("hi"), []box.Value{box.Number(1)}, true)
	})

	t.Run("non-string val", func(t *testing.T) {
		bad := box.Number(1)
		for _, key := range []string{
			"email", "url", "uuid", "alphanumeric", "alpha", "numeric",
			"lowercase", "uppercase", "trimmed", "not_empty",
		} {
			t.Run(key, func(t *testing.T) {
				c := constraints.StringContraintCheckers[key]
				runChecker(t, c, bad, nil, true)
			})
		}
	})

	t.Run("not_one_of non-string val", func(t *testing.T) {
		c := constraints.StringContraintCheckers["not_one_of"]
		runChecker(t, c, box.Number(1), []box.Value{box.String("a")}, true)
	})
}

func TestTrinaryRemainingBranches(t *testing.T) {
	t.Run("neq wrong arg count", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["neq"]
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{
			box.Trinary(trinary.False), box.Trinary(trinary.True),
		}, true)
	})

	t.Run("neq non-trinary val", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["neq"]
		runChecker(t, c, box.String("x"), []box.Value{box.Trinary(trinary.False)}, true)
	})

	t.Run("is_true non-trinary val", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["is_true"]
		runChecker(t, c, box.String("x"), nil, true)
	})

	t.Run("is_false non-trinary val", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["is_false"]
		runChecker(t, c, box.Number(1), nil, true)
	})
}
