// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package box

import "github.com/sentrie-sh/sentrie/trinary"

func (s *BoxTestSuite) TestTrinaryFrom_matchesFromAny() {
	cases := []Value{
		Undefined(),
		Null(),
		Bool(true),
		Bool(false),
		Number(0),
		Number(3.14),
		String(""),
		String("hello"),
		String("true"),
		Trinary(trinary.True),
		Trinary(trinary.False),
		Trinary(trinary.Unknown),
		List(nil),
		List([]Value{}),
		List([]Value{Number(0)}),
		Dict(nil),
		Dict(map[string]Value{}),
		Dict(map[string]Value{"a": Number(1)}),
		Object(struct{ X int }{1}),
		Object(map[string]any{"k": 1}),
	}
	for _, b := range cases {
		s.Run(b.String(), func() {
			s.Equal(trinary.From(b.Any()), TrinaryFrom(b))
		})
	}
}

func (s *BoxTestSuite) TestTrinaryFrom_decisionEnvelopeUsesState() {
	for _, state := range []trinary.Value{trinary.True, trinary.False, trinary.Unknown} {
		env := DecisionEnvelope(map[string]Value{"state": Trinary(state), "value": Bool(true)}, state)
		got, ok := env.DecisionState()
		s.True(ok)
		s.Equal(state, got)
		s.Equal(state, TrinaryFrom(env))
		s.Equal(ValueDict, env.Kind())
	}

	plain := Dict(map[string]Value{"state": Bool(false)})
	_, ok := plain.DecisionState()
	s.False(ok)
	s.Equal(trinary.True, TrinaryFrom(plain))
}

func (s *BoxTestSuite) TestEqualValues_decisionEnvelopeComparesState() {
	f := DecisionEnvelope(map[string]Value{"value": Bool(false)}, trinary.False)
	u := DecisionEnvelope(map[string]Value{"value": Undefined()}, trinary.Unknown)
	s.False(EqualValues(f, Bool(true)))
	s.True(EqualValues(Bool(false), f))
	s.True(EqualValues(f, Trinary(trinary.False)))
	s.False(EqualValues(u, Bool(true)))
	s.False(EqualValues(u, Bool(false)))
	s.False(EqualValues(Dict(map[string]Value{"state": Bool(false)}), Bool(false)))
}

func (s *BoxTestSuite) TestFromBoundaryAny_keepsAValueAndNeverTagsBoundaryData() {
	env := DecisionEnvelope(map[string]Value{"state": Trinary(trinary.False)}, trinary.False)
	state, ok := FromBoundaryAny(env).DecisionState()
	s.True(ok)
	s.Equal(trinary.False, state)

	// what JSON or a JS call hands over is plain data and decodes untagged
	_, ok = FromBoundaryAny(ToBoundaryAny(env)).DecisionState()
	s.False(ok)
	_, ok = FromBoundaryAny(map[string]any{"state": false, "value": false}).DecisionState()
	s.False(ok)
}
