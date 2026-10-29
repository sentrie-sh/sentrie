// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/capabilities"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *RuntimeTestSuite) mustGrant(src string) *capabilities.Capabilities {
	c, err := capabilities.Parse([]byte(src))
	s.Require().NoError(err)
	return c
}

// The builtin gate runs at dispatch even when the static check was skipped.
func (s *RuntimeTestSuite) TestGetTargetRefusesUngrantedBuiltin() {
	ctx := s.T().Context()
	p := newEvalTestPolicy()
	exec := &executorImpl{capabilities: s.mustGrant(`{"version": 1, "builtins": ["count"]}`)}
	ec := NewExecutionContext(p, exec)

	call := ast.NewCallExpression(ast.NewIdentifier("now", stubRange()), nil, false, nil, stubRange())
	_, err := getTarget(ctx, ec, exec, p, call)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))
	s.Contains(err.Error(), `builtin "now"`)

	call = ast.NewCallExpression(ast.NewIdentifier("count", stubRange()), nil, false, nil, stubRange())
	target, err := getTarget(ctx, ec, exec, p, call)
	s.Require().NoError(err)
	s.NotNil(target)

	var nilExec *executorImpl
	s.Nil(nilExec.grant())
}

func (s *RuntimeTestSuite) TestExecRuleRefusesUngrantedModule() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy pol {
  use { v4 } from @sentrie/uuid as uuid
  rule ok = { yield uuid.v4() != "" }
  export decision of ok
}`
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "caps.sentrie", src: src})
	s.Require().NoError(idx.Validate(ctx))

	denied, err := NewExecutor(idx, WithCapabilities(s.mustGrant(`{"version": 1, "modules": ["@sentrie/hash"]}`)))
	s.Require().NoError(err)
	_, err = denied.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))

	granted, err := NewExecutor(idx, WithCapabilities(s.mustGrant(`{"version": 1, "modules": ["@sentrie/uuid"]}`)))
	s.Require().NoError(err)
	out, err := granted.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}

// A granted local module that requires an ungranted builtin module fails at
// load time through CheckModuleGrants, and, if that check is skipped, when the
// require runs, with ErrNotGranted surviving goja's error wrapping.
func (s *RuntimeTestSuite) TestLocalModuleRequireOfUngrantedModule() {
	ctx := s.T().Context()
	packRoot := s.T().TempDir()
	s.Require().NoError(os.WriteFile(filepath.Join(packRoot, "helper.ts"),
		[]byte("import { v4 } from \"@sentrie/uuid\";\nexport const id = () => v4();"), 0o644))
	src := `namespace com/ex
policy pol {
  use { id } from "./helper.ts" as helper
  rule ok = { yield helper.id() != "" }
  export decision of ok
}`
	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, &pack.PackFile{Location: packRoot}))
	prog, err := parser.NewParserFromString(src, filepath.Join(packRoot, "policy.sentrie")).ParseProgram(ctx)
	s.Require().NoError(err)
	s.Require().NoError(idx.AddProgram(ctx, prog))
	s.Require().NoError(idx.Validate(ctx))

	grant := s.mustGrant(`{"version": 1, "modules": [], "local_modules": true}`)
	s.Require().NoError(idx.CheckCapabilities(ctx, grant), "the index cannot see require() in TypeScript")

	exec, err := NewExecutor(idx, WithCapabilities(grant))
	s.Require().NoError(err)
	err = CheckModuleGrants(exec)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted))
	s.Contains(err.Error(), `module "@sentrie/uuid"`)

	_, err = exec.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().Error(err)
	s.True(errors.Is(err, xerr.ErrNotGranted), "%T: %v", err, err)

	open, err := NewExecutor(idx)
	s.Require().NoError(err)
	s.NoError(CheckModuleGrants(open))
}

// With deterministic_js, Date.now inside module code reads the same pinned
// execution clock as @sentrie/time.now (milliseconds against seconds, so the
// check does not depend on the now() builtin's unit), and Math.random is
// reproducible per execution and function. @sentrie/js has no now since #207,
// so the clock is read from pack TypeScript.
func (s *RuntimeTestSuite) TestExecRuleDeterministicJavaScriptShims() {
	exec := s.newTempPackExecutor(map[string]string{
		"clock.ts": `export function nowSeconds(): number { return Math.floor(Date.now() / 1000); }`,
		"policy.sentrie": `namespace com/ex
policy pol {
  use { random } from @sentrie/js as js
  use { now } from @sentrie/time as clock
  use { nowSeconds } from "./clock" as local
  rule ok = { yield local.nowSeconds() == clock.now() and js.random() == js.random() }
  export decision of ok
}
`,
	}, WithCapabilities(s.mustGrant(`{"version": 1, "deterministic_js": true}`)))
	out, err := exec.ExecRule(s.T().Context(), "com/ex", "pol", "ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}
