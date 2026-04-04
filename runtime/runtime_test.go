// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2025 Binaek Sarkar
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

package runtime

import (
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/suite"
)

// RuntimeTestSuite is the single suite for all runtime package tests.
type RuntimeTestSuite struct {
	suite.Suite
	ctx    context.Context
	ec     *ExecutionContext
	exec   *executorImpl
	policy *index.Policy
}

func (s *RuntimeTestSuite) SetupSuite() {
	s.ctx = context.Background()
	s.ec = &ExecutionContext{}
	s.exec = &executorImpl{}
	s.policy = &index.Policy{
		Namespace: &index.Namespace{
			FQN: ast.NewFQN([]string{"test", "namespace"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 1, Offset: 0}, To: tokens.Pos{Line: 1, Column: 1, Offset: 0}}),
		},
	}
}

func (s *RuntimeTestSuite) SetupTest() {
	s.ctx = context.Background()
}

func TestRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(RuntimeTestSuite))
}
