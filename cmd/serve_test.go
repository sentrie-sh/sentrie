// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import "context"

func runServeCLI(ctx context.Context, args []string) error {
	cli := Setup(ctx, "test")
	return Execute(ctx, cli, append([]string{"sentrie", "serve"}, args...))
}

func (s *CmdTestSuite) TestServeCmdHelpUsesHTTPScopedFlags() {
	out := s.captureStdout(func() {
		err := runServeCLI(context.Background(), []string{"--help"})
		s.Require().NoError(err)
	})

	s.Contains(out, "--http-port")
	s.Contains(out, "--http-listen")
	s.NotContains(out, "\n --port")
	s.NotContains(out, "\n --listen")
}

func (s *CmdTestSuite) TestServeCmdRejectsLegacyPortFlag() {
	err := runServeCLI(context.Background(), []string{"--port", "9999"})
	s.Require().Error(err)
	s.Contains(err.Error(), "flag '--port' is not defined for command 'serve'")
}

func (s *CmdTestSuite) TestServeCmdRejectsLegacyListenFlag() {
	err := runServeCLI(context.Background(), []string{"--listen", "0.0.0.0"})
	s.Require().Error(err)
	s.Contains(err.Error(), "flag '--listen' is not defined for command 'serve'")
}
