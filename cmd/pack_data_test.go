// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// writeDataPackDir writes a pack whose manifest carries dataTable as its [data]
// body, plus the given files (relative path to content).
func (s *CmdTestSuite) writeDataPackDir(dataTable string, files map[string]string) string {
	dir := s.T().TempDir()
	manifest := testPackToml + "\n[data]\n" + dataTable
	s.Require().NoError(os.WriteFile(filepath.Join(dir, "sentrie.pack.toml"), []byte(manifest), 0o600))
	for name, content := range files {
		path := filepath.Join(dir, name)
		s.Require().NoError(os.MkdirAll(filepath.Dir(path), 0o755))
		s.Require().NoError(os.WriteFile(path, []byte(content), 0o600))
	}
	return dir
}

const cmdRolesJSON = `{"admin": ["read", "write"], "viewer": ["read"], "limits": {"max": {"rows": 10}}}`

// execDecision runs `sentrie exec <rule> --output json` and returns the decision
// value of the single output.
func (s *CmdTestSuite) execDecision(dir, rule, facts string) any {
	s.T().Helper()
	var runErr error
	out := s.captureStdout(func() {
		runErr = runExecCLI(s.T().Context(), []string{rule, "--pack-location", dir, "--output", "json", "--facts", facts})
	})
	s.Require().NoError(runErr)
	var parsed []struct {
		Decision struct {
			Value any `json:"Value"`
		} `json:"Decision"`
	}
	s.Require().NoError(json.Unmarshal([]byte(out), &parsed), out)
	s.Require().Len(parsed, 1, out)
	return parsed[0].Decision.Value
}

func (s *CmdTestSuite) TestExecCmdReadsPackData() {
	dir := s.writeDataPackDir("roles = \"data/roles.json\"\n", map[string]string{
		"data/roles.json": cmdRolesJSON,
		"policy.sentrie": `namespace com/ex

shape Roles dict[any]

policy typed {
  fact user: string
  data roles: Roles as table
  data roles as raw

  rule canWrite = "write" in table[user]
  rule maxRows = raw.limits.max.rows
  rule readerCount = count(raw["viewer"])
  export decision of canWrite
  export decision of maxRows
  export decision of readerCount
}
`,
	})

	s.Equal(true, s.execDecision(dir, "com/ex/typed/canWrite", `{"user":"admin"}`))
	s.Equal(false, s.execDecision(dir, "com/ex/typed/canWrite", `{"user":"viewer"}`))
	s.Equal(float64(10), s.execDecision(dir, "com/ex/typed/maxRows", `{"user":"x"}`))
	s.Equal(float64(1), s.execDecision(dir, "com/ex/typed/readerCount", `{"user":"x"}`))
}

func (s *CmdTestSuite) TestExecCmdPackDataListMembership() {
	dir := s.writeDataPackDir("allowlist = \"allow.json\"\n", map[string]string{
		"allow.json": `["10.0.0.1", "10.0.0.2"]`,
		"policy.sentrie": `namespace com/ex
policy ips {
  fact ip: string
  data allowlist
  rule allowed = ip in allowlist
  export decision of allowed
}
`,
	})
	s.Equal(true, s.execDecision(dir, "com/ex/ips/allowed", `{"ip":"10.0.0.2"}`))
	s.Equal(false, s.execDecision(dir, "com/ex/ips/allowed", `{"ip":"8.8.8.8"}`))
}

// `data` is a contextual keyword: it must still work as a fact name, a field
// name and a let name next to a data declaration in the same policy.
func (s *CmdTestSuite) TestExecCmdDataStaysUsableAsIdentifier() {
	dir := s.writeDataPackDir("cfg = \"cfg.json\"\n", map[string]string{
		"cfg.json": `{"limit": 5}`,
		"policy.sentrie": `namespace com/ex
policy ids {
  fact data: dict[any]
  data cfg
  rule fieldRead = data.data
  rule bothSources = data.n + cfg.limit
  export decision of fieldRead
  export decision of bothSources
}
`,
	})
	s.Equal("inner", s.execDecision(dir, "com/ex/ids/fieldRead", `{"data":{"data":"inner","n":2}}`))
	s.Equal(float64(7), s.execDecision(dir, "com/ex/ids/bothSources", `{"data":{"data":"inner","n":2}}`))
}

