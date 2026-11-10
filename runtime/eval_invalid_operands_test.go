// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"
	"math"
	"strings"

	"github.com/sentrie-sh/sentrie/trinary"
)

// invalidOperandCase is one rule body evaluated through the full pipeline.
type invalidOperandCase struct {
	name    string
	expr    string
	facts   map[string]any
	want    any    // expected value when wantErr is empty and wantUndef is false
	wantErr string // substring of the expected error
	// wantUnknown expects a trinary unknown result.
	wantUnknown bool
	wantUndef   bool
}

func (s *RuntimeTestSuite) runInvalidOperandCases(cases []invalidOperandCase) {
	ctx := s.T().Context()
	var b strings.Builder
	b.WriteString("namespace inv/ops\n\npolicy p {\n")
	b.WriteString("  fact xs: list[number]\n  fact i: number\n  fact u?: number\n  fact idx_s: string\n  fact name: string\n  fact bag: list[string]\n\n")
	for n, tc := range cases {
		fmt.Fprintf(&b, "  rule r%d = { yield %s }\n", n, tc.expr)
	}
	b.WriteString("\n")
	for n := range cases {
		fmt.Fprintf(&b, "  export decision of r%d\n", n)
	}
	b.WriteString("}\n")
	exec := s.newTempPackExecutor(map[string]string{"ops.sentrie": b.String()})
	defer func() { s.NoError(exec.Close()) }()

	for n, tc := range cases {
		s.Run(tc.name, func() {
			facts := map[string]any{
				"xs":    []any{10.0, 20.0, 30.0},
				"i":     0.0,
				"idx_s": "foo",
				"name":  "a",
				"bag":   []any{"x"},
			}
			for k, v := range tc.facts {
				facts[k] = v
			}
			out, err := exec.ExecRule(ctx, "inv/ops", "p", fmt.Sprintf("r%d", n), facts)
			if tc.wantErr != "" {
				s.Require().Error(err, tc.expr)
				s.Contains(err.Error(), tc.wantErr, tc.expr)
				return
			}
			s.Require().NoError(err, tc.expr)
			switch {
			case tc.wantUndef:
				s.True(out.Decision.Value.IsUndefined(), "%s: got %v", tc.expr, out.Decision.Value)
			case tc.wantUnknown:
				s.Equal(trinary.Unknown, out.Decision.State, tc.expr)
			default:
				s.Equal(tc.want, out.Decision.Value.Any(), tc.expr)
			}
		})
	}
}

// TestListIndexMustBeAnIntegerEndToEnd covers both list paths: a literal list
// (boxed) and a list injected as a fact (boundary []any).
func (s *RuntimeTestSuite) TestListIndexMustBeAnIntegerEndToEnd() {
	const want = "list index must be an integer, got "
	s.runInvalidOperandCases([]invalidOperandCase{
		{name: "boxed string index", expr: `[10, 20, 30]["foo"]`, wantErr: want + "string"},
		{name: "boxed fractional index", expr: `[10, 20, 30][1.7]`, wantErr: want},
		{name: "boxed integral float index", expr: `[10, 20, 30][2.0]`, want: int64(30)},
		{name: "boxed integer index", expr: `[10, 20, 30][1]`, want: int64(20)},
		{name: "boxed out of range", expr: `[10, 20, 30][5]`, wantUndef: true},
		{name: "boxed negative index", expr: `[10, 20, 30][-1]`, wantUndef: true},
		{name: "boxed undefined index", expr: `[10, 20, 30][u]`, wantUndef: true},
		{name: "boxed bool index", expr: `[10, 20, 30][true]`, wantErr: want + "bool"},
		{name: "boxed null index", expr: `[10, 20, 30][null]`, wantErr: want},
		{name: "boxed huge index", expr: `[10, 20, 30][1e300]`, wantUndef: true},
		{name: "boxed index beyond int64", expr: `[10, 20, 30][1e19]`, wantUndef: true},
		{name: "boxed infinite index", expr: `[10, 20, 30][i]`, facts: map[string]any{"i": math.Inf(1)}, wantUndef: true},
		{name: "boxed negative infinite index", expr: `[10, 20, 30][i]`, facts: map[string]any{"i": math.Inf(-1)}, wantUndef: true},
		{name: "boxed NaN index", expr: `[10, 20, 30][i]`, facts: map[string]any{"i": math.NaN()}, wantErr: want},
		{name: "dict missing key", expr: `{"a": 1}["b"]`, wantUndef: true},
		{name: "dict present key", expr: `{"a": 1}["a"]`, want: int64(1)},

		{name: "boundary string index", expr: `xs[idx_s]`, wantErr: want + "string"},
		{name: "boundary fractional index", expr: `xs[1.7]`, wantErr: want},
		{name: "boundary integral float index", expr: `xs[2.0]`, want: 30.0},
		{name: "boundary integer index", expr: `xs[i]`, facts: map[string]any{"i": 1.0}, want: 20.0},
		{name: "boundary out of range", expr: `xs[5]`, wantUndef: true},
		{name: "boundary undefined index", expr: `xs[u]`, wantUndef: true},
		{name: "boundary huge index", expr: `xs[i]`, facts: map[string]any{"i": 1e300}, wantUndef: true},
		{name: "boundary infinite index", expr: `xs[i]`, facts: map[string]any{"i": math.Inf(1)}, wantUndef: true},
		{name: "boundary NaN index", expr: `xs[i]`, facts: map[string]any{"i": math.NaN()}, wantErr: want},
		{name: "boundary bool index", expr: `xs[true]`, wantErr: want + "bool"},
	})
}

