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

	"github.com/binaek/perch"
	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/runtime/js"
	"github.com/sentrie-sh/sentrie/trinary"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (s *RuntimeTestSuite) TestEvaluateRuleOutcomeWhenFalseDefaultBranches() {
	p := newEvalTestPolicy()
	ruleStmt := ast.NewRuleStatement("r", nil, ast.NewTrinaryLiteral(trinary.False, stubRange()), ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	rule := &index.Rule{
		Node:    ruleStmt,
		Policy:  p,
		Name:    "r",
		FQN:     ast.CreateFQN(p.FQN, "r"),
		When:    ruleStmt.When,
		Body:    ruleStmt.Body,
		Default: nil,
	}
	ec := NewExecutionContext(p, &executorImpl{})
	decision, _, err := evaluateRuleOutcome(context.Background(), ec, &executorImpl{}, p, rule)
	s.Require().NoError(err)
	s.Require().Equal(trinary.Unknown, decision.State)

	rule.Default = ast.NewTrinaryLiteral(trinary.True, stubRange())
	decision, _, err = evaluateRuleOutcome(context.Background(), ec, &executorImpl{}, p, rule)
	s.Require().NoError(err)
	s.Require().Equal(trinary.True, decision.State)
}

func newExecutorAndPolicyWithFact(fact *ast.FactStatement) (*executorImpl, *index.Policy) {
	idx := index.CreateIndex()
	nsFQN := ast.NewFQN([]string{"test", "ns"}, stubRange())
	ns := &index.Namespace{
		FQN:          nsFQN,
		Policies:     map[string]*index.Policy{},
		Shapes:       map[string]*index.Shape{},
		ShapeExports: map[string]*index.ExportedShape{},
		Children:     []*index.Namespace{},
	}
	idx.Namespaces[nsFQN.String()] = ns

	p := &index.Policy{
		Namespace:   ns,
		Name:        "pol",
		FQN:         ast.CreateFQN(nsFQN, "pol"),
		Facts:       map[string]*ast.FactStatement{fact.Alias: fact},
		Rules:       map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{},
		Lets:        map[string]*ast.VarDeclaration{},
		Uses:        map[string]*ast.UseStatement{},
		Shapes:      map[string]*index.Shape{},
	}
	ruleStmt := ast.NewRuleStatement("allow", nil, nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	rule := &index.Rule{
		Node:   ruleStmt,
		Policy: p,
		Name:   "allow",
		FQN:    ast.CreateFQN(p.FQN, "allow"),
		Body:   ruleStmt.Body,
	}
	p.Rules["allow"] = rule
	p.RuleExports["allow"] = &index.ExportedRule{RuleName: "allow"}
	ns.Policies["pol"] = p

	return &executorImpl{index: idx}, p
}

func (s *RuntimeTestSuite) TestExecRuleFactNullBranchesWrapInvalidInvocation() {
	fact := ast.NewFactStatement("user", ast.NewStringTypeRef(stubRange()), "user", nil, false, stubRange())
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{"user": nil})
	s.Require().Error(err)
	s.Contains(err.Error(), "fact 'user' cannot be null")
	s.ErrorIs(err, xerr.InvalidInvocationError{})
}

func (s *RuntimeTestSuite) TestExecRuleNullableFactAcceptsNull() {
	fact := ast.NewFactStatement(
		"user",
		ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
		"user",
		nil,
		false,
		stubRange(),
	)
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{"user": nil})
	s.Require().NoError(err)
}

func (s *RuntimeTestSuite) TestExecRuleDefaultFactEvalErrorWrapsUnresolvableFact() {
	fact := ast.NewFactStatement("user", ast.NewStringTypeRef(stubRange()), "user", ast.NewIdentifier("missing", stubRange()), true, stubRange())
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), "unresolvable fact: user")
	s.ErrorIs(err, xerr.InvalidInvocationError{})
}

func (s *RuntimeTestSuite) TestExecRuleDefaultFactNullWrapsInvalidInvocation() {
	fact := ast.NewFactStatement("user", ast.NewStringTypeRef(stubRange()), "user", ast.NewNullLiteral(stubRange()), true, stubRange())
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), "fact 'user' cannot have null default value")
	s.ErrorIs(err, xerr.InvalidInvocationError{})
}

