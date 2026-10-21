// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"fmt"

	"github.com/sentrie-sh/sentrie/trinary"
)

// exportRestrictionPack has two policies that `use` different named exports of
// one module. Each also has a rule that reaches for the other policy's export.
var exportRestrictionPack = map[string]string{
	"shared.ts": `
export function a(): string { return "a"; }
export function b(): string { return "b"; }
`,
	"policies.sentrie": `namespace exports/test

policy pa {
  use { a } from "./shared" as m

  rule own = default false {
    yield m.a() == "a"
  }

  rule other = default false {
    yield m.b() == "b"
  }

  export decision of own
  export decision of other
}

policy pb {
  use { b } from "./shared" as m

  rule own = default false {
    yield m.b() == "b"
  }

  rule other = default false {
    yield m.a() == "a"
  }

  export decision of own
  export decision of other
}
`,
}

// TestUseExportRestrictionIsPerUseStatement covers #102: two policies that use
// different named exports of the same module, evaluated through one executor in
// both orders, each reach their own export and never the other's.
func (s *RuntimeTestSuite) TestUseExportRestrictionIsPerUseStatement() {
	orders := map[string][]string{
		"pa-then-pb": {"pa", "pb"},
		"pb-then-pa": {"pb", "pa"},
	}
	for name, order := range orders {
		s.Run(name, func() {
			ctx := s.T().Context()
			exec := s.newTempPackExecutor(exportRestrictionPack)
			for range 3 {
				for _, policy := range order {
					own, err := exec.ExecRule(ctx, "exports/test", policy, "own", map[string]any{})
					s.Require().NoError(err, "policy %s must reach its own export", policy)
					s.Equal(trinary.True, own.Decision.State)

					_, err = exec.ExecRule(ctx, "exports/test", policy, "other", map[string]any{})
					s.Require().Error(err, "policy %s must not reach the other policy's export", policy)
					s.Contains(err.Error(), "not found in module")
				}
			}
		})
	}
}

// TestUseExportRestrictionUnderExecPolicy evaluates both policies concurrently
// through ExecPolicy, where rule goroutines race to load the shared module.
func (s *RuntimeTestSuite) TestUseExportRestrictionUnderExecPolicy() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"shared.ts": exportRestrictionPack["shared.ts"],
		"policies.sentrie": `namespace exports/test

policy pa {
  use { a } from "./shared" as m
  rule r1 = { yield m.a() == "a" }
  rule r2 = { yield m.a() == "a" }
  export decision of r1
  export decision of r2
}

policy pb {
  use { b } from "./shared" as m
  rule r1 = { yield m.b() == "b" }
  rule r2 = { yield m.b() == "b" }
  export decision of r1
  export decision of r2
}
`,
	})
	for range 5 {
		for _, policy := range []string{"pa", "pb"} {
			outputs, err := exec.ExecPolicy(ctx, "exports/test", policy, map[string]any{})
			s.Require().NoError(err)
			s.Require().Len(outputs, 2)
			for _, out := range outputs {
				s.Equal(trinary.True, out.Decision.State)
			}
		}
	}
}

// TestUseMissingRequiredExportIsReportedPerUse checks that a `use` naming an
// export the module lacks fails at bind time even after another `use` of the
// same module has loaded it successfully.
func (s *RuntimeTestSuite) TestUseMissingRequiredExportIsReportedPerUse() {
	ctx := s.T().Context()
	exec := s.newTempPackExecutor(map[string]string{
		"shared.ts": exportRestrictionPack["shared.ts"],
		"policies.sentrie": `namespace exports/test

policy good {
  use { a } from "./shared" as m
  rule r = { yield m.a() == "a" }
  export decision of r
}

policy bad {
  use { c } from "./shared" as m
  rule r = { yield true }
  export decision of r
}
`,
	})

	out, err := exec.ExecRule(ctx, "exports/test", "good", "r", map[string]any{})
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)

	_, err = exec.ExecRule(ctx, "exports/test", "bad", "r", map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), `missing required export "c"`)
}

// TestTwoUsesOfOneModuleInOnePolicy is the layout from #102's update: one policy
// with two aliases of the same module, each naming a different export. Under
// ExecPolicy the rules bind concurrently, so before the fix one alias received
// the other's export set and its calls failed with "not found in module".
func (s *RuntimeTestSuite) TestTwoUsesOfOneModuleInOnePolicy() {
	ctx := s.T().Context()
	policy := "namespace exports/test\n\npolicy p {\n" +
		"  use { a } from \"./shared\" as ma\n" +
		"  use { b } from \"./shared\" as mb\n"
	for i := range 8 {
		policy += fmt.Sprintf("  rule ra%d = { yield ma.a() == \"a\" }\n  export decision of ra%d\n", i, i)
		policy += fmt.Sprintf("  rule rb%d = { yield mb.b() == \"b\" }\n  export decision of rb%d\n", i, i)
	}
	policy += "}\n"

	exec := s.newTempPackExecutor(map[string]string{
		"shared.ts":        exportRestrictionPack["shared.ts"],
		"policies.sentrie": policy,
	})
	for range 10 {
		outputs, err := exec.ExecPolicy(ctx, "exports/test", "p", map[string]any{})
		s.Require().NoError(err)
		s.Require().Len(outputs, 16)
		for _, out := range outputs {
			s.Equal(trinary.True, out.Decision.State, out.RuleName)
		}
	}
}
