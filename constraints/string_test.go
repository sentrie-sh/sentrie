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
)

func (s *ConstraintsTestSuite) TestStringLengthFamily() {
	s.Run("length", func() {
		c := constraints.StringContraintCheckers["length"]
		s.runChecker(c, box.String("abc"), []box.Value{box.Number(3)}, false)
		s.runChecker(c, box.String("ab"), []box.Value{box.Number(3)}, true)
		s.runChecker(c, box.Number(1), []box.Value{box.Number(3)}, true)
		s.runChecker(c, box.String("abc"), []box.Value{box.String("x")}, true)
		s.runChecker(c, box.String("abc"), nil, true)
		s.runChecker(c, box.String("abc"), []box.Value{box.Number(3), box.Number(3)}, true)
	})
	s.Run("minlength", func() {
		c := constraints.StringContraintCheckers["minlength"]
		s.runChecker(c, box.String("abc"), []box.Value{box.Number(2)}, false)
		s.runChecker(c, box.String("a"), []box.Value{box.Number(2)}, true)
		s.runChecker(c, box.Number(1), []box.Value{box.Number(2)}, true)
	})
	s.Run("maxlength", func() {
		c := constraints.StringContraintCheckers["maxlength"]
		s.runChecker(c, box.String("ab"), []box.Value{box.Number(3)}, false)
		s.runChecker(c, box.String("abcd"), []box.Value{box.Number(3)}, true)
	})
}

func (s *ConstraintsTestSuite) TestStringRegexp() {
	c := constraints.StringContraintCheckers["regexp"]
	s.runChecker(c, box.String("hello"), []box.Value{box.String(`^h.*o$`)}, false)
	s.runChecker(c, box.String("nope"), []box.Value{box.String(`^h.*o$`)}, true)
	s.runChecker(c, box.String("x"), []box.Value{box.String(`(`)}, true)
	s.runChecker(c, box.Number(1), []box.Value{box.String(`.*`)}, true)
	s.runChecker(c, box.String("x"), []box.Value{box.Number(1)}, true)
}

func (s *ConstraintsTestSuite) TestStringPrefixSuffixSubstring() {
	s.Run("starts_with", func() {
		c := constraints.StringContraintCheckers["starts_with"]
		s.runChecker(c, box.String("hello"), []box.Value{box.String("he")}, false)
		s.runChecker(c, box.String("hello"), []box.Value{box.String("lo")}, true)
		s.runChecker(c, box.Number(1), []box.Value{box.String("x")}, true)
	})
	s.Run("ends_with", func() {
		c := constraints.StringContraintCheckers["ends_with"]
		s.runChecker(c, box.String("hello"), []box.Value{box.String("lo")}, false)
		s.runChecker(c, box.String("hello"), []box.Value{box.String("he")}, true)
	})
	s.Run("has_substring", func() {
		c := constraints.StringContraintCheckers["has_substring"]
		s.runChecker(c, box.String("hello"), []box.Value{box.String("ell")}, false)
		s.runChecker(c, box.String("hello"), []box.Value{box.String("xyz")}, true)
	})
	s.Run("not_has_substring", func() {
		c := constraints.StringContraintCheckers["not_has_substring"]
		s.runChecker(c, box.String("hello"), []box.Value{box.String("xyz")}, false)
		s.runChecker(c, box.String("hello"), []box.Value{box.String("ell")}, true)
	})
}

func (s *ConstraintsTestSuite) TestStringEmailURLUUID() {
	s.Run("email", func() {
		c := constraints.StringContraintCheckers["email"]
		s.runChecker(c, box.String("a@b.co"), nil, false)
		s.runChecker(c, box.String("not-an-email"), nil, true)
		s.runChecker(c, box.Number(1), nil, true)
	})
	s.Run("url", func() {
		c := constraints.StringContraintCheckers["url"]
		s.runChecker(c, box.String("https://example.com/foo"), nil, false)
		s.runChecker(c, box.String("ftp://example.com/foo"), nil, true)
		s.runChecker(c, box.String("not a url"), nil, true)
	})
	s.Run("uuid", func() {
		c := constraints.StringContraintCheckers["uuid"]
		s.runChecker(c, box.String("550e8400-e29b-41d4-a716-446655440000"), nil, false)
		s.runChecker(c, box.String("not-a-uuid"), nil, true)
	})
}

func (s *ConstraintsTestSuite) TestStringAlphaNumericCase() {
	s.Run("alphanumeric", func() {
		c := constraints.StringContraintCheckers["alphanumeric"]
		s.runChecker(c, box.String("ab12"), nil, false)
		s.runChecker(c, box.String("a b"), nil, true)
	})
	s.Run("alpha", func() {
		c := constraints.StringContraintCheckers["alpha"]
		s.runChecker(c, box.String("abZ"), nil, false)
		s.runChecker(c, box.String("a1"), nil, true)
	})
	s.Run("numeric", func() {
		c := constraints.StringContraintCheckers["numeric"]
		s.runChecker(c, box.String("12"), nil, false)
		s.runChecker(c, box.String("1e10"), nil, false)
		s.runChecker(c, box.String("x"), nil, true)
	})
	s.Run("lowercase", func() {
		c := constraints.StringContraintCheckers["lowercase"]
		s.runChecker(c, box.String("abc"), nil, false)
		s.runChecker(c, box.String("Ab"), nil, true)
	})
	s.Run("uppercase", func() {
		c := constraints.StringContraintCheckers["uppercase"]
		s.runChecker(c, box.String("ABC"), nil, false)
		s.runChecker(c, box.String("Ab"), nil, true)
	})
	s.Run("trimmed", func() {
		c := constraints.StringContraintCheckers["trimmed"]
		s.runChecker(c, box.String("abc"), nil, false)
		s.runChecker(c, box.String(" abc "), nil, true)
		s.runChecker(c, box.String("a c"), nil, false)
	})
}

func (s *ConstraintsTestSuite) TestStringNotEmptyOneOf() {
	s.Run("not_empty", func() {
		c := constraints.StringContraintCheckers["not_empty"]
		s.runChecker(c, box.String("x"), nil, false)
		s.runChecker(c, box.String(""), nil, true)
		s.runChecker(c, box.Number(1), nil, true)
	})
	s.Run("one_of", func() {
		c := constraints.StringContraintCheckers["one_of"]
		s.runChecker(c, box.String("b"), []box.Value{box.String("a"), box.String("b"), box.String("c")}, false)
		s.runChecker(c, box.String("z"), []box.Value{box.String("a"), box.String("b")}, true)
		s.runChecker(c, box.String("a"), []box.Value{}, true)
		s.runChecker(c, box.Number(1), []box.Value{box.String("a")}, true)
		s.runChecker(c, box.String("a"), []box.Value{box.Number(1)}, true)
	})
	s.Run("not_one_of", func() {
		c := constraints.StringContraintCheckers["not_one_of"]
		s.runChecker(c, box.String("z"), []box.Value{box.String("a"), box.String("b")}, false)
		s.runChecker(c, box.String("a"), []box.Value{box.String("a"), box.String("b")}, true)
		s.runChecker(c, box.String("z"), []box.Value{}, true)
		s.runChecker(c, box.String("z"), []box.Value{box.Number(1)}, true)
	})
}
