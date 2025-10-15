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

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateValueAgainstTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v any, typeRef ast.TypeRef, pos tokens.Position) error {
	switch t := typeRef.(type) {
	case *ast.IntTypeRef:
		return validateAgainstIntTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.StringTypeRef:
		return validateAgainstStringTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.BoolTypeRef:
		return validateAgainstBoolTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.FloatTypeRef:
		return validateAgainstFloatTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.ListTypeRef:
		return validateAgainstListTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.MapTypeRef:
		return validateAgainstMapTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.ShapeTypeRef:
		return validateAgainstShapeTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.DocumentTypeRef:
		return validateAgainstDocumentTypeRef(ctx, ec, exec, p, v, t, pos)
	case *ast.RecordTypeRef:
		return validateAgainstRecordTypeRef(ctx, ec, exec, p, v, t, pos)
	}

	return nil
}
