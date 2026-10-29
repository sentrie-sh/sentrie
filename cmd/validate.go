// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"os"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
)

func addValidateCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("validate", validateCmd).
			WithArgument(cling.NewStringCmdInput("rule").
				WithDescription("Rule to execute").
				AsArgument(),
			).
			WithFlag(cling.
				NewStringCmdInput("pack-location").
				WithDefault(".").
				WithDescription("Pack directory to load").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("facts").
				WithDefault("{}").
				WithDescription("Facts to execute the rule with").
				AsFlag(),
			).
			WithFlag(capabilitiesFlag()),
	)
}

type validateCmdArgs struct {
	PackLocation string `cling-name:"pack-location"`
	Rule         string `cling-name:"rule"`
	Facts        string `cling-name:"facts"`
	Capabilities string `cling-name:"capabilities"`
}

func validateCmd(ctx context.Context, args []string) error {
	input := validateCmdArgs{}
	if err := cling.Hydrate(ctx, args, &input); err != nil {
		return err
	}

	pack, err := loader.LoadPack(ctx, input.PackLocation)
	if err != nil {
		return err
	}

	idx := index.CreateIndex()

	if err := idx.SetPack(ctx, pack); err != nil {
		return err
	}

	programs, err := loader.LoadPrograms(ctx, pack)
	if err != nil {
		return err
	}

	for _, program := range programs {
		if err := idx.AddProgram(ctx, program); err != nil {
			return err
		}
	}

	if err := idx.Validate(ctx); err != nil {
		return err
	}

	exec, err := newGrantedExecutor(ctx, idx, input.Capabilities)
	if err != nil {
		return err
	}
	if err := exec.Close(); err != nil {
		return err
	}

	return printNondeterminismWarnings(ctx, os.Stderr, idx)
}
