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

package box

func (s *BoxTestSuite) TestMustNumbers() {
	s.Run("returns both numbers when operands are numeric", func() {
		lhs, rhs, err := MustNumbers(Number(12), Number(2.5))
		s.Require().NoError(err)
		s.Equal(12.0, lhs)
		s.Equal(2.5, rhs)
	})
	s.Run("returns error when left operand is not a number", func() {
		lhs, rhs, err := MustNumbers(String("12"), Number(2))
		s.EqualError(err, "left operand is not a number")
		s.Zero(lhs)
		s.Zero(rhs)
	})
	s.Run("returns error when right operand is not a number", func() {
		lhs, rhs, err := MustNumbers(Number(2), Bool(true))
		s.EqualError(err, "right operand is not a number")
		s.Zero(lhs)
		s.Zero(rhs)
	})
}

func (s *BoxTestSuite) TestEqualValues() {
	s.Run("supports cross kind numeric equality only for numbers", func() {
		s.True(EqualValues(Number(42), Number(42.0)))
		s.False(EqualValues(Number(42), String("42")))
	})
	s.Run("treats undefined and null as equal within each kind", func() {
		s.True(EqualValues(Undefined(), Undefined()))
		s.True(EqualValues(Null(), Null()))
		s.False(EqualValues(Undefined(), Null()))
	})
	s.Run("compares nested lists recursively", func() {
		left := List([]Value{
			Number(1),
			List([]Value{String("x"), Number(2)}),
			Dict(map[string]Value{"k": Bool(true)}),
		})
		rightEqual := List([]Value{
			Number(1.0),
			List([]Value{String("x"), Number(2)}),
			Dict(map[string]Value{"k": Bool(true)}),
		})
		rightDifferent := List([]Value{
			Number(1),
			List([]Value{String("x"), Number(3)}),
			Dict(map[string]Value{"k": Bool(true)}),
		})
		s.True(EqualValues(left, rightEqual))
		s.False(EqualValues(left, rightDifferent))
	})
	s.Run("compares maps recursively and checks key set", func() {
		left := Dict(map[string]Value{
			"a": Number(7),
			"b": Dict(map[string]Value{"nested": String("ok")}),
		})
		rightEqual := Dict(map[string]Value{
			"a": Number(7.0),
			"b": Dict(map[string]Value{"nested": String("ok")}),
		})
		rightMissingKey := Dict(map[string]Value{"a": Number(7)})
		rightDifferentValue := Dict(map[string]Value{
			"a": Number(7),
			"b": Dict(map[string]Value{"nested": String("nope")}),
		})
		s.True(EqualValues(left, rightEqual))
		s.False(EqualValues(left, rightMissingKey))
		s.False(EqualValues(left, rightDifferentValue))
	})
	s.Run("uses document reference identity semantics", func() {
		type doc struct {
			id string
		}
		shared := &doc{id: "same"}
		equalByValue := &doc{id: "same"}
		s.True(EqualValues(Document(shared), Document(shared)))
		s.False(EqualValues(Document(shared), Document(equalByValue)))
	})
}

func (s *BoxTestSuite) TestMatchesValue() {
	s.Run("returns error when haystack is not a string", func() {
		matched, err := MatchesValue(Number(12), String("^12$"))
		s.EqualError(err, "haystack must be a string")
		s.False(matched)
	})
	s.Run("returns error when pattern is not a string", func() {
		matched, err := MatchesValue(String("12"), Number(12))
		s.EqualError(err, "pattern must be a string")
		s.False(matched)
	})
	s.Run("returns true when regex matches", func() {
		matched, err := MatchesValue(String("abc123"), String("^[a-z]+\\d+$"))
		s.Require().NoError(err)
		s.True(matched)
	})
	s.Run("returns false when regex does not match", func() {
		matched, err := MatchesValue(String("abc"), String("^\\d+$"))
		s.Require().NoError(err)
		s.False(matched)
	})
	s.Run("returns regex compile error for invalid pattern", func() {
		matched, err := MatchesValue(String("abc"), String("("))
		s.Error(err)
		s.False(matched)
	})
}

func (s *BoxTestSuite) TestContainsValue() {
	s.Run("string contains requires non empty string needle", func() {
		haystack := String("hello world")
		s.True(ContainsValue(haystack, String("world")))
		s.False(ContainsValue(haystack, String("")))
		s.False(ContainsValue(haystack, Number(1)))
	})
	s.Run("list uses semantic equality for contains", func() {
		haystack := List([]Value{
			Number(1),
			Dict(map[string]Value{"x": Number(2)}),
		})
		s.True(ContainsValue(haystack, Number(1.0)))
		s.True(ContainsValue(haystack, Dict(map[string]Value{"x": Number(2.0)})))
		s.False(ContainsValue(haystack, Dict(map[string]Value{"x": Number(3)})))
	})
	s.Run("map string needle performs key lookup", func() {
		haystack := Dict(map[string]Value{
			"k1": Number(1),
			"k2": String("v2"),
		})
		s.True(ContainsValue(haystack, String("k1")))
		s.False(ContainsValue(haystack, String("missing")))
	})
	s.Run("map needle as map checks subset semantics", func() {
		haystack := Dict(map[string]Value{
			"a": Number(1),
			"b": Dict(map[string]Value{"nested": Bool(true)}),
		})
		s.True(ContainsValue(haystack, Dict(map[string]Value{"a": Number(1.0)})))
		s.True(ContainsValue(haystack, Dict(map[string]Value{"b": Dict(map[string]Value{"nested": Bool(true)})})))
		s.False(ContainsValue(haystack, Dict(map[string]Value{"missing": Number(1)})))
		s.False(ContainsValue(haystack, Dict(map[string]Value{"b": Dict(map[string]Value{"nested": Bool(false)})})))
	})
	s.Run("map rejects non string and non map needles", func() {
		haystack := Dict(map[string]Value{
			"a": Number(1),
			"b": List([]Value{String("x")}),
		})
		s.False(ContainsValue(haystack, Number(1.0)))
		s.False(ContainsValue(haystack, List([]Value{String("x")})))
		s.False(ContainsValue(haystack, Number(2)))
		reviewerExample := Dict(map[string]Value{"a": Number(42)})
		s.False(ContainsValue(reviewerExample, Number(42)))
	})
	s.Run("returns false for unsupported haystack kinds", func() {
		s.False(ContainsValue(Number(123), Number(123)))
		s.False(ContainsValue(Bool(true), Bool(true)))
		s.False(ContainsValue(Null(), Null()))
	})
}
