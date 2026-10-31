// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"context"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/box"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/tokens"
)

func validateValueAgainstTypeRef(ctx context.Context, ec *ExecutionContext, exec Executor, p *index.Policy, v box.Value, typeRef ast.TypeRef, valueRange tokens.Range) error {
	if ast.IsNullableTypeRef(typeRef) {
		typeRef = ast.UnwrapNullableTypeRef(typeRef)
		if v.IsNull() {
			// null conforms to T?, but a broken constraint argument on T is
			// still an error, whatever the value (#273).
			_, err := bindConstraints(ctx, ec, exec, p, typeRef)
			return err
		}
	}

	switch t := typeRef.(type) {
	case *ast.StringTypeRef:
		return validateAgainstStringTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.TrinaryTypeRef:
		return validateAgainstTrinaryTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.NumberTypeRef:
		return validateAgainstNumberTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.ListTypeRef:
		return validateAgainstListTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.DictTypeRef:
		return validateAgainstDictTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.ShapeTypeRef:
		return validateAgainstShapeTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.DocumentTypeRef:
		return validateAgainstDocumentTypeRef(ctx, ec, exec, p, v, t, valueRange)
	case *ast.RecordTypeRef:
		return validateAgainstRecordTypeRef(ctx, ec, exec, p, v, t, valueRange)
	}

	return nil
}
