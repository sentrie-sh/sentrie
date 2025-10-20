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
)

func evalCast(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, cast *ast.CastExpression) (any, *trace.Node, error) {
	node, done := trace.New("cast", cast.TargetType.String(), cast, map[string]any{})
	defer done()

	val, child, err := eval(ctx, ec, e, p, cast.Expr)
	node.Attach(child)
	if err != nil {
		return nil, node.SetErr(err), err
	}

	var result any
	target := cast.TargetType

	defer func() {
		if r := recover(); r != nil {
			// we are doing type casting on an unknown entity
			// catch panics and return as error
			node.SetErr(fmt.Errorf("cast: %v", r))
			err = fmt.Errorf("cast: %v", r)
			return
		}

		if result != nil {
			// validate the result before returning
			if validateErr := validateValueAgainstTypeRef(ctx, ec, e, p, result, target, cast.Span()); validateErr != nil {
				node.SetErr(validateErr)
				err = validateErr
				result = nil
			}
		}

	}()
	switch target.(type) {
	case *ast.StringTypeRef:
		result = fmt.Sprintf("%v", val)

	case *ast.NumberTypeRef:
		switch v := val.(type) {
		case float32:
			result = float64(v)
		case float64:
			result = float64(v)
		case int:
			result = float64(v)
		case int64:
			result = float64(v)
		case string:
			atof, parseErr := strconv.ParseFloat(v, 64)
			if parseErr != nil {
				return nil, node.SetErr(parseErr), parseErr
			}
			result = atof
		default:
			err = fmt.Errorf("cannot cast %T to float", val)
			return nil, node.SetErr(err), err
		}

	case *ast.BoolTypeRef:
		switch v := val.(type) {
		case bool:
			result = v
		case string:
			parsed, parseErr := strconv.ParseBool(v)
			if parseErr != nil {
				return nil, node.SetErr(parseErr), parseErr
			}
			result = parsed
		case int:
			result = v != 0
		case float64:
			result = v != 0
		default:
			err = fmt.Errorf("cannot cast %T to bool", val)
			return nil, node.SetErr(err), err
		}

	case *ast.ListTypeRef:
		switch v := val.(type) {
		case []any:
			// Already an array, return as-is
			result = v
		case []string:
			// Convert []string to []any
			arr := make([]any, len(v))
			for i, s := range v {
				arr[i] = s
			}
			result = arr
		case []int:
			// Convert []int to []any
			arr := make([]any, len(v))
			for i, n := range v {
				arr[i] = n
			}
			result = arr
		case []float64:
			// Convert []float64 to []any
			arr := make([]any, len(v))
			for i, f := range v {
				arr[i] = f
			}
			result = arr
		default:
			err = fmt.Errorf("cannot cast %T to array", val)
			return nil, node.SetErr(err), err
		}

	case *ast.MapTypeRef:
		switch v := val.(type) {
		case map[string]any:
			// Already a map, return as-is
			result = v
		default:
			err = fmt.Errorf("cannot cast %T to map", val)
			return nil, node.SetErr(err), err
		}

	case *ast.ShapeTypeRef:
		// For shape types, we just return the value as-is
		// Shape validation would typically happen elsewhere
		result = val

	default:
		// Unknown type, return value as-is
		result = val
	}

	return result, node.SetResult(result).SetErr(err), err
}
