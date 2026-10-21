// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package capabilities

import (
	"os"
	"path/filepath"
	"strings"
)

func (s *CapabilitiesTestSuite) TestNilGrantsEverything() {
	var c *Capabilities
	s.True(c.AllowsBuiltin("now"))
	s.True(c.AllowsModule("@sentrie/uuid"))
	s.True(c.AllowsLocalModules())
}

func (s *CapabilitiesTestSuite) TestAbsentFieldsGrantEverything() {
	c, err := Parse([]byte(`{"version": 1}`))
	s.Require().NoError(err)
	s.True(c.AllowsBuiltin("now"))
	s.True(c.AllowsModule("@sentrie/uuid"))
	s.True(c.AllowsLocalModules())

	c, err = Parse([]byte(`{"version": 1, "builtins": null, "modules": null, "local_modules": null}`))
	s.Require().NoError(err)
	s.True(c.AllowsBuiltin("now"))
	s.True(c.AllowsModule("@sentrie/uuid"))
	s.True(c.AllowsLocalModules())
}

func (s *CapabilitiesTestSuite) TestListsRestrict() {
	c, err := Parse([]byte(`{
  "version": 1,
  "builtins": ["count", "filter"],
  "modules": ["@sentrie/hash"],
  "local_modules": false
}`))
	s.Require().NoError(err)
	s.True(c.AllowsBuiltin("count"))
	s.False(c.AllowsBuiltin("now"))
	s.True(c.AllowsModule("@sentrie/hash"))
	s.False(c.AllowsModule("@sentrie/uuid"))
	s.False(c.AllowsLocalModules())
}

func (s *CapabilitiesTestSuite) TestEmptyListsGrantNothing() {
	c, err := Parse([]byte(`{"version": 1, "builtins": [], "modules": [], "local_modules": true}`))
	s.Require().NoError(err)
	s.False(c.AllowsBuiltin("count"))
	s.False(c.AllowsModule("@sentrie/hash"))
	s.True(c.AllowsLocalModules())
}

func (s *CapabilitiesTestSuite) TestUnknownNamesLoad() {
	c, err := Parse([]byte(`{"version": 1, "builtins": ["now_ms"], "modules": ["@sentrie/future"]}`))
	s.Require().NoError(err)
	s.True(c.AllowsBuiltin("now_ms"))
	s.False(c.AllowsBuiltin("now"))
}

func (s *CapabilitiesTestSuite) TestParseErrors() {
	cases := map[string]string{
		`{}`:                                       `missing "version"`,
		`{"version": 2}`:                           "unsupported version 2",
		`{"version": 1, "extra": true}`:            `unknown field "extra"`,
		`{"version": 1} {}`:                        "unexpected data after the top-level object",
		`{"version": 1, "builtins": [""]}`:         "builtins: empty name",
		`{"version": 1, "modules": ["hash"]}`:      `modules: "hash" must be a full specifier`,
		`{"version": 1, "modules": ["@sentrie/"]}`: `modules: "@sentrie/" must be a full specifier`,
		`{"version": 1, "builtins": "count"}`:      "cannot unmarshal",
		`not json`:                                 "capabilities:",
	}
	for input, want := range cases {
		_, err := Parse([]byte(input))
		s.Require().Error(err, input)
		s.Contains(err.Error(), want, input)
	}
}

func (s *CapabilitiesTestSuite) TestParseErrorsNamePositionsAndShape() {
	cases := map[string]string{
		"":        "capabilities: empty file",
		"  \n\t ": "capabilities: empty file",
		"null":    "the top-level value must be an object",
		"[1]":     "the top-level value must be an object",
		"{\n  \"version\": 1,\n  \"builtins\": [\"a\",]\n}":       "capabilities: 3:",
		"{\"version\": 1,\n \"builtins\": \"count\"}":             "capabilities: 2:",
		"{\"version\": 1, \"builtins\": [],\n\"builtins\": null}": `capabilities: 2:1: duplicate field "builtins"`,
		`{"version": 1, "Builtins": []}`:                          `unknown field "Builtins"`,
		`{"Version": 1}`:                                          `unknown field "Version"`,
		`{"version": 1`:                                           "1:14: unexpected end of JSON input",
	}
	for input, want := range cases {
		_, err := Parse([]byte(input))
		s.Require().Error(err, input)
		s.Contains(err.Error(), want, input)
	}
}

func (s *CapabilitiesTestSuite) TestLoad() {
	c, err := Load("")
	s.Require().NoError(err)
	s.Nil(c)

	dir := s.T().TempDir()
	good := filepath.Join(dir, "caps.json")
	s.Require().NoError(os.WriteFile(good, []byte(`{"version": 1, "builtins": ["count"]}`), 0o600))
	c, err = Load(good)
	s.Require().NoError(err)
	s.True(c.AllowsBuiltin("count"))
	s.False(c.AllowsBuiltin("now"))

	bad := filepath.Join(dir, "bad.json")
	s.Require().NoError(os.WriteFile(bad, []byte(`{"version": 9}`), 0o600))
	_, err = Load(bad)
	s.Require().Error(err)
	s.Contains(err.Error(), bad)

	missing := filepath.Join(dir, "missing.json")
	_, err = Load(missing)
	s.Require().Error(err)
	s.True(strings.HasPrefix(err.Error(), missing+": capabilities: "), err.Error())
	_, err = Load(bad)
	s.True(strings.HasPrefix(err.Error(), bad+": capabilities: "), err.Error())
}
