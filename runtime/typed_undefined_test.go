// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// A typed let, a derive return, a lambda return and an optional shape field
// given undefined pass it through, as an untyped one does; a rule `default` does
// not cover a body that yields undefined, and an error raised in a `default`
// expression is the rule's error (#148). A dict[T] entry holding undefined
// counts as absent, so it passes a dict[T] check wherever one is written.

const typedUndefinedSrc = `namespace com/ex

derive missingNum = (d: document): number => { yield d.missing }

policy pol {
  let doc = { "present": 1 }
  shape Holder {
    n: number
    o?: number
  }
  shape Bag { m: dict[number] }
  rule typedLet = {
    let x: number = doc.missing
    yield x
  }
  rule untypedLet = {
    let x = doc.missing
    yield x
  }
  rule deriveReturn = { yield missingNum(doc) }
  rule lambdaReturn = {
    let f = (d: document): number => { yield d.missing }
    yield f(doc)
  }
  rule shapeField = {
    let h: Holder = { "n": 1, "o": doc.missing }
    yield h.o
  }
  rule dictIs = { yield { "a": doc.missing } is dict[number] }
  rule dictTypedLet = {
    let d: dict[number] = { "a": doc.missing, "b": 2 }
    yield d.b == 2
  }
  rule dictNested = {
    let d: list[dict[dict[number]]] = [{ "o": { "a": doc.missing } }]
    yield true
  }
  rule dictShapeField = {
    let b: Bag = { "m": { "a": doc.missing } }
    yield true
  }
  rule dictWrongValue = { yield { "a": doc.missing, "b": "x" } is dict[number] }
  rule defaultBodyUndefined = default false when true { yield doc.missing }
  rule defaultGateFalse = default false when false { yield true }
  rule defaultErrors = default 1 / 0 when false { yield true }
  export decision of typedLet
  export decision of untypedLet
  export decision of deriveReturn
  export decision of lambdaReturn
  export decision of shapeField
  export decision of defaultBodyUndefined
  export decision of dictIs
  export decision of dictTypedLet
  export decision of dictNested
  export decision of dictShapeField
  export decision of dictWrongValue
  export decision of defaultGateFalse
  export decision of defaultErrors
}
`

func (s *RuntimeTestSuite) TestTypedUndefinedPassesThrough() {
	ctx := s.T().Context()
	for _, rule := range []string{"typedLet", "untypedLet", "deriveReturn", "lambdaReturn", "shapeField", "defaultBodyUndefined"} {
		out, err := s.execRuleUnvalidated(ctx, typedUndefinedSrc, rule)
		if !s.NoError(err, rule) {
			continue
		}
		s.Equal(trinary.Unknown, out.Decision.State, rule)
	}
}

func (s *RuntimeTestSuite) TestDictEntryHoldingUndefinedIsAbsent() {
	ctx := s.T().Context()
	for _, rule := range []string{"dictIs", "dictTypedLet", "dictNested", "dictShapeField"} {
		out, err := s.execRuleUnvalidated(ctx, typedUndefinedSrc, rule)
		if !s.NoError(err, rule) {
			continue
		}
		s.Equal(trinary.True, out.Decision.State, rule)
	}

	out, err := s.execRuleUnvalidated(ctx, typedUndefinedSrc, "dictWrongValue")
	if s.NoError(err, "dictWrongValue") {
		s.Equal(trinary.False, out.Decision.State, "a real value of the wrong type still fails")
	}
}

func (s *RuntimeTestSuite) TestRuleDefaultExpressionErrorIsReturned() {
	ctx := s.T().Context()
	out, err := s.execRuleUnvalidated(ctx, typedUndefinedSrc, "defaultGateFalse")
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)

	_, err = s.execRuleUnvalidated(ctx, typedUndefinedSrc, "defaultErrors")
	s.Require().ErrorContains(err, "divide by zero")
}
