// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"

	"github.com/Masterminds/semver/v3"
	"github.com/sentrie-sh/sentrie/pack"
)

func enginePack(constraint string) *pack.PackFile {
	c, err := semver.NewConstraint(constraint)
	if err != nil {
		panic(err)
	}
	return &pack.PackFile{Engine: &pack.Engine{Sentrie: c}}
}

func (s *CmdTestSuite) TestCheckEngine() {
	s.T().Cleanup(func() { SetEngineVersion("") })
	cases := []struct {
		name, running, constraint string
		skip                      bool
		wantErr, wantWarn         string
	}{
		{name: "satisfied", running: "v0.0.5", constraint: ">=0.0.5"},
		{name: "mismatch", running: "0.0.4", constraint: ">=0.0.5", wantErr: "pack requires sentrie >=0.0.5, running 0.0.4"},
		{name: "strict prerelease", running: "0.0.5-rc1", constraint: ">=0.0.5", wantErr: "running 0.0.5-rc1"},
		{name: "skip", running: "0.0.4", constraint: ">=0.0.5", skip: true},
		{name: "dev build warns", running: "", constraint: ">=99.0.0", wantWarn: "dev or unversioned"},
		{name: "pseudo-version warns", running: "v0.0.5-0.20261110181122-5f9f2a7446cb+dirty", constraint: ">=0.0.5", wantWarn: "built from source"},
		{name: "unparseable warns", running: "garbage", constraint: ">=1.0.0", wantWarn: "not a semantic version"},
	}
	for _, c := range cases {
		s.Run(c.name, func() {
			SetEngineVersion(c.running)
			var w bytes.Buffer
			err := checkEngine(enginePack(c.constraint), c.skip, &w)
			if c.wantErr != "" {
				s.Require().ErrorContains(err, c.wantErr)
			} else {
				s.Require().NoError(err)
			}
			if c.wantWarn != "" {
				s.Contains(w.String(), c.wantWarn)
			} else {
				s.Empty(w.String())
			}
		})
	}
	s.Require().NoError(checkEngine(&pack.PackFile{}, false, &bytes.Buffer{}), "no [engine] table")
}
