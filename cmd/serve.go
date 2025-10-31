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

	"github.com/binaek/cling"
	"github.com/sentrie-sh/sentrie/api"
	"github.com/sentrie-sh/sentrie/constants"
	"github.com/sentrie-sh/sentrie/index"
	"github.com/sentrie-sh/sentrie/loader"
	"github.com/sentrie-sh/sentrie/otel"
	"github.com/sentrie-sh/sentrie/runtime"
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
			).
			WithFlag(
				cling.NewBoolCmdInput("otel-enabled").
					WithDefault(false).
					WithDescription("Enable OpenTelemetry tracing").
					AsFlag().
					FromEnv([]string{constants.EnvOtelEnabled}),
			).
			WithFlag(
				cling.NewStringCmdInput("otel-endpoint").
					WithDefault("http://localhost:4317").
					WithDescription("OpenTelemetry endpoint to send traces to").
					AsFlag().
					FromEnv([]string{constants.EnvOtelEndpoint}),
			).
			WithFlag(
				cling.NewStringCmdInput("otel-protocol").
					WithDefault("grpc").
					WithValidator(cling.NewEnumValidator("http", "grpc")).
					WithDescription("OpenTelemetry protocol. Allowed values: http, grpc.").
					AsFlag().
					FromEnv([]string{constants.EnvOtelProtocol}),
			).
			WithFlag(
				cling.NewBoolCmdInput("otel-trace-execution").
					WithDefault(false).
					WithDescription("Enable OpenTelemetry tracing for detailed policy execution.").
					AsFlag().
					FromEnv([]string{constants.EnvOtelTraceExecution}),
			),
	)
}

type serveCmdArgs struct {
	Port               int      `cling-name:"port"`
	PackLocation       string   `cling-name:"pack-location"`
	Listen             []string `cling-name:"listen"`
	OtelEnabled        bool     `cling-name:"otel-enabled"`
	OtelEndpoint       string   `cling-name:"otel-endpoint"`
	OtelProtocol       string   `cling-name:"otel-protocol"`
	OtelTraceExecution bool     `cling-name:"otel-trace-execution"`
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

	// Initialize OpenTelemetry if enabled
	var otelCleanup otel.ShutdownFn
	otelConfig := otel.OTelConfig{
		Enabled:        input.OtelEnabled,
		Endpoint:       input.OtelEndpoint,
		Protocol:       input.OtelProtocol,
		ServiceName:    constants.APPNAME,
		ServiceVersion: constants.APPVERSION,
		PackName:       pack.Name,
		TraceExecution: input.OtelEnabled && input.OtelTraceExecution,
	}

	if otelConfig.Enabled {
		otelCleanup, err = otel.InitProvider(ctx, otelConfig)
		if err != nil {
			return err
		}

		defer func() {
			if otelCleanup != nil {
				_ = otelCleanup(context.WithoutCancel(ctx))
			}
		}()
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

	// Create executor with OpenTelemetry options
	exec, err := runtime.NewExecutor(
		idx,
		runtime.WithOTelConfig(&otelConfig),
	)
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