// TestCountRejectsUncountableOperandsEndToEnd covers count at runtime, where
// the kind is only known from a fact or a call result.
func (s *RuntimeTestSuite) TestCountRejectsUncountableOperandsEndToEnd() {
	s.runInvalidOperandCases([]invalidOperandCase{
		{name: "count of a number reached through a lambda", expr: `collect([5], (x) => { yield count(x) })`, wantErr: "count expects a list, dict or string"},
		{name: "count of a null literal", expr: `count(null)`, wantErr: "count"},
		{name: "count of null reached through a lambda", expr: `collect([null], (x) => { yield count(x) })`, wantErr: "count"},
		{name: "count of undefined reached through a lambda", expr: `collect([u], (x) => { yield count(x) })[0]`, wantUndef: true},
		{name: "count of a list", expr: `count([1, 2, 3])`, want: int64(3)},
		{name: "count of a dict", expr: `count({"a": 1, "b": 2})`, want: int64(2)},
		{name: "count of a string", expr: `count("hello")`, want: int64(5)},
		{name: "count of a list fact", expr: `count(xs)`, want: int64(3)},
	})
}

// TestStringPlusFollowsJavaScriptEndToEnd covers + with a string operand.
func (s *RuntimeTestSuite) TestStringPlusFollowsJavaScriptEndToEnd() {
	s.runInvalidOperandCases([]invalidOperandCase{
		{name: "string plus int", expr: `"a" + 1`, want: "a1"},
		{name: "int plus string", expr: `1 + "a"`, want: "1a"},
		{name: "string plus integral float", expr: `"a" + 1.0`, want: "a1"},
		{name: "string plus fraction", expr: `"a" + 1.5`, want: "a1.5"},
		{name: "string plus true", expr: `"a" + true`, want: "atrue"},
		{name: "false plus string", expr: `false + "a"`, want: "falsea"},
		{name: "string plus null", expr: `"a" + null`, want: "anull"},
		{name: "string plus 1e20", expr: `"a" + 1e20`, want: "a100000000000000000000"},
		{name: "string plus 1e21", expr: `"a" + 1e21`, want: "a1e+21"},
		{name: "string plus 0.000001", expr: `"a" + 0.000001`, want: "a0.000001"},
		{name: "string plus 1e-7", expr: `"a" + 1e-7`, want: "a1e-7"},
		{name: "string plus negative", expr: `"a" + -2.5`, want: "a-2.5"},
		{name: "string plus string", expr: `"a" + "b"`, want: "ab"},
		{name: "fact string plus number fact", expr: `name + i`, facts: map[string]any{"i": 7.0}, want: "a7"},

		{name: "string plus undefined", expr: `"a" + u`, wantUnknown: true},
		{name: "undefined plus string", expr: `u + "a"`, wantUnknown: true},
		{name: "string plus unknown", expr: `"a" + unknown`, wantUnknown: true},
		{name: "unknown plus string", expr: `unknown + "a"`, wantUnknown: true},
		{name: "number plus undefined stays undefined", expr: `1 + u`, wantUndef: true},
		{name: "undefined plus unknown stays undefined", expr: `u + unknown`, wantUndef: true},

		{name: "string plus list", expr: `"a" + [1]`, wantErr: "+ cannot join a string with a list"},
		{name: "list plus string", expr: `[1] + "a"`, wantErr: "+ cannot join a string with a list"},
		{name: "string plus dict", expr: `"a" + {"k": 1}`, wantErr: "+ cannot join a string with a dict"},
		{name: "dict plus string", expr: `{"k": 1} + "a"`, wantErr: "+ cannot join a string with a dict"},
	})
}
