// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MainTestSuite struct {
	suite.Suite
}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

const nonSemverPackToml = `[schema]
version = 1

[pack]
name = "nonsemver"
version = "0.0.1"

[engine]
sentrie = ">=0.0.5"
`

const nonSemverPolicy = `namespace com/ex
policy pol {
  rule ok = { yield count([1]) > 0 }
  export decision of ok
}
`

// A binary stamped with a version that is not semver starts, warns that the
// pack's [engine] constraint is not checked, and loads the pack (#157). This
// builds the real binary so main's startup path (versionOnce().String()) runs.
func (s *MainTestSuite) TestNonSemverBuildVersionWarnsAndLoads() {
	if testing.Short() {
		s.T().Skip("builds the sentrie binary")
	}
	t := s.T()
	bin := filepath.Join(t.TempDir(), "sentrie")
	build := exec.CommandContext(t.Context(), "go", "build", "-ldflags", "-X main.version=not-semver", "-o", bin, ".")
	build.Env = append(os.Environ(), "GOWORK=off")
	out, err := build.CombinedOutput()
	s.Require().NoError(err, string(out))

	dir := t.TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "sentrie.pack.toml"), []byte(nonSemverPackToml), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "pol.sentrie"), []byte(nonSemverPolicy), 0o600))

	for _, args := range [][]string{
		{"validate", "--pack-location", dir},
		{"exec", "com/ex/pol/ok", "--pack-location", dir},
	} {
		var stdout, stderr bytes.Buffer
		run := exec.CommandContext(t.Context(), bin, args...)
		run.Stdout, run.Stderr = &stdout, &stderr
		err := run.Run()
		s.Require().NoError(err, "%v\nstdout: %s\nstderr: %s", args, stdout.String(), stderr.String())
		s.NotContains(stderr.String(), "panic", args)
		s.Contains(stderr.String(), `warning: pack requires sentrie >=0.0.5; the binary version "not-semver" is not a semantic version, so the engine version is not checked`, args)
	}

	var stdout bytes.Buffer
	run := exec.CommandContext(t.Context(), bin, "--version")
	run.Stdout = &stdout
	s.Require().NoError(run.Run())
	s.Contains(stdout.String(), "Sentrie not-semver\n")
}
