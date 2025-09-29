package index

import "github.com/sentrie-sh/sentrie/ast"

type Program struct {
	Reference    *ast.Program
	Namespace    *ast.NamespaceStatement
	Policies     []*ast.PolicyStatement
	Shapes       []*ast.ShapeStatement
	ShapeExports []*ast.ShapeExportStatement
}

func createProgram(astProgram *ast.Program) *Program {
	p := &Program{
		Reference:    astProgram,
		Namespace:    nil,
		Policies:     make([]*ast.PolicyStatement, 0),
		Shapes:       make([]*ast.ShapeStatement, 0),
		ShapeExports: make([]*ast.ShapeExportStatement, 0),
	}

	for _, stmt := range astProgram.Statements {
		switch stmt := stmt.(type) {
		case *ast.NamespaceStatement:
			p.Namespace = stmt
		case *ast.PolicyStatement:
			p.Policies = append(p.Policies, stmt)
		case *ast.ShapeStatement:
			p.Shapes = append(p.Shapes, stmt)
		case *ast.ShapeExportStatement:
			p.ShapeExports = append(p.ShapeExports, stmt)
		}
	}

	return p
}
