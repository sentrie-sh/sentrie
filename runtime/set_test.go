// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"encoding/json"

	"github.com/sentrie-sh/sentrie/trinary"
)

const setPolicy = `namespace set/app

policy p {
  fact scopes: set[string]
  fact required: set[string]
  fact maybe?: set[number]? default null

  let typed: set[number] = [3, 1, 2]

  rule fact_is_set = { yield scopes is set[string] and not (scopes is set[number]) }
  rule has_all = { yield is_subset(required, scopes) and is_superset(scopes, required) }
  rule member = { yield "read" in scopes and scopes contains "write" and not ("admin" in scopes) }
  rule counted = { yield count(scopes) == 3 }
  rule list_is_set = { yield [1, 2] is set[number] and not ([1, 1] is set[number]) }
  rule nullable = { yield maybe is set[number]? }
  rule cast_dedupes = { yield count(["b", "a", "b"] as set[string]) == 2 }
  rule typed_let = { yield typed is set[number] and typed[0] == 1 }
  rule all_ops = { yield union(required, ["z", "a"]) }
  rule common = { yield intersection(scopes, ["write", "admin", "read"]) }
  rule missing = { yield difference(["admin", "read"], scopes) }
  rule built = { yield set([{"b": 1}, [2], "x", 1, true, null, 1]) }
  rule fact_out = { yield scopes }

  export decision of fact_is_set
  export decision of has_all
  export decision of member
  export decision of counted
  export decision of list_is_set
  export decision of nullable
  export decision of cast_dedupes
  export decision of typed_let
  export decision of all_ops
  export decision of common
  export decision of missing
  export decision of built
  export decision of fact_out
}
`

var setFacts = map[string]any{
	"scopes":   []any{"write", "read", "list"},
	"required": []any{"read", "list"},
}

// TestSetTypeAndBuiltins covers #162 end to end: a JSON array bound to a
// set[T] fact is a set, the set builtins, `in`/`contains`/`count`, `is set[T]`,
// a cast, and sorted JSON output.
func (s *RuntimeTestSuite) TestSetTypeAndBuiltins() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": setPolicy})
	defer func() { s.NoError(exec.Close()) }()

	for _, rule := range []string{"fact_is_set", "has_all", "member", "counted", "list_is_set", "nullable", "cast_dedupes", "typed_let"} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "set/app", "p", rule, setFacts)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}

	values := map[string]string{
		"all_ops":  `["a","list","read","z"]`,
		"common":   `["read","write"]`,
		"missing":  `["admin"]`,
		"built":    `[null,true,1,"x",[2],{"b":1}]`,
		"fact_out": `["list","read","write"]`,
	}
	for rule, want := range values {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "set/app", "p", rule, setFacts)
			s.Require().NoError(err)
			raw, err := json.Marshal(out.Decision.Value)
			s.Require().NoError(err)
			s.JSONEq(want, string(raw))
			s.Equal(want, string(raw), "set output must be in sorted order")
		})
	}
}

// TestSetFactRejectsDuplicatesAndWrongElements checks the fact boundary: a
// JSON array bound to set[T] is validated, never silently deduplicated.
func (s *RuntimeTestSuite) TestSetFactRejectsDuplicatesAndWrongElements() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": setPolicy})
	defer func() { s.NoError(exec.Close()) }()

	_, err := exec.ExecRule(ctx, "set/app", "p", "counted", map[string]any{
		"scopes":   []any{"read", "read"},
		"required": []any{},
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "duplicate set element read")

	_, err = exec.ExecRule(ctx, "set/app", "p", "counted", map[string]any{
		"scopes":   []any{"read", 1},
		"required": []any{},
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "item is not valid")
}

// TestSetBuiltinElementErrors checks that a value with no set order is
// rejected with a clear error.
func (s *RuntimeTestSuite) TestSetBuiltinElementErrors() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{"policy.sentrie": `namespace set/err

policy p {
  rule unknown_elem = { yield set([unknown]) }
  rule second_bad = { yield is_superset(["a"], [unknown]) }

  export decision of unknown_elem
  export decision of second_bad
}
`})
	defer func() { s.NoError(exec.Close()) }()

	_, err := exec.ExecRule(ctx, "set/err", "p", "unknown_elem", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "set: a set cannot hold an unknown value")

	_, err = exec.ExecRule(ctx, "set/err", "p", "second_bad", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "is_superset: second argument: a set cannot hold an unknown value")
}

// TestSetOfBooleanHoldsImportedDecisionStates covers a set[boolean] that is
// given imported decisions (#136 with #162): each element is the decision's
// state, not its dict, at a typed let and at a cast; a decision and the bool
// it equals are duplicates; an unknown state cannot be held.
func (s *RuntimeTestSuite) TestSetOfBooleanHoldsImportedDecisionStates() {
	ctx := s.T().Context()
	src := importEnvelopeBase + `policy sp {
  rule impF = import decision f from t/q
  rule impT = import decision tr from t/q
  rule impU = import decision u from t/q

  let states: set[boolean] = [impT, impF]
  let dup: set[boolean] = [impT, true]
  let unk: set[boolean] = [impU]

  rule let_states = { yield states }
  rule let_member = { yield true in states and false in states and count(states) == 2 }
  rule cast_states = { yield [impT, true, impF] as set[boolean] }
  rule nullable_cast = { yield ["b", "a", "b"] as set[string]? }
  rule nullable_cast_null = { yield (null as set[string]?) == null }
  rule let_dup = { yield count(dup) }
  rule let_unk = { yield count(unk) }

  export decision of let_states
  export decision of let_member
  export decision of cast_states
  export decision of nullable_cast
  export decision of nullable_cast_null
  export decision of let_dup
  export decision of let_unk
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "set_envelope.sentrie", src: src})

	values := map[string]string{
		"let_states":    `[false,true]`,
		"cast_states":   `[false,true]`,
		"nullable_cast": `["a","b"]`,
	}
	for rule, want := range values {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "t", "sp", rule, nil)
			s.Require().NoError(err)
			s.True(out.Decision.Value.IsSet(), "%s must yield a set", rule)
			raw, err := json.Marshal(out.Decision.Value)
			s.Require().NoError(err)
			s.Equal(want, string(raw))
		})
	}
	for _, rule := range []string{"let_member", "nullable_cast_null"} {
		out, err := exec.ExecRule(ctx, "t", "sp", rule, nil)
		s.Require().NoError(err)
		s.Equal(trinary.True, out.Decision.State, rule)
	}

	_, err := exec.ExecRule(ctx, "t", "sp", "let_dup", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "duplicate set element true")

	_, err = exec.ExecRule(ctx, "t", "sp", "let_unk", nil)
	s.Require().Error(err)
	s.Contains(err.Error(), "a set cannot hold an unknown value")
}
