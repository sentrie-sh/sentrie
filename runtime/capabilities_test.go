// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"errors"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/capabilities"
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

// With deterministic_js, Date.now inside module code reads the same pinned
// clock as the now() builtin, and Math.random is reproducible per execution
// and function.
func (s *RuntimeTestSuite) TestExecRuleDeterministicJavaScriptShims() {
	ctx := s.T().Context()
	src := `namespace com/ex
policy pol {
  use { now, random } from @sentrie/js as js
  rule ok = { yield now() == js.now() and js.random() == js.random() }
  export decision of ok
}`
	idx := s.mustBuildDeriveIndex(ctx, deriveTestProgram{name: "det.sentrie", src: src})
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx, WithCapabilities(s.mustGrant(`{"version": 1, "deterministic_js": true}`)))
	s.Require().NoError(err)
	out, err := exec.ExecRule(ctx, "com/ex", "pol", "ok", nil)
	s.Require().NoError(err)
	s.Equal(trinary.True, out.Decision.State)
}
