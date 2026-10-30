// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
)

// isCase is one rule body and the decision it must reach.
type isCase struct {
	name string
	expr string
	want trinary.Value
}

// buildIsExecutor compiles one policy whose rules are the given cases. The policy
// declares a document fact `doc` (supplied by the test), an omitted-by-default
// optional fact `extra`, an optional fact with a default `dflt`, and provided
// empty-string fact `blank`.
func (s *RuntimeTestSuite) buildIsExecutor(cases []isCase, extraSrc string) *executorImpl {
	src := "namespace com/ex\n\n" + extraSrc + "\npolicy pol {\n" +
		"  fact doc: document as doc\n" +
		"  fact extra?: document as extra\n" +
		"  fact dflt?: string as dflt default \"d\"\n" +
		"  fact blank: string as blank\n" +
		"  fact emptyObj: document as emptyObj\n" +
		"  fact emptyArr: list[number] as emptyArr\n"
	for i, c := range cases {
		src += fmt.Sprintf("  rule r%d = { yield %s }\n  export decision of r%d\n", i, c.expr, i)
	}
	src += "}\n"
	_, exec := s.mustBuildDeriveExecutor(s.T().Context(), deriveTestProgram{name: "is_ops.sentrie", src: src})
	return exec
}

func isFacts() map[string]any {
	return map[string]any{
		"doc": map[string]any{
			"s":       "",
			"str":     "x",
			"n":       nil,
			"zero":    0,
			"one":     1,
			"t":       true,
			"f":       false,
			"list0":   []any{},
			"list1":   []any{1},
			"obj0":    map[string]any{},
			"obj1":    map[string]any{"a": 1},
			"nonempt": "y",
		},
		"blank":    "",
		"emptyObj": map[string]any{},
		"emptyArr": []any{},
	}
}

