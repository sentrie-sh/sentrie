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

package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/binaek/cling"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/sentrie-sh/sentrie/pack"
)

func addInitCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("init", initCmd).
			WithFlag(cling.NewStringCmdInput("directory").WithDefault(".").WithDescription("The directory to initialize in MUST be empty.").AsFlag()).
			WithArgument(cling.NewStringCmdInput("name").WithDescription("The name of the pack.").AsArgument()),
	)
}

type initCmdArgs struct {
	Directory string `cling-name:"directory"`
	Name      string `cling-name:"name"`
}

func initCmd(ctx context.Context, args []string) error {
	input := initCmdArgs{}
	if err := cling.Hydrate(ctx, args, &input); err != nil {
		return err
	}

	packFile := pack.NewPackFile(input.Name)

	stat, err := os.Stat(input.Directory)
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		return errors.New("directory is not a directory")
	}

	// if the directory is not empty, we return an error
	entries, err := os.ReadDir(input.Directory)
	if err != nil {
		return errors.Wrapf(err, "could not read directory")
	}
	if len(entries) > 0 {
		return errors.New("directory is not empty - please choose a different directory")
	}

	f, err := os.OpenFile(filepath.Join(input.Directory, loader.PackFileName), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return errors.Wrapf(err, "could not create pack file")
	}
	defer func() { _ = f.Close() }()

	encoder := toml.NewEncoder(f)
	encoder.SetTablesInline(true)
	if err := encoder.Encode(packFile); err != nil {
		return errors.Wrapf(err, "could not encode pack file")
	}

	return nil
}
