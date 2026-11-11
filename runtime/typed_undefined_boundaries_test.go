// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// What #148 leaves strict: `is` and casts on a bare undefined, a typed lambda
// parameter given undefined, an absent required shape field, a required shape
// field holding undefined (undefined is not a value, so it counts as missing),
// and a default that must not mask a body error.

const typedUndefinedBoundarySrc = `namespace com/ex

policy pol {
  let doc = { "present": 1 }
  shape Holder { n: number }
  rule isShape = { yield doc.missing is Holder }
  rule castNumber = { yield doc.missing as number }
  rule lambdaParam = {
    let f = (n: number): number => { yield 1 }
    yield f(doc.missing)
  }
  rule absentField = {
    let h: Holder = { "other": 1 }
    yield h
  }
  rule undefinedField = {
    let h: Holder = { "n": doc.missing }
    yield h
  }
  rule isShapeUndefinedField = { yield { "n": doc.missing } is Holder }
  rule defaultNoMask = default false when true { yield 1 / 0 }
  export decision of isShape
  export decision of castNumber
  export decision of lambdaParam
  export decision of absentField
  export decision of undefinedField
  export decision of isShapeUndefinedField
  export decision of defaultNoMask
}
`

func (s *RuntimeTestSuite) TestTypedUndefinedStaysStrictAtBoundaries() {
	ctx := s.T().Context()

	out, err := s.execRuleUnvalidated(ctx, typedUndefinedBoundarySrc, "isShape")
	if s.NoError(err, "isShape") {
		s.NotEqual(trinary.True, out.Decision.State, "undefined is not a Holder")
	}

	_, err = s.execRuleUnvalidated(ctx, typedUndefinedBoundarySrc, "castNumber")
	s.Error(err, "castNumber: a cast rejects undefined")

	_, err = s.execRuleUnvalidated(ctx, typedUndefinedBoundarySrc, "lambdaParam")
	s.Error(err, "lambdaParam: a required typed parameter given undefined still errors")

	_, err = s.execRuleUnvalidated(ctx, typedUndefinedBoundarySrc, "absentField")
	s.Error(err, "absentField: an absent required shape field still fails")

	_, err = s.execRuleUnvalidated(ctx, typedUndefinedBoundarySrc, "undefinedField")
	if s.Error(err, "undefinedField: a required shape field holding undefined is missing") {
		s.Contains(err.Error(), "field n is required")
	}

	out, err = s.execRuleUnvalidated(ctx, typedUndefinedBoundarySrc, "isShapeUndefinedField")
	if s.NoError(err, "isShapeUndefinedField") {
		s.Equal(trinary.False, out.Decision.State, "a required field holding undefined is not a Holder")
	}

	_, err = s.execRuleUnvalidated(ctx, typedUndefinedBoundarySrc, "defaultNoMask")
	s.Error(err, "defaultNoMask: default never masks a body error")
}