func (s *CmdTestSuite) TestValidateCmdPackData() {
	okPolicy := `namespace com/ex
policy pol {
  data roles
  rule ok = count(roles) > 0
  export decision of ok
}
`
	cases := []struct {
		name    string
		table   string
		files   map[string]string
		wantErr string // empty means validation succeeds
	}{
		{
			name:  "untyped data validates",
			table: "roles = \"roles.json\"\n",
			files: map[string]string{"roles.json": cmdRolesJSON, "policy.sentrie": okPolicy},
		},
		{
			name:  "typed data validates",
			table: "roles = \"roles.json\"\n",
			files: map[string]string{"roles.json": cmdRolesJSON, "policy.sentrie": `namespace com/ex
policy pol {
  data roles: dict[any]
  rule ok = count(roles) > 0
  export decision of ok
}
`},
		},
		{
			name:    "type mismatch fails before any request",
			table:   "roles = \"roles.json\"\n",
			files:   map[string]string{"roles.json": cmdRolesJSON, "policy.sentrie": "namespace com/ex\npolicy pol {\n  data roles: list[string]\n  rule ok = count(roles) > 0\n  export decision of ok\n}\n"},
			wantErr: `pack data "roles" in policy com/ex/pol does not match its declared type`,
		},
		{
			name:    "policy declares data the manifest does not",
			table:   "roles = \"roles.json\"\n",
			files:   map[string]string{"roles.json": cmdRolesJSON, "policy.sentrie": "namespace com/ex\npolicy pol {\n  data nothere\n  rule ok = count(nothere) > 0\n  export decision of ok\n}\n"},
			wantErr: "nothere",
		},
		{
			name:  "data inside a derive is refused",
			table: "roles = \"roles.json\"\n",
			files: map[string]string{"roles.json": cmdRolesJSON, "policy.sentrie": `namespace com/ex
policy pol {
  data roles
  let base = 1
  derive pick = (k: string) => { yield roles[k] }
  rule ok = count(pick("admin")) > 0
  export decision of ok
}
`},
			wantErr: "pack data is not available inside a derive",
		},
		{
			name:    "missing file",
			table:   "roles = \"nope.json\"\n",
			files:   map[string]string{"policy.sentrie": okPolicy},
			wantErr: `data "roles" (nope.json)`,
		},
		{
			name:    "absolute path",
			table:   "roles = \"/etc/hosts\"\n",
			files:   map[string]string{"policy.sentrie": okPolicy},
			wantErr: "relative to the pack root",
		},
		{
			name:    "parent escape",
			table:   "roles = \"../roles.json\"\n",
			files:   map[string]string{"policy.sentrie": okPolicy},
			wantErr: "escapes the pack root",
		},
		{
			name:    "directory instead of file",
			table:   "roles = \"data\"\n",
			files:   map[string]string{"data/x.json": "{}", "policy.sentrie": okPolicy},
			wantErr: "not a regular file",
		},
		{
			name:    "malformed JSON",
			table:   "roles = \"roles.json\"\n",
			files:   map[string]string{"roles.json": `{"a": [`, "policy.sentrie": okPolicy},
			wantErr: "invalid JSON",
		},
		{
			name:    "trailing content",
			table:   "roles = \"roles.json\"\n",
			files:   map[string]string{"roles.json": `{"a": 1} 2`, "policy.sentrie": okPolicy},
			wantErr: "unexpected content",
		},
	}
	for _, tc := range cases {
		s.Run(tc.name, func() {
			dir := s.writeDataPackDir(tc.table, tc.files)
			err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
			if tc.wantErr == "" {
				s.Require().NoError(err)
				return
			}
			s.Require().Error(err)
			s.Contains(err.Error(), tc.wantErr)
		})
	}
}

func (s *CmdTestSuite) TestValidateCmdPackDataSymlinkOutsideRootRejected() {
	outside := filepath.Join(s.T().TempDir(), "secret.json")
	s.Require().NoError(os.WriteFile(outside, []byte(`{"k": 1}`), 0o600))
	dir := s.writeDataPackDir("roles = \"link.json\"\n", map[string]string{
		"policy.sentrie": "namespace com/ex\npolicy pol {\n  data roles\n  rule ok = count(roles) > 0\n  export decision of ok\n}\n",
	})
	if err := os.Symlink(outside, filepath.Join(dir, "link.json")); err != nil {
		s.T().Skipf("symlinks unavailable: %v", err)
	}
	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), `data "roles" (link.json)`)
}

func (s *CmdTestSuite) TestValidateCmdPackDataOversizeRejected() {
	dir := s.writeDataPackDir("roles = \"big.json\"\n", map[string]string{
		"policy.sentrie": "namespace com/ex\npolicy pol {\n  data roles\n  rule ok = count(roles) > 0\n  export decision of ok\n}\n",
	})
	f, err := os.Create(filepath.Join(dir, "big.json"))
	s.Require().NoError(err)
	s.Require().NoError(f.Truncate(16<<20 + 1))
	s.Require().NoError(f.Close())
	err = runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), "larger than the")
}

// A manifest entry that no policy declares is not visible to it: using the
// bare name is an unresolved identifier, not an implicit read of the data.
func (s *CmdTestSuite) TestExecCmdUndeclaredPackDataNameIsNotVisible() {
	dir := s.writeDataPackDir("roles = \"roles.json\"\n", map[string]string{
		"roles.json":     cmdRolesJSON,
		"policy.sentrie": "namespace com/ex\npolicy pol {\n  rule ok = count(roles) > 0\n  export decision of ok\n}\n",
	})
	var runErr error
	s.captureStdout(func() {
		runErr = runExecCLI(s.T().Context(), []string{"com/ex/pol/ok", "--pack-location", dir, "--output", "json"})
	})
	s.Require().Error(runErr)
	s.Contains(runErr.Error(), "roles")
}

// Pack data reaches a derive only as an argument, and then evaluates normally.
func (s *CmdTestSuite) TestExecCmdPackDataPassedIntoDerive() {
	dir := s.writeDataPackDir("roles = \"roles.json\"\n", map[string]string{
		"roles.json": cmdRolesJSON,
		"policy.sentrie": `namespace com/ex
policy pol {
  data roles
  let z = 1
  derive has = (m: dict[any], k: string) => { yield m[k] is defined }
  rule ok = has(roles, "admin")
  export decision of ok
}
`,
	})
	s.Equal(true, s.execDecision(dir, "com/ex/pol/ok", "{}"))
}
