// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

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
