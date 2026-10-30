// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/trinary"
)

// importEnvelopeStates are the three imported decisions every matrix case runs
// against: the source rule in policy q, the bare-import rule in policy p, and
// the state the imported decision carries (#136).
var importEnvelopeStates = []struct {
	suffix string
	source string
	imp    string
	state  trinary.Value
}{
	{suffix: "F", source: "f", imp: "impF", state: trinary.False},
	{suffix: "T", source: "tr", imp: "impT", state: trinary.True},
	{suffix: "U", source: "u", imp: "impU", state: trinary.Unknown},
}

// importEnvelopeBase is the exporting policy plus the bare imports the matrix
// rules use. Every rule built on top of it is generated per state.
const importEnvelopeBase = `namespace t
policy q {
  rule f = { yield false }
  rule tr = { yield true }
  rule u = { yield unknown }
  export decision of f attach why as "denied"
  export decision of tr attach why as "granted"
  export decision of u attach why as "undecided"
}
`

func notOf(v trinary.Value) trinary.Value {
	switch v {
	case trinary.True:
		return trinary.False
	case trinary.False:
		return trinary.True
	default:
		return trinary.Unknown
	}
}

func boolState(b bool) trinary.Value {
	if b {
		return trinary.True
	}
	return trinary.False
}

// TestImportedDecisionIsItsStateWherePolicyUsesATruthValue is the #136
// regression matrix: a false or unknown imported decision must never be read as
// true, whether the import is the rule body, an operand of a logical operator, a
// ternary condition, a `when` gate or an operand of ==, !=.
func (s *RuntimeTestSuite) TestImportedDecisionIsItsStateWherePolicyUsesATruthValue() {
	ctx := s.T().Context()

	// Each case is an expression over the import `%s`; want maps the imported
	// state to the state of the rule that yields (or is gated by) it.
	cases := []struct {
		name string
		rule string // full rule text with %[1]s = import name, %[2]s = rule name
		want func(imported trinary.Value) trinary.Value
	}{
		{"direct alias of the import", "rule %[2]s = %[1]s", func(v trinary.Value) trinary.Value { return v }},
		{"and true", "rule %[2]s = { yield %[1]s and true }", func(v trinary.Value) trinary.Value { return v }},
		{"and false", "rule %[2]s = { yield %[1]s and false }", func(trinary.Value) trinary.Value { return trinary.False }},
		{"or false", "rule %[2]s = { yield %[1]s or false }", func(v trinary.Value) trinary.Value { return v }},
		{"or true", "rule %[2]s = { yield %[1]s or true }", func(trinary.Value) trinary.Value { return trinary.True }},
		{"xor true", "rule %[2]s = { yield %[1]s xor true }", notOf},
		{"xor false", "rule %[2]s = { yield %[1]s xor false }", func(v trinary.Value) trinary.Value { return v }},
		{"not", "rule %[2]s = { yield not %[1]s }", notOf},
		{"bang", "rule %[2]s = { yield !%[1]s }", notOf},
		{"ternary condition", "rule %[2]s = { yield %[1]s ? true : false }", func(v trinary.Value) trinary.Value { return boolState(v.IsTrue()) }},
		{"when gate", "rule %[2]s = when (%[1]s) { yield true }", func(v trinary.Value) trinary.Value {
			if v.IsTrue() {
				return trinary.True
			}
			return trinary.Unknown // a closed gate is the default, not a pass
		}},
		{"== true", "rule %[2]s = { yield %[1]s == true }", func(v trinary.Value) trinary.Value { return boolState(v.IsTrue()) }},
		{"== false", "rule %[2]s = { yield %[1]s == false }", func(v trinary.Value) trinary.Value { return boolState((v == trinary.False)) }},
		{"!= false", "rule %[2]s = { yield %[1]s != false }", func(v trinary.Value) trinary.Value { return boolState(v != trinary.False) }},
		{"!= true", "rule %[2]s = { yield %[1]s != true }", func(v trinary.Value) trinary.Value { return boolState(!v.IsTrue()) }},
	}

	var b strings.Builder
	b.WriteString(importEnvelopeBase)
	b.WriteString("policy p {\n")
	for _, st := range importEnvelopeStates {
		fmt.Fprintf(&b, "  rule %s = import decision %s from t/q\n  export decision of %s\n", st.imp, st.source, st.imp)
	}
	for i, tc := range cases {
		for _, st := range importEnvelopeStates {
			name := fmt.Sprintf("c%d%s", i, st.suffix)
			fmt.Fprintf(&b, "  %s\n  export decision of %s\n", fmt.Sprintf(tc.rule, st.imp, name), name)
		}
	}
	b.WriteString("}\n")
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "import_matrix.sentrie", src: b.String()})

	// the bare import itself decides the imported state
	for _, st := range importEnvelopeStates {
		out, err := exec.ExecRule(ctx, "t", "p", st.imp, nil)
		s.Require().NoError(err)
		s.Equal(st.state, out.Decision.State, "bare import of %s", st.source)
	}

	for i, tc := range cases {
		for _, st := range importEnvelopeStates {
			name := fmt.Sprintf("c%d%s", i, st.suffix)
			s.Run(fmt.Sprintf("%s/%s", tc.name, st.suffix), func() {
				out, err := exec.ExecRule(ctx, "t", "p", name, nil)
				s.Require().NoError(err)
				s.Equal(tc.want(st.state), out.Decision.State, "imported state %s", st.state)
			})
		}
	}
}

