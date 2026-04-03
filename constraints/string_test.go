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
	"testing"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constraints"
)

func TestStringLengthFamily(t *testing.T) {
	t.Run("length", func(t *testing.T) {
		c := constraints.StringContraintCheckers["length"]
		runChecker(t, c, box.String("abc"), []box.Value{box.Number(3)}, false)
		runChecker(t, c, box.String("ab"), []box.Value{box.Number(3)}, true)
		runChecker(t, c, box.Number(1), []box.Value{box.Number(3)}, true)
		runChecker(t, c, box.String("abc"), []box.Value{box.String("x")}, true)
		runChecker(t, c, box.String("abc"), nil, true)
		runChecker(t, c, box.String("abc"), []box.Value{box.Number(3), box.Number(3)}, true)
	})
	t.Run("minlength", func(t *testing.T) {
		c := constraints.StringContraintCheckers["minlength"]
		runChecker(t, c, box.String("abc"), []box.Value{box.Number(2)}, false)
		runChecker(t, c, box.String("a"), []box.Value{box.Number(2)}, true)
		runChecker(t, c, box.Number(1), []box.Value{box.Number(2)}, true)
	})
	t.Run("maxlength", func(t *testing.T) {
		c := constraints.StringContraintCheckers["maxlength"]
		runChecker(t, c, box.String("ab"), []box.Value{box.Number(3)}, false)
		runChecker(t, c, box.String("abcd"), []box.Value{box.Number(3)}, true)
	})
}

func TestStringRegexp(t *testing.T) {
	c := constraints.StringContraintCheckers["regexp"]
	runChecker(t, c, box.String("hello"), []box.Value{box.String(`^h.*o$`)}, false)
	runChecker(t, c, box.String("nope"), []box.Value{box.String(`^h.*o$`)}, true)
	runChecker(t, c, box.String("x"), []box.Value{box.String(`(`)}, true)
	runChecker(t, c, box.Number(1), []box.Value{box.String(`.*`)}, true)
	runChecker(t, c, box.String("x"), []box.Value{box.Number(1)}, true)
}

func TestStringPrefixSuffixSubstring(t *testing.T) {
	t.Run("starts_with", func(t *testing.T) {
		c := constraints.StringContraintCheckers["starts_with"]
		runChecker(t, c, box.String("hello"), []box.Value{box.String("he")}, false)
		runChecker(t, c, box.String("hello"), []box.Value{box.String("lo")}, true)
		runChecker(t, c, box.Number(1), []box.Value{box.String("x")}, true)
	})
	t.Run("ends_with", func(t *testing.T) {
		c := constraints.StringContraintCheckers["ends_with"]
		runChecker(t, c, box.String("hello"), []box.Value{box.String("lo")}, false)
		runChecker(t, c, box.String("hello"), []box.Value{box.String("he")}, true)
	})
	t.Run("has_substring", func(t *testing.T) {
		c := constraints.StringContraintCheckers["has_substring"]
		runChecker(t, c, box.String("hello"), []box.Value{box.String("ell")}, false)
		runChecker(t, c, box.String("hello"), []box.Value{box.String("xyz")}, true)
	})
	t.Run("not_has_substring", func(t *testing.T) {
		c := constraints.StringContraintCheckers["not_has_substring"]
		runChecker(t, c, box.String("hello"), []box.Value{box.String("xyz")}, false)
		runChecker(t, c, box.String("hello"), []box.Value{box.String("ell")}, true)
	})
}

func TestStringEmailURLUUID(t *testing.T) {
	t.Run("email", func(t *testing.T) {
		c := constraints.StringContraintCheckers["email"]
		runChecker(t, c, box.String("a@b.co"), nil, false)
		runChecker(t, c, box.String("not-an-email"), nil, true)
		runChecker(t, c, box.Number(1), nil, true)
	})
	t.Run("url", func(t *testing.T) {
		c := constraints.StringContraintCheckers["url"]
		runChecker(t, c, box.String("https://example.com/foo"), nil, false)
		runChecker(t, c, box.String("ftp://example.com/foo"), nil, true)
		runChecker(t, c, box.String("not a url"), nil, true)
	})
	t.Run("uuid", func(t *testing.T) {
		c := constraints.StringContraintCheckers["uuid"]
		runChecker(t, c, box.String("550e8400-e29b-41d4-a716-446655440000"), nil, false)
		runChecker(t, c, box.String("not-a-uuid"), nil, true)
	})
}

func TestStringAlphaNumericCase(t *testing.T) {
	t.Run("alphanumeric", func(t *testing.T) {
		c := constraints.StringContraintCheckers["alphanumeric"]
		runChecker(t, c, box.String("ab12"), nil, false)
		runChecker(t, c, box.String("a b"), nil, true)
	})
	t.Run("alpha", func(t *testing.T) {
		c := constraints.StringContraintCheckers["alpha"]
		runChecker(t, c, box.String("abZ"), nil, false)
		runChecker(t, c, box.String("a1"), nil, true)
	})
	t.Run("numeric", func(t *testing.T) {
		c := constraints.StringContraintCheckers["numeric"]
		runChecker(t, c, box.String("12"), nil, false)
		runChecker(t, c, box.String("1e10"), nil, false)
		runChecker(t, c, box.String("x"), nil, true)
	})
	t.Run("lowercase", func(t *testing.T) {
		c := constraints.StringContraintCheckers["lowercase"]
		runChecker(t, c, box.String("abc"), nil, false)
		runChecker(t, c, box.String("Ab"), nil, true)
	})
	t.Run("uppercase", func(t *testing.T) {
		c := constraints.StringContraintCheckers["uppercase"]
		runChecker(t, c, box.String("ABC"), nil, false)
		runChecker(t, c, box.String("Ab"), nil, true)
	})
	t.Run("trimmed", func(t *testing.T) {
		c := constraints.StringContraintCheckers["trimmed"]
		runChecker(t, c, box.String("abc"), nil, false)
		runChecker(t, c, box.String(" abc "), nil, true)
		runChecker(t, c, box.String("a c"), nil, false)
	})
}

func TestStringNotEmptyOneOf(t *testing.T) {
	t.Run("not_empty", func(t *testing.T) {
		c := constraints.StringContraintCheckers["not_empty"]
		runChecker(t, c, box.String("x"), nil, false)
		runChecker(t, c, box.String(""), nil, true)
		runChecker(t, c, box.Number(1), nil, true)
	})
	t.Run("one_of", func(t *testing.T) {
		c := constraints.StringContraintCheckers["one_of"]
		runChecker(t, c, box.String("b"), []box.Value{box.String("a"), box.String("b"), box.String("c")}, false)
		runChecker(t, c, box.String("z"), []box.Value{box.String("a"), box.String("b")}, true)
		runChecker(t, c, box.String("a"), []box.Value{}, true)
		runChecker(t, c, box.Number(1), []box.Value{box.String("a")}, true)
		runChecker(t, c, box.String("a"), []box.Value{box.Number(1)}, true)
	})
	t.Run("not_one_of", func(t *testing.T) {
		c := constraints.StringContraintCheckers["not_one_of"]
		runChecker(t, c, box.String("z"), []box.Value{box.String("a"), box.String("b")}, false)
		runChecker(t, c, box.String("a"), []box.Value{box.String("a"), box.String("b")}, true)
		runChecker(t, c, box.String("z"), []box.Value{}, true)
		runChecker(t, c, box.String("z"), []box.Value{box.Number(1)}, true)
	})
}
