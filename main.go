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

package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/google/uuid"
	"github.com/sentrie-sh/sentrie/cmd"
	"github.com/sentrie-sh/sentrie/constants"
)

var version = "0.1.0"

func main() {
	ctx := context.Background()

	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	defer stop()

	// set an exit code
	exitCode := 0

	// setup logger
	logger := setupDefaultLogger()
	slog.SetDefault(logger)

	cli := cmd.Setup(ctx, version)
	if err := cmd.Execute(ctx, cli, os.Args); err != nil {
		// pretty print the error in the forn <red>Error</red>: <error>
		fmt.Printf("Error: %s\n", err)
		exitCode = 1
	}
	os.Exit(exitCode)
}

func setupDefaultLogger() *slog.Logger {
	logLevel := slog.LevelVar{}
	if _, ok := os.LookupEnv(constants.EnvDebug); ok {
		// force debug log if we are running in DEBUG mode
		os.Setenv(constants.EnvLogLevel, "DEBUG")
	}
	// set log level from env
	switch strings.ToUpper(os.Getenv(constants.EnvLogLevel)) { // DEBUG, INFO, WARN, ERROR
	case "DEBUG":
		logLevel.Set(slog.LevelDebug)
	case "INFO":
		logLevel.Set(slog.LevelInfo)
	case "WARN":
		logLevel.Set(slog.LevelWarn)
	case "ERROR":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}

	attrs := []slog.Attr{
		slog.String("app", "sentinel"),
		slog.String("version", version),

		// generate a unique instance id - so that we may track logs from a separate instances (if at all)
		slog.String("instance", uuid.NewString()),
	}
	if _, ok := os.LookupEnv(constants.EnvDebug); ok {
		attrs = append(
			attrs,
			slog.Bool("debug", true),
			slog.Any("args", os.Args),
		)
		if exec, err := os.Executable(); err == nil {
			attrs = append(attrs, slog.String("executable", exec))
		}
	}

	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     &logLevel,
	}).WithAttrs(attrs)

	logger := slog.New(logHandler)

	return logger
}
