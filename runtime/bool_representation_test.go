// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/trinary"
)

// boolProducer is one way a policy can produce a known truth value. yes and no
// are Sentrie expressions that evaluate to true and false.
type boolProducer struct {
	name string
	yes  string
	no   string
}

// boolProducers covers every route into the runtime from #135: literals,
// comparisons, JSON facts (untyped document and trinary-typed), a TypeScript
// return value, logical operators, derive returns and casts. Each pair must
// agree under ==, in and contains.
var boolProducers = []boolProducer{
	{name: "literal", yes: "true", no: "false"},
	{name: "comparison", yes: "(1 == 1)", no: "(1 == 2)"},
	{name: "document fact", yes: "user.on", no: "user.off"},
	{name: "typed fact", yes: "flagOn", no: "flagOff"},
	{name: "js return", yes: "m.yes()", no: "m.no()"},
	{name: "not", yes: "(not false)", no: "(not true)"},
	{name: "and", yes: "(true and true)", no: "(true and false)"},
	{name: "derive return", yes: "alwaysYes()", no: "alwaysNo()"},
	{name: "cast", yes: "(1 as trinary)", no: "(0 as trinary)"},
}

const boolMatrixModule = `export function yes(): boolean { return true; }
export function no(): boolean { return false; }
export function kind(a: any): string { return typeof a + ":" + String(a); }
export function pick(a: any): number { return a ? 1 : 0; }
export function kinds(a: any[]): string { return a.map((x) => typeof x).join(","); }
export function fieldKind(a: any): string { return typeof a.flag; }
`

var boolMatrixFacts = map[string]any{
	"user":    map[string]any{"on": true, "off": false},
	"flagOn":  true,
	"flagOff": false,
}

const boolMatrixHeader = `namespace bool/matrix

derive alwaysYes = (): trinary => { yield true }
derive alwaysNo = (): trinary => { yield false }

policy p {
  fact user: document as user
  fact flagOn: trinary
  fact flagOff: trinary
  use { yes, no, kind, pick, kinds, fieldKind } from "./bools" as m
`

// buildBoolMatrix returns the policy source and the rule names. Each rule
// yields one boolean per (left, right) truth-value combination so a failure
// names the producer pair and the operator.
func buildBoolMatrix() (string, map[string][]bool) {
	var b strings.Builder
	b.WriteString(boolMatrixHeader)
	want := map[string][]bool{}
	for i, l := range boolProducers {
		for j, r := range boolProducers {
			combos := []struct{ a, b string }{
				{l.yes, r.yes}, {l.yes, r.no}, {l.no, r.yes}, {l.no, r.no},
			}
			expected := []bool{true, false, false, true}
			for _, op := range []string{"eq", "in", "contains"} {
				var terms []string
				for _, c := range combos {
					switch op {
					case "eq":
						terms = append(terms, fmt.Sprintf("%s == %s", c.a, c.b))
					case "in":
						// the second element keeps the list from being a one-item shortcut
						terms = append(terms, fmt.Sprintf("%s in [%s, 99]", c.a, c.b))
					case "contains":
						terms = append(terms, fmt.Sprintf("[%s, 99] contains %s", c.b, c.a))
					}
				}
				name := fmt.Sprintf("%s_%d_%d", op, i, j)
				fmt.Fprintf(&b, "  rule %s = { yield [%s] }\n  export decision of %s\n", name, strings.Join(terms, ", "), name)
				want[name] = expected
			}
		}
	}
	b.WriteString("}\n")
	return b.String(), want
}

// TestEveryTruthValueProducerAgreesOnEqualityMembershipAndContains is #135's
// matrix: true from any producer equals true from any other and differs from
// every false, under ==, in and contains.
func (s *RuntimeTestSuite) TestEveryTruthValueProducerAgreesOnEqualityMembershipAndContains() {
	ctx := s.T().Context()
	src, want := buildBoolMatrix()
	exec := s.newTempPackExecutor(map[string]string{"bools.ts": boolMatrixModule, "policy.sentrie": src})
	defer func() { s.NoError(exec.Close()) }()

	for name, expected := range want {
		out, err := exec.ExecRule(ctx, "bool/matrix", "p", name, boolMatrixFacts)
		s.Require().NoError(err, name)
		got, ok := out.Decision.Value.Any().([]any)
		s.Require().True(ok, "%s: decision value %v is not a list", name, out.Decision.Value)
		s.Require().Len(got, len(expected), name)
		for k := range expected {
			s.Equal(expected[k], got[k], "%s combination %d (yes/yes, yes/no, no/yes, no/no)", describeMatrixRule(name), k)
		}
	}
}

func describeMatrixRule(name string) string {
	var op string
	var i, j int
	if _, err := fmt.Sscanf(strings.ReplaceAll(name, "_", " "), "%s %d %d", &op, &i, &j); err != nil {
		return name
	}
	return fmt.Sprintf("%s: %s vs %s", op, boolProducers[i].name, boolProducers[j].name)
}

