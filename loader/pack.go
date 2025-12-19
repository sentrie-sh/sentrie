// SPDX-License-Identifier: Apache-2.0

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
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/pack"
)

var (
	ErrPackFileNotFound   = errors.New("pack file not found")
	ErrPackFileLoadFailed = errors.New("pack file load failed")
)

var (
	PackFileName = (constants.APPNAME + "." + constants.PackFileExtension)
)

func LoadPack(ctx context.Context, root string) (_ *pack.PackFile, e error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	packPath, err := locatePackFile(ctx, root)
	if err != nil {
		return nil, errors.Wrap(err, "locate pack file")
	}

	stat, err := os.Stat(packPath)
	if err != nil {
		return nil, errors.Wrap(err, "stat pack file")
	}

	if stat.Size() == 0 {
		return nil, errors.New("pack file is empty")
	}

	f, err := os.Open(packPath)
	if err != nil {
		return nil, errors.Wrap(err, "open pack file")
	}
	defer func() { _ = f.Close() }()

	var p pack.PackFile
	decoder := toml.NewDecoder(f)
	if err := decoder.Decode(&p); err != nil {
		return nil, errors.Wrap(err, "parse pack file failed")
	}

	if p.SchemaVersion == nil {
		return nil, errors.New("schema version is required")
	}

	if p.Name == "" {
		return nil, errors.New("name is required")
	}

	p.Location = filepath.Dir(packPath)

	return &p, nil
}

func locatePackFile(ctx context.Context, root string) (string, error) {
	if root == "/" {
		return "", errors.New("cannot search from filesystem root")
	}

	if len(strings.TrimSpace(root)) == 0 {
		return "", errors.New("root is empty")
	}

	// get the absolute path to the root
	root, err := filepath.Abs(root)
	if err != nil {
		return "", errors.Wrap(err, "failed to get absolute path to root")
	}

	// locate the pack file
	// if the root is a file, we take the containing directory of the file
	// then we check if the pack file exists in the root directory
	// if it does, we load it and return
	// if it doesn't, we walk up the directory tree
	// till we find one - if we reach the root and don't find it, we return an error
	info, err := os.Stat(root)
	if err != nil {
		return "", errors.Wrap(err, "failed to locate pack file")
	}

	// if the name is "sentrie.pack.toml", we use it
	if info.Name() == PackFileName {
		return root, nil
	}

	// the name is not "sentrie.pack.toml", we try to find it in the parent directory
	packFilePath := filepath.Join(root, PackFileName)

	// if we have a packfile here - we use it
	if _, err := os.Stat(packFilePath); err == nil {
		return packFilePath, nil
	}

	// otherwise, we walk up the directory tree till we find it or we reach root
	for {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		root = filepath.Dir(root)
		if root == "/" || (runtime.GOOS == "windows" && strings.HasSuffix(root, `:\` /* a drive letter */)) {
			break
		}
		if _, err := os.Stat(filepath.Join(root, PackFileName)); err == nil {
			return filepath.Join(root, PackFileName), nil
		}
	}

	return "", ErrPackFileNotFound
}
