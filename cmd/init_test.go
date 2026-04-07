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

package cmd

import (
	"context"
	"os"
	"path/filepath"
)

func runInitCLI(ctx context.Context, args []string) error {
	cli := Setup(ctx, "test")
	return Execute(ctx, cli, append([]string{"sentrie", "init"}, args...))
}

func (s *CmdTestSuite) TestInitCmdRejectsInvalidPackNameMessage() {
	dir := s.T().TempDir()
	err := runInitCLI(context.Background(), []string{"--directory", dir, "1bad-name"})
	s.Require().Error(err)
	s.Equal("name needs to be a valid identity. It must start with a letter and can only contain letters, numbers, underscores and `dot`", err.Error())
}

func (s *CmdTestSuite) TestInitCmdWrapsCreatePackFileError() {
	dir := s.T().TempDir()
	s.Require().NoError(os.Chmod(dir, 0o500))
	defer func() { _ = os.Chmod(dir, 0o700) }()

	err := runInitCLI(context.Background(), []string{"--directory", dir, "valid.pack"})
	s.Require().Error(err)
	s.Contains(err.Error(), "could not create pack file")
}

func (s *CmdTestSuite) TestInitCmdCreatesPackFileOnSuccess() {
	dir := s.T().TempDir()
	err := runInitCLI(context.Background(), []string{"--directory", dir, "valid.pack"})
	s.Require().NoError(err)

	_, statErr := os.Stat(filepath.Join(dir, "sentrie.pack.toml"))
	s.Require().NoError(statErr)
}
