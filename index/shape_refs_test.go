// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package index

import (
	"errors"
	"strings"

	"github.com/sentrie-sh/sentrie/xerr"
)

// shapeRefIsPositions are the constraintArgPositions whose type is the type
// of `x is T`; resolveIsExpressions reports their shape names.
var shapeRefIsPositions = map[string]bool{
	"is": true, "list element": true, "dict value": true, "record field": true,
	"rule default": true, "derive body is": true, "is inside a block": true,
}

// shapeRefSource puts typ in position, in a namespace that declares the alias
// shape Known.
func (s *IndexTestSuite) shapeRefSource(position, typ, value string) string {
	return strings.Replace(s.constraintArgSource(position, typ, value), "namespace com/ex\n", "namespace com/ex\nshape Known number\n", 1)
}

// TestValidateRejectsMissingShapeInEveryTypePosition is #282: a shape name
// that does not exist is an index error wherever a type ref can appear, at
// the type ref's span, and an `is` type is reported once, by
// resolveIsExpressions.
func (s *IndexTestSuite) TestValidateRejectsMissingShapeInEveryTypePosition() {
	for position := range constraintArgPositions {
		s.Run(position, func() {
			err := s.parseAndValidate(s.shapeRefSource(position, "Nope", "1"), "c.sentrie")
			s.Require().Error(err)
			s.True(errors.Is(err, xerr.ErrIndex), err.Error())
			s.Equal(1, strings.Count(err.Error(), "index error"), err.Error())
			if shapeRefIsPositions[position] {
				s.Contains(err.Error(), "`is ")
				return
			}
			s.Contains(err.Error(), "shape 'Nope' not found at c.sentrie:")
		})
	}
}

// TestValidateAcceptsKnownShapeInEveryTypePosition checks that a shape in
// scope still validates in every position.
func (s *IndexTestSuite) TestValidateAcceptsKnownShapeInEveryTypePosition() {
	for position := range constraintArgPositions {
		for _, typ := range []string{"Known", "Known?"} {
			s.Run(position+"/"+typ, func() {
				s.NoError(s.parseAndValidate(s.shapeRefSource(position, typ, "1"), "c.sentrie"))
			})
		}
	}
}

// TestValidateResolvesShapeRefsInTheirScope checks the scope each position
// resolves in: a policy's shapes are visible to its facts, lets, shapes and
// policy-level derives, not to namespace shapes or namespace-level derives,
// and a qualified name must name a shape its namespace exports.
func (s *IndexTestSuite) TestValidateResolvesShapeRefsInTheirScope() {
	const lib = "namespace com/lib\nshape Account { id: string }\nexport shape Account\nshape Secret { id: string }\n"
	const rule = "  rule r = { yield true }\n  export decision of r\n"
	cases := map[string]struct {
		src string
		msg string // empty when the source validates
	}{
		"policy shape in a fact": {
			src: "namespace com/app\npolicy p {\n  fact x: Local as x\n  shape Local number\n" + rule + "}\n",
		},
		"policy shape in a policy shape field": {
			src: "namespace com/app\npolicy p {\n  shape Local number\n  shape Outer { f: Local }\n" + rule + "}\n",
		},
		"policy shape in a policy-level derive": {
			src: "namespace com/app\npolicy p {\n  shape Local number\n  derive d = (a: Local): Local => { yield (a as Local) }\n" + rule + "}\n",
		},
		"policy shape in a namespace shape field": {
			src: "namespace com/app\nshape Outer { f: Local }\npolicy p {\n  shape Local number\n" + rule + "}\n",
			msg: "shape 'Local' not found at s.sentrie:2:18-23",
		},
		"policy shape in a namespace-level derive": {
			src: "namespace com/app\nderive d = (a: number): trinary => { yield (a as Local) == a }\npolicy p {\n  shape Local number\n" + rule + "}\n",
			msg: "shape 'Local' not found at s.sentrie:2:50-55",
		},
		"exported shape of another namespace": {
			src: "namespace com/app\npolicy p {\n  fact x: com/lib/Account as x\n" + rule + "}\n",
		},
		"unexported shape of another namespace": {
			src: "namespace com/app\npolicy p {\n  fact x: com/lib/Secret as x\n" + rule + "}\n",
			msg: "shape 'com/lib/Secret' not exported at s.sentrie:3:11-25",
		},
		"missing shape of another namespace": {
			src: "namespace com/app\npolicy p {\n  let v: list[com/lib/Gone] = []\n" + rule + "}\n",
			msg: "shape 'com/lib/Gone' not found at s.sentrie:3:15-27",
		},
		"shape of a namespace that does not exist": {
			src: "namespace com/app\npolicy p {\n  fact x: com/nowhere/Account as x\n" + rule + "}\n",
			msg: "shape 'com/nowhere/Account' not found at s.sentrie:3:11-30",
		},
	}
	for name, tc := range cases {
		s.Run(name, func() {
			err := s.validateSources([2]string{"lib.sentrie", lib}, [2]string{"s.sentrie", tc.src})
			if tc.msg == "" {
				s.NoError(err)
				return
			}
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
			s.True(errors.Is(err, xerr.ErrIndex), err.Error())
		})
	}
}

// TestValidateReportsEveryMissingShapeInSourceOrder checks that every missing
// shape is reported, in source order, whatever the map iteration order.
func (s *IndexTestSuite) TestValidateReportsEveryMissingShapeInSourceOrder() {
	src := "namespace com/ex\npolicy p {\n  fact b: Second as b\n  let a: First = 1\n  rule r = { yield (1 as Third) == 1 }\n  export decision of r\n}\n"
	want := "validation error: shape 'Second' not found at c.sentrie:3:11-17: index error\n" +
		"shape 'First' not found at c.sentrie:4:10-15: index error\n" +
		"shape 'Third' not found at c.sentrie:5:26-31: index error"
	for range 5 {
		err := s.parseAndValidate(src, "c.sentrie")
		s.Require().Error(err)
		s.Equal(want, err.Error())
	}
}