func (s *RuntimeTestSuite) boolBoundaryRules() (map[string]string, []string) {
	var b strings.Builder
	b.WriteString(boolMatrixHeader)
	var rules []string
	add := func(name, expr string) {
		fmt.Fprintf(&b, "  rule %s = { yield %s }\n  export decision of %s\n", name, expr, name)
		rules = append(rules, name)
	}
	for i, p := range boolProducers {
		add(fmt.Sprintf("kind_yes_%d", i), fmt.Sprintf(`m.kind(%s) == "boolean:true"`, p.yes))
		add(fmt.Sprintf("kind_no_%d", i), fmt.Sprintf(`m.kind(%s) == "boolean:false"`, p.no))
		add(fmt.Sprintf("pick_yes_%d", i), fmt.Sprintf(`m.pick(%s) == 1`, p.yes))
		add(fmt.Sprintf("pick_no_%d", i), fmt.Sprintf(`m.pick(%s) == 0`, p.no))
	}
	// Three ways to produce unknown all arrive as undefined.
	add("unknown_literal", `m.kind(unknown) == "undefined:undefined"`)
	add("unknown_not", `m.kind(not unknown) == "undefined:undefined"`)
	add("unknown_and", `m.kind(true and unknown) == "undefined:undefined"`)
	add("unknown_pick", `m.pick(unknown) == 0`)
	// Nested values convert recursively.
	add("list_kinds", `m.kinds([true, 1 == 2, not false, user.off, unknown]) == "boolean,boolean,boolean,boolean,undefined"`)
	add("dict_kind", `m.fieldKind({ "flag": not true }) == "boolean"`)
	b.WriteString("}\n")
	return map[string]string{"bools.ts": boolMatrixModule, "policy.sentrie": b.String()}, rules
}

// TestTruthValuesCrossIntoJavaScriptAsBooleans covers #135's JS defect: true
// and false are real booleans from every producer (so false is falsy in a
// ternary) and unknown is undefined.
func (s *RuntimeTestSuite) TestTruthValuesCrossIntoJavaScriptAsBooleans() {
	ctx := s.T().Context()
	files, rules := s.boolBoundaryRules()
	exec := s.newTempPackExecutor(files)
	defer func() { s.NoError(exec.Close()) }()

	for _, name := range rules {
		out, err := exec.ExecRule(ctx, "bool/matrix", "p", name, boolMatrixFacts)
		s.Require().NoError(err, name)
		s.Equal(trinary.True, out.Decision.State, name)
	}
}

// TestDecisionJSONCarriesKnownTruthValuesAsBooleans covers the output half of
// #135 through the encoder exec --output json and the serve API both use: a
// known value is a JSON boolean, state keeps its string enum, unknown stays a
// string.
func (s *RuntimeTestSuite) TestDecisionJSONCarriesKnownTruthValuesAsBooleans() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"policy.sentrie": `namespace bool/json

policy p {
  rule negated = { yield not false }
  rule negatedTrue = { yield not true }
  rule comparison = { yield 1 == 1 }
  rule failedComparison = { yield 1 == 2 }
  rule literal = { yield true }
  rule falseLiteral = { yield false }
  rule logical = { yield true and true }
  rule open = { yield unknown }

  export decision of negated
  export decision of negatedTrue
  export decision of comparison
  export decision of failedComparison
  export decision of literal
  export decision of falseLiteral
  export decision of logical
  export decision of open
}
`,
	})
	defer func() { s.NoError(exec.Close()) }()

	outputs, err := exec.ExecPolicy(ctx, "bool/json", "p", nil)
	s.Require().NoError(err)
	raw, err := json.Marshal(outputs)
	s.Require().NoError(err)

	var decoded []struct {
		Rule     string `json:"rule"`
		Decision struct {
			State string `json:"state"`
			Value any    `json:"value"`
		} `json:"decision"`
	}
	s.Require().NoError(json.Unmarshal(raw, &decoded))

	type result struct {
		state string
		value any
	}
	want := map[string]result{
		"negated":          {"true", true},
		"negatedTrue":      {"false", false},
		"comparison":       {"true", true},
		"failedComparison": {"false", false},
		"literal":          {"true", true},
		"falseLiteral":     {"false", false},
		"logical":          {"true", true},
		"open":             {"unknown", "unknown"},
	}
	s.Require().Len(decoded, len(want))
	for _, d := range decoded {
		w, ok := want[d.Rule]
		s.Require().True(ok, "unexpected rule %s", d.Rule)
		s.Equal(w.state, d.Decision.State, "%s state", d.Rule)
		s.Equal(w.value, d.Decision.Value, "%s value must keep its JSON type", d.Rule)
	}
}
