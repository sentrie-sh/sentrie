// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// factBindingCmdPolicy echoes its facts, so a test reads what was bound.
const factBindingCmdPolicy = `namespace com/ex
policy pol {
  fact user: number as u
  fact opts?: any as o default {}
  rule userEcho = { yield u }
  rule optsEcho = { yield o }
  export decision of userEcho
  export decision of optsEcho
}
`

func (s *CmdTestSuite) writeFactFile(content string) string {
	path := filepath.Join(s.T().TempDir(), "facts.json")
	s.Require().NoError(os.WriteFile(path, []byte(content), 0o600))
	return path
}

// execRuleValue runs `exec com/ex/pol/<rule>` and returns the rule's value as JSON.
func (s *CmdTestSuite) execRuleValue(rule string, args ...string) (string, string, error) {
	dir := s.writeTestPack(factBindingCmdPolicy)
	var err error
	out := s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), append([]string{"com/ex/pol/" + rule, "--pack-location", dir, "--output", "json"}, args...))
	})
	if err != nil {
		return "", out, err
	}
	var got []struct {
		Decision struct {
			Value json.RawMessage `json:"value"`
		} `json:"decision"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &got), out)
	s.Require().Len(got, 1)
	return string(got[0].Decision.Value), out, nil
}

// exec takes a fact by its declared name and binds it under the alias; the
// alias alone does not supply it.
func (s *CmdTestSuite) TestExecFactsAreSuppliedByDeclaredName() {
	value, _, err := s.execRuleValue("userEcho", "--facts", `{"user": 41}`)
	s.Require().NoError(err)
	s.JSONEq(`41`, value)

	_, out, err := s.execRuleValue("userEcho", "--facts", `{"u": 41}`)
	s.Require().Error(err)
	s.Contains(err.Error(), "required fact not found: user")
	s.Empty(out)
}

// A wrongly typed fact is an error on exec, for a rule and for a policy, not
// a per-rule result.
func (s *CmdTestSuite) TestExecWronglyTypedFactIsAnError() {
	_, out, err := s.execRuleValue("userEcho", "--facts", `{"user": "x"}`)
	s.Require().Error(err)
	s.Contains(err.Error(), "fact 'user'")
	s.Contains(err.Error(), "is not a number")
	s.Empty(out)

	dir := s.writeTestPack(factBindingCmdPolicy)
	out = s.captureStdout(func() {
		err = runExecCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir, "--facts", `{"user": "x"}`})
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "fact 'user'")
	s.NotContains(err.Error(), "rules failed")
	s.Empty(out)
}

// An imported policy's wrongly typed `with` fact is an error on exec too, the
// same as a caller's wrongly typed fact (#147, user decision 11), for the
// importing rule and for the whole policy.
func (s *CmdTestSuite) TestExecImportedWronglyTypedFactIsAnError() {
	dir := s.writeTestPack(`namespace com/ex
policy target {
  fact n?: number
  rule ok = { yield true }
  export decision of ok
}
policy importer {
  rule fine = { yield true }
  rule viaImport = import decision ok from com/ex/target with n as "x"
  export decision of fine
  export decision of viaImport
}
`)
	for _, target := range []string{"com/ex/importer/viaImport", "com/ex/importer"} {
		var err error
		out := s.captureStdout(func() {
			err = runExecCLI(s.T().Context(), []string{target, "--pack-location", dir})
		})
		s.Require().Error(err, target)
		s.Contains(err.Error(), "imported policy com/ex/target: fact 'n'", target)
		s.Contains(err.Error(), "is not a number", target)
		s.NotContains(err.Error(), "rules failed", target)
		s.Empty(out, target)
	}
}

// --facts is merged over --fact-file: objects deeply, anything else replaced.
func (s *CmdTestSuite) TestExecFactFileAndFactsMergeDeeply() {
	file := s.writeFactFile(`{"user": 1, "opts": {"a": {"x": 1, "y": 0}, "b": 3, "items": [1, 2, 3]}}`)

	value, _, err := s.execRuleValue("optsEcho", "--fact-file", file,
		"--facts", `{"opts": {"a": {"y": 2}, "items": [9]}}`)
	s.Require().NoError(err)
	s.JSONEq(`{"a": {"x": 1, "y": 2}, "b": 3, "items": [9]}`, value)

	value, _, err = s.execRuleValue("userEcho", "--fact-file", file, "--facts", `{"user": 41}`)
	s.Require().NoError(err)
	s.JSONEq(`41`, value, "a scalar from --facts wins")

	value, _, err = s.execRuleValue("userEcho", "--fact-file", file)
	s.Require().NoError(err)
	s.JSONEq(`1`, value, "the file alone still supplies")
}

// An object in --facts replaces a non-object from the file, and the reverse.
func (s *CmdTestSuite) TestMergeFactsReplacesMismatchedKinds() {
	dst := map[string]any{"a": map[string]any{"x": 1}, "b": 2, "c": []any{1, 2}}
	src := map[string]any{"a": 5, "b": map[string]any{"y": 1}, "c": []any{3}}
	s.Equal(map[string]any{"a": 5, "b": map[string]any{"y": 1}, "c": []any{3}}, mergeFacts(dst, src))
}

// Merging never writes into the nested objects of its inputs.
func (s *CmdTestSuite) TestMergeFactsDoesNotMutateNestedInputs() {
	file := map[string]any{"opts": map[string]any{"a": 1}}
	flag := map[string]any{"opts": map[string]any{"b": 2}}
	merged := mergeFacts(mergeFacts(nil, file), flag)
	s.Equal(map[string]any{"opts": map[string]any{"a": 1, "b": 2}}, merged)
	s.Equal(map[string]any{"opts": map[string]any{"a": 1}}, file)
	s.Equal(map[string]any{"opts": map[string]any{"b": 2}}, flag)
}

// bench merges the fact file and --facts the same way exec does.
func (s *CmdTestSuite) TestBenchFactFileAndFlagMergeDeeply() {
	file := s.writeFactFile(`{"user": 1, "opts": {"a": {"x": 1, "y": 0}, "items": [1, 2, 3]}}`)
	facts, err := loadBenchFacts(file, `{"opts": {"a": {"y": 2}, "items": [9]}}`)
	s.Require().NoError(err)
	s.Equal(map[string]any{
		"user": float64(1),
		"opts": map[string]any{"a": map[string]any{"x": float64(1), "y": float64(2)}, "items": []any{float64(9)}},
	}, facts)
}

// A wrongly typed literal default fails `validate` before any request.
func (s *CmdTestSuite) TestValidateRejectsWronglyTypedFactDefault() {
	for name, fact := range map[string]string{
		"literal of the wrong kind": `fact n?: number default "x"`,
		"null for a non-nullable":   `fact n?: number default null`,
	} {
		s.Run(name, func() {
			dir := s.writeTestPack("namespace com/ex\npolicy pol {\n  " + fact + "\n  rule r = { yield true }\n  export decision of r\n}\n")
			err := runValidateCLI(s.T().Context(), []string{"--pack-location", dir})
			s.Require().Error(err)
			s.Contains(err.Error(), "fact 'n'")
		})
	}
}
