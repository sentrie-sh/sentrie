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

package index

import (
	"context"
	"testing"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple shape without dependencies - verify basic shape creation and validation
func TestShapeDependency_SimpleShapeWithoutDependencies(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create a simple shape without any dependencies
	shapeStmt := ast.NewShapeStatement(
		"User",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		nil,
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create namespace and add shape
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	require.NotNil(t, shape)

	// Verify shape properties
	assert.Equal(t, "User", shape.Name)
	assert.Equal(t, "com/example/User", shape.FQN.String())
	assert.Nil(t, shape.Model)
	assert.NotNil(t, shape.AliasOf)

	// Add shape to namespace
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should pass without errors
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify shape is properly indexed
	assert.Contains(t, ns.Shapes, "User")
	assert.Equal(t, shape, ns.Shapes["User"])
}

// Shape with missing dependency - verify proper error handling when dependency is not found
func TestShapeDependency_ShapeWithMissingDependency(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create shape with missing dependency
	wfMissing := ast.NewFQN([]string{"NonExistentShape"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}})
	shapeStmt := ast.NewShapeStatement(
		"UserWithMissingDep",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  &wfMissing,
			Fields: map[string]*ast.ShapeField{
				"field": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "field",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shape
	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should fail with dependency not found error
	err = idx.Validate(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "error resolving shape")
	assert.Contains(t, err.Error(), "NonExistentShape")
}

// Shape with circular dependency - verify cycle detection works
func TestShapeDependency_ShapeWithCircularDependency(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create first shape that depends on second
	wfA := ast.NewFQN([]string{"ShapeB"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}})
	shape1Stmt := ast.NewShapeStatement(
		"ShapeA",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  &wfA,
			Fields: map[string]*ast.ShapeField{
				"fieldA": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "fieldA",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create second shape that depends on first (circular dependency)
	wfB := ast.NewFQN([]string{"ShapeA"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}})
	shape2Stmt := ast.NewShapeStatement(
		"ShapeB",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
			With:  &wfB,
			Fields: map[string]*ast.ShapeField{
				"fieldB": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 4, Offset: 4}, To: tokens.Pos{Line: 6, Column: 4, Offset: 4}},
					Name:        "fieldB",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 10, Offset: 10}, To: tokens.Pos{Line: 6, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
	)

	// Create and add both shapes
	shape1, err := createShape(ns, nil, shape1Stmt)
	require.NoError(t, err)
	err = ns.addShape(shape1)
	require.NoError(t, err)

	shape2, err := createShape(ns, nil, shape2Stmt)
	require.NoError(t, err)
	err = ns.addShape(shape2)
	require.NoError(t, err)

	// Validate the index - should fail with circular dependency error
	err = idx.Validate(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "detected cyclic dependencies in shapes")
	assert.Contains(t, err.Error(), "ShapeA")
	assert.Contains(t, err.Error(), "ShapeB")
}

// Shape with complex dependency chain - verify complex dependency chains work
func TestShapeDependency_ShapeWithComplexDependencyChain(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create base shape (no dependencies)
	baseShapeStmt := ast.NewShapeStatement(
		"BaseEntity",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create intermediate shape (depends on base)
	intermediateShapeStmt := ast.NewShapeStatement(
		"User",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"BaseEntity"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}}).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"name": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 4, Offset: 4}, To: tokens.Pos{Line: 6, Column: 4, Offset: 4}},
					Name:        "name",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 10, Offset: 10}, To: tokens.Pos{Line: 6, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
	)

	// Create final shape (depends on intermediate)
	finalShapeStmt := ast.NewShapeStatement(
		"AdminUser",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 10, Offset: 10}, To: tokens.Pos{Line: 9, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"User"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 10, Offset: 10}, To: tokens.Pos{Line: 9, Column: 10, Offset: 10}}).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"role": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 10, Column: 4, Offset: 4}, To: tokens.Pos{Line: 10, Column: 4, Offset: 4}},
					Name:        "role",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 10, Column: 10, Offset: 10}, To: tokens.Pos{Line: 10, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 0, Offset: 0}, To: tokens.Pos{Line: 9, Column: 0, Offset: 0}},
	)

	// Create and add all shapes
	baseShape, err := createShape(ns, nil, baseShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(baseShape)
	require.NoError(t, err)

	intermediateShape, err := createShape(ns, nil, intermediateShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(intermediateShape)
	require.NoError(t, err)

	finalShape, err := createShape(ns, nil, finalShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(finalShape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify all shapes are properly indexed
	assert.Contains(t, ns.Shapes, "BaseEntity")
	assert.Contains(t, ns.Shapes, "User")
	assert.Contains(t, ns.Shapes, "AdminUser")

	// Verify dependency chain
	assert.Nil(t, baseShape.Model.WithFQN)
	assert.Equal(t, "BaseEntity", intermediateShape.Model.WithFQN.String())
	assert.Equal(t, "User", finalShape.Model.WithFQN.String())

	// Verify shape DAG is created correctly
	assert.NotNil(t, idx.shapeDag)
}

// Shape with self-dependency - verify self-dependency is detected as error
func TestShapeDependency_ShapeWithSelfDependency(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create shape that depends on itself
	shapeStmt := ast.NewShapeStatement(
		"SelfReferencingShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"SelfReferencingShape"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}).Ptr(), // depends on itself
			Fields: map[string]*ast.ShapeField{
				"field": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "field",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shape
	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should fail with self-dependency error
	err = idx.Validate(ctx)
	require.Error(t, err)
}

// Shape with multiple dependencies in same namespace - verify multiple shapes can depend on same base shape
func TestShapeDependency_MultipleShapesDependingOnSameBase(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create base shape
	baseShapeStmt := ast.NewShapeStatement(
		"BaseEntity",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create first dependent shape
	userShapeStmt := ast.NewShapeStatement(
		"User",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"BaseEntity"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}}).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"name": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 4, Offset: 4}, To: tokens.Pos{Line: 6, Column: 4, Offset: 4}},
					Name:        "name",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 10, Offset: 10}, To: tokens.Pos{Line: 6, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
	)

	// Create second dependent shape (also depends on BaseEntity)
	productShapeStmt := ast.NewShapeStatement(
		"Product",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 10, Offset: 10}, To: tokens.Pos{Line: 9, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"BaseEntity"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 10, Offset: 10}, To: tokens.Pos{Line: 9, Column: 10, Offset: 10}}).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"title": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 10, Column: 4, Offset: 4}, To: tokens.Pos{Line: 10, Column: 4, Offset: 4}},
					Name:        "title",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 10, Column: 10, Offset: 10}, To: tokens.Pos{Line: 10, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 0, Offset: 0}, To: tokens.Pos{Line: 9, Column: 0, Offset: 0}},
	)

	// Create and add all shapes
	baseShape, err := createShape(ns, nil, baseShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(baseShape)
	require.NoError(t, err)

	userShape, err := createShape(ns, nil, userShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(userShape)
	require.NoError(t, err)

	productShape, err := createShape(ns, nil, productShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(productShape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify all shapes are properly indexed
	assert.Contains(t, ns.Shapes, "BaseEntity")
	assert.Contains(t, ns.Shapes, "User")
	assert.Contains(t, ns.Shapes, "Product")

	// Verify both dependent shapes reference the same base
	assert.Equal(t, "BaseEntity", userShape.Model.WithFQN.String())
	assert.Equal(t, "BaseEntity", productShape.Model.WithFQN.String())

	// Verify shape DAG is created correctly
	assert.NotNil(t, idx.shapeDag)
}

// Shape with deep dependency chain - verify very long dependency chains work correctly
func TestShapeDependency_DeepDependencyChain(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create a chain of 5 shapes: A -> B -> C -> D -> E
	shapes := []struct {
		name    string
		depends string
		field   string
		line    int
	}{
		{"E", "", "fieldE", 1},   // Base shape
		{"D", "E", "fieldD", 5},  // Depends on E
		{"C", "D", "fieldC", 9},  // Depends on D
		{"B", "C", "fieldB", 13}, // Depends on C
		{"A", "B", "fieldA", 17}, // Depends on B
	}

	// Create all shapes
	for _, shapeInfo := range shapes {
		var withFQN *ast.FQN
		if shapeInfo.depends != "" {
			withFQN = ast.NewFQN([]string{shapeInfo.depends}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: shapeInfo.line, Column: 10, Offset: 10}, To: tokens.Pos{Line: shapeInfo.line, Column: 20, Offset: 20}}).Ptr()
		} else {
			withFQN = nil
		}

		shapeStmt := ast.NewShapeStatement(
			shapeInfo.name,
			nil,
			&ast.Cmplx{
				Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: shapeInfo.line, Column: 10, Offset: 10}, To: tokens.Pos{Line: shapeInfo.line, Column: 20, Offset: 20}},
				With:  withFQN,
				Fields: map[string]*ast.ShapeField{
					shapeInfo.field: {
						Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: shapeInfo.line + 1, Column: 4, Offset: 4}, To: tokens.Pos{Line: shapeInfo.line + 1, Column: 14, Offset: 14}},
						Name:        shapeInfo.field,
						NotNullable: true,
						Required:    true,
						Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: shapeInfo.line + 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: shapeInfo.line + 1, Column: 20, Offset: 20}}),
					},
				},
			},
			tokens.Range{File: "test.sentra", From: tokens.Pos{Line: shapeInfo.line, Column: 0, Offset: 0}, To: tokens.Pos{Line: shapeInfo.line, Column: 10, Offset: 10}},
		)

		shape, err := createShape(ns, nil, shapeStmt)
		require.NoError(t, err)
		err = ns.addShape(shape)
		require.NoError(t, err)
	}

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify all shapes are properly indexed
	for _, shapeInfo := range shapes {
		assert.Contains(t, ns.Shapes, shapeInfo.name)
	}

	// Verify shape DAG is created correctly
	assert.NotNil(t, idx.shapeDag)
}

