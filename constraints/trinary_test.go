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
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/trinary"
)

func (s *ConstraintsTestSuite) TestTrinaryNotUnknown() {
	c := constraints.TrinaryConstraintCheckers["not_unknown"]
	s.runChecker(c, box.Trinary(trinary.True), nil, false)
	s.runChecker(c, box.Trinary(trinary.Unknown), nil, true)
	s.runChecker(c, box.String("x"), nil, true)
}

func (s *ConstraintsTestSuite) TestTrinaryEqNeq() {
	s.Run("eq", func() {
		c := constraints.TrinaryConstraintCheckers["eq"]
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.True)}, false)
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.False)}, true)
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.True), box.Trinary(trinary.False)}, true)
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{box.String("x")}, true)
		s.runChecker(c, box.String("x"), []box.Value{box.Trinary(trinary.True)}, true)
	})
	s.Run("neq", func() {
		c := constraints.TrinaryConstraintCheckers["neq"]
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.False)}, false)
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.True)}, true)
		s.runChecker(c, box.Trinary(trinary.True), []box.Value{box.String("x")}, true)
	})
}

func (s *ConstraintsTestSuite) TestTrinaryIsTrueIsFalse() {
	s.Run("is_true", func() {
		c := constraints.TrinaryConstraintCheckers["is_true"]
		s.runChecker(c, box.Trinary(trinary.True), nil, false)
		s.runChecker(c, box.Trinary(trinary.False), nil, true)
		s.runChecker(c, box.Trinary(trinary.Unknown), nil, true)
	})
	s.Run("is_false", func() {
		c := constraints.TrinaryConstraintCheckers["is_false"]
		s.runChecker(c, box.Trinary(trinary.False), nil, false)
		s.runChecker(c, box.Trinary(trinary.True), nil, true)
	})
}
