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
	"fmt"
	"strconv"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
	"github.com/sentrie-sh/sentrie/trinary"
)

func evalCast(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, cast *ast.CastExpression) (Value, *trace.Node, error) {
	ctx, node, done := trace.New(ctx, cast, "cast", map[string]any{
		"target": cast.TargetType.String(),
	})
	defer done()

	val, child, err := eval(ctx, ec, e, p, cast.Expr)
	node.Attach(child)
	if err != nil {
		return Value{}, node.SetErr(err), err
	}
	result := val
	target := cast.TargetType

	defer func() {
		if r := recover(); r != nil {
			// we are doing type casting on an unknown entity
			// catch panics and return as error
			node.SetErr(fmt.Errorf("cast: %v", r))
			err = fmt.Errorf("cast: %v", r)
			return
		}

		if result.IsValid() {
			// validate the result before returning
			if validateErr := validateValueAgainstTypeRef(ctx, ec, e, p, result.Any(), target, cast.Span()); validateErr != nil {
				node.SetErr(validateErr)
				err = validateErr
				result = Value{}
			}
		}

	}()
	switch target.(type) {
	case *ast.StringTypeRef:
		result = String(val.String())

	case *ast.NumberTypeRef:
		if n, ok := val.NumberValue(); ok {
			result = Number(n)
		} else if s, ok := val.StringValue(); ok {
			atof, parseErr := strconv.ParseFloat(s, 64)
			if parseErr != nil {
				return Value{}, node.SetErr(parseErr), parseErr
			}
			result = Number(atof)
		} else if b, ok := val.BoolValue(); ok {
			if b {
				result = Number(1)
			} else {
				result = Number(0)
			}
		} else {
			err = fmt.Errorf("cannot cast %s to number", val.Kind())
			return Value{}, node.SetErr(err), err
		}

	case *ast.TrinaryTypeRef:
		result = Trinary(trinary.From(val.Any()))

	case *ast.ListTypeRef:
		if val.Kind() != ValueList {
			err = fmt.Errorf("cannot cast %s to list", val.Kind())
			return Value{}, node.SetErr(err), err
		}
		result = val

	case *ast.MapTypeRef:
		if val.Kind() != ValueMap {
			err = fmt.Errorf("cannot cast %s to map", val.Kind())
			return Value{}, node.SetErr(err), err
		}
		result = val

	case *ast.ShapeTypeRef:
		result = val

	default:
		result = val
	}

	return result, node.SetResult(result.Any()).SetErr(err), err
}
