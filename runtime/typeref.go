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

	"github.com/binaek/sentra/ast"
	"github.com/binaek/sentra/index"
)

type constraintChecker[T any] func(ctx context.Context, p *index.Policy, val T, args []any) error

func validateValueAgainstTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef ast.TypeRef) error {
	switch t := typeRef.(type) {
	case *ast.IntTypeRef:
		return validateAgainstIntTypeRef(ctx, ec, exec, p, v, t)
	case *ast.StringTypeRef:
		return validateAgainstStringTypeRef(ctx, ec, exec, p, v, t)
	case *ast.BoolTypeRef:
		return validateAgainstBoolTypeRef(ctx, ec, exec, p, v, t)
	case *ast.FloatTypeRef:
		return validateAgainstFloatTypeRef(ctx, ec, exec, p, v, t)
	case *ast.ListTypeRef:
		return validateAgainstListTypeRef(ctx, ec, exec, p, v, t)
	case *ast.MapTypeRef:
		return validateAgainstMapTypeRef(ctx, ec, exec, p, v, t)
	case *ast.ShapeTypeRef:
		return validateAgainstShapeTypeRef(ctx, ec, exec, p, v, t)
	case *ast.DocumentTypeRef:
		return validateAgainstDocumentTypeRef(ctx, ec, exec, p, v, t)
	case *ast.RecordTypeRef:
		return validateAgainstRecordTypeRef(ctx, ec, exec, p, v, t)
	}

	return nil
}
