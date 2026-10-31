// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import "fmt"

// TestSetBuiltinsAreArityAndKindChecked covers #162 at validate time: the set
// builtins reject a wrong argument count or a literal of the wrong kind before
// anything runs.
func (s *IndexTestSuite) TestSetBuiltinsAreArityAndKindChecked() {
	cases := []struct {
		name string
		expr string
		want string
	}{
		{"set no args", `set()`, "set requires 1 argument"},
		{"set two args", `set([1], [2])`, "set requires 1 argument"},
		{"set string", `set("abc")`, "set: argument must be a list"},
		{"union one arg", `union([1])`, "union requires 2 arguments"},
		{"union three args", `union([1], [2], [3])`, "union requires 2 arguments"},
		{"union first number", `union(1, [2])`, "union: first argument must be a set or list"},
		{"union second string", `union([1], "x")`, "union: second argument must be a set or list"},
		{"intersection one arg", `intersection([1])`, "intersection requires 2 arguments"},
		{"intersection second dict", `intersection([1], {"a": 1})`, "intersection: second argument must be a set or list"},
		{"difference first bool", `difference(true, [1])`, "difference: first argument must be a set or list"},
		{"is_subset one arg", `is_subset([1])`, "is_subset requires 2 arguments"},
		{"is_subset first string", `is_subset("a", [1])`, "is_subset: first argument must be a set or list"},
		{"is_superset three args", `is_superset([1], [2], [3])`, "is_superset requires 2 arguments"},
		{"is_superset second number", `is_superset([1], 2)`, "is_superset: second argument must be a set or list"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			err := s.parseAndValidate(fmt.Sprintf(`namespace com/ex
policy p {
  rule r = { yield %s }
  export decision of r
}
`, tc.expr), "f.sentrie")
			s.Require().Error(err)
			s.Contains(err.Error(), tc.want)
		})
	}
}

// TestSetBuiltinsAcceptSetsListsAndTypedFacts: a set[T] or list[T] fact, a list
// literal and a set() call are all valid operands; a string fact is not, so
// the set[T] typeref resolves to the list kind.
func (s *IndexTestSuite) TestSetBuiltinsAcceptSetsListsAndTypedFacts() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  fact a: set[string]
  fact b: list[string]
  rule r = { yield is_subset(a, b) and is_superset(union(a, b), set(b)) and count(difference(a, ["x"])) > 0 }
  rule q = { yield count(intersection(set(["x"]), a)) == 0 }
  export decision of r
  export decision of q
}
`, "f.sentrie")
	s.Require().NoError(err)

	err = s.parseAndValidate(`namespace com/ex
policy p {
  fact a: set[string]
  fact s: string
  rule r = { yield is_subset(a, s) }
  export decision of r
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "is_subset: second argument must be a set or list")

	// a set[T] fact is a list, so a builtin that wants a dict rejects it
	err = s.parseAndValidate(`namespace com/ex
policy p {
  fact a: set[string]
  rule r = { yield count(merge(a, {"x": 1})) }
  export decision of r
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "first argument is not a dict")
}

// TestIsSetTypeChecksItsElementShape: `x is set[T]` resolves T like any other
// type, so an unknown shape is a validate-time error.
func (s *IndexTestSuite) TestIsSetTypeChecksItsElementShape() {
	err := s.parseAndValidate(`namespace com/ex
policy p {
  rule r = { yield [1] is set[number] }
  export decision of r
}
`, "f.sentrie")
	s.Require().NoError(err)

	err = s.parseAndValidate(`namespace com/ex
policy p {
  rule r = { yield [1] is set[Missing] }
  export decision of r
}
`, "f.sentrie")
	s.Require().Error(err)
	s.Contains(err.Error(), "Missing")
}
