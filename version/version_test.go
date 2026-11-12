// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type VersionTestSuite struct {
	suite.Suite
}

func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}

// String must not panic on a version stamped with -ldflags that is not semver
// (#157): it prints the version as given.
func (s *VersionTestSuite) TestStringNonSemverVersion() {
	info := Info{Name: "Sentrie", GitVersion: "not-semver"}
	var out string
	s.Require().NotPanics(func() { out = info.String() })
	s.Contains(out, "Sentrie not-semver\n")
}

func (s *VersionTestSuite) TestStringSemverVersion() {
	for _, v := range []string{"0.0.5", "v0.0.5"} {
		out := Info{Name: "Sentrie", GitVersion: v}.String()
		s.Contains(out, "Sentrie v0.0.5\n", v)
	}
}

// An empty GitVersion prints only the name.
func (s *VersionTestSuite) TestStringEmptyVersion() {
	out := Info{Name: "Sentrie"}.String()
	s.Contains(out, "Sentrie\n")
	s.NotContains(out, "Sentrie v")
}