func (s *RuntimeTestSuite) TestEvalIsDefinedAndIsEmptyTable() {
	ctx := s.T().Context()
	T, F := trinary.True, trinary.False

	cases := []isCase{
		// is defined: literals, null, and every value that is merely falsy
		{"empty string is defined", `"" is defined`, T},
		{"empty list is defined", `[] is defined`, T},
		{"empty dict is defined", `{} is defined`, T},
		{"null is defined", `null is defined`, T},
		{"zero is defined", `0 is defined`, T},
		{"false is defined", `false is defined`, T},
		{"null field is defined", `doc.n is defined`, T},
		{"present field is defined", `doc.str is defined`, T},
		{"missing field is not defined", `doc.nope is defined`, F},
		{"missing field is not defined (negated)", `doc.nope is not defined`, T},
		{"missing nested field is not defined", `doc.nope.deeper is defined`, F},
		{"out of range index is not defined", `doc.list1[5] is defined`, F},
		{"in range index is defined", `doc.list1[0] is defined`, T},
		{"omitted optional fact is not defined", `extra is defined`, F},
		{"omitted optional fact is not defined (negated)", `extra is not defined`, T},
		{"field of omitted optional fact is not defined", `extra.k is defined`, F},
		{"field of omitted optional fact is not defined (negated)", `extra.k is not defined`, T},
		{"optional fact with default is defined", `dflt is defined`, T},
		{"optional fact with default is defined (negated)", `dflt is not defined`, F},
		{"provided fact is defined", `blank is defined`, T},
		{"provided fact is defined (negated)", `blank is not defined`, F},
		{"negated defined on present", `doc.str is not defined`, F},
		{"negated defined on null", `null is not defined`, F},

		// is empty: the empty values
		{"empty string is empty", `"" is empty`, T},
		{"empty list is empty", `[] is empty`, T},
		{"empty dict is empty", `{} is empty`, T},
		{"null is empty", `null is empty`, T},
		{"missing field is empty", `doc.nope is empty`, T},
		{"omitted optional fact is empty", `extra is empty`, T},
		{"field of omitted optional fact is empty", `extra.k is empty`, T},
		{"empty string field is empty", `doc.s is empty`, T},
		{"null field is empty", `doc.n is empty`, T},
		{"empty list field is empty", `doc.list0 is empty`, T},
		{"empty dict field is empty", `doc.obj0 is empty`, T},
		{"provided empty string fact is empty", `blank is empty`, T},
		{"document fact that is an empty object is empty", `emptyObj is empty`, T},
		{"list fact that is an empty array is empty", `emptyArr is empty`, T},
		{"negated empty on empty string", `"" is not empty`, F},
		{"negated empty on empty list", `[] is not empty`, F},
		{"negated empty on empty dict", `{} is not empty`, F},
		{"negated empty on null", `null is not empty`, F},
		{"negated empty on missing field", `doc.nope is not empty`, F},
		{"negated empty on omitted optional fact", `extra is not empty`, F},
		{"negated empty on empty document object", `emptyObj is not empty`, F},
		{"negated empty on empty list fact", `emptyArr is not empty`, F},

		// is empty: non-empty values, scalars included, are false and never an error
		{"non-empty string is not empty", `"a" is empty`, F},
		{"non-empty list is not empty", `[1] is empty`, F},
		{"non-empty dict is not empty", `{"a": 1} is empty`, F},
		{"zero is not empty", `0 is empty`, F},
		{"one is not empty", `1 is empty`, F},
		{"true is not empty", `true is empty`, F},
		{"false is not empty", `false is empty`, F},
		{"non-empty string field is not empty", `doc.str is empty`, F},
		{"zero field is not empty", `doc.zero is empty`, F},
		{"false field is not empty", `doc.f is empty`, F},
		{"non-empty list field is not empty", `doc.list1 is empty`, F},
		{"non-empty dict field is not empty", `doc.obj1 is empty`, F},
		{"optional fact with default is not empty", `dflt is empty`, F},
		{"negated empty on non-empty string", `"a" is not empty`, T},
		{"negated empty on non-empty list", `[1] is not empty`, T},
		{"negated empty on non-empty dict", `{"a": 1} is not empty`, T},
		{"negated empty on zero", `0 is not empty`, T},
		{"negated empty on one", `1 is not empty`, T},
		{"negated empty on true", `true is not empty`, T},
		{"negated empty on false", `false is not empty`, T},
		{"negated empty on optional fact with default", `dflt is not empty`, T},
		{"negated empty on non-empty dict field", `doc.obj1 is not empty`, T},

		// composition: the predicates are ordinary boolean operands
		{"predicates combine with and", `extra is not defined and doc.str is not empty`, T},
		{"predicates combine with or", `extra is defined or doc.nope is not defined`, T},
		{"empty guard over a field chain", `doc.str is not empty and doc.obj1 is not empty`, T},
	}

	exec := s.buildIsExecutor(cases, "")
	for i, c := range cases {
		s.Run(c.name, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", fmt.Sprintf("r%d", i), isFacts())
			s.Require().NoError(err)
			s.Equal(c.want, out.Decision.State, c.expr)
		})
	}
}

func (s *RuntimeTestSuite) TestEvalIsInsideDeriveBody() {
	ctx := s.T().Context()
	T, F := trinary.True, trinary.False
	derives := `derive blank = (v?: string): trinary => { yield v is empty }
derive present = (v?: string): trinary => { yield v is defined }
derive missingKey = (m: dict[number]): trinary => { yield m["k"] is not defined }
derive nonEmpty = (xs: list[number]): trinary => { yield xs is not empty }
`
	cases := []isCase{
		{"is empty on an empty argument", `blank("")`, T},
		{"is empty on an omitted optional parameter", `blank()`, T},
		{"is empty on a non-empty argument", `blank("a")`, F},
		{"is defined on an omitted optional parameter", `present()`, F},
		{"is defined on an empty argument", `present("")`, T},
		{"is not defined on a missing key", `missingKey({"a": 1})`, T},
		{"is not defined on a present key", `missingKey({"k": 1})`, F},
		{"is not empty on an empty list", `nonEmpty([])`, F},
		{"is not empty on a non-empty list", `nonEmpty([1])`, T},
	}
	exec := s.buildIsExecutor(cases, derives)
	for i, c := range cases {
		s.Run(c.name, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", fmt.Sprintf("r%d", i), isFacts())
			s.Require().NoError(err)
			s.Equal(c.want, out.Decision.State, c.expr)
		})
	}
}

