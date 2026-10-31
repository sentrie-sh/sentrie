// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

// TestValidateCmdChecksStdlibBuiltinCalls is the user-visible side of #164:
// `sentrie validate` rejects a malformed glob, object or units call without
// executing the policy.
func (s *CmdTestSuite) TestValidateCmdChecksStdlibBuiltinCalls() {
	cases := []struct {
		name string
		expr string
		msg  string
	}{
		{"glob arity", `glob_match("a", ["/"])`, "glob_match requires 3 arguments"},
		{"glob delimiters kind", `glob_match("a", "/", "b")`, "glob_match: delimiters must be a list"},
		{"get arity", `object_get({}, "a")`, "object_get requires 3 arguments"},
		{"get path kind", `object_get({}, true, 0)`, "object_get: path must be a string, a number or a list of segments"},
		{"keys kind", `object_keys([1])`, "object_keys: argument must be a dict"},
		{"remove arity", `object_remove({})`, "object_remove requires 2 arguments"},
		{"filter keys kind", `object_filter({}, "a")`, "object_filter: keys must be a list of strings"},
		{"union kind", `object_union({}, 1)`, "object_union: second argument must be a dict"},
		{"quantity kind", `parse_quantity(10)`, "parse_quantity: argument must be a string"},
		{"duration arity", `parse_duration()`, "parse_duration requires 1 argument"},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			dir := s.writeTestPack("namespace com/ex\npolicy pol {\n  rule r = { yield " + tc.expr + " }\n  export decision of r\n}\n")
			err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
			s.Require().Error(err)
			s.Contains(err.Error(), tc.msg)
		})
	}

	s.Run("well formed calls validate", func() {
		dir := s.writeTestPack(`namespace com/ex
policy pol {
  rule r = {
    yield glob_match("/a/**", ["/"], "/a/b") and object_get({ "a": [1] }, ["a", 0], 0) == 1 and parse_quantity("1Gi") > 0 and parse_duration("1s") == 1
  }
  export decision of r
}
`)
		s.Require().NoError(runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir}))
	})
}