// TestImportedDecisionKeepsStateValueAndAttachmentAccess: the envelope is still
// a dict inside the policy, so .state, .value and attachments read through it.
func (s *RuntimeTestSuite) TestImportedDecisionKeepsStateValueAndAttachmentAccess() {
	ctx := s.T().Context()
	src := importEnvelopeBase + `policy p {
  rule impF = import decision f from t/q
  rule impU = import decision u from t/q
  rule why = { yield impF.why }
  rule val = { yield impF.value }
  rule stateEq = { yield impF.state == false }
  rule stateUnknown = { yield impU.state }
  rule whyUnknown = { yield impU.why }
  rule held = {
    let kept = impF
    yield kept or false
  }
  rule heldUnknown = {
    let kept = impU
    yield kept or false
  }
  rule viaDeriveF = { yield isTrue(impF) }
  rule viaDeriveU = { yield isTrue(impU) }
  export decision of why
  export decision of val
  export decision of stateEq
  export decision of stateUnknown
  export decision of whyUnknown
  export decision of held
  export decision of heldUnknown
  export decision of viaDeriveF
  export decision of viaDeriveU
}
derive isTrue = (x: dict[trinary]): trinary => { yield x or false }
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "import_access.sentrie", src: src})

	run := func(rule string) *ExecutorOutput {
		out, err := exec.ExecRule(ctx, "t", "p", rule, nil)
		s.Require().NoError(err, rule)
		return out
	}

	s.Equal("denied", run("why").Decision.Value.Any())
	s.Equal(false, run("val").Decision.Value.Any())
	s.Equal(trinary.False, run("val").Decision.State)
	s.Equal(trinary.True, run("stateEq").Decision.State)
	s.Equal(trinary.Unknown, run("stateUnknown").Decision.Value.Any())
	s.Equal("undecided", run("whyUnknown").Decision.Value.Any())

	// the tag survives a `let` and a typed derive argument inside the policy
	s.Equal(trinary.False, run("held").Decision.State)
	s.Equal(trinary.Unknown, run("heldUnknown").Decision.State)
	s.Equal(trinary.False, run("viaDeriveF").Decision.State)
	s.Equal(trinary.Unknown, run("viaDeriveU").Decision.State)
}

// TestUserDictWithStateKeyKeepsDictTruthiness: only an import envelope is read
// as its state. A dict a policy builds with a "state" key, even one shaped like
// the envelope, stays an ordinary non-empty dict.
func (s *RuntimeTestSuite) TestUserDictWithStateKeyKeepsDictTruthiness() {
	ctx := s.T().Context()
	src := `namespace t
policy p {
  rule lookalike = { yield {"state": false} or false }
  rule lookalikeEq = { yield {"state": false} == false }
  rule forged = { yield {"state": false, "value": 1} }
  rule field = { yield {"state": false}.state }
  export decision of lookalike
  export decision of lookalikeEq
  export decision of forged
  export decision of field
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "user_dict.sentrie", src: src})

	run := func(rule string) *ExecutorOutput {
		out, err := exec.ExecRule(ctx, "t", "p", rule, nil)
		s.Require().NoError(err, rule)
		return out
	}

	s.Equal(trinary.True, run("lookalike").Decision.State)
	s.Equal(trinary.False, run("lookalikeEq").Decision.State)
	s.Equal(trinary.False, run("field").Decision.State)

	// a dict that merely looks like an envelope is reported as the dict it is
	forged := run("forged").Decision
	s.Equal(trinary.True, forged.State)
	raw, err := json.Marshal(forged)
	s.Require().NoError(err)
	s.JSONEq(`{"state":"true","value":{"state":false,"value":1}}`, string(raw))
}