// An operand that fails to evaluate is an error, never a quiet false: swallowing it
// would hide typos and real failures behind a definite answer.
func (s *RuntimeTestSuite) TestEvalIsPropagatesOperandError() {
	ctx := s.T().Context()
	cases := []isCase{
		{"is defined", `(1 / 0) is defined`, trinary.Unknown},
		{"is not defined", `(1 / 0) is not defined`, trinary.Unknown},
		{"is empty", `(1 / 0) is empty`, trinary.Unknown},
		{"is not empty", `(1 / 0) is not empty`, trinary.Unknown},
	}
	exec := s.buildIsExecutor(cases, "")
	for i, c := range cases {
		s.Run(c.name, func() {
			_, err := exec.ExecRule(ctx, "com/ex", "pol", fmt.Sprintf("r%d", i), isFacts())
			s.Require().Error(err, c.expr)
		})
	}
}

// Resolving an omitted optional fact as undefined changed what an ordinary
// expression over it does: it propagates undefined and the rule decides Unknown
// instead of failing the request with "identifier not found".
func (s *RuntimeTestSuite) TestOmittedOptionalFactDecidesUnknownInOrdinaryExpression() {
	ctx := s.T().Context()
	cases := []isCase{
		{"field comparison", `extra.role == "admin"`, trinary.Unknown},
		{"bare identifier comparison", `extra == "admin"`, trinary.Unknown},
	}
	exec := s.buildIsExecutor(cases, "")
	for i, c := range cases {
		s.Run(c.name, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", fmt.Sprintf("r%d", i), isFacts())
			s.Require().NoError(err, c.expr)
			s.Equal(c.want, out.Decision.State, c.expr)
		})
	}

	s.Run("the same fact provided still compares", func() {
		facts := isFacts()
		facts["extra"] = map[string]any{"role": "admin"}
		out, err := exec.ExecRule(ctx, "com/ex", "pol", "r0", facts)
		s.Require().NoError(err)
		s.Equal(trinary.True, out.Decision.State)
	})
}

func (s *RuntimeTestSuite) TestEvalIsTraceNode() {
	ctx := s.T().Context()
	cases := []isCase{
		{"defined", `doc.nope is defined`, trinary.False},
		{"empty", `doc.str is empty`, trinary.False},
	}
	exec := s.buildIsExecutor(cases, "")

	var find func(n *trace.Node) *trace.Node
	find = func(n *trace.Node) *trace.Node {
		if n == nil {
			return nil
		}
		if n.Op == "is" {
			return n
		}
		for _, ch := range n.Children {
			if got := find(ch); got != nil {
				return got
			}
		}
		return nil
	}

	for i, c := range cases {
		s.Run(c.name, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", fmt.Sprintf("r%d", i), isFacts())
			s.Require().NoError(err)
			s.Equal(c.want, out.Decision.State)

			node := find(out.RuleNode)
			s.Require().NotNil(node, "no 'is' node in the rule trace")
			s.Equal(c.name, node.Meta["operator"])
			s.Require().Len(node.Children, 1, "the operand is attached as the only child")
			s.Empty(node.Err)
			b, ok := node.Result.BoolValue()
			s.True(ok, "result is a bool, got %v", node.Result.Kind())
			s.False(b)
		})
	}
}

// A host document (box.Document) is not a dict or list, so emptiness reads the
// wrapped value's length by reflection. Facts arrive as dicts and lists; documents
// that reach a policy wrap typed Go values a builtin returned, such as a []string or
// map[string]string, so those must read as empty when they have no elements.
func (s *RuntimeTestSuite) TestIsEmptyValueHostDocuments() {
	type hostStruct struct{ A int }

	s.True(isEmptyValue(box.Document(map[string]any{})))
	s.False(isEmptyValue(box.Document(map[string]any{"a": 1})))
	s.True(isEmptyValue(box.Document([]any{})))
	s.False(isEmptyValue(box.Document([]any{1})))
	s.False(isEmptyValue(box.Document(hostStruct{})), "a host shape with no length is never empty")

	s.True(isEmptyValue(box.Document([]string{})), "empty typed slice")
	s.False(isEmptyValue(box.Document([]string{"a"})))
	s.True(isEmptyValue(box.Document(map[string]string{})), "empty typed map")
	s.False(isEmptyValue(box.Document(map[string]string{"k": "v"})))
	s.True(isEmptyValue(box.Document([0]int{})), "empty array")
	s.True(isEmptyValue(box.Document((map[string]int)(nil))), "nil typed map")
	s.False(isEmptyValue(box.Document(42)), "a host number is never empty")
	s.False(isEmptyValue(box.Document[any](nil)), "an untyped nil document has no length")
}
