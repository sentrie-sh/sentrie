// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/sentrie-sh/sentrie/pack"
)

const packDataRolesJSON = `{"admin": ["read", "write"], "viewer": ["read"]}`

// buildPackDataExecutor writes a pack whose manifest carries dataTable as its
// [data] body, loads and validates it, and returns NewExecutor's result.
func (s *RuntimeTestSuite) buildPackDataExecutor(dataTable string, files map[string]string) (Executor, error) {
	ctx := s.T().Context()
	dir := s.T().TempDir()
	manifest := testPackManifest + "\n[data]\n" + dataTable
	s.Require().NoError(os.WriteFile(filepath.Join(dir, loader.PackFileName), []byte(manifest), 0o644))
	for name, content := range files {
		path := filepath.Join(dir, name)
		s.Require().NoError(os.MkdirAll(filepath.Dir(path), 0o755))
		s.Require().NoError(os.WriteFile(path, []byte(content), 0o644))
	}
	packFile, err := loader.LoadPack(ctx, dir)
	s.Require().NoError(err)
	programs, err := loader.LoadPrograms(ctx, packFile)
	s.Require().NoError(err)
	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, packFile))
	for _, program := range programs {
		s.Require().NoError(idx.AddProgram(ctx, program))
	}
	s.Require().NoError(idx.Validate(ctx))
	return NewExecutor(idx)
}

func (s *RuntimeTestSuite) TestPackDataReadByPolicy() {
	ctx := s.T().Context()
	exec, err := s.buildPackDataExecutor(`roles = "data/roles.json"
`, map[string]string{
		"data/roles.json": packDataRolesJSON,
		"policy.sentrie": `namespace data/test

policy p {
  fact user: string
  data roles
  data roles as r2

  rule canWrite = "write" in roles[user]
  rule viewerPerms = count(r2["viewer"])
  rule sameValue = roles == r2

  export decision of canWrite
  export decision of viewerPerms
  export decision of sameValue
}
`,
	})
	s.Require().NoError(err)
	defer func() { _ = exec.Close() }()

	out, err := exec.ExecRule(ctx, "data/test", "p", "canWrite", map[string]any{"user": "admin"})
	s.Require().NoError(err)
	s.Equal(box.Bool(true).Any(), out.Decision.Value.Any())

	out, err = exec.ExecRule(ctx, "data/test", "p", "canWrite", map[string]any{"user": "viewer"})
	s.Require().NoError(err)
	s.Equal(box.Bool(false).Any(), out.Decision.Value.Any())

	out, err = exec.ExecRule(ctx, "data/test", "p", "viewerPerms", map[string]any{"user": "x"})
	s.Require().NoError(err)
	s.Equal(float64(1), s.decisionNumber(out))

	out, err = exec.ExecRule(ctx, "data/test", "p", "sameValue", map[string]any{"user": "x"})
	s.Require().NoError(err)
	s.Equal(box.Bool(true).Any(), out.Decision.Value.Any())
}

func (s *RuntimeTestSuite) TestPackDataTypedDeclarationValidatedAtExecutorBuild() {
	policy := func(typ string) string {
		return `namespace data/test

shape RoleTable dict[list[string]]

policy p {
  data roles: ` + typ + `
  rule r = count(roles) > 0
  export decision of r
}
`
	}

	s.Run("matching type builds", func() {
		exec, err := s.buildPackDataExecutor(`roles = "roles.json"`+"\n", map[string]string{
			"roles.json":     packDataRolesJSON,
			"policy.sentrie": policy("RoleTable"),
		})
		s.Require().NoError(err)
		_ = exec.Close()
	})

	s.Run("mismatching type fails the build", func() {
		_, err := s.buildPackDataExecutor(`roles = "roles.json"`+"\n", map[string]string{
			"roles.json":     packDataRolesJSON,
			"policy.sentrie": policy("list[string]"),
		})
		s.Require().Error(err)
		s.Contains(err.Error(), `pack data "roles" in policy data/test/p does not match its declared type`)
	})
}

// A TypeScript module receives a copy of pack data: mutating its argument does
// not change the value the policy, or any later evaluation, reads.
func (s *RuntimeTestSuite) TestPackDataImmutableFromModules() {
	ctx := s.T().Context()
	exec, err := s.buildPackDataExecutor(`roles = "roles.json"`+"\n", map[string]string{
		"roles.json": packDataRolesJSON,
		"mut.ts": `export function mutate(x: any): number {
  x.admin[0] = "changed";
  x.extra = ["y"];
  delete x.viewer;
  return 1;
}
`,
		"policy.sentrie": `namespace data/test

policy p {
  data roles
  use { mutate } from "./mut" as m

  rule after = {
    let touched = m.mutate(roles)
    yield touched == 1 and count(roles["admin"]) == 2 and roles["admin"][0] == "read" and roles["extra"] is not defined and roles["viewer"] is defined
  }
  rule admin = roles["admin"]

  export decision of after
  export decision of admin
}
`,
	})
	s.Require().NoError(err)
	defer func() { _ = exec.Close() }()

	for range 2 {
		out, err := exec.ExecRule(ctx, "data/test", "p", "after", map[string]any{})
		s.Require().NoError(err)
		s.Equal(box.Bool(true).Any(), out.Decision.Value.Any())
	}
	out, err := exec.ExecRule(ctx, "data/test", "p", "admin", map[string]any{})
	s.Require().NoError(err)
	s.Equal([]any{"read", "write"}, out.Decision.Value.Any())
}

// The runtime half of the derive purity pair: evalIdent refuses pack data
// inside a derive body even when the static check is bypassed.
func (s *RuntimeTestSuite) TestEvalIdentDeniesPackDataInsideDerive() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	d := ast.NewDataStatement("roles", nil, "roles", stubRange())
	p.Data = map[string]*ast.DataStatement{"roles": d}

	idx := index.CreateIndex()
	pk := newPackDataTestPackFile(map[string]box.Value{"roles": box.List([]box.Value{box.String("a")})})
	s.Require().NoError(idx.SetPack(ctx, pk))
	exec := &executorImpl{index: idx}

	s.Run("outside a derive the value resolves", func() {
		ec := NewExecutionContext(p, exec)
		v, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("roles", stubRange()))
		s.Require().NoError(err)
		s.Equal([]any{"a"}, v.Any())
	})

	s.Run("inside a derive it is refused", func() {
		ec := NewExecutionContext(p, exec)
		ec.evalDerive = &index.Derive{Name: "d"}
		_, _, err := evalIdent(ctx, ec, exec, p, ast.NewIdentifier("roles", stubRange()))
		s.Require().Error(err)
		s.Contains(err.Error(), `pack data is not available inside a derive ("roles")`)
	})

	s.Run("unloaded entry is an error, not undefined", func() {
		empty := index.CreateIndex()
		s.Require().NoError(empty.SetPack(ctx, newPackDataTestPackFile(nil)))
		ec := NewExecutionContext(p, &executorImpl{index: empty})
		_, _, err := evalIdent(ctx, ec, &executorImpl{index: empty}, p, ast.NewIdentifier("roles", stubRange()))
		s.Require().Error(err)
		s.Contains(err.Error(), `pack data "roles" is not loaded`)
	})
}

func newPackDataTestPackFile(values map[string]box.Value) *pack.PackFile {
	p := pack.NewPackFile("data_pack")
	p.DataValues = values
	return p
}
