// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
)

func addValidateCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("validate", validateCmd).
			WithDescription("Check a pack without evaluating it: sentrie validate [--pack-location <dir>]").
			WithLongDescription(validateLongDescription).
			WithFlag(cling.
				NewStringCmdInput("pack-location").
				WithDefault(".").
				WithDescription("Pack directory to load").
				AsFlag(),
			).
			WithFlag(capabilitiesFlag()).
			WithFlag(skipEngineCheckFlag()),
	)
}

const validateLongDescription = `Loads the pack, indexes every policy and runs the static checks, without
evaluating any rule: parse errors, unknown names, cycles and self-references,
type and shape names, call targets and builtin calls, constraint arguments,
pack data and the capabilities grant. It exits non-zero on the first failing
check and prints nondeterminism warnings on stderr. It takes no arguments and
no facts; use sentrie exec to evaluate a rule.
`

type validateCmdArgs struct {
	PackLocation    string `cling-name:"pack-location"`
	Capabilities    string `cling-name:"capabilities"`
	SkipEngineCheck bool   `cling-name:"skip-engine-check"`
}

func validateCmd(ctx context.Context, args []string) error {
	input := validateCmdArgs{}
	if err := cling.Hydrate(ctx, args, &input); err != nil {
		return err
	}
	// validate checks the whole pack; it used to accept a rule target and
	// facts and ignore both (#155). cling hands the action its positionals
	// followed by every flag as --name=value, and binds positionals without
	// complaint; a single-dash word such as -x is a positional to cling, so
	// only a "--" prefix marks a flag.
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			continue
		}
		if strings.HasPrefix(a, "-") {
			return fmt.Errorf("validate: unknown flag %q: flags take two dashes, as in --pack-location", a)
		}
		return fmt.Errorf("validate takes no arguments, got %q: it checks the whole pack given by --pack-location", a)
	}

	pack, err := loader.LoadPack(ctx, input.PackLocation)
	if err != nil {
		return err
	}
	if err := checkEngine(pack, input.SkipEngineCheck, os.Stderr); err != nil {
		return err
	}
	if err := printUnenforcedPermissionWarnings(os.Stderr, pack); err != nil {
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
