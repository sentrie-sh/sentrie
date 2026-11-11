// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"bytes"
	"os"
	"path/filepath"

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

const engineTestPolicy = `namespace com/ex
policy pol {
  rule ok = { yield count([1]) > 0 }
  export decision of ok
}
`

// writeEnginePack writes a pack whose [engine] table requires constraint.
func (s *CmdTestSuite) writeEnginePack(constraint, policy string) string {
	dir := s.writeTestPack(policy)
	manifest := testPackToml + "\n[engine]\nsentrie = \"" + constraint + "\"\n"
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "sentrie.pack.toml"), []byte(manifest), 0o600))
	return dir
}

const engineSkipSuffix = "(use --skip-engine-check to load it anyway)"

func (s *CmdTestSuite) TestEngineCheckWiredIntoExecServeValidate() {
	s.T().Cleanup(func() { SetEngineVersion("") })
	ctx := s.T().Context()
	run := map[string]func([]string) error{
		"exec": func(a []string) error {
			return runExecCLI(ctx, append([]string{"com/ex/pol/ok"}, a...))
		},
		"validate": func(a []string) error {
			return runValidateCLI(ctx, append([]string{"com/ex/pol"}, a...))
		},
	}
	for name, fn := range run {
		s.Run(name+" satisfied", func() {
			SetEngineVersion("0.0.5")
			dir := s.writeEnginePack(">=0.0.5", engineTestPolicy)
			var err error
			out := s.captureStderr(func() { err = fn([]string{"--pack-location", dir}) })
			s.Require().NoError(err)
			s.NotContains(out, "engine version is not checked")
		})
		s.Run(name+" unsatisfied", func() {
			SetEngineVersion("0.0.4")
			dir := s.writeEnginePack(">=0.0.5", engineTestPolicy)
			err := fn([]string{"--pack-location", dir})
			s.Require().Error(err)
			s.Contains(err.Error(), "pack requires sentrie >=0.0.5, running 0.0.4")
			s.Contains(err.Error(), engineSkipSuffix)
		})
		s.Run(name+" strict prerelease", func() {
			SetEngineVersion("0.0.5-rc1")
			dir := s.writeEnginePack(">=0.0.5", engineTestPolicy)
			err := fn([]string{"--pack-location", dir})
			s.Require().ErrorContains(err, "running 0.0.5-rc1")
		})
		s.Run(name+" skip flag", func() {
			SetEngineVersion("0.0.4")
			dir := s.writeEnginePack(">=0.0.5", engineTestPolicy)
			s.Require().NoError(fn([]string{"--pack-location", dir, "--skip-engine-check"}))
		})
		s.Run(name+" dev build warns and passes", func() {
			SetEngineVersion("")
			dir := s.writeEnginePack(">=99.0.0", engineTestPolicy)
			var err error
			out := s.captureStderr(func() { err = fn([]string{"--pack-location", dir}) })
			s.Require().NoError(err)
			s.Contains(out, "warning: pack requires sentrie >=99.0.0")
		})
		s.Run(name+" pseudo-version warns and passes", func() {
			SetEngineVersion("v0.0.5-0.20261110181122-5f9f2a7446cb+dirty")
			dir := s.writeEnginePack(">=0.0.5", engineTestPolicy)
			var err error
			out := s.captureStderr(func() { err = fn([]string{"--pack-location", dir}) })
			s.Require().NoError(err)
			s.Contains(out, "built from source")
		})
		s.Run(name+" unparseable version warns and passes", func() {
			SetEngineVersion("not-a-version")
			dir := s.writeEnginePack(">=1.0.0", engineTestPolicy)
			var err error
			out := s.captureStderr(func() { err = fn([]string{"--pack-location", dir}) })
			s.Require().NoError(err)
			s.Contains(out, "not a semantic version")
		})
	}

	// serve blocks on success, so its check is observed through a policy that
	// fails to load: the engine error comes first, and --skip-engine-check lets
	// the load error through instead.
	s.Run("serve unsatisfied", func() {
		SetEngineVersion("0.0.4")
		dir := s.writeEnginePack(">=0.0.5", "not a policy")
		err := runServeCLI(ctx, []string{"--pack-location", dir})
		s.Require().Error(err)
		s.Contains(err.Error(), "pack requires sentrie >=0.0.5, running 0.0.4")
		s.Contains(err.Error(), engineSkipSuffix)
	})
	s.Run("serve skip flag", func() {
		SetEngineVersion("0.0.4")
		dir := s.writeEnginePack(">=0.0.5", "not a policy")
		err := runServeCLI(ctx, []string{"--pack-location", dir, "--skip-engine-check"})
		s.Require().Error(err)
		s.NotContains(err.Error(), "pack requires sentrie")
		s.NotContains(err.Error(), "skip-engine-check", "the flag must be known to serve")
	})
	s.Run("serve satisfied reaches the load step", func() {
		SetEngineVersion("0.0.5")
		dir := s.writeEnginePack(">=0.0.5", "not a policy")
		err := runServeCLI(ctx, []string{"--pack-location", dir})
		s.Require().Error(err)
		s.NotContains(err.Error(), "pack requires sentrie")
	})
}

func (s *CmdTestSuite) TestBenchDoesNotCheckEngine() {
	s.T().Cleanup(func() { SetEngineVersion("") })
	SetEngineVersion("0.0.4")
	dir := s.writeEnginePack(">=0.0.5", engineTestPolicy)
	err := runBenchCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", dir, "--benchtime", "1x"})
	s.Require().NoError(err)
}
