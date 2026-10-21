// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package capabilities

import (
	"os"
	"path/filepath"
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
		`{}`:                                       "unsupported version 0",
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

	_, err = Load(filepath.Join(dir, "missing.json"))
	s.Require().Error(err)
}

func (s *CapabilitiesTestSuite) TestDeterministicJS() {
	var c *Capabilities
	s.False(c.DeterministicJS())

	for src, want := range map[string]bool{
		`{"version": 1}`: false,
		`{"version": 1, "deterministic_js": false}`: false,
		`{"version": 1, "deterministic_js": null}`:  false,
		`{"version": 1, "deterministic_js": true}`:  true,
	} {
		c, err := Parse([]byte(src))
		s.Require().NoError(err, src)
		s.Equal(want, c.DeterministicJS(), src)
	}

	_, err := Parse([]byte(`{"version": 1, "deterministic_js": "yes"}`))
	s.Require().Error(err)
}
