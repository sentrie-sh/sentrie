// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package format

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FormatTestSuite struct {
	suite.Suite
}

func TestFormatTestSuite(t *testing.T) {
	suite.Run(t, new(FormatTestSuite))
}

// corpus lists every .sentrie file under lang_test/ and example_pack/,
// subdirectories included.
func (s *FormatTestSuite) corpus() []string {
	var files []string
	for _, dir := range []string{"../lang_test", "../example_pack"} {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && filepath.Ext(path) == ".sentrie" {
				files = append(files, path)
			}
			return err
		})
		s.Require().NoError(err)
	}
	s.Require().NotEmpty(files)
	return files
}

// TestCorpusIsStable formats every corpus file twice. Source itself checks
// that the output parses to a program equal to the input, comments included;
// this adds that formatting is idempotent.
func (s *FormatTestSuite) TestCorpusIsStable() {
	for _, file := range s.corpus() {
		s.Run(filepath.Base(file), func() {
			src, err := os.ReadFile(file)
			s.Require().NoError(err)
			once, err := Source(s.T().Context(), file, src)
			s.Require().NoError(err)
			before, err := parse(s.T().Context(), file, src)
			s.Require().NoError(err)
			after, err := parse(s.T().Context(), file, once)
			s.Require().NoError(err)
			s.NoError(Equal(before, after), "formatting changed the program")
			twice, err := Source(s.T().Context(), file, once)
			s.Require().NoError(err)
			s.Equal(string(once), string(twice), "formatting is not idempotent")
		})
	}
}
