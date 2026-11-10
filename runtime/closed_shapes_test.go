// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

const closedShapesSrc = `namespace t

shape Inner { x: number }
shape Closed { name: string }
shape Open { name: string ... }
shape Outer { inner: Inner }
shape OuterOpenInner { inner: Inner ... }
shape OpenOuter { inner: Inner ... }
shape OpenBase { a: number ... }
shape ClosedChildOfOpen with OpenBase { b: number }
shape ClosedBase { a: number }
shape OpenChildOfClosed with ClosedBase { b: number ... }
shape Tagged { v: number @min(2) }
shape Dicts { d: dict[Inner] }

policy p {
  let okClosed: Closed = {"name": "a"}
  let extraClosed = {"name": "a", "extra": 1}
  let extraOpen = {"name": "a", "extra": 1}
  let nestedExtra = {"inner": {"x": 1, "y": 2}}
  let nestedOK = {"inner": {"x": 1}}
  let outerExtra = {"inner": {"x": 1}, "z": 1}
  let childExtraBase = {"a": 1, "b": 2, "c": 3}
  let childClosedOK = {"a": 1, "b": 2}
  let closedChildExtra = {"a": 1, "b": 2, "c": 3}
  let dictBad = {"d": {"k": {"x": 1, "y": 2}}}
  let listBad = [{"x": 1, "y": 2}]
  let dupKeys = {"a": 1, "a": 2}

  rule closed_ok = { let v: Closed = okClosed  yield v.name == "a" }
  rule closed_extra = { let v: Closed = extraClosed  yield v is defined }
  rule open_extra = { let v: Open = extraOpen  yield v.name == "a" }
  rule nested_ok = { let v: Outer = nestedOK  yield v.inner.x == 1 }
  rule nested_extra = { let v: Outer = nestedExtra  yield v is defined }
  rule outer_extra = { let v: Outer = outerExtra  yield v is defined }
  rule open_outer_closed_inner = { let v: OpenOuter = nestedExtra  yield v is defined }
  rule open_outer_extra = { let v: OpenOuter = outerExtra  yield v.inner.x == 1 }
  rule open_base_not_inherited = { let v: ClosedChildOfOpen = childExtraBase  yield v is defined }
  rule open_base_child_fields_ok = { let v: ClosedChildOfOpen = childClosedOK  yield v.b == 2 }
  rule open_child_of_closed = { let v: OpenChildOfClosed = closedChildExtra  yield v.c == 3 }
  rule dict_value_closed = { let v: Dicts = dictBad  yield v is defined }
  rule dict_value_bad_type = { let v: dict[number] = {"a": "x"}  yield v is defined }
  rule dict_value_constraint = { let v: dict[number @min(2)] = {"a": 1}  yield v is defined }
  rule list_item_closed = { let v: list[Inner] = listBad  yield v is defined }
  rule dup_key_last_wins = { yield dupKeys.a == 2 }

  export decision of closed_ok
  export decision of closed_extra
  export decision of open_extra
  export decision of nested_ok
  export decision of nested_extra
  export decision of outer_extra
  export decision of open_outer_closed_inner
  export decision of open_outer_extra
  export decision of open_base_not_inherited
  export decision of open_base_child_fields_ok
  export decision of open_child_of_closed
  export decision of dict_value_closed
  export decision of dict_value_bad_type
  export decision of dict_value_constraint
  export decision of list_item_closed
  export decision of dup_key_last_wins
}
`

// TestClosedShapes covers #145: shapes reject undeclared fields unless their
// own body ends in `...`, at every depth, and openness is per declaration.
func (s *RuntimeTestSuite) TestClosedShapes() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "closed.sentrie", src: closedShapesSrc})

	accepted := []string{
		"closed_ok",                 // declared fields only
		"open_extra",                // `...` accepts an undeclared field
		"nested_ok",                 // nested shape with declared fields only
		"open_outer_extra",          // an open outer shape accepts its own extra field
		"open_base_child_fields_ok", // child of an open base still accepts the declared fields
		"open_child_of_closed",      // an open child of a closed base is open
	}
	for _, rule := range accepted {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "t", "p", rule, nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}

	rejected := map[string]string{
		"closed_extra":            "field extra is not declared in shape t/Closed",
		"nested_extra":            "field y is not declared in shape t/Inner",
		"outer_extra":             "field z is not declared in shape t/Outer",
		"open_outer_closed_inner": "field y is not declared in shape t/Inner", // openness is not inherited by nested shapes
		"open_base_not_inherited": "field c is not declared in shape t/ClosedChildOfOpen",
		"dict_value_closed":       "field y is not declared in shape t/Inner",
		"list_item_closed":        "field y is not declared in shape t/Inner",
	}
	for rule, msg := range rejected {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "t", "p", rule, nil)
			s.Require().Error(err)
			s.Contains(err.Error(), msg)
		})
	}
}

// TestClosedShapeUndeclaredFieldReportedAfterDeclaredOnes pins the order of
// checks: a missing or wrongly typed declared field is reported before an
// undeclared one.
func (s *RuntimeTestSuite) TestClosedShapeUndeclaredFieldReportedAfterDeclaredOnes() {
	ctx := s.T().Context()
	src := `namespace t
shape S { a: number }
policy p {
  let missing = {"extra": 1}
  rule r = { let v: S = missing  yield v is defined }
  export decision of r
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "order.sentrie", src: src})
	_, err := exec.ExecRule(ctx, "t", "p", "r", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "field a is required")
	s.NotContains(err.Error(), "not declared")
}

// TestDictValuesCheckedAgainstElementType pins what #299 already did on the
// base: each dict[T] value is checked, constraints included, and the error
// names the key. It passes without the #145 change (noted in the journal).
func (s *RuntimeTestSuite) TestDictValuesCheckedAgainstElementType() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "closed.sentrie", src: closedShapesSrc})
	for _, rule := range []string{"dict_value_bad_type", "dict_value_constraint"} {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "t", "p", rule, nil)
			s.Require().Error(err)
			s.Contains(err.Error(), `value for key "a"`)
		})
	}
}

// TestDuplicateDictLiteralKeyLastWins pins that a duplicate key in a dict
// literal is not an error and the last value wins (#145, unchanged).
func (s *RuntimeTestSuite) TestDuplicateDictLiteralKeyLastWins() {
	ctx := s.T().Context()
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "closed.sentrie", src: closedShapesSrc})
	out, err := exec.ExecRule(ctx, "t", "p", "dup_key_last_wins", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}
