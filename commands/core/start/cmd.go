// Copyright 2021 The sacloud Authors
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
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "start [flags]...",
	Short: "start autoscaler's core server",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.Struct(param)
	},
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
}

func run(cmd *cobra.Command, args []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	return core.Start(ctx, param.ListenAddress, param.ConfigPath, flags.NewLogger())
}