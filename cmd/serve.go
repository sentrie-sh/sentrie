package cmd

import (
	"context"

	"github.com/binaek/cling"
	"github.com/binaek/sentra/api"
	"github.com/binaek/sentra/index"
	"github.com/binaek/sentra/loader"
	"github.com/binaek/sentra/runtime"
)

func addServeCmd(cli *cling.CLI) {
	cli.WithCommand(
		cling.NewCommand("serve", serveCmd).
			WithFlag(cling.
				NewIntCmdInput("port").
				WithDefault(7529 /* PLCY - keypad */).
				WithDescription("Port to listen on").
				AsFlag(),
			).
			WithFlag(cling.
				NewStringCmdInput("pack-location").
				WithDefault("./").
				WithDescription("Pack directory to serve").
				AsFlag(),
			).
			WithFlag(cling.
				NewCmdSliceInput[string]("listen").
				WithDefault([]string{"local"}).
				WithDescription("Address(es) to listen on").
				AsFlag(),
			),
	)
}

type serveCmdArgs struct {
	Port         int      `cling-name:"port"`
	PackLocation string   `cling-name:"pack-location"`
	Listen       []string `cling-name:"listen"`
}

func serveCmd(ctx context.Context, args []string) error {
	input := serveCmdArgs{}
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

	exec, err := runtime.NewExecutor(idx)
	if err != nil {
		return err
	}

	server := api.NewHTTPAPI(exec)
	if err := server.Setup(ctx, input.Port, input.Listen); err != nil {
		return err
	}

	go func() {
		server.StartServer(ctx, input.Port, input.Listen)
	}()

	<-ctx.Done()

	return server.StopServer(ctx)
}