// TestBareImportRuleReportsTheImportedDecision: the decision a bare-import rule
// exports is the imported state and value, not the envelope dict, and a policy
// that imports an importer sees it flat.
func (s *RuntimeTestSuite) TestBareImportRuleReportsTheImportedDecision() {
	ctx := s.T().Context()
	src := importEnvelopeBase + `policy p {
  rule impF = import decision f from t/q
  rule impT = import decision tr from t/q
  rule impU = import decision u from t/q
  export decision of impF attach layer as "p"
  export decision of impT
  export decision of impU
}
policy p2 {
  rule n = import decision impF from t/p
  rule layer = { yield n.layer }
  rule value = { yield n.value }
  rule inherited = { yield n.why }
  export decision of n
  export decision of layer
  export decision of value
  export decision of inherited
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "import_export.sentrie", src: src})

	for _, st := range importEnvelopeStates {
		source, err := exec.ExecRule(ctx, "t", "q", st.source, nil)
		s.Require().NoError(err)
		imported, err := exec.ExecRule(ctx, "t", "p", st.imp, nil)
		s.Require().NoError(err)

		s.Equal(st.state, imported.Decision.State, "state of imported %s", st.source)
		s.Equal(box.Trinary(st.state).Any(), imported.Decision.Value.Any(), "value of imported %s", st.source)

		// byte for byte what the exporting rule itself reports
		want, err := json.Marshal(source.Decision)
		s.Require().NoError(err)
		got, err := json.Marshal(imported.Decision)
		s.Require().NoError(err)
		s.JSONEq(string(want), string(got), "exported JSON of imported %s", st.source)
	}

	// imported attachments are not re-exported; the importing rule's own are
	imported, err := exec.ExecRule(ctx, "t", "p", "impF", nil)
	s.Require().NoError(err)
	s.Require().Contains(imported.Attachments, "layer")
	s.Equal("p", imported.Attachments["layer"].Any())
	s.Len(imported.Attachments, 1)

	// nested: p2 imports p's import of q, and gets the flat decision
	n, err := exec.ExecRule(ctx, "t", "p2", "n", nil)
	s.Require().NoError(err)
	s.Equal(trinary.False, n.Decision.State)
	s.Equal(false, n.Decision.Value.Any())
	s.Empty(n.Attachments, "attachments read through the import are not re-exported")

	layer, err := exec.ExecRule(ctx, "t", "p2", "layer", nil)
	s.Require().NoError(err)
	s.Equal("p", layer.Decision.Value.Any())

	value, err := exec.ExecRule(ctx, "t", "p2", "value", nil)
	s.Require().NoError(err)
	s.Equal(trinary.False, value.Decision.State)
	s.Equal(false, value.Decision.Value.Any(), "the nested value is the imported value, not an envelope")

	// q's own attachment is not carried through p, so p2 cannot see it
	inherited, err := exec.ExecRule(ctx, "t", "p2", "inherited", nil)
	s.Require().NoError(err)
	s.Equal(trinary.Unknown, inherited.Decision.State)
}

// TestImportedDecisionThroughAWithFactBinding: an import bound by `with` keeps
// its state in the target policy. Before the fix the binding crossed
// TryToBoundaryAny and arrived as a plain, truthy dict, so a false or unknown
// import read as true there (#136). A trinary fact takes the import's state.
func (s *RuntimeTestSuite) TestImportedDecisionThroughAWithFactBinding() {
	ctx := s.T().Context()
	src := importEnvelopeBase + `policy rt {
  fact flag: trinary
  rule via = { yield flag or false }
  rule isBool = { yield flag == false }
  export decision of via
  export decision of isBool
}
policy rd {
  fact flag: document
  rule via = { yield flag or false }
  export decision of via
}
policy rb {
  fact flag: dict[bool]
  rule via = { yield flag or false }
  export decision of via
}
policy rl {
  fact flag: dict[trinary]
  rule via = { yield flag or false }
  rule nested = { yield flag.state == false }
  export decision of via
  export decision of nested
}
policy rn {
  fact flags: list[dict[trinary]]
  rule via = { yield flags[0] or false }
  export decision of via
}
policy p {
  rule impF = import decision f from t/q
  rule impU = import decision u from t/q
  rule rtF = import decision via from t/rt with flag as impF
  rule rtU = import decision via from t/rt with flag as impU
  rule rtBool = import decision isBool from t/rt with flag as impF
  rule rdF = import decision via from t/rd with flag as impF
  rule rdU = import decision via from t/rd with flag as impU
  rule rbF = import decision via from t/rb with flag as impF
  rule rbU = import decision via from t/rb with flag as impU
  rule rlF = import decision via from t/rl with flag as impF
  rule rlU = import decision via from t/rl with flag as impU
  rule rnF = import decision via from t/rn with flags as [impF]
  rule rnU = import decision via from t/rn with flags as [impU]
  export decision of rtF
  export decision of rtU
  export decision of rtBool
  export decision of rdF
  export decision of rdU
  export decision of rbF
  export decision of rbU
  export decision of rlF
  export decision of rlU
  export decision of rnF
  export decision of rnU
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "import_with.sentrie", src: src})

	for _, c := range []struct {
		rule string
		want trinary.Value
	}{
		{"rtF", trinary.False}, {"rtU", trinary.Unknown}, {"rtBool", trinary.True},
		{"rdF", trinary.False}, {"rdU", trinary.Unknown},
		{"rbF", trinary.False}, {"rbU", trinary.Unknown},
		{"rlF", trinary.False}, {"rlU", trinary.Unknown},
		{"rnF", trinary.False}, {"rnU", trinary.Unknown},
	} {
		out, err := exec.ExecRule(ctx, "t", "p", c.rule, nil)
		s.Require().NoError(err, c.rule)
		s.Equal(c.want, out.Decision.State, c.rule)
	}
}

