// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/api"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/sentrie-sh/sentrie/runtime"
)

func addServeCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("serve", serveCmd).
			WithFlag(cling.
				NewIntCmdInput("http-port").
				WithDefault(7529 /* PLCY - keypad */).
				WithDescription("HTTP port to listen on").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("pack-location").
				WithDefault("./").
				WithDescription("Pack directory to serve").
				AsFlag(),
			).
			WithFlag(cling.
				NewCmdSliceInput[string]("http-listen").
				WithDefault([]string{"local"}).
				WithDescription("HTTP address(es) to listen on").
				AsFlag(),
			).
			WithFlag(cling.
				NewIntCmdInput("http-max-body-bytes").
				WithDefault(int(api.DefaultMaxBodyBytes)).
				WithDescription("Largest request body to accept, in bytes; a larger one is answered with 413. 0 disables the limit").
				AsFlag(),
			).
			WithFlag(evalTimeoutFlag()).
			WithFlag(evalMaxTimeoutFlag()).
			WithFlag(capabilitiesFlag()).
			WithFlag(skipEngineCheckFlag()),
	)
}

type serveCmdArgs struct {
	Port            int      `cling-name:"http-port"`
	PackLocation    string   `cling-name:"pack-location"`
	Listen          []string `cling-name:"http-listen"`
	MaxBodyBytes    int      `cling-name:"http-max-body-bytes"`
	EvalTimeout     string   `cling-name:"eval-timeout"`
	EvalMaxTimeout  string   `cling-name:"eval-max-timeout"`
	Capabilities    string   `cling-name:"capabilities"`
	SkipEngineCheck bool     `cling-name:"skip-engine-check"`
}

// validateMaxBodyBytes checks an `--http-max-body-bytes` value (#115): a byte
// count, where 0 disables the limit.
func validateMaxBodyBytes(n int) error {
	if n < 0 {
		return fmt.Errorf("invalid --http-max-body-bytes %d: must not be negative; use 0 for no limit", n)
	}
	return nil
}

func serveCmd(ctx context.Context, args []string) error {
	input := serveCmdArgs{}
	if err := cling.Hydrate(ctx, args, &input); err != nil {
		return err
	}

	if err := validateMaxBodyBytes(input.MaxBodyBytes); err != nil {
		return err
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
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := idx.AddProgram(ctx, program); err != nil {
			return err
		}
	}

	if err := idx.Validate(ctx); err != nil {
		return err
	}

	evalTimeout, err := parseEvalTimeout(input.EvalTimeout)
	if err != nil {
		return err
	}
	evalMaxTimeout, err := parseEvalMaxTimeout(input.EvalMaxTimeout, evalTimeout)
	if err != nil {
		return err
	}

	// The API hands every request its deadline (the default, or Sentrie-Timeout
	// within the ceiling), which replaces the executor's own: one deadline per
	// request.
	exec, err := newGrantedExecutor(ctx, idx, input.Capabilities, runtime.WithEvalTimeout(evalTimeout))
	if err != nil {
		return err
	}
	defer func() { _ = exec.Close() }()

	server := api.NewHTTPAPI(exec,
		api.WithRequestTimeout(evalTimeout),
		api.WithMaxRequestTimeout(evalMaxTimeout),
		api.WithMaxBodyBytes(int64(input.MaxBodyBytes)),
	)
	if err := server.Setup(ctx, input.Port, input.Listen); err != nil {
		return fmt.Errorf("invalid or unusable --http-listen/--http-port: %w", err)
	}

	go func() {
		server.StartServer(ctx, input.Port, input.Listen)
	}()

	<-ctx.Done()

	// ctx is already cancelled, which is what ended the wait, so the shutdown
	// gets its own bounded context: in-flight requests finish within the grace
	// period, then the connections are closed.
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), api.DefaultShutdownGracePeriod)
	defer cancel()
	slog.InfoContext(shutdownCtx, "Shutting down", "grace_period", api.DefaultShutdownGracePeriod.String())

	return server.StopServer(shutdownCtx)
}
