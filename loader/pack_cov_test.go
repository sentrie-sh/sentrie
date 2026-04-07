// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2026 Binaek Sarkar

package loader

import (
	"context"
	"os"
	"path/filepath"
)

func (s *LoaderTestSuite) TestLocatePackFileAbsPathFailure_Cov() {
	cwd, err := os.Getwd()
	s.Require().NoError(err)

	dir := s.T().TempDir()
	s.Require().NoError(os.Chdir(dir))
	s.Require().NoError(os.RemoveAll(dir))
	defer func() { _ = os.Chdir(cwd) }()

	_, err = locatePackFile(context.Background(), "relative/path")
	s.Require().Error(err)
}

func (s *LoaderTestSuite) TestLoadPackFailsForInvalidToml_Cov() {
	dir := s.T().TempDir()
	packPath := filepath.Join(dir, PackFileName)
	s.Require().NoError(os.WriteFile(packPath, []byte("[pack\nname=\"bad\""), 0o644))

	_, err := LoadPack(context.Background(), dir)
	s.Require().Error(err)
	s.Contains(err.Error(), "failed to parse pack file")
}
