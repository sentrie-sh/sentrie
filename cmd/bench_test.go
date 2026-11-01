// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// benchLineRE is the Go benchmark result line benchstat reads: a name starting
// with Benchmark, an iteration count, then value/unit pairs.
var benchLineRE = regexp.MustCompile(`^(Benchmark\S+)\s+(\d+)\s+([0-9.]+) ns/op\s+(\d+) B/op\s+(\d+) allocs/op$`)

const benchAdminFacts = `{"user":{"role":"admin","status":"active"}}`

func benchExamplePack() string {
	return filepath.Join("..", "example_pack")
}

func runBenchCLI(ctx context.Context, args []string) error {
	cli := Setup(ctx, "test")
	return Execute(ctx, cli, append([]string{"sentrie", "bench"}, args...))
}

// benchLines returns the benchmark result lines of out, ignoring the goos and
// goarch configuration lines.
func benchLines(out string) []string {
	var lines []string
	for _, l := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.HasPrefix(l, "Benchmark") {
			lines = append(lines, l)
		}
	}
	return lines
}

func (s *CmdTestSuite) runBenchOutput(args ...string) (string, error) {
	var runErr error
	out := s.captureStdout(func() {
		runErr = runBenchCLI(s.T().Context(), args)
	})
	return out, runErr
}

func (s *CmdTestSuite) TestBenchCmdFixedIterationsAndLineFormat() {
	out, err := s.runBenchOutput("user_management/user_access/allow_admin",
		"--pack-location", benchExamplePack(), "--facts", benchAdminFacts, "--benchtime", "7x")
	s.Require().NoError(err)
	s.Contains(out, "goos: "+runtime.GOOS)
	s.Contains(out, "goarch: "+runtime.GOARCH)

	lines := benchLines(out)
	s.Require().Len(lines, 1)
	m := benchLineRE.FindStringSubmatch(lines[0])
	s.Require().NotNil(m, "line is not Go benchmark format: %q", lines[0])
	s.Equal("7", m[2], "Nx must run exactly N iterations")
	s.True(strings.HasPrefix(m[1], "BenchmarkExec/user_management/user_access/allow_admin"))
	// every evaluation allocates, so the per-op figures must be non-zero
	s.NotEqual("0", m[4])
	s.NotEqual("0", m[5])
}

func (s *CmdTestSuite) TestBenchCmdCountPrintsOneLinePerRun() {
	out, err := s.runBenchOutput("user_management/user_access/allow_admin",
		"--pack-location", benchExamplePack(), "--facts", benchAdminFacts,
		"--benchtime", "3x", "--count", "3")
	s.Require().NoError(err)
	lines := benchLines(out)
	s.Require().Len(lines, 3)
	for _, l := range lines {
		m := benchLineRE.FindStringSubmatch(l)
		s.Require().NotNil(m, "line is not Go benchmark format: %q", l)
		s.Equal("3", m[2])
	}
}

func (s *CmdTestSuite) TestBenchCmdDurationBenchtime() {
	out, err := s.runBenchOutput("user_management/user_access/allow_admin",
		"--pack-location", benchExamplePack(), "--facts", benchAdminFacts, "--benchtime", "20ms")
	s.Require().NoError(err)
	lines := benchLines(out)
	s.Require().Len(lines, 1)
	m := benchLineRE.FindStringSubmatch(lines[0])
	s.Require().NotNil(m)
	n, convErr := strconv.Atoi(m[2])
	s.Require().NoError(convErr)
	s.Positive(n)
}

func (s *CmdTestSuite) TestBenchCmdWholePolicyTarget() {
	out, err := s.runBenchOutput("user_management/user_access",
		"--pack-location", benchExamplePack(), "--facts", benchAdminFacts, "--benchtime", "2x")
	s.Require().NoError(err)
	lines := benchLines(out)
	s.Require().Len(lines, 1)
	s.Contains(lines[0], "BenchmarkExec/user_management/user_access")
}

func (s *CmdTestSuite) TestBenchCmdFactFileAndFlagOverride() {
	factFile := filepath.Join(s.T().TempDir(), "facts.json")
	s.Require().NoError(os.WriteFile(factFile, []byte(`{"user":{"role":"guest","status":"inactive"}}`), 0o600))
	_, err := s.runBenchOutput("user_management/user_access/allow_admin",
		"--pack-location", benchExamplePack(), "--fact-file", factFile,
		"--facts", benchAdminFacts, "--benchtime", "1x")
	s.Require().NoError(err)

	facts, err := loadBenchFacts(factFile, `{"user":"flag"}`)
	s.Require().NoError(err)
	s.Equal("flag", facts["user"], "--facts overrides the fact file")
}

