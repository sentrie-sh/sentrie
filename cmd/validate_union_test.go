// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

func (s *CmdTestSuite) TestValidateCmdAcceptsUnionAndAnyTypes() {
	dir := s.writeTestPack(`namespace com/ex
shape Item { qty: number | list[number] }
policy pol {
  fact f?: Item | list[Item]
  fact g?: any
  rule ok = { yield [1] is number | list[number] and null is any }
  export decision of ok
}
`)
	s.Require().NoError(runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir}))
}

func (s *CmdTestSuite) TestValidateCmdRejectsUnknownShapeInUnionMember() {
	dir := s.writeTestPack(`namespace com/ex
policy pol {
  rule bad = { yield 1 is number | Missing }
  export decision of bad
}
`)
	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), `shape "Missing" is not visible here`)
}

func (s *CmdTestSuite) TestValidateCmdRejectsBarePipeInExpression() {
	dir := s.writeTestPack(`namespace com/ex
policy pol {
  rule bad = { yield 1 | 2 }
  export decision of bad
}
`)
	err := runValidateCLI(s.T().Context(), []string{"com/ex/pol", "--pack-location", dir})
	s.Require().Error(err)
	s.Contains(err.Error(), "only '|>' is supported")
}
