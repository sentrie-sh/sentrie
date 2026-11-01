// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// benchNameRE is the name field of a result line: one token with the optional
// -GOMAXPROCS suffix that go test adds.
var benchNameRE = regexp.MustCompile(`^BenchmarkExec/\S+?(-\d+)?$`)

func (s *CmdTestSuite) TestBenchCmdLineFieldsAreTabSeparatedAndOrdered() {
	out, err := s.runBenchOutput("user_management/user_access/allow_admin",
		"--pack-location", benchExamplePack(), "--facts", benchAdminFacts, "--benchtime", "5x")
	s.Require().NoError(err)
	lines := benchLines(out)
	s.Require().Len(lines, 1)

	fields := strings.Split(lines[0], "\t")
	s.Require().Len(fields, 5, "name, N, ns/op, B/op, allocs/op: %q", lines[0])
	s.Regexp(benchNameRE, fields[0])
	s.False(strings.ContainsAny(fields[0], " \t"), "the name is one field")
	n, err := strconv.Atoi(strings.TrimSpace(fields[1]))
	s.Require().NoError(err)
	s.Equal(5, n)
	for i, unit := range []string{" ns/op", " B/op", " allocs/op"} {
		f := strings.TrimSpace(fields[2+i])
		s.Require().True(strings.HasSuffix(f, unit), "field %q must end with %q", f, unit)
		_, err := strconv.ParseFloat(strings.TrimSuffix(f, unit), 64)
		s.Require().NoError(err, "field %q value", f)
	}
}

func (s *CmdTestSuite) TestBenchCmdFactFileAloneSuppliesFacts() {
	factFile := filepath.Join(s.T().TempDir(), "facts.json")
	s.Require().NoError(os.WriteFile(factFile, []byte(benchAdminFacts), 0o600))
	out, err := s.runBenchOutput("user_management/user_access/allow_admin",
		"--pack-location", benchExamplePack(), "--fact-file", factFile, "--benchtime", "2x")
	s.Require().NoError(err)
	s.Len(benchLines(out), 1)
}

func (s *CmdTestSuite) TestBenchCmdFailuresPrintNothingToStdout() {
	pack := benchExamplePack()
	badFactFile := filepath.Join(s.T().TempDir(), "bad.json")
	s.Require().NoError(os.WriteFile(badFactFile, []byte("{nope"), 0o600))

	tests := []struct {
		name string
		args []string
	}{
		{"unknown rule in known policy", []string{"user_management/user_access/no_such_rule", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime", "1x"}},
		{"unknown policy", []string{"user_management/no_such_policy", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime", "1x"}},
		{"unknown namespace", []string{"no_such_ns/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime", "1x"}},
		{"required fact missing", []string{"user_management/user_access/allow_admin", "--pack-location", pack, "--benchtime", "1x"}},
		{"facts json array", []string{"user_management/user_access", "--pack-location", pack, "--facts", "[1,2]"}},
		{"facts json scalar", []string{"user_management/user_access", "--pack-location", pack, "--facts", "3"}},
		{"fact file not json", []string{"user_management/user_access", "--pack-location", pack, "--fact-file", badFactFile}},
		{"benchtime overflow", []string{"user_management/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime", "99999999999999999999x"}},
		{"benchtime scientific count", []string{"user_management/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime", "1e3x"}},
		{"benchtime only x twice", []string{"user_management/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime", "xx"}},
		{"benchtime bare number", []string{"user_management/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime", "5"}},
		{"benchtime negative duration", []string{"user_management/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime=-1s"}},
		{"benchtime negative count", []string{"user_management/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--benchtime=-3x"}},
		{"count negative equals form", []string{"user_management/user_access", "--pack-location", pack, "--facts", benchAdminFacts, "--count=-1"}},
	}
	for _, tc := range tests {
		s.Run(tc.name, func() {
			out, err := s.runBenchOutput(tc.args...)
			s.Require().Error(err)
			s.Empty(strings.TrimSpace(out), "failure must print nothing, not even configuration lines")
		})
	}
}

func (s *CmdTestSuite) TestParseBenchtimeEdges() {
	for _, bad := range []string{" 1s", "1s ", "1h30", "1 x", "+0x", "0000x", "9223372036854775808x", "NaNs", "-0s"} {
		_, err := parseBenchtime(bad)
		s.Require().Error(err, "benchtime %q", bad)
	}
	spec, err := parseBenchtime("+5x")
	s.Require().NoError(err, "a leading plus is an integer to Atoi, as go test reads it")
	s.Equal(benchSpec{N: 5}, spec)
	spec, err = parseBenchtime("1ns")
	s.Require().NoError(err)
	s.Equal(benchSpec{Duration: time.Nanosecond}, spec)
}

func (s *CmdTestSuite) TestRunBenchDurationModeStopsOnLaterError() {
	boom := errors.New("boom")
	calls := 0
	_, err := runBench(benchSpec{Duration: time.Hour}, func() error {
		calls++
		if calls == 4 {
			return boom
		}
		return nil
	})
	s.Require().ErrorIs(err, boom)
	s.Equal(4, calls, "the run stops at the failing call, it does not keep growing")
}

func (s *CmdTestSuite) TestRunBenchMeasuresAllocationsPerOp() {
	var sink [][]byte
	res, err := runBench(benchSpec{N: 50}, func() error {
		sink = append(sink, make([]byte, 1024))
		return nil
	})
	s.Require().NoError(err)
	s.NotEmpty(sink)
	s.GreaterOrEqual(res.Bytes, uint64(50*1024), "each op allocates 1 KiB")
	s.GreaterOrEqual(res.Mallocs, uint64(50))
}
