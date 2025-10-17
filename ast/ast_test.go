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
