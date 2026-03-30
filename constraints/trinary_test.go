// SPDX-License-Identifier: Apache-2.0
//
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

package constraints_test

import (
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
	"github.com/sentrie-sh/sentrie/trinary"
)

func TestTrinaryNotUnknown(t *testing.T) {
	c := constraints.TrinaryConstraintCheckers["not_unknown"]
	runChecker(t, c, box.Trinary(trinary.True), nil, false)
	runChecker(t, c, box.Trinary(trinary.Unknown), nil, true)
	runChecker(t, c, box.String("x"), nil, true)
}

func TestTrinaryEqNeq(t *testing.T) {
	t.Run("eq", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["eq"]
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.True)}, false)
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.False)}, true)
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.True), box.Trinary(trinary.False)}, true)
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{box.String("x")}, true)
		runChecker(t, c, box.String("x"), []box.Value{box.Trinary(trinary.True)}, true)
	})
	t.Run("neq", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["neq"]
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.False)}, false)
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{box.Trinary(trinary.True)}, true)
		runChecker(t, c, box.Trinary(trinary.True), []box.Value{box.String("x")}, true)
	})
}

func TestTrinaryIsTrueIsFalse(t *testing.T) {
	t.Run("is_true", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["is_true"]
		runChecker(t, c, box.Trinary(trinary.True), nil, false)
		runChecker(t, c, box.Trinary(trinary.False), nil, true)
		runChecker(t, c, box.Trinary(trinary.Unknown), nil, true)
	})
	t.Run("is_false", func(t *testing.T) {
		c := constraints.TrinaryConstraintCheckers["is_false"]
		runChecker(t, c, box.Trinary(trinary.False), nil, false)
		runChecker(t, c, box.Trinary(trinary.True), nil, true)
	})
}
