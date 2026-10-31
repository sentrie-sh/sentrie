// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

// TestStdlibBuiltinsInDeriveAreDeterministic calls every glob, object and
// units builtin from inside derive bodies and checks the produced values (not
// only a truth state) and that repeated executions agree, including the
// key order of object_keys.
func (s *RuntimeTestSuite) TestStdlibBuiltinsInDeriveAreDeterministic() {
	ctx := s.T().Context()
	src := `namespace com/ex

derive d_glob = (p: string) => { yield glob_match("**/issues/*", ["/"], p) }
derive d_keys = (m: dict[number]) => { yield object_keys(m) }
derive d_get = (m: document) => { yield object_get(m, ["rows", 1, "id"], -1) }
derive d_get_missing = (m: document) => { yield object_get(m, ["rows", 7, "id"], "none") }
derive d_remove = (m: dict[number]) => { yield object_remove(m, ["b"]) }
derive d_filter = (m: dict[number]) => { yield object_filter(m, ["a", "c", "zz"]) }
derive d_union = () => { yield object_union({ "n": { "p": 1, "q": 2 }, "l": [1, 2] }, { "n": { "q": 9 }, "l": [3] }) }
derive d_quantity = (q: string) => { yield parse_quantity(q) }
derive d_duration = (q: string) => { yield parse_duration(q) }

policy pol {
  rule all_values = {
    let m = { "c": 3, "a": 1, "b": 2, "d": 4 }
    yield [
      d_glob("/x/issues/1"),
      d_glob("/x/y/issues/1/z"),
      d_keys(m),
      d_get({ "rows": [{ "id": 1 }, { "id": 2 }] }),
      d_get_missing({ "rows": [{ "id": 1 }] }),
      d_remove(m),
      d_filter(m),
      d_union(),
      d_quantity("1.5Gi"),
      d_quantity("300m"),
      d_duration("1h30m"),
      d_duration("250ms")
    ]
  }
  export decision of all_values
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "stdlib_derive.sentrie", src: src})

	want := []any{
		true,
		false,
		[]any{"a", "b", "c", "d"},
		2.0,
		"none",
		map[string]any{"a": 1.0, "c": 3.0, "d": 4.0},
		map[string]any{"a": 1.0, "c": 3.0},
		map[string]any{"n": map[string]any{"p": 1.0, "q": 9.0}, "l": []any{3.0}},
		1610612736.0,
		0.3,
		5400.0,
		0.25,
	}
	for i := range 10 {
		out, err := exec.ExecRule(ctx, "com/ex", "pol", "all_values", nil)
		s.Require().NoError(err, "run %d", i)
		s.Equal(want, out.Decision.Value.Any(), "run %d", i)
	}
}

// TestStdlibBuiltinsUndefinedPropagatesThroughRules checks that an undefined
// argument yields an undefined result rather than an error, and that null is
// an error, through the real dispatcher.
func (s *RuntimeTestSuite) TestStdlibBuiltinsUndefinedPropagatesThroughRules() {
	ctx := s.T().Context()
	src := `namespace com/ex

policy pol {
  fact doc: document? as doc
  rule undef_glob = { yield glob_match("a", ["/"], doc.missing) }
  rule undef_quantity = { yield parse_quantity(doc.missing) }
  rule undef_keys = { yield object_keys(doc.missing) }
  rule null_quantity = { yield parse_quantity(null) }
  rule null_glob = { yield glob_match(null, ["/"], "a") }
  export decision of undef_glob
  export decision of undef_quantity
  export decision of undef_keys
  export decision of null_quantity
  export decision of null_glob
}
`
	_, exec := s.mustBuildDeriveExecutor(ctx, deriveTestProgram{name: "stdlib_undef.sentrie", src: src})
	facts := map[string]any{"doc": map[string]any{"present": 1.0}}

	for _, rule := range []string{"undef_glob", "undef_quantity", "undef_keys"} {
		s.Run(rule, func() {
			out, err := exec.ExecRule(ctx, "com/ex", "pol", rule, facts)
			s.Require().NoError(err)
			s.True(out.Decision.Value.IsUndefined(), "got %v", out.Decision.Value)
		})
	}
	for rule, msg := range map[string]string{
		"null_quantity": "parse_quantity: argument must be a string",
		"null_glob":     "glob_match: pattern must be a string",
	} {
		s.Run(rule, func() {
			_, err := exec.ExecRule(ctx, "com/ex", "pol", rule, facts)
			s.Require().Error(err)
			s.Contains(err.Error(), msg)
		})
	}
}
