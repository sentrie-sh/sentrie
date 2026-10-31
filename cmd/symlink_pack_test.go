// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"strings"
)

const symlinkPackPolicy = `namespace com/ex
policy pol {
  use { total } from "./helper" as helper
  rule ok = { yield helper.total() == 3 }
  export decision of ok
}
`

// symlinkedPack writes a pack under releases/1 and returns the path of a
// symlink to it (current -> releases/1), the way a deploy directory is laid
// out, along with the real directory.
func (s *CmdTestSuite) symlinkedPack(policy string) (link, real string) {
	real = s.writeTestPack(policy)
	link = filepath.Join(s.T().TempDir(), "current")
	s.Require().NoError(os.Symlink(real, link))
	return link, real
}

func (s *CmdTestSuite) writePackSource(root, rel, src string) {
	p := filepath.Join(root, rel)
	s.Require().NoError(os.MkdirAll(filepath.Dir(p), 0o755))
	s.Require().NoError(os.WriteFile(p, []byte(src), 0o600))
}

// helper.ts -> ./lib/mid -> ../leaf: each require after the first resolves
// from a module's own symlink-resolved directory.
func (s *CmdTestSuite) writeNestedTSChain(root string) {
	s.writePackSource(root, "helper.ts", "const m = require('./lib/mid');\nexport const total = () => m.n + 1;")
	s.writePackSource(root, "lib/mid.ts", "const l = require('../leaf');\nexport const n = l.n + 1;")
	s.writePackSource(root, "leaf.ts", "export const n = 1;")
}

func (s *CmdTestSuite) TestExecCmdRunsPackBehindSymlinkWithNestedRequires() {
	link, real := s.symlinkedPack(symlinkPackPolicy)
	s.writeNestedTSChain(real)

	for _, loc := range []string{link, real} {
		out := s.captureStdout(func() {
			s.Require().NoError(runExecCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", loc}))
		})
		s.Contains(out, "true", loc)
	}
	s.Require().NoError(runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", link}))
}

// A pack-local module that is a symlink out of the pack stays refused when
// the pack itself is reached through a symlink, whether the policy names it
// with its extension or relies on extension probing.
func (s *CmdTestSuite) TestExecCmdRefusesModuleLinkedOutsideSymlinkedPack() {
	outside := s.T().TempDir()
	s.writePackSource(outside, "secret.ts", "export const total = () => 3;")

	for _, spec := range []string{"./helper", "./helper.ts"} {
		policy := strings.Replace(symlinkPackPolicy, `"./helper"`, `"`+spec+`"`, 1)
		link, real := s.symlinkedPack(policy)
		s.Require().NoError(os.Symlink(filepath.Join(outside, "secret.ts"), filepath.Join(real, "helper.ts")))

		err := runExecCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", link})
		s.Require().Error(err, spec)
		s.Contains(err.Error(), "outside the packroot", spec)
	}
}
