package runtime

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RuntimeTestSuite struct {
	suite.Suite
}

func (r *RuntimeTestSuite) SetupSuite() {
	r.T().Log("SetupSuite")
}

func (r *RuntimeTestSuite) TearDownSuite() {
	r.T().Log("TearDownSuite")
}

func (r *RuntimeTestSuite) BeforeTest(suiteName, testName string) {
	r.T().Log("BeforeTest", suiteName, testName)
}

func (r *RuntimeTestSuite) AfterTest(suiteName, testName string) {
	r.T().Log("AfterTest", suiteName, testName)
}

func TestRuntimeTestSuite(t *testing.T) {
	suite.Run(t, new(RuntimeTestSuite))
}
