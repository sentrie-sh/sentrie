package cmd

import (
	"context"
	"log/slog"

	"github.com/binaek/cling"
)

func Setup(ctx context.Context, version string) *cling.CLI {
	cli := cling.NewCLI("sentrie", version).
		WithDescription("Sentrie is the policy enforcement engine").
		WithPreRun(func(ctx context.Context, args []string) error {
			slog.DebugContext(ctx, "==> Starting Sentrie", slog.String("version", version))
			return nil
		}).
		WithPostRun(func(ctx context.Context, args []string) error {
			slog.DebugContext(ctx, "==> Exiting Sentrie")
			return nil
		})

	addServeCmd(cli)

	return cli
}

func Execute(ctx context.Context, cli *cling.CLI, args []string) error {
	if cli == nil {
		panic("CLI cannot be NIL")
	}
	return cli.Run(ctx, args)
}
