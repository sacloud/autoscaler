// Copyright 2021-2023 The sacloud/autoscaler Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package start

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/core"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/go-otelsetup"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "start [flags]...",
	Short: "start autoscaler's core server",
	PreRunE: flags.ValidateMultiFunc(true,
		func(*cobra.Command, []string) error {
			return validate.Struct(param)
		},
		flags.ValidateStrictModeFlags,
	),
	RunE: run,
}

type parameter struct {
	ListenAddress string `name:"--addr" validate:"required"`
	ConfigPath    string `name:"--config" validate:"required,file"`
}

var param = &parameter{
	ListenAddress: defaults.CoreSocketAddr,
	ConfigPath:    defaults.CoreConfigPath,
}

func init() {
	Command.Flags().StringVar(&param.ListenAddress, "addr", param.ListenAddress, "Address of the gRPC endpoint to listen to")
	Command.Flags().StringVar(&param.ConfigPath, "config", param.ConfigPath, "File path of configuration of AutoScaler Core")
	flags.SetStrictModeFlag(Command)
}

func run(*cobra.Command, []string) error {
	ctx, shutdown := context.WithCancel(otelsetup.ContextForTrace(context.Background()))
	defer shutdown()

	logger := flags.NewLogger()
	coreInstance, err := core.New(ctx, param.ListenAddress, param.ConfigPath, flags.StrictMode(), logger)
	if err != nil {
		return err
	}

	// シグナルを受け取った際のgraceful shutdown
	signalCtx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-signalCtx.Done()
		if ctx.Err() != nil {
			logger.Info("signal received. waiting for shutdown...")
		}
		if err := coreInstance.Stop(); err != nil {
			logger.Error(err.Error())
		}
		shutdown()
	}()

	// 本体(core)の起動
	errChan := make(chan error)
	go func() {
		errChan <- coreInstance.Run(ctx)
	}()

	return <-errChan
}
