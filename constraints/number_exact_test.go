// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints_test

import (
	"math"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
)

// TestNumberConstraintsAreExactForLargeIntegers covers #166: integers above
// 2^53 are compared and divided exactly, where float64 would merge neighbours.
func (s *ConstraintsTestSuite) TestNumberConstraintsAreExactForLargeIntegers() {
	const big = int64(1)<<53 + 1 // 9007199254740993
	type row struct {
		name    string
		key     string
		val     box.Value
		args    []box.Value
		wantErr bool
	}
	rows := []row{
		{"min above", "min", box.Int(big), []box.Value{box.Int(big - 1)}, false},
		{"min below", "min", box.Int(big - 1), []box.Value{box.Int(big)}, true},
		{"max below", "max", box.Int(big - 1), []box.Value{box.Int(big)}, false},
		{"max above", "max", box.Int(big), []box.Value{box.Int(big - 1)}, true},
		{"eq neighbour", "eq", box.Int(big), []box.Value{box.Int(big - 1)}, true},
		{"eq same", "eq", box.Int(big), []box.Value{box.Int(big)}, false},
		{"eq int and float", "eq", box.Int(5), []box.Value{box.Float(5)}, false},
		{"neq neighbour", "neq", box.Int(big), []box.Value{box.Int(big - 1)}, false},
		{"gt neighbour", "gt", box.Int(big), []box.Value{box.Int(big - 1)}, false},
		{"lt neighbour", "lt", box.Int(big - 1), []box.Value{box.Int(big)}, false},
		{"gt against rounded float", "gt", box.Int(big), []box.Value{box.Float(float64(big - 1))}, false},
		{"in exact", "in", box.Int(big), []box.Value{box.List([]box.Value{box.Int(big)})}, false},
		{"in neighbour only", "in", box.Int(big), []box.Value{box.List([]box.Value{box.Int(big - 1)})}, true},
		{"in scalar", "in", box.Int(3), []box.Value{box.Float(3)}, false},
		{"not_in neighbour only", "not_in", box.Int(big), []box.Value{box.List([]box.Value{box.Int(big - 1)})}, false},
		{"range inside", "range", box.Int(big), []box.Value{box.Int(big), box.Int(big + 1)}, false},
		{"range below", "range", box.Int(big - 1), []box.Value{box.Int(big), box.Int(big + 1)}, true},
		{"range inverted", "range", box.Int(1), []box.Value{box.Int(big), box.Int(big - 1)}, true},
		{"odd big", "odd", box.Int(big), nil, false},
		{"even big", "even", box.Int(big), nil, true},
		{"even big minus one", "even", box.Int(big - 1), nil, false},
		{"multiple_of exact", "multiple_of", box.Int(big), []box.Value{box.Int(3)}, false},
		{"multiple_of exact miss", "multiple_of", box.Int(big + 1), []box.Value{box.Int(3)}, true},
		{"multiple_of max int", "multiple_of", box.Int(math.MaxInt64), []box.Value{box.Int(7)}, false},
		{"multiple_of mixed float", "multiple_of", box.Int(9), []box.Value{box.Float(1.5)}, false},
		{"multiple_of integral float divisor", "multiple_of", box.Int(big), []box.Value{box.Float(2)}, true},
		{"multiple_of integral float divisor hit", "multiple_of", box.Int(big - 1), []box.Value{box.Float(2)}, false},
		{"multiple_of integral float value", "multiple_of", box.Float(float64(big - 1)), []box.Value{box.Int(3)}, true},
		{"even integral float", "even", box.Float(float64(big - 1)), nil, false},
		{"odd integral float", "odd", box.Float(3), nil, false},
		{"min NaN never fails", "min", box.Float(math.NaN()), []box.Value{box.Int(1)}, false},
		{"eq NaN fails", "eq", box.Float(math.NaN()), []box.Value{box.Int(1)}, true},
	}
	for _, r := range rows {
		s.Run(r.name, func() {
			s.runChecker(constraints.NumberContraintCheckers[r.key], r.val, r.args, r.wantErr)
		})
	}
}
