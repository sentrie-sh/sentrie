// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package constraints

import "github.com/sentrie-sh/sentrie/ast"

// ForTypeRef returns the definitions that apply to typeRef's own constraints,
// or nil for a type ref that takes none. A nullable type ref has no
// constraints of its own (they belong to its inner type), so callers unwrap
// it first. The runtime validators and the index's static argument check
// both select a table here.
func ForTypeRef(typeRef ast.TypeRef) map[string]ConstraintDefinition {
	switch typeRef.(type) {
	case *ast.StringTypeRef:
		return StringContraintCheckers
	case *ast.NumberTypeRef:
		return NumberContraintCheckers
	case *ast.TrinaryTypeRef:
		return TrinaryConstraintCheckers
	case *ast.ListTypeRef:
		return ListContraintCheckers
	case *ast.DictTypeRef:
		return DictContraintCheckers
	case *ast.DocumentTypeRef:
		return DocumentContraintCheckers
	case *ast.RecordTypeRef:
		return RecordContraintCheckers
	case *ast.ShapeTypeRef:
		return ShapeContraintCheckers
	}
	return nil
}
