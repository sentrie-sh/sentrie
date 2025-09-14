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
	"fmt"

	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/runtime/trace"
	"github.com/pkg/errors"
)

func evalIdent(ctx context.Context, ec *ExecutionContext, exec *executorImpl, p *index.Policy, i string) (any, *trace.Node, error) {
	n, done := trace.New("identifier", "", nil, map[string]any{"name": i})
	defer done()

	// check in the local scope
	if v, ok := ec.GetLocal(i); ok {
		return v, n.SetResult(v), nil
	}

	// check whether this has been passed in as a FACT
	if v, ok := ec.GetFact(i); ok {
		return v, n.SetResult(v), nil
	}

	// we couldn't find anything yet - look for a let declaration in the ExecutionContext
	if v, ok := ec.GetLet(i); ok {
		// we found a let declaration - evaluate it and set the local
		val, letEvalNode, err := eval(ctx, ec, exec, p, v.Value)
		n.Attach(letEvalNode)
		if err != nil {
			return nil, n.SetErr(err), err
		}

		// check the type of the let declaration
		if v.Type != nil {
			if err := validateValueAgainstTypeRef(ctx, ec, exec, p, val, v.Type); err != nil {
				return nil, n.SetErr(errors.Wrapf(err, "invalid value for let declaration %s", i)), err
			}
		}

		ec.SetLocal(i, val)
		return val, n.SetResult(val), nil
	}

	if r, found := p.Rules[i]; found {
		decision, _, node, err := exec.execRule(ctx, ec, p.Namespace.FQN.String(), p.Name, r.Name)
		n.Attach(node)
		if err != nil {
			return nil, n.SetErr(err), err
		}
		ec.SetLocal(i, decision)
		return decision, n.SetResult(decision), nil
	}

	err := fmt.Errorf("identifier not found: %s", i)
	return nil, n.SetErr(err), err
}
