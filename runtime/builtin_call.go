// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"time"

	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/builtins"
	"github.com/sentrie-sh/sentrie/index"
)

// CallSite is the evaluation frame passed to every builtin so higher-order
// builtins can invoke lambdas with the correct policy and executor.
type CallSite struct {
	EC     *ExecutionContext
	Exec   *executorImpl
	Policy *index.Policy
}

var _ builtins.Env = (*CallSite)(nil)

func (s *CallSite) Call(ctx context.Context, fn box.Value, args []box.Value) (box.Value, error) {
	c, err := callableFromValue(fn)
	if err != nil {
		return box.Undefined(), err
	}
	return c.Invoke(ctx, s, args)
}

func (s *CallSite) CallableArity(fn box.Value) (int, error) {
	c, err := callableFromValue(fn)
	if err != nil {
		return 0, err
	}
	return c.Arity(), nil
}

func (s *CallSite) ExecutionStart() time.Time {
	return s.EC.CreatedAt()
}