func (s *CmdTestSuite) TestBenchCmdErrors() {
	pack := benchExamplePack()
	tests := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"missing target", []string{"--pack-location", pack}, "required argument"},
		{"unknown target", []string{"no/such/target", "--pack-location", pack}, "not found"},
		{"missing pack", []string{"user_management/user_access", "--pack-location", filepath.Join(s.T().TempDir(), "absent")}, ""},
		{"missing required fact", []string{"user_management/user_access", "--pack-location", pack, "--benchtime", "1x"}, "failed to evaluate"},
		{"bad facts json", []string{"user_management/user_access", "--pack-location", pack, "--facts", "{nope"}, "invalid --facts"},
		{"missing fact file", []string{"user_management/user_access", "--pack-location", pack, "--fact-file", filepath.Join(s.T().TempDir(), "absent.json")}, ""},
		{"zero count", []string{"user_management/user_access", "--pack-location", pack, "--count", "0"}, "invalid --count"},
		{"negative count", []string{"user_management/user_access", "--pack-location", pack, "--count", "-2"}, "invalid --count"},
		{"non numeric count", []string{"user_management/user_access", "--pack-location", pack, "--count", "many"}, ""},
		{"zero x benchtime", []string{"user_management/user_access", "--pack-location", pack, "--benchtime", "0x"}, "invalid --benchtime"},
		{"bad benchtime", []string{"user_management/user_access", "--pack-location", pack, "--benchtime", "soon"}, "invalid --benchtime"},
		{"zero duration benchtime", []string{"user_management/user_access", "--pack-location", pack, "--benchtime", "0s"}, "invalid --benchtime"},
		{"bad eval timeout", []string{"user_management/user_access", "--pack-location", pack, "--eval-timeout", "x"}, "invalid --eval-timeout"},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			out, err := s.runBenchOutput(tc.args...)
			s.Require().Error(err)
			if tc.wantErr != "" {
				s.Contains(err.Error(), tc.wantErr)
			}
			s.Empty(benchLines(out), "a failing run must not print a result line")
		})
	}
}

func (s *CmdTestSuite) TestParseBenchtime() {
	spec, err := parseBenchtime("100x")
	s.Require().NoError(err)
	s.Equal(benchSpec{N: 100}, spec)

	spec, err = parseBenchtime("1500ms")
	s.Require().NoError(err)
	s.Equal(benchSpec{Duration: 1500 * time.Millisecond}, spec)

	for _, bad := range []string{"", "x", "-1x", "1.5x", "0x", "0", "-1s", "abc"} {
		_, err := parseBenchtime(bad)
		s.Require().Error(err, "benchtime %q", bad)
	}
}

func (s *CmdTestSuite) TestRunBenchFixedCountRunsExactlyN() {
	calls := 0
	res, err := runBench(benchSpec{N: 13}, func() error { calls++; return nil })
	s.Require().NoError(err)
	s.Equal(13, calls)
	s.Equal(13, res.N)
}

func (s *CmdTestSuite) TestRunBenchDurationRunsAtLeastTheDuration() {
	res, err := runBench(benchSpec{Duration: 15 * time.Millisecond}, func() error {
		time.Sleep(time.Millisecond)
		return nil
	})
	s.Require().NoError(err)
	s.GreaterOrEqual(res.Elapsed, 15*time.Millisecond)
	s.GreaterOrEqual(res.N, 2)
}

func (s *CmdTestSuite) TestRunBenchStopsOnError() {
	boom := errors.New("boom")
	calls := 0
	_, err := runBench(benchSpec{N: 10}, func() error {
		calls++
		if calls == 3 {
			return boom
		}
		return nil
	})
	s.Require().ErrorIs(err, boom)
	s.Equal(3, calls)
}

func (s *CmdTestSuite) TestBenchResultLineFormat() {
	line := benchResult{N: 4, Elapsed: 4 * time.Microsecond, Bytes: 400, Mallocs: 8}.line("BenchmarkExec/x")
	m := benchLineRE.FindStringSubmatch(line)
	s.Require().NotNil(m, "line: %q", line)
	s.Equal("4", m[2])
	s.Equal("1000.00", m[3])
	s.Equal("100", m[4])
	s.Equal("2", m[5])
}

func (s *CmdTestSuite) TestBenchNameHasNoWhitespace() {
	name := benchName("a b/c\td")
	s.NotRegexp(`\s`, name)
	s.True(strings.HasPrefix(name, "BenchmarkExec/a_b/c_d"))
}