// TestImportedDecisionBoundToATruthType: a let, a derive parameter and a derive
// return declared bool or trinary accept an import and hold its state, not the
// envelope (#136). So does a fact default; an injected fact is covered through
// `with` above.
func (s *RuntimeTestSuite) TestImportedDecisionBoundToATruthType() {
	ctx := s.T().Context()
	src := importEnvelopeBase + `policy p {
  fact fl?: trinary default impF
  fact flu?: boolean default impU
  rule defF = { yield fl or false }
  rule defU = { yield flu or false }
  rule defIsState = { yield fl == false }
  rule defIsDict = { yield fl contains "state" }
  export decision of defF
  export decision of defU
  export decision of defIsState
  export decision of defIsDict
  rule impF = import decision f from t/q
  rule impU = import decision u from t/q
  rule letF = { let x: trinary = impF  yield x or false }
  rule letU = { let x: boolean = impU  yield x or false }
  rule letIsState = { let x: trinary = impF  yield x == false }
  rule paramF = { yield asIs(impF) }
  rule paramU = { yield asIs(impU) }
  rule paramIsState = { yield asIs(impF) == false }
  rule retF = { yield back(impF) == false }
  export decision of letF
  export decision of letU
  export decision of letIsState
  export decision of paramF
  export decision of paramU
  export decision of paramIsState
  export decision of retF
}
derive asIs = (x: trinary): trinary => { yield x }
derive back = (x: dict[trinary]): trinary => { yield x }
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "import_typed.sentrie", src: src})

	for _, c := range []struct {
		rule string
		want trinary.Value
	}{
		{"letF", trinary.False}, {"letU", trinary.Unknown}, {"letIsState", trinary.True},
		{"paramF", trinary.False}, {"paramU", trinary.Unknown}, {"paramIsState", trinary.True},
		{"retF", trinary.True},
		{"defF", trinary.False}, {"defU", trinary.Unknown}, {"defIsState", trinary.True},
		{"defIsDict", trinary.False},
	} {
		out, err := exec.ExecRule(ctx, "t", "p", c.rule, nil)
		s.Require().NoError(err, c.rule)
		s.Equal(c.want, out.Decision.State, c.rule)
		_, isEnvelope := out.Decision.Value.DecisionState()
		s.False(isEnvelope, c.rule)
	}
}