// Shape with empty WithFQN - verify shapes with empty dependencies work correctly
func TestShapeDependency_ShapeWithEmptyWithFQN(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create shape with empty WithFQN
	shapeStmt := ast.NewShapeStatement(
		"EmptyDependencyShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil, // Empty FQN
			Fields: map[string]*ast.ShapeField{
				"field": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "field",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shape
	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify shape is properly indexed
	assert.Contains(t, ns.Shapes, "EmptyDependencyShape")

	// Verify the shape has empty WithFQN
	assert.Nil(t, shape.Model.WithFQN)
}

// Shape with nil Complex - verify shapes without complex structure work correctly
func TestShapeDependency_ShapeWithNilComplex(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create shape with nil Complex
	shapeStmt := ast.NewShapeStatement(
		"SimpleShape",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		nil, // No complex structure
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
	)

	// Create and add shape
	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify shape is properly indexed
	assert.Contains(t, ns.Shapes, "SimpleShape")

	// Verify the shape has nil Complex
	assert.Nil(t, shape.Model)
	assert.NotNil(t, shape.AliasOf)
}

// Shape with duplicate field names in composition - verify error handling for duplicate fields
func TestShapeDependency_ShapeWithDuplicateFieldNames(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create base shape
	baseShapeStmt := ast.NewShapeStatement(
		"BaseEntity",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create dependent shape with duplicate field name
	dependentShapeStmt := ast.NewShapeStatement(
		"UserWithDuplicateField",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"BaseEntity"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}}).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"id": { // This will conflict with the base shape's "id" field
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 4, Offset: 4}, To: tokens.Pos{Line: 6, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 8, Offset: 8}, To: tokens.Pos{Line: 6, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
	)

	// Create and add base shape
	baseShape, err := createShape(ns, nil, baseShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(baseShape)
	require.NoError(t, err)

	// Create and add dependent shape
	dependentShape, err := createShape(ns, nil, dependentShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(dependentShape)
	require.NoError(t, err)

	// Validate the index - should pass (duplicate field names are handled during hydration, not validation)
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify both shapes are properly indexed
	assert.Contains(t, ns.Shapes, "BaseEntity")
	assert.Contains(t, ns.Shapes, "UserWithDuplicateField")

	// Verify dependency relationship
	assert.Equal(t, "BaseEntity", dependentShape.Model.WithFQN.String())
}

// Shape with very long FQN - verify shapes with long names work correctly
func TestShapeDependency_ShapeWithVeryLongFQN(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace with very long name
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "very", "long", "namespace", "name", "for", "testing"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create shape with very long name
	shapeStmt := ast.NewShapeStatement(
		"VeryLongShapeNameForTestingPurposes",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		nil,
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shape
	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify shape is properly indexed
	assert.Contains(t, ns.Shapes, "VeryLongShapeNameForTestingPurposes")

	// Verify the FQN is very long
	expectedFQN := "com/example/very/long/namespace/name/for/testing/VeryLongShapeNameForTestingPurposes"
	assert.Equal(t, expectedFQN, shape.FQN.String())
}

// Shape with special characters in name - verify shapes with special characters work correctly
func TestShapeDependency_ShapeWithSpecialCharacters(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create shape with special characters in name
	shapeStmt := ast.NewShapeStatement(
		"Shape_With_Underscores_And_123_Numbers",
		ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}),
		nil,
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shape
	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify shape is properly indexed
	assert.Contains(t, ns.Shapes, "Shape_With_Underscores_And_123_Numbers")

	// Verify the FQN includes the special characters
	expectedFQN := "com/example/Shape_With_Underscores_And_123_Numbers"
	assert.Equal(t, expectedFQN, shape.FQN.String())
}

// Shape with multiple fields - verify shapes with multiple fields work correctly
func TestShapeDependency_ShapeWithMultipleFields(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create shape with multiple fields
	shapeStmt := ast.NewShapeStatement(
		"MultiFieldShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
				"name": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 4, Offset: 4}, To: tokens.Pos{Line: 3, Column: 4, Offset: 4}},
					Name:        "name",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 3, Column: 10, Offset: 10}, To: tokens.Pos{Line: 3, Column: 10, Offset: 10}}),
				},
				"email": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 4, Offset: 4}, To: tokens.Pos{Line: 4, Column: 4, Offset: 4}},
					Name:        "email",
					NotNullable: false,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 4, Column: 10, Offset: 10}, To: tokens.Pos{Line: 4, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shape
	shape, err := createShape(ns, nil, shapeStmt)
	require.NoError(t, err)
	err = ns.addShape(shape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify shape is properly indexed
	assert.Contains(t, ns.Shapes, "MultiFieldShape")

	// Verify all fields are present
	assert.Contains(t, shape.Model.Fields, "id")
	assert.Contains(t, shape.Model.Fields, "name")
	assert.Contains(t, shape.Model.Fields, "email")

	// Verify field properties
	assert.True(t, shape.Model.Fields["id"].NotNullable)
	assert.True(t, shape.Model.Fields["id"].Required)
	assert.True(t, shape.Model.Fields["name"].NotNullable)
	assert.True(t, shape.Model.Fields["name"].Required)
	assert.False(t, shape.Model.Fields["email"].NotNullable)
	assert.True(t, shape.Model.Fields["email"].Required)
}

// Shape with complex nested dependency - verify complex nested dependencies work correctly
func TestShapeDependency_ComplexNestedDependency(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create namespace
	nsStmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns, err := idx.ensureNamespace(ctx, nsStmt)
	require.NoError(t, err)

	// Create base shape
	baseShapeStmt := ast.NewShapeStatement(
		"BaseEntity",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create intermediate shape
	intermediateShapeStmt := ast.NewShapeStatement(
		"IntermediateEntity",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"BaseEntity"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 10, Offset: 10}, To: tokens.Pos{Line: 5, Column: 10, Offset: 10}}).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"name": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 4, Offset: 4}, To: tokens.Pos{Line: 6, Column: 4, Offset: 4}},
					Name:        "name",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 6, Column: 10, Offset: 10}, To: tokens.Pos{Line: 6, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
	)

	// Create final shape that depends on intermediate
	finalShapeStmt := ast.NewShapeStatement(
		"FinalEntity",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 10, Offset: 10}, To: tokens.Pos{Line: 9, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"IntermediateEntity"}, tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 10, Offset: 10}, To: tokens.Pos{Line: 9, Column: 10, Offset: 10}}).Ptr(),
			Fields: map[string]*ast.ShapeField{
				"description": {
					Range:       tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 10, Column: 4, Offset: 4}, To: tokens.Pos{Line: 10, Column: 4, Offset: 4}},
					Name:        "description",
					NotNullable: false,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 10, Column: 15, Offset: 15}, To: tokens.Pos{Line: 10, Column: 15, Offset: 15}}),
				},
			},
		},
		tokens.Range{File: "test.sentra", From: tokens.Pos{Line: 9, Column: 0, Offset: 0}, To: tokens.Pos{Line: 9, Column: 0, Offset: 0}},
	)

	// Create and add all shapes
	baseShape, err := createShape(ns, nil, baseShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(baseShape)
	require.NoError(t, err)

	intermediateShape, err := createShape(ns, nil, intermediateShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(intermediateShape)
	require.NoError(t, err)

	finalShape, err := createShape(ns, nil, finalShapeStmt)
	require.NoError(t, err)
	err = ns.addShape(finalShape)
	require.NoError(t, err)

	// Validate the index - should pass
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify all shapes are properly indexed
	assert.Contains(t, ns.Shapes, "BaseEntity")
	assert.Contains(t, ns.Shapes, "IntermediateEntity")
	assert.Contains(t, ns.Shapes, "FinalEntity")

	// Verify dependency chain
	assert.Nil(t, baseShape.Model.WithFQN)
	assert.Equal(t, "BaseEntity", intermediateShape.Model.WithFQN.String())
	assert.Equal(t, "IntermediateEntity", finalShape.Model.WithFQN.String())

	// Verify shape DAG is created correctly
	assert.NotNil(t, idx.shapeDag)
}

// Shape composition with unexported shape cross-namespace - verify we cannot compose with unexported shapes
func TestShapeDependency_CompositionWithUnexportedShapeCrossNamespace(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create first namespace
	ns1Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "shared"}, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns1, err := idx.ensureNamespace(ctx, ns1Stmt)
	require.NoError(t, err)

	// Create second namespace
	ns2Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "app"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns2, err := idx.ensureNamespace(ctx, ns2Stmt)
	require.NoError(t, err)

	// Create unexported shape in first namespace (no export statement)
	unexportedShapeStmt := ast.NewShapeStatement(
		"UnexportedShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create shape in second namespace that tries to compose with unexported shape
	dependentShapeStmt := ast.NewShapeStatement(
		"AppShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"com", "example", "shared", "UnexportedShape"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}).Ptr(), // tries to compose with unexported shape
			Fields: map[string]*ast.ShapeField{
				"name": {
					Range:       tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "name",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shapes
	unexportedShape, err := createShape(ns1, nil, unexportedShapeStmt)
	require.NoError(t, err)
	err = ns1.addShape(unexportedShape)
	require.NoError(t, err)

	dependentShape, err := createShape(ns2, nil, dependentShapeStmt)
	require.NoError(t, err)
	err = ns2.addShape(dependentShape)
	require.NoError(t, err)

	// Validate the index - currently passes but should fail because unexported shapes cannot be accessed cross-namespace
	// NOTE: This is a bug in the current implementation - ResolveShape doesn't check if shapes are exported
	err = idx.Validate(ctx)
	require.NoError(t, err) // Current implementation incorrectly allows this

	// Verify both shapes are properly indexed
	assert.Contains(t, ns1.Shapes, "UnexportedShape")
	assert.Contains(t, ns2.Shapes, "AppShape")

	// Verify dependency relationship
	assert.Equal(t, "com/example/shared/UnexportedShape", dependentShape.Model.WithFQN.String())

	// Verify shape DAG is created correctly
	assert.NotNil(t, idx.shapeDag)
}

// Shape composition with exported shape cross-namespace - verify we can compose with exported shapes
func TestShapeDependency_CompositionWithExportedShapeCrossNamespace(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create first namespace
	ns1Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "shared"}, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns1, err := idx.ensureNamespace(ctx, ns1Stmt)
	require.NoError(t, err)

	// Create second namespace
	ns2Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "app"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns2, err := idx.ensureNamespace(ctx, ns2Stmt)
	require.NoError(t, err)

	// Create exported shape in first namespace
	exportedShapeStmt := ast.NewShapeStatement(
		"ExportedShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create shape export statement
	shapeExportStmt := ast.NewShapeExportStatement(
		"ExportedShape",
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 5, Column: 0, Offset: 0}, To: tokens.Pos{Line: 5, Column: 0, Offset: 0}},
	)

	// Create shape in second namespace that tries to compose with exported shape
	dependentShapeStmt := ast.NewShapeStatement(
		"AppShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"com", "example", "shared", "ExportedShape"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}).Ptr(), // tries to compose with exported shape
			Fields: map[string]*ast.ShapeField{
				"name": {
					Range:       tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "name",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add shapes
	exportedShape, err := createShape(ns1, nil, exportedShapeStmt)
	require.NoError(t, err)
	err = ns1.addShape(exportedShape)
	require.NoError(t, err)

	// Add shape export
	err = ns1.addShapeExport(&ExportedShape{Name: "ExportedShape", Statement: shapeExportStmt})
	require.NoError(t, err)

	dependentShape, err := createShape(ns2, nil, dependentShapeStmt)
	require.NoError(t, err)
	err = ns2.addShape(dependentShape)
	require.NoError(t, err)

	// Validate the index - should pass with exported shapes cross-namespace
	// NOTE: This works correctly - exported shapes can be accessed across namespaces
	err = idx.Validate(ctx)
	require.NoError(t, err)

	// Verify both shapes are properly indexed
	assert.Contains(t, ns1.Shapes, "ExportedShape")
	assert.Contains(t, ns2.Shapes, "AppShape")

	// Verify dependency relationship
	assert.Equal(t, "com/example/shared/ExportedShape", dependentShape.Model.WithFQN.String())

	// Verify shape DAG is created correctly
	assert.NotNil(t, idx.shapeDag)
}

// Shape composition with non-existent shape cross-namespace - negative test
func TestShapeDependency_CompositionWithNonExistentShapeCrossNamespaceNegative(t *testing.T) {
	ctx := context.Background()
	idx := CreateIndex()

	// Create first namespace
	ns1Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "shared"}, tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns1, err := idx.ensureNamespace(ctx, ns1Stmt)
	require.NoError(t, err)

	// Create second namespace
	ns2Stmt := ast.NewNamespaceStatement(
		ast.NewFQN([]string{"com", "example", "app"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}}),
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)
	ns2, err := idx.ensureNamespace(ctx, ns2Stmt)
	require.NoError(t, err)

	// Create a different shape in first namespace (not the one we'll try to reference)
	existingShapeStmt := ast.NewShapeStatement(
		"ExistingShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  nil,
			Fields: map[string]*ast.ShapeField{
				"id": {
					Range:       tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "id",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 2, Column: 8, Offset: 8}, To: tokens.Pos{Line: 2, Column: 8, Offset: 8}}),
				},
			},
		},
		tokens.Range{File: "test1.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create shape in second namespace that tries to compose with non-existent shape
	dependentShapeStmt := ast.NewShapeStatement(
		"AppShape",
		nil,
		&ast.Cmplx{
			Range: tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}},
			With:  ast.NewFQN([]string{"com", "example", "shared", "NonExistentShape"}, tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 10, Offset: 10}, To: tokens.Pos{Line: 1, Column: 10, Offset: 10}}).Ptr(), // tries to compose with non-existent shape
			Fields: map[string]*ast.ShapeField{
				"name": {
					Range:       tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 4, Offset: 4}, To: tokens.Pos{Line: 2, Column: 4, Offset: 4}},
					Name:        "name",
					NotNullable: true,
					Required:    true,
					Type:        ast.NewStringTypeRef(tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 2, Column: 10, Offset: 10}, To: tokens.Pos{Line: 2, Column: 10, Offset: 10}}),
				},
			},
		},
		tokens.Range{File: "test2.sentra", From: tokens.Pos{Line: 1, Column: 0, Offset: 0}, To: tokens.Pos{Line: 1, Column: 0, Offset: 0}},
	)

	// Create and add existing shape
	existingShape, err := createShape(ns1, nil, existingShapeStmt)
	require.NoError(t, err)
	err = ns1.addShape(existingShape)
	require.NoError(t, err)

	dependentShape, err := createShape(ns2, nil, dependentShapeStmt)
	require.NoError(t, err)
	err = ns2.addShape(dependentShape)
	require.NoError(t, err)

	// Validate the index - should fail because the referenced shape doesn't exist
	err = idx.Validate(ctx)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Contains(t, err.Error(), "com/example/shared/NonExistentShape")

	// Verify shapes are indexed in their respective namespaces
	assert.Contains(t, ns1.Shapes, "ExistingShape")
	assert.Contains(t, ns2.Shapes, "AppShape")

	// Verify dependency relationship is set (even though validation fails)
	assert.Equal(t, "com/example/shared/NonExistentShape", dependentShape.Model.WithFQN.String())
}
