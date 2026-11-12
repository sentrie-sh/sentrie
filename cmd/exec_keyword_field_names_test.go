// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
)

const keywordFieldsCmdPolicy = `namespace com/kw
shape Doc {
  title: string
  version: string
  default?: number
  in: boolean
  not: boolean
  is: number
  contains: list[number]
}
policy pol {
  fact d: Doc
  fact obj: document
  rule get_title = { yield d.title }
  rule nested = { yield obj.true.null }
  rule is_set = { yield d.is is defined }
  rule is_member = { yield 2 in [d.is] }
  rule has_seven = { yield d.contains contains 7 }
  rule ops = { yield d.in and not d.not }
  rule literal_key = { yield {"title": 1}.title }
  export decision of is_set
    attach version as d.version
    attach default as d.default
    attach as as d.title
  export decision of get_title
  export decision of nested
  export decision of is_member
  export decision of has_seven
  export decision of ops
  export decision of literal_key
}
`

const keywordFieldsFacts = `{"d": {"title": "T", "version": "1.2", "default": 5, "in": true, "not": false, "is": 2, "contains": [1, 7]}, "obj": {"true": {"null": "deep"}}}`

func (s *CmdTestSuite) execKeywordFields(dir, rule, facts string) (map[string]any, error) {
	s.T().Helper()
	var runErr error
	out := s.captureStdout(func() {
		runErr = runExecCLI(s.T().Context(), []string{"com/kw/pol/" + rule, "--pack-location", dir, "--facts", facts, "--output", "json"})
	})
	if runErr != nil {
		return nil, runErr
	}
	var decoded []map[string]any
	s.Require().NoError(json.Unmarshal([]byte(out), &decoded), out)
	s.Require().Len(decoded, 1)
	return decoded[0], nil
}

// TestExecKeywordFieldNamesEvaluate covers #159: keyword field names after `.`
// read the right values, and `is`, `in` and `contains` after a keyword field
// still act as operators.
func (s *CmdTestSuite) TestExecKeywordFieldNamesEvaluate() {
	dir := s.writeTestPack(keywordFieldsCmdPolicy)
	value := func(rule string) any {
		res, err := s.execKeywordFields(dir, rule, keywordFieldsFacts)
		s.Require().NoError(err, rule)
		return res["decision"].(map[string]any)["value"]
	}
	s.Equal("T", value("get_title"))
	s.Equal("deep", value("nested"))
	s.Equal(true, value("is_set"))
	s.Equal(true, value("is_member"))
	s.Equal(true, value("has_seven"))
	s.Equal(true, value("ops"))
	s.EqualValues(1, value("literal_key"))
}

// TestExecKeywordAttachmentNames covers #159: keyword attachment names appear
// in the output attachments.
func (s *CmdTestSuite) TestExecKeywordAttachmentNames() {
	dir := s.writeTestPack(keywordFieldsCmdPolicy)
	res, err := s.execKeywordFields(dir, "is_set", keywordFieldsFacts)
	s.Require().NoError(err)
	att := res["attachments"].(map[string]any)
	s.Equal("1.2", att["version"])
	s.EqualValues(5, att["default"])
	s.Equal("T", att["as"])
}

// TestExecKeywordShapeFieldsValidatedAgainstFacts covers #159: shape fields
// named with keywords are enforced when the facts are validated.
func (s *CmdTestSuite) TestExecKeywordShapeFieldsValidatedAgainstFacts() {
	dir := s.writeTestPack(keywordFieldsCmdPolicy)

	// optional `default` may be omitted
	_, err := s.execKeywordFields(dir, "get_title", `{"d": {"title": "T", "version": "1", "in": true, "not": false, "is": 2, "contains": []}, "obj": {}}`)
	s.Require().NoError(err)

	// required `version` missing
	_, err = s.execKeywordFields(dir, "get_title", `{"d": {"title": "T", "in": true, "not": false, "is": 2, "contains": []}, "obj": {}}`)
	s.Require().Error(err)
	s.Contains(err.Error(), "version")

	// wrong type for `default`
	_, err = s.execKeywordFields(dir, "get_title", `{"d": {"title": "T", "version": "1", "default": "x", "in": true, "not": false, "is": 2, "contains": []}, "obj": {}}`)
	s.Require().Error(err)
	s.Contains(err.Error(), "default")
}
