// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sentrie-sh/sentrie/pack"
)

func (s *CmdTestSuite) TestInitCmdReadDirFailure() {
	if runtime.GOOS == "windows" {
		s.T().Skip("permission-based ReadDir failure is not deterministic on windows")
	}

	parent := s.T().TempDir()
	target := filepath.Join(parent, "restricted")
	s.Require().NoError(os.Mkdir(target, 0o700))

	s.Require().NoError(os.Chmod(target, 0o000))
	defer func() { _ = os.Chmod(target, 0o700) }()

	err := runInitCLI(context.Background(), []string{"--directory", target, "valid_name"})
	s.Require().Error(err)
	s.Contains(err.Error(), "could not read directory")
}

func (s *CmdTestSuite) TestInitCmdWrapsEncodePackFileError() {
	dir := s.T().TempDir()
	prev := encodePackFile
	encodePackFile = func(*os.File, *pack.PackFile) error {
		return errors.New("encode failed")
	}
	defer func() { encodePackFile = prev }()

	err := runInitCLI(context.Background(), []string{"--directory", dir, "valid.pack"})
	s.Require().Error(err)
	s.Contains(err.Error(), "could not encode pack file")
	s.Contains(err.Error(), "encode failed")
}