func (s *RuntimeTestSuite) TestExecRuleNullableFactDefaultAcceptsNull() {
	fact := ast.NewFactStatement(
		"user",
		ast.NewNullableTypeRef(ast.NewStringTypeRef(stubRange()), stubRange()),
		"user",
		ast.NewNullLiteral(stubRange()),
		true,
		stubRange(),
	)
	exec, _ := newExecutorAndPolicyWithFact(fact)
	_, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{})
	s.Require().NoError(err)
}

func (s *RuntimeTestSuite) TestExecRuleValidationErrorReturnsUnknownDecision() {
	fact := ast.NewFactStatement("age", ast.NewNumberTypeRef(stubRange()), "age", nil, false, stubRange())
	exec, _ := newExecutorAndPolicyWithFact(fact)
	out, err := exec.ExecRule(context.Background(), "test/ns", "pol", "allow", map[string]any{"age": "bad"})
	s.Require().Error(err)
	s.Require().NotNil(out)
	s.Require().NotNil(out.Decision)
	s.Equal(trinary.Unknown, out.Decision.State)
}

func (s *RuntimeTestSuite) TestExecRuleInternalRuleLookupFailureBranch() {
	p := newEvalTestPolicy()
	ruleStmt := ast.NewRuleStatement("allow", nil, nil, ast.NewTrinaryLiteral(trinary.True, stubRange()), stubRange())
	rule := &index.Rule{
		Node:   ruleStmt,
		Policy: p,
		Name:   "allow",
		FQN:    ast.CreateFQN(p.FQN, "allow"),
		Body:   ruleStmt.Body,
	}
	p.Rules["allow"] = rule
	p.RuleExports["allow"] = &index.ExportedRule{RuleName: "allow"}
	idx := index.CreateIndex()
	ns := p.Namespace
	ns.Policies = map[string]*index.Policy{p.Name: p}
	idx.Namespaces[ns.FQN.String()] = ns

	exec := &executorImpl{index: idx}
	ec := NewExecutionContext(p, exec)
	_, _, _, err := exec.execRule(context.Background(), ec, ns.FQN.String(), p.Name, "missing")
	s.Require().Error(err)
	s.ErrorIs(err, xerr.NotFoundError{})
}

func (s *RuntimeTestSuite) TestEvaluateRuleOutcomeDefaultExpressionErrorKeepsDefaultUnknown() {
	p := newEvalTestPolicy()
	ruleStmt := ast.NewRuleStatement(
		"r",
		nil,
		ast.NewTrinaryLiteral(trinary.False, stubRange()),
		ast.NewTrinaryLiteral(trinary.True, stubRange()),
		stubRange(),
	)
	rule := &index.Rule{
		Node:    ruleStmt,
		Policy:  p,
		Name:    "r",
		FQN:     ast.CreateFQN(p.FQN, "r"),
		When:    ruleStmt.When,
		Body:    ruleStmt.Body,
		Default: ast.NewIdentifier("missing_default", stubRange()),
	}
	ec := NewExecutionContext(p, &executorImpl{})
	decision, _, err := evaluateRuleOutcome(context.Background(), ec, &executorImpl{}, p, rule)
	s.Require().NoError(err)
	s.Require().NotNil(decision)
	s.Equal(trinary.Unknown, decision.State)
}

func (s *RuntimeTestSuite) TestEvaluateRuleOutcomeWhenEvaluationFailureReturnsError() {
	p := newEvalTestPolicy()
	ruleStmt := ast.NewRuleStatement(
		"r",
		nil,
		ast.NewIdentifier("missing_when", stubRange()),
		ast.NewTrinaryLiteral(trinary.True, stubRange()),
		stubRange(),
	)
	rule := &index.Rule{
		Node: ruleStmt, Policy: p, Name: "r", FQN: ast.CreateFQN(p.FQN, "r"),
		When: ruleStmt.When, Body: ruleStmt.Body,
	}
	ec := NewExecutionContext(p, &executorImpl{})
	decision, _, err := evaluateRuleOutcome(context.Background(), ec, &executorImpl{}, p, rule)
	s.Require().Error(err)
	s.Nil(decision)
}

func (s *RuntimeTestSuite) TestWithCallMemoizeCacheSizeAndToTrinaryHelpers() {
	exec := &executorImpl{}
	WithCallMemoizeCacheSize(2)(exec)
	s.NotNil(exec.callMemoizePerch)

	out := &ExecutorOutput{Decision: DecisionOf(box.Trinary(trinary.True))}
	s.Equal(trinary.True, out.ToTrinary())

	s.Panics(func() {
		_ = (&ExecutorOutput{}).ToTrinary()
	})
}

