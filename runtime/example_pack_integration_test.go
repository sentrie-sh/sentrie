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

package runtime

import (
	"context"
	"path/filepath"
	"runtime"

	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
)

func examplePackDir() string {
	_, current, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(current), "..", "example_pack")
}

func (s *RuntimeTestSuite) TestExamplePackExecPolicySmoke() {
	ctx := context.Background()

	packFile, err := loader.LoadPack(ctx, examplePackDir())
	s.Require().NoError(err)

	programs, err := loader.LoadPrograms(ctx, packFile)
	s.Require().NoError(err)
	s.Require().NotEmpty(programs)

	idx := index.CreateIndex()
	s.Require().NoError(idx.SetPack(ctx, packFile))
	for _, program := range programs {
		s.Require().NoError(idx.AddProgram(ctx, program))
	}
	s.Require().NoError(idx.Validate(ctx))

	exec, err := NewExecutor(idx)
	s.Require().NoError(err)

	testCases := []struct {
		namespace string
		policy    string
		facts     map[string]any
		expectErr string
	}{
		{
			namespace: "sh/sentrie/example",
			policy:    "user_access",
			facts: map[string]any{
				"user": map[string]any{"role": "admin", "status": "active"},
			},
		},
		{
			namespace: "user_management",
			policy:    "user_access",
			facts: map[string]any{
				"user": map[string]any{"role": "user", "status": "active"},
			},
		},
		{
			namespace: "sh/sentrie/example/shapes",
			policy:    "example",
			facts:     map[string]any{},
			expectErr: "invalid value for let declaration user",
		},
		{
			namespace: "sh/sentrie/example",
			policy:    "var_test",
			facts:     map[string]any{},
		},
		{
			namespace: "sh/sentrie/example",
			policy:    "jsglobalpolicy",
			facts:     map[string]any{},
			expectErr: "conflict: let declaration",
		},
		{
			namespace: "sh/sentrie/example/pipeline",
			policy:    "basics",
			facts:     map[string]any{},
		},
		{
			namespace: "sh/sentrie/example/pipeline/placeholder",
			policy:    "placeholder_pipeline",
			facts:     map[string]any{},
		},
		{
			namespace: "sh/sentrie/example/pipeline/module",
			policy:    "module_pipeline",
			facts:     map[string]any{},
		},
		{
			namespace: "sh/sentrie/example/pipeline/memoized",
			policy:    "memoized_pipeline",
			facts:     map[string]any{},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.namespace+"/"+tc.policy, func() {
			outputs, execErr := exec.ExecPolicy(ctx, tc.namespace, tc.policy, tc.facts)
			if tc.expectErr != "" {
				s.Require().Error(execErr)
				s.Contains(execErr.Error(), tc.expectErr)
				return
			}
			s.Require().NoError(execErr)
			s.Require().NotEmpty(outputs)
		})
	}
}
