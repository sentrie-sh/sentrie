// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/sentrie-sh/sentrie/trinary"

// TestStdlibBuiltinsEndToEnd runs the glob, object and units builtins through
// the real dispatcher, including from inside a derive (they are pure).
func (s *RuntimeTestSuite) TestStdlibBuiltinsEndToEnd() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  let cfg = { "a": { "b": 1, "c": 2 }, "z": [10, 20], "k": "v" }

  derive pure_glob = (p: string) => { yield glob_match("/repos/*/issues/**", ["/"], p) }
  derive pure_get = (m: dict[number]) => { yield object_get(m, ["a", "b"], 0) }

  rule glob_yes = { yield glob_match("/repos/*/issues/**", ["/"], "/repos/x/issues/1/c") }
  rule glob_no = { yield not glob_match("/repos/*/issues/**", ["/"], "/repos/x/y/issues/1") }
  rule glob_derive = { yield pure_glob("/repos/q/issues/9") and pure_get({ "a": { "b": 1 } }) == 1 }
  rule get_hit = { yield object_get(cfg, ["z", 1], 0) == 20 }
  rule get_default = { yield object_get(cfg, ["a", "nope"], "d") == "d" }
  rule keys_sorted = { yield object_keys(cfg) == ["a", "k", "z"] }
  rule remove_ok = { yield object_keys(object_remove(cfg, ["a", "z"])) == ["k"] }
  rule filter_ok = { yield object_keys(object_filter(cfg, ["a", "k", "missing"])) == ["a", "k"] }
  rule union_ok = { yield object_get(object_union(cfg, { "a": { "c": 9, "d": 4 } }), ["a", "c"], 0) == 9 }
  rule union_keeps = { yield object_get(object_union(cfg, { "a": { "c": 9, "d": 4 } }), ["a", "b"], 0) == 1 }
  rule quantity_ok = { yield parse_quantity("10Gi") == 10737418240 and parse_quantity("500m") == 0.5 and parse_quantity("2k") == 2000 }
  rule duration_ok = { yield parse_duration("2h") == 7200 and parse_duration("1h30m") == 5400 and parse_duration("250ms") == 0.25 }

  rule bad_glob = { yield glob_match("[", ["/"], "x") }
  rule bad_quantity = { yield parse_quantity("10GB") > 0 }
  rule bad_duration = { yield parse_duration("1d") > 0 }

  export decision of glob_yes
  export decision of glob_no
  export decision of glob_derive
  export decision of get_hit
  export decision of get_default
  export decision of keys_sorted
  export decision of remove_ok
  export decision of filter_ok
  export decision of union_ok
  export decision of union_keeps
  export decision of quantity_ok
  export decision of duration_ok
  export decision of bad_glob
  export decision of bad_quantity
  export decision of bad_duration
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "stdlib.sentrie", src: src})
	for _, rule := range []string{
		"glob_yes", "glob_no", "glob_derive", "get_hit", "get_default", "keys_sorted",
		"remove_ok", "filter_ok", "union_ok", "union_keeps", "quantity_ok", "duration_ok",
	} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().NoError(err)
			s.Equal(trinary.True, out.Decision.State)
		})
	}
	for rule, msg := range map[string]string{
		"bad_glob":     "unterminated character class",
		"bad_quantity": "unknown suffix",
		"bad_duration": "parse_duration",
	} {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "com/ex", "pol", rule, nil)
			s.Require().Error(err)
			s.Contains(err.Error(), msg)
		})
	}
}