func (s *RuntimeTestSuite) TestExecPolicyRecoversPanicFromExecRule() {
	idx := index.CreateIndex()
	nsFQN := ast.NewFQN([]string{"panic", "ns"}, stubRange())
	ns := &index.Namespace{
		FQN:          nsFQN,
		Policies:     map[string]*index.Policy{},
		Shapes:       map[string]*index.Shape{},
		ShapeExports: map[string]*index.ExportedShape{},
		Children:     []*index.Namespace{},
	}
	idx.Namespaces[nsFQN.String()] = ns

	p := &index.Policy{
		Namespace:   ns,
		Name:        "pol",
		FQN:         ast.CreateFQN(nsFQN, "pol"),
		Facts:       map[string]*ast.FactStatement{},
		Rules:       map[string]*index.Rule{},
		RuleExports: map[string]*index.ExportedRule{},
		Lets:        map[string]*ast.VarDeclaration{},
		Uses:        map[string]*ast.UseStatement{},
		Shapes:      map[string]*index.Shape{},
	}
	p.RuleExports["panicRule"] = nil
	ns.Policies[p.Name] = p

	exec := &executorImpl{index: idx}
	_, err := exec.ExecPolicy(context.Background(), nsFQN.String(), p.Name, map[string]any{})
	s.Require().Error(err)
	s.Contains(err.Error(), "panic in ExecRule")
}

func testExecutorForModuleBinding() *executorImpl {
	idx := index.CreateIndex()
	idx.Pack = &pack.PackFile{Location: "."}
	exec := &executorImpl{
		index:              idx,
		jsRegistry:         js.NewRegistry("."),
		moduleBindingPerch: perch.New[*ModuleBinding](1 << 20),
		callMemoizePerch:   perch.New[any](1 << 20),
	}
	exec.jsRegistry.RegisterGoBuiltin("hash", js.BuiltinHashGo)
	exec.moduleBindingPerch.Reserve()
	exec.callMemoizePerch.Reserve()
	return exec
}

func (s *RuntimeTestSuite) TestGetModuleBindingCachesBindings() {
	exec := testExecutorForModuleBinding()
	use := ast.NewUseStatement([]string{"md5"}, "", []string{constants.APPNAME, "hash"}, "hash", stubRange())
	ms, err := exec.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, ".")
	s.Require().NoError(err)

	first, loaded, err := exec.getModuleBinding(context.Background(), use, ms)
	s.Require().NoError(err)
	_ = loaded
	s.NotNil(first)
	s.Equal(ms.KeyOrPath(), first.CanonicalKey)
	s.Equal(use.As, first.Alias)

	second, loaded, err := exec.getModuleBinding(context.Background(), use, ms)
	s.Require().NoError(err)
	_ = loaded
	s.NotNil(second)
	s.Equal(ms.KeyOrPath(), second.CanonicalKey)
	s.Equal(use.As, second.Alias)
}

func (s *RuntimeTestSuite) TestJSBindingConstructorRejectsMissingRequestedExport() {
	exec := testExecutorForModuleBinding()
	use := ast.NewUseStatement([]string{"doesNotExist"}, "", []string{constants.APPNAME, "hash"}, "hash", stubRange())
	ms, err := exec.jsRegistry.PrepareUse(use.RelativeFrom, use.LibFrom, ".")
	s.Require().NoError(err)

	_, err = exec.jsBindingConstructor(context.Background(), use, ms)
	s.Require().Error(err)
	s.Contains(err.Error(), "missing required export")
}

func (s *RuntimeTestSuite) TestBindUsesBindsPreparedModule() {
	exec := testExecutorForModuleBinding()
	use := ast.NewUseStatement([]string{"md5"}, "", []string{constants.APPNAME, "hash"}, "hash", stubRange())
	policy := &index.Policy{
		FilePath: "policy.sentra",
		Uses:     map[string]*ast.UseStatement{"hash": use},
	}
	ec := NewExecutionContext(policy, exec)

	err := exec.bindUses(context.Background(), ec, policy)
	s.Require().NoError(err)
	binding, ok := ec.Module("hash")
	s.True(ok)
	s.NotNil(binding)
}

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
