// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"
	"fmt"
	"strconv"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/runtime/trace"
)

func evalCast(ctx context.Context, ec *ExecutionContext, e *executorImpl, p *index.Policy, cast *ast.CastExpression) (result box.Value, node *trace.Node, err error) {
	ctx, node, done := trace.New(ctx, cast, "cast", map[string]any{
		"target": cast.TargetType.String(),
	})
	defer done()

	val, child, evalErr := eval(ctx, ec, e, p, cast.Expr)
	node.Attach(child)
	if evalErr != nil {
		return box.Value{}, node.SetErr(evalErr), evalErr
	}
	result = val
	target := cast.TargetType

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("cast: %v", r)
			node.SetErr(err)
			result = box.Value{}
			return
		}

		if result.IsValid() {
			if validateErr := validateValueAgainstTypeRef(ctx, ec, e, p, result, target, cast.Span()); validateErr != nil {
				err = validateErr
				node.SetErr(validateErr)
				result = box.Value{}
			}
		}
	}()
	switch target.(type) {
	case *ast.StringTypeRef:
		result = box.String(val.String())

	case *ast.NumberTypeRef:
		if n, ok := val.NumberValue(); ok {
			result = box.Number(n)
		} else if s, ok := val.StringValue(); ok {
			atof, parseErr := strconv.ParseFloat(s, 64)
			if parseErr != nil {
				return box.Value{}, node.SetErr(parseErr), parseErr
			}
			result = box.Number(atof)
		} else if b, ok := val.BoolValue(); ok {
			if b {
				result = box.Number(1)
			} else {
				result = box.Number(0)
			}
		} else {
			err = fmt.Errorf("cannot cast %s to number", val.Kind())
			return box.Value{}, node.SetErr(err), err
		}

	case *ast.TrinaryTypeRef:
		result = box.Trinary(box.TrinaryFrom(val))

	case *ast.ListTypeRef:
		if val.Kind() != box.ValueList {
			err = fmt.Errorf("cannot cast %s to list", val.Kind())
			return box.Value{}, node.SetErr(err), err
		}
		result = val

	case *ast.DictTypeRef:
		if val.Kind() != box.ValueDict {
			err = fmt.Errorf("cannot cast %s to dict", val.Kind())
			return box.Value{}, node.SetErr(err), err
		}
		result = val

	case *ast.ShapeTypeRef:
		result = val

	default:
		result = val
	}

	node = node.SetResult(result).SetErr(err)
	return
}
