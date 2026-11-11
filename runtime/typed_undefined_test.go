// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"

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

const dictUndefinedPositionsSrc = `namespace com/ex

shape Item { n: number }

policy pol {
  let doc = { "present": 1 }
  rule lambdaParam = {
    let f = (d: dict[number]): boolean => { yield true }
    yield f({ "a": doc.missing })
  }
  rule lambdaReturnDict = {
    let f = (d: document): dict[number] => { yield { "a": d.missing, "b": 2 } }
    yield f(doc).b == 2
  }
  rule unionDict = { yield { "a": doc.missing } is dict[number] | string }
  rule listOfDict = { yield [{ "a": doc.missing }] is list[dict[number]] }
  rule dictOfShape = {
    let d: dict[Item] = { "a": doc.missing, "b": { "n": 1 } }
    yield d.b.n == 1
  }
  rule shapeWrong = { yield { "a": doc.missing, "b": { "n": "x" } } is dict[Item] }
  rule listElementUndefined = { yield [doc.missing] is list[number] }
  rule onlyUndefined = { yield { "a": doc.missing } is dict[number] }
  rule outputKeepsNull = {
    let d: dict[number] = { "a": doc.missing, "b": 2 }
    yield d
  }
  export decision of lambdaParam
  export decision of lambdaReturnDict
  export decision of unionDict
  export decision of listOfDict
  export decision of dictOfShape
  export decision of shapeWrong
  export decision of listElementUndefined
  export decision of onlyUndefined
  export decision of outputKeepsNull
}
`

func (s *RuntimeTestSuite) TestDictEntryHoldingUndefinedInEveryPosition() {
	ctx := s.T().Context()
	for _, rule := range []string{"lambdaParam", "lambdaReturnDict", "unionDict", "listOfDict", "dictOfShape", "onlyUndefined"} {
		out, err := s.execRuleUnvalidated(ctx, dictUndefinedPositionsSrc, rule)
		if !s.NoError(err, rule) {
			continue
		}
		s.Equal(trinary.True, out.Decision.State, rule)
	}
}

func (s *RuntimeTestSuite) TestDictEntryHoldingUndefinedStaysTheSame() {
	ctx := s.T().Context()

	// A real wrong-typed value beside the undefined entry still fails.
	out, err := s.execRuleUnvalidated(ctx, dictUndefinedPositionsSrc, "shapeWrong")
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)

	// A list[number] element holding undefined is not a number.
	out, err = s.execRuleUnvalidated(ctx, dictUndefinedPositionsSrc, "listElementUndefined")
	s.Require().NoError(err)
	s.Equal(trinary.False, out.Decision.State)

	// The output keeps the key as JSON null.
	out, err = s.execRuleUnvalidated(ctx, dictUndefinedPositionsSrc, "outputKeepsNull")
	s.Require().NoError(err)
	js, err := json.Marshal(out.Decision.Value)
	s.Require().NoError(err)
	s.JSONEq(`{"a":null,"b":2}`, string(js))
}
