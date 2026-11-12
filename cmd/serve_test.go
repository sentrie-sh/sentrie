// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func runServeCLI(ctx context.Context, args []string) error {
	cli := Setup(ctx, "test")
	return Execute(ctx, cli, append([]string{"sentrie", "serve"}, args...))
}

func (s *CmdTestSuite) TestServeCmdHelpUsesHTTPScopedFlags() {
	out := s.captureStdout(func() {
		err := runServeCLI(s.T().Context(), []string{"--help"})
		s.Require().NoError(err)
	})

	s.Contains(out, "--http-port")
	s.Contains(out, "--http-listen")
	s.NotContains(out, "\n --port")
	s.NotContains(out, "\n --listen")
}

func (s *CmdTestSuite) TestServeCmdRejectsLegacyPortFlag() {
	err := runServeCLI(s.T().Context(), []string{"--port", "9999"})
	s.Require().Error(err)
	s.Contains(err.Error(), "flag '--port' is not defined for command 'serve'")
}

func (s *CmdTestSuite) TestServeCmdRejectsLegacyListenFlag() {
	err := runServeCLI(s.T().Context(), []string{"--listen", "0.0.0.0"})
	s.Require().Error(err)
	s.Contains(err.Error(), "flag '--listen' is not defined for command 'serve'")
}

func (s *CmdTestSuite) TestServeCmdHelpListsMaxBodyBytes() {
	out := s.captureStdout(func() {
		s.Require().NoError(runServeCLI(s.T().Context(), []string{"--help"}))
	})
	s.Contains(out, "--http-max-body-bytes")
	s.Contains(out, "1048576", "the default is 1 MiB")
}

func (s *CmdTestSuite) TestValidateMaxBodyBytes() {
	for _, n := range []int{0, 1, 1 << 20} {
		s.Require().NoError(validateMaxBodyBytes(n), "%d", n)
	}
	err := validateMaxBodyBytes(-1)
	s.Require().ErrorContains(err, "invalid --http-max-body-bytes -1")
	s.Require().ErrorContains(err, "must not be negative")
}

// A negative limit is refused before the pack is loaded: the pack location
// does not exist, yet the error names the flag, not the pack.
func (s *CmdTestSuite) TestServeCmdRejectsNegativeMaxBodyBytesBeforePackLoad() {
	err := runServeCLI(s.T().Context(), []string{
		"--pack-location", s.T().TempDir() + "/no-such-pack",
		"--http-max-body-bytes=-5",
	})
	s.Require().ErrorContains(err, "invalid --http-max-body-bytes -5")
}

// A listen address the process cannot bind is reported with the flag name.
func (s *CmdTestSuite) TestServeCmdWrapsUnusableListenAddress() {
	packDir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(packDir, "sentrie.pack.toml"), []byte(signalTestPack), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(packDir, "signal.sentrie"), []byte(signalTestPolicy), 0o600))

	// Hold a free localhost port so the bind in serve fails.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)
	defer func() { _ = ln.Close() }()
	port := strconv.Itoa(ln.Addr().(*net.TCPAddr).Port)

	err = runServeCLI(s.T().Context(), []string{
		"--pack-location", packDir,
		"--http-port", port,
		"--http-listen", "127.0.0.1",
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid or unusable --http-listen/--http-port: ")
	s.Contains(err.Error(), "failed to listen on 127.0.0.1:"+port)
}

// An empty --http-listen value must be rejected, not bind every interface.
func (s *CmdTestSuite) TestServeCmdRejectsEmptyListenAddress() {
	packDir := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(packDir, "sentrie.pack.toml"), []byte(signalTestPack), 0o600))
	s.Require().NoError(os.WriteFile(filepath.Join(packDir, "signal.sentrie"), []byte(signalTestPolicy), 0o600))

	// A regression would bind and block in serve; the deadline turns that into a failure.
	ctx, cancel := context.WithTimeout(s.T().Context(), 5*time.Second)
	defer cancel()
	err := runServeCLI(ctx, []string{
		"--pack-location", packDir,
		"--http-port", "0",
		"--http-listen=",
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "invalid or unusable --http-listen")
	s.Contains(err.Error(), "listen address must not be empty")
}
