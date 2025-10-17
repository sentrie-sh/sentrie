package cmd

import (
	"context"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/sentrie-sh/sentrie/runtime"
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
			),
	)
}

type validateCmdArgs struct {
	PackLocation string `cling-name:"pack-location"`
	Rule         string `cling-name:"rule"`
	Facts        string `cling-name:"facts"`
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

	_, err = runtime.NewExecutor(idx)
	return err
}
