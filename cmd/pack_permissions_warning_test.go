// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

const permissionsWarnPolicy = `namespace com/ex
policy pol {
  rule ok = { yield 1 }
  export decision of ok
}
`

// writePermissionsPack writes a pack whose manifest carries the given
// [permissions] lines (none when permissions is empty).
func (s *CmdTestSuite) writePermissionsPack(policy, permissions string) string {
	dir := s.writeTestPack(policy)
	manifest := testPackToml
	if permissions != "" {
		manifest += "\n[permissions]\n" + permissions + "\n"
	}
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "sentrie.pack.toml"), []byte(manifest), 0o600))
	return dir
}

// TestPackPermissionWarningsAtCLI covers #316 through the real commands: every
// command that loads a pack warns on stderr for a non-empty fs_read or net,
// and stays silent otherwise. serve is pointed at a policy that does not parse
// so it returns after loading instead of listening.
func (s *CmdTestSuite) TestPackPermissionWarningsAtCLI() {
	const fsWarn = `warning: pack permission "fs_read" is declared but not enforced yet; it gates nothing in this release`
	const netWarn = `warning: pack permission "net" is declared but not enforced yet; it gates nothing in this release`

	perms := []struct {
		name    string
		toml    string
		wantFS  bool
		wantNet bool
	}{
		{"fs_read", `fs_read = ["/data"]`, true, false},
		{"net", `net = ["https://example.com"]`, false, true},
		{"both", "fs_read = [\"/data\"]\nnet = [\"https://example.com\"]", true, true},
		{"empty lists", "fs_read = []\nnet = []", false, false},
		{"env only", `env = ["HOME"]`, false, false},
		{"none", "", false, false},
	}
	cmds := []struct {
		name   string
		policy string
		run    func(dir string) error
	}{
		{"validate", permissionsWarnPolicy, func(dir string) error {
			return runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
		}},
		{"exec", permissionsWarnPolicy, func(dir string) error {
			return runExecCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", dir})
		}},
		{"bench", permissionsWarnPolicy, func(dir string) error {
			return runBenchCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", dir, "--benchtime", "1x"})
		}},
		{"serve", "this is not a policy {{{", func(dir string) error {
			return runServeCLI(s.T().Context(), []string{"--pack-location", dir})
		}},
	}

	for _, c := range cmds {
		for _, p := range perms {
			s.Run(c.name+"/"+p.name, func() {
				dir := s.writePermissionsPack(c.policy, p.toml)
				var runErr error
				var out string
				// stdout is captured too so exec and bench output stays off the test log.
				s.captureStdout(func() {
					out = s.captureStderr(func() { runErr = c.run(dir) })
				})
				if c.name == "serve" {
					s.Require().Error(runErr, "serve must fail on the unparsable policy")
				} else {
					s.Require().NoError(runErr)
				}
				s.Equal(p.wantFS, containsLine(out, fsWarn), out)
				s.Equal(p.wantNet, containsLine(out, netWarn), out)
			})
		}
	}
}

func containsLine(out, line string) bool {
	return slices.Contains(strings.Split(out, "\n"), line)
}
