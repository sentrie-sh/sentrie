// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
)

// kleeneOperand is one truth value a policy can put on either side of a logic
// operator. undefined is a missing field of a document fact.
type kleeneOperand struct {
	name  string
	src   string
	state trinary.Value // what the operand reads as; undefined reads as unknown
}

var kleeneOperands = []kleeneOperand{
	{"true", "true", trinary.True},
	{"false", "false", trinary.False},
	{"unknown", "unknown", trinary.Unknown},
	{"undefined", "user.missing", trinary.Unknown},
}

const kleeneHeader = `namespace logic/kleene

policy p {
  fact user: document as user
  use { boom } from "./counter" as m
`

const kleeneThrowModule = `export function boom(): boolean { throw new Error("right operand evaluated"); }
`

var kleeneFacts = map[string]any{"user": map[string]any{}}

func (s *RuntimeTestSuite) kleeneExecutor(rules map[string]string) *executorImpl {
	var b strings.Builder
	b.WriteString(kleeneHeader)
	for name, body := range rules {
		fmt.Fprintf(&b, "  rule %s = %s\n  export decision of %s\n", name, body, name)
	}
	b.WriteString("}\n")
	return s.newTempPackExecutor(map[string]string{"counter.ts": kleeneThrowModule, "policy.sentrie": b.String()})
}

func kleeneOf(op string, l, r trinary.Value) trinary.Value {
	switch op {
	case "and":
		return l.And(r)
	case "or":
		return l.Or(r)
	default:
		return l.Or(r).And(l.And(r).Not())
	}
}

// TestKleeneMatrixForAndOrXor is #137's truth table: every left and right of
// true, false, unknown and undefined under and, or and xor gives the Kleene
// result with undefined read as unknown, and a known result is the same boolean
// a comparison produces.
func (s *RuntimeTestSuite) TestKleeneMatrixForAndOrXor() {
	ctx := s.T().Context()
	rules := map[string]string{}
	type key struct {
		name string
		op   string
		l, r kleeneOperand
	}
	var keys []key
	for _, op := range []string{"and", "or", "xor"} {
		for _, l := range kleeneOperands {
			for _, r := range kleeneOperands {
				name := fmt.Sprintf("%s_%s_%s", op, l.name, r.name)
				rules[name] = fmt.Sprintf("{ yield %s %s %s }", l.src, op, r.src)
				keys = append(keys, key{name, op, l, r})
			}
		}
	}
	exec := s.kleeneExecutor(rules)
	defer func() { s.NoError(exec.Close()) }()

	for _, k := range keys {
		s.Run(k.name, func() {
			out, err := exec.ExecRule(ctx, "logic/kleene", "p", k.name, kleeneFacts)
			s.Require().NoError(err)
			want := kleeneOf(k.op, k.l.state, k.r.state)
			s.Equal(want, out.Decision.State)
			if want == trinary.Unknown {
				return
			}
			s.Equal(box.ValueBool, out.Decision.Value.Kind(), "a known result is a boolean")
			b, ok := out.Decision.Value.BoolValue()
			s.Require().True(ok)
			s.Equal(want == trinary.True, b)
		})
	}
}

// TestShortCircuitSkipsTheRightOperandWhenTheLeftDecides: an erroring right
// side does not fail the expression, and the right side is not evaluated at all
// (no trace child, no call into a module).
func (s *RuntimeTestSuite) TestShortCircuitSkipsTheRightOperandWhenTheLeftDecides() {
	ctx := s.T().Context()
	rules := map[string]string{
		"andErr":  `{ yield false and ("a" < 1) }`,
		"orErr":   `{ yield true or ("a" < 1) }`,
		"andBoom": `{ yield false and m.boom() }`,
		"orBoom":  `{ yield true or m.boom() }`,
	}
	exec := s.kleeneExecutor(rules)
	defer func() { s.NoError(exec.Close()) }()

	var find func(n *trace.Node, op string) *trace.Node
	find = func(n *trace.Node, op string) *trace.Node {
		if n == nil {
			return nil
		}
		if n.Op == "infix" && n.Meta["operator"] == op {
			return n
		}
		for _, ch := range n.Children {
			if got := find(ch, op); got != nil {
				return got
			}
		}
		return nil
	}

	for _, c := range []struct {
		rule, op string
		want     trinary.Value
		children int
	}{
		{"andErr", "and", trinary.False, 1},
		{"orErr", "or", trinary.True, 1},
	} {
		s.Run(c.rule, func() {
			out, err := exec.ExecRule(ctx, "logic/kleene", "p", c.rule, kleeneFacts)
			s.Require().NoError(err)
			s.Equal(c.want, out.Decision.State)
			node := find(out.RuleNode, c.op)
			s.Require().NotNil(node, "no %q node in the rule trace", c.op)
			s.Len(node.Children, c.children, "only the left operand is in the trace")
		})
	}

	// A module function that throws observes evaluation directly: the call
	// never happens, so there is no error. (A counting module cannot be used:
	// module state does not survive between calls in one rule.)
	for name, want := range map[string]trinary.Value{"andBoom": trinary.False, "orBoom": trinary.True} {
		s.Run(name, func() {
			out, err := exec.ExecRule(ctx, "logic/kleene", "p", name, kleeneFacts)
			s.Require().NoError(err)
			s.Equal(want, out.Decision.State)
		})
	}
}

