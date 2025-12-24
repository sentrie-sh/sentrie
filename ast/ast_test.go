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

package ast

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AstTestSuite struct {
	suite.Suite
}

func (suite *AstTestSuite) SetupSuite() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(suite.T().Output(), nil)))
}

func (suite *AstTestSuite) BeforeTest(suiteName, testName string) {
	slog.InfoContext(suite.T().Context(), "BeforeTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.T().Context(), "BeforeTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *AstTestSuite) AfterTest(suiteName, testName string) {
	slog.InfoContext(suite.T().Context(), "AfterTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.T().Context(), "AfterTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *AstTestSuite) TearDownSuite() {
	slog.InfoContext(suite.T().Context(), "TearDownSuite")
	defer slog.InfoContext(suite.T().Context(), "TearDownSuite end")
}

// TestAstTestSuite runs the test suite
func TestAstTestSuite(t *testing.T) {
	suite.Run(t, new(AstTestSuite))
}
