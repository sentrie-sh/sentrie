package trinary

import (
	"context"
	"io"
	"log/slog"

	"github.com/stretchr/testify/suite"
)

type TristateTestSuite struct {
	suite.Suite
	ctx context.Context
}

func (s *TristateTestSuite) TestTristateLogicalOperators() {
	s.Equal(True.And(True), True)
	s.Equal(Unknown.And(True), Unknown)
	s.Equal(Unknown.And(False), False)
	s.Equal(Unknown.And(Unknown), Unknown)
	s.Equal(True.And(Unknown), Unknown)
	s.Equal(False.And(Unknown), Unknown)
	s.Equal(Unknown.And(Unknown), Unknown)

	s.Equal(True.Or(True), True)
	s.Equal(Unknown.Or(True), True)
	s.Equal(Unknown.Or(False), Unknown)
	s.Equal(Unknown.Or(Unknown), Unknown)
	s.Equal(True.Or(Unknown), True)
	s.Equal(False.Or(Unknown), Unknown)
	s.Equal(Unknown.Or(Unknown), Unknown)

	s.Equal(True.Not(), False)
	s.Equal(False.Not(), True)
	s.Equal(Unknown.Not(), Unknown)
}

func (suite *TristateTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// set slog to discard logs
	// this is so that we don't have logs getting spit out in the test output
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, nil)))
}

func (suite *TristateTestSuite) BeforeTest(suiteName, testName string) {
	slog.InfoContext(suite.ctx, "BeforeTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.ctx, "BeforeTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *TristateTestSuite) AfterTest(suiteName, testName string) {
	slog.InfoContext(suite.ctx, "AfterTest start", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
	defer slog.InfoContext(suite.ctx, "AfterTest end", slog.String("TestSuite", suiteName), slog.String("TestName", testName))
}

func (suite *TristateTestSuite) TearDownSuite() {
}