// TestRightOperandErrorStillFailsWhenTheLeftDoesNotDecide: the right side was
// evaluated, so its error propagates as before.
func (s *RuntimeTestSuite) TestRightOperandErrorStillFailsWhenTheLeftDoesNotDecide() {
	ctx := s.T().Context()
	rules := map[string]string{
		"trueAnd":      `{ yield true and ("a" < 1) }`,
		"falseOr":      `{ yield false or ("a" < 1) }`,
		"unknownAnd":   `{ yield unknown and ("a" < 1) }`,
		"undefinedOr":  `{ yield user.missing or ("a" < 1) }`,
		"undefinedAnd": `{ yield user.missing and ("a" < 1) }`,
		"trueAndBoom":  `{ yield true and m.boom() }`,
		"falseOrBoom":  `{ yield false or m.boom() }`,
		"xorFalse":     `{ yield false xor ("a" < 1) }`,
		"xorTrue":      `{ yield true xor ("a" < 1) }`,
	}
	exec := s.kleeneExecutor(rules)
	defer func() { s.NoError(exec.Close()) }()

	for name := range rules {
		s.Run(name, func() {
			_, err := exec.ExecRule(ctx, "logic/kleene", "p", name, kleeneFacts)
			s.Require().Error(err)
		})
	}
}

// TestUndefinedStillPropagatesOutsideLogicOperators: only and/or/xor read
// undefined as unknown. Arithmetic, comparison and not keep it undefined, which
// a rule decision reads as unknown.
func (s *RuntimeTestSuite) TestUndefinedStillPropagatesOutsideLogicOperators() {
	ctx := s.T().Context()
	rules := map[string]string{
		"add":     `{ yield user.missing + 1 }`,
		"eq":      `{ yield user.missing == 1 }`,
		"lt":      `{ yield user.missing < 1 }`,
		"negated": `{ yield not user.missing }`,
		"banged":  `{ yield !user.missing }`,
		"notWrap": `{ yield (not user.missing) or false }`,
	}
	exec := s.kleeneExecutor(rules)
	defer func() { s.NoError(exec.Close()) }()

	for name := range rules {
		s.Run(name, func() {
			out, err := exec.ExecRule(ctx, "logic/kleene", "p", name, kleeneFacts)
			s.Require().NoError(err)
			s.Equal(trinary.Unknown, out.Decision.State)
			if name == "notWrap" {
				return // an or over it is a logic result: unknown, not undefined
			}
			s.True(out.Decision.Value.IsUndefined(), "the raw value stays undefined, got %v", out.Decision.Value.Kind())
		})
	}
}

// TestIssue137Repros is the three expressions from the issue, which failed open
// or errored before.
func (s *RuntimeTestSuite) TestIssue137Repros() {
	ctx := s.T().Context()
	rules := map[string]string{
		"orMissing":  `{ yield true or user.missing }`,
		"andMissing": `{ yield false and user.missing }`,
		"orErr":      `{ yield true or ("a" < 1) }`,
	}
	exec := s.kleeneExecutor(rules)
	defer func() { s.NoError(exec.Close()) }()

	for name, want := range map[string]trinary.Value{"orMissing": trinary.True, "andMissing": trinary.False, "orErr": trinary.True} {
		s.Run(name, func() {
			out, err := exec.ExecRule(ctx, "logic/kleene", "p", name, kleeneFacts)
			s.Require().NoError(err)
			s.Equal(want, out.Decision.State)
		})
	}
}

// TestWhenGateTakesTheBodyWhenOrIsDecidedByTheLeft: `true or <undefined>` used
// to be unknown, so the gate closed and the default ran. An undefined gate
// still closes it.
func (s *RuntimeTestSuite) TestWhenGateTakesTheBodyWhenOrIsDecidedByTheLeft() {
	ctx := s.T().Context()
	rules := map[string]string{
		"gateOr":        `default false when true or user.missing { yield true }`,
		"gateUndefined": `default false when user.missing { yield true }`,
	}
	exec := s.kleeneExecutor(rules)
	defer func() { s.NoError(exec.Close()) }()

	for name, want := range map[string]trinary.Value{"gateOr": trinary.True, "gateUndefined": trinary.False} {
		s.Run(name, func() {
			out, err := exec.ExecRule(ctx, "logic/kleene", "p", name, kleeneFacts)
			s.Require().NoError(err)
			s.Equal(want, out.Decision.State)
		})
	}
}
