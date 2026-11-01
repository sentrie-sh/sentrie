// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"bytes"
	"encoding/json"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
)

// exactIntegersPolicy exercises #166 end to end: integer literals and JSON
// facts beyond 2^53 stay exact through arithmetic, comparison and output.
const exactIntegersPolicy = `namespace exact/ints

policy p {
  fact id: number
  fact other: number

  rule literal = { yield 9007199254740993 }
  rule incremented = { yield 9007199254740992 + 1 }
  rule literal_ne = { yield 9007199254740993 != 9007199254740992 }
  rule ids_equal = { yield id == other }
  rule ids_ordered = { yield id > other }
  rule id_echo = { yield id }
  rule int_float_equal = { yield 5 == 5.0 }
  rule halves = { yield 7 / 2 }
  rule overflow = { yield 9223372036854775807 + 1 }
  rule negate_overflow = { yield -(-9223372036854775807 - 1) }
  rule huge_literal = { yield 18446744073709551616 }
  rule cast_string = { yield "9007199254740993" as number }

  export decision of literal
  export decision of incremented
  export decision of literal_ne
  export decision of ids_equal
  export decision of ids_ordered
  export decision of id_echo
  export decision of int_float_equal
  export decision of halves
  export decision of overflow
  export decision of negate_overflow
  export decision of huge_literal
  export decision of cast_string
}
`

func (s *RuntimeTestSuite) exactIntegerFacts() map[string]any {
	var facts map[string]any
	dec := json.NewDecoder(bytes.NewReader([]byte(`{"id": 9007199254740993, "other": 9007199254740992}`)))
	dec.UseNumber()
	s.Require().NoError(dec.Decode(&facts))
	return facts
}

func (s *RuntimeTestSuite) TestExactIntegersEndToEnd() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": exactIntegersPolicy})
	defer func() { s.NoError(exec.Close()) }()
	facts := s.exactIntegerFacts()

	run := func(rule string) box.Value {
		out, err := exec.ExecRule(ctx, "exact/ints", "p", rule, facts)
		s.Require().NoError(err, rule)
		return out.Decision.Value
	}
	asJSON := func(v box.Value) string {
		b, err := json.Marshal(v)
		s.Require().NoError(err)
		return string(b)
	}

	s.Equal("9007199254740993", asJSON(run("literal")))
	s.Equal("9007199254740993", asJSON(run("incremented")))
	s.Equal("true", asJSON(run("literal_ne")))
	s.Equal("false", asJSON(run("ids_equal")), "two IDs one apart must not compare equal")
	s.Equal("true", asJSON(run("ids_ordered")))
	s.Equal("9007199254740993", asJSON(run("id_echo")))
	s.Equal("true", asJSON(run("int_float_equal")))
	s.Equal("3.5", asJSON(run("halves")))
	s.Equal("9007199254740993", asJSON(run("cast_string")))

	huge := run("huge_literal")
	s.False(huge.IsInt(), "an integer literal beyond int64 falls back to float")
	s.Equal(box.ValueNumber, huge.Kind())

	for _, rule := range []string{"overflow", "negate_overflow"} {
		_, err := exec.ExecRule(ctx, "exact/ints", "p", rule, facts)
		s.Require().Error(err, rule)
		s.ErrorIs(err, box.ErrIntegerOverflow, rule)
	}
}

// TestExactIntegerCrossingIntoJS covers the JS boundary choice for #166: a safe
// integer crosses as a number, an unsafe one is an error instead of being
// rounded by goja.
func (s *RuntimeTestSuite) TestExactIntegerCrossingIntoJS() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"echo.ts": `export function echo(x: number): number { return x; }
`,
		"policy.sentrie": `namespace exact/js

policy p {
  use { echo } from "./echo" as m

  rule safe = { yield m.echo(9007199254740991) }
  rule unsafe = { yield m.echo(9007199254740993) }
  rule nested_unsafe = { yield m.echo({ "id": 9007199254740993 }) }

  export decision of safe
  export decision of unsafe
  export decision of nested_unsafe
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	out, err := exec.ExecRule(ctx, "exact/js", "p", "safe", map[string]any{})
	s.Require().NoError(err)
	i, ok := out.Decision.Value.IntValue()
	s.True(ok)
	s.Equal(int64(9007199254740991), i)

	for _, rule := range []string{"unsafe", "nested_unsafe"} {
		_, err = exec.ExecRule(ctx, "exact/js", "p", rule, map[string]any{})
		s.Require().Error(err, rule)
		s.Contains(err.Error(), "outside the JavaScript safe integer range", rule)
	}
}

// TestMemoKeyTreatsEqualNumbersAlike: 5 and 5.0 are ==, so a memoized call
// must hit the same cache entry for both, nested ones included.
func (s *RuntimeTestSuite) TestMemoKeyTreatsEqualNumbersAlike() {
	node := ast.NewCallExpression(ast.NewIdentifier("f", stubRange()), nil, true, nil, stubRange())
	key := func(args ...box.Value) string {
		k, err := calculateHashKey(node, args)
		s.Require().NoError(err)
		return k
	}
	s.Equal(key(box.Int(5)), key(box.Float(5)))
	s.Equal(key(box.List([]box.Value{box.Int(1)})), key(box.List([]box.Value{box.Float(1)})))
	s.Equal(key(box.Dict(map[string]box.Value{"a": box.Int(2)})), key(box.Dict(map[string]box.Value{"a": box.Float(2)})))
	s.NotEqual(key(box.Int(5)), key(box.Float(5.5)))
	s.NotEqual(key(box.Int(1<<53+1)), key(box.Int(1<<53)))
}

func (s *RuntimeTestSuite) TestNormalizeBoundaryForJSIntegers() {
	out, err := normalizeBoundaryForJS(int64(maxSafeJSInteger))
	s.Require().NoError(err)
	s.Equal(int64(maxSafeJSInteger), out)
	out, err = normalizeBoundaryForJS(int64(-maxSafeJSInteger))
	s.Require().NoError(err)
	s.Equal(int64(-maxSafeJSInteger), out)
	_, err = normalizeBoundaryForJS(int64(maxSafeJSInteger + 1))
	s.Error(err)
	// 2^53 itself is out: 2^53+1 rounds onto it, so it is not a safe integer.
	_, err = normalizeBoundaryForJS(int64(1) << 53)
	s.Error(err)
	_, err = normalizeBoundaryForJS(-(int64(1) << 53))
	s.Error(err)
	_, err = normalizeBoundaryForJS([]any{int64(-maxSafeJSInteger - 1)})
	s.Error(err)
}
