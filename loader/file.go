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

package loader

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sentrie-sh/sentrie/ast"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/pack"
	"github.com/sentrie-sh/sentrie/parser"
)

func LoadPrograms(ctx context.Context, packFile *pack.PackFile) ([]*ast.Program, error) {
	// walk the directory tree - starting from root
	// if we find a .sentra file, we load it
	programs := make([]*ast.Program, 0)
	err := fs.WalkDir(os.DirFS(packFile.Location), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(filepath.Ext(d.Name()), constants.PolicyFileExtension) {
			return nil
		}

		path = filepath.Join(packFile.Location, path)
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		parser := parser.NewParser(file, path)
		program, err := parser.ParseProgram(ctx)
		if err != nil {
			return err
		}
		if program == nil {
			return nil
		}

		programs = append(programs, program)

		return nil
	})

	return programs, err
}
