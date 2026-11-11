// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"strconv"
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
	s.Contains(err.Error(), "invalid or unusable --http-listen: ")
	s.Contains(err.Error(), "failed to listen on 127.0.0.1:"+port)
}
