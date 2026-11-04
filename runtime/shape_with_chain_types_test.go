// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"github.com/sentrie-sh/sentrie/trinary"
)

// The type of a field inherited through a `with` chain is enforced, not only
// its presence. Every level of the chain contributes a differently typed field,
// and an optional field two levels up stays optional.
func (s *RuntimeTestSuite) TestShapeWithChainEnforcesInheritedFieldTypes() {
	ctx := s.T().Context()

	decls := map[string]string{
		"A": "shape A { fa: number  fo?: string }",
		"B": "shape B with A { fb: string }",
		"C": "shape C with B { fc: boolean }",
	}
	orders := map[string][]string{
		"in order": {"A", "B", "C"},
		"reversed": {"C", "B", "A"},
	}
	want := map[string]trinary.Value{
		"ok":                  trinary.True,
		"ok_with_optional":    trinary.True,
		"grandparent_type":    trinary.False,
		"parent_type":         trinary.False,
		"own_type":            trinary.False,
		"optional_wrong_type": trinary.False,
		"missing_required":    trinary.False,
	}

	for name, order := range orders {
		s.Run(name, func() {
			src := "namespace com/app\n"
			for _, n := range order {
				src += decls[n] + "\n"
			}
			src += `
policy pol {
  rule ok = { yield { "fa": 1, "fb": "x", "fc": true } is C }
  rule ok_with_optional = { yield { "fa": 1, "fo": "y", "fb": "x", "fc": true } is C }
  rule grandparent_type = { yield { "fa": "1", "fb": "x", "fc": true } is C }
  rule parent_type = { yield { "fa": 1, "fb": 2, "fc": true } is C }
  rule own_type = { yield { "fa": 1, "fb": "x", "fc": "true" } is C }
  rule optional_wrong_type = { yield { "fa": 1, "fo": 5, "fb": "x", "fc": true } is C }
  rule missing_required = { yield { "fo": "y", "fb": "x", "fc": true } is C }
  export decision of ok
  export decision of ok_with_optional
  export decision of grandparent_type
  export decision of parent_type
  export decision of own_type
  export decision of optional_wrong_type
  export decision of missing_required
}
`
			_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "types.sentrie", src: src})
			for rule, state := range want {
				out, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
				s.Require().NoError(err)
				s.Equal(state, out.Decision.State, rule)
			}
		})
	}
}

// A policy-local shape composed with the top of a namespace-level chain gets
// every field of the chain, whatever order the chain is declared in.
func (s *RuntimeTestSuite) TestShapeWithChainPolicyLocalOverNamespaceChain() {
	ctx := s.T().Context()
	chain := map[string][]string{
		"in order": {"shape A { fa: number }", "shape B with A { fb: number }", "shape C with B { fc: number }"},
		"reversed": {"shape C with B { fc: number }", "shape B with A { fb: number }", "shape A { fa: number }"},
	}
	want := map[string]trinary.Value{
		"ok":         trinary.True,
		"no_a":       trinary.False,
		"no_b":       trinary.False,
		"no_c":       trinary.False,
		"no_own":     trinary.False,
		"a_wrong":    trinary.False,
		"own_wrong":  trinary.False,
		"all_absent": trinary.False,
	}
	for name, decls := range chain {
		s.Run(name, func() {
			src := "namespace com/app\n" + decls[0] + "\n" + decls[1] + "\n" + decls[2] + `
policy pol {
  shape L with C { fl: string }
  rule ok = { yield { "fa": 1, "fb": 1, "fc": 1, "fl": "x" } is L }
  rule no_a = { yield { "fb": 1, "fc": 1, "fl": "x" } is L }
  rule no_b = { yield { "fa": 1, "fc": 1, "fl": "x" } is L }
  rule no_c = { yield { "fa": 1, "fb": 1, "fl": "x" } is L }
  rule no_own = { yield { "fa": 1, "fb": 1, "fc": 1 } is L }
  rule a_wrong = { yield { "fa": "1", "fb": 1, "fc": 1, "fl": "x" } is L }
  rule own_wrong = { yield { "fa": 1, "fb": 1, "fc": 1, "fl": 1 } is L }
  rule all_absent = { yield {} is L }
  export decision of ok
  export decision of no_a
  export decision of no_b
  export decision of no_c
  export decision of no_own
  export decision of a_wrong
  export decision of own_wrong
  export decision of all_absent
}
`
			_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "local.sentrie", src: src})
			for rule, state := range want {
				out, err := exec.ExecRule(ctx, "com/app", "pol", rule, nil)
				s.Require().NoError(err)
				s.Equal(state, out.Decision.State, rule)
			}
		})
	}
}
