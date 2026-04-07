// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar
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
	"errors"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/xerr"
)

func (suite *IndexTestSuite) TestValidateWrapsTopLevelValidationError() {
	idx := CreateIndex()
	suite.Require().NoError(idx.AddProgram(context.Background(), programWithRichRuleGraph(false)))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := idx.Validate(ctx)
	suite.Require().Error(err)
	suite.Contains(err.Error(), "validation error")
	suite.ErrorIs(err, xerr.ErrIndex)
}

func (suite *IndexTestSuite) TestShapeResolveDependencyWrapsCrossNamespaceNotExportedAsIndexError() {
	ctx := context.Background()
	idx := CreateIndex()

	appStmt := ast.NewNamespaceStatement(ast.NewFQN([]string{"com", "example", "app"}, rng(10)), rng(10))
	appNS, err := idx.ensureNamespace(ctx, appStmt)
	suite.Require().NoError(err)
	withMissing := ast.NewFQN([]string{"com", "example", "shared", "MissingShape"}, rng(11)).Ptr()
	userStmt := ast.NewShapeStatement("User", nil, &ast.Cmplx{
		Range: rng(11),
		With:  withMissing,
		Fields: map[string]*ast.ShapeField{
			"name": {Range: rng(12), Name: "name", Required: true, NotNullable: true, Type: ast.NewStringTypeRef(rng(12))},
		},
	}, rng(11))
	user, err := createShape(appNS, nil, userStmt)
	suite.Require().NoError(err)

	err = user.resolveDependency(idx, nil)
	suite.Require().Error(err)
	suite.True(errors.Is(err, xerr.ErrIndex))
	suite.Contains(err.Error(), "not found")
}
