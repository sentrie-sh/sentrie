// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/suite"
)

// RuntimeTestSuite is the shared testify suite for runtime package tests.
//
// Feature-specific runtime tests should be added as methods on this suite, even
// when they live in separate *_test.go files. Suite methods run sequentially
// under suite.Run (one shared *RuntimeTestSuite), so do not call s.T().Parallel()
// from suite methods: testify updates SetT on the shared suite and parallel
// subtests race.
type RuntimeTestSuite struct {
	suite.Suite
	policy *index.Policy
}

func (s *RuntimeTestSuite) SetupSuite() {
	s.policy = &index.Policy{
		Namespace: &index.Namespace{
			FQN: ast.NewFQN([]string{"test", "namespace"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
	}
}

// builtinSite is used by builtin unit tests that need a CallSite frame.
func (s *RuntimeTestSuite) builtinSite() *CallSite {
	ec := NewExecutionContext(s.policy, &executorImpl{})
	return &CallSite{EC: ec, Exec: &executorImpl{}, Policy: s.policy}
}

func (s *RuntimeTestSuite) builtinArgs(parts ...any) []box.Value {
	out := make([]box.Value, len(parts))
	for i := range parts {
		out[i] = box.FromAny(parts[i])
	}
	return out
}

func TestRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(RuntimeTestSuite))
}
