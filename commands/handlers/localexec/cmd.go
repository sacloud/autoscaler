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

package localexec

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/localexec"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "local-exec [flags]...",
	Short: "Handler for executing a shell script",
	Args:  cobra.NoArgs,
	PreRunE: flags.ValidateMultiFunc(true,
		func(*cobra.Command, []string) error {
			return validate.Struct(param)
		},
	),
	RunE: run,
}

type parameter struct {
	ExecutablePath string `name:"--executable-path" validate:"required,file"`
	HandlerType    string `name:"--handler-type" validate:"required,oneof=pre-handle handle post-handle"`
	ListenAddr     string `name:"--addr" validate:"required,printascii"`
	Config         string `name:"--config" validate:"omitempty,file"`
}

var param = &parameter{
	ListenAddr: "unix:autoscaler-handlers-local-exec.sock",
}

func init() {
	Command.Flags().StringVarP(&param.ListenAddr, "addr", "", param.ListenAddr, "the address for the server to listen on")
	Command.Flags().StringVarP(&param.ExecutablePath, "executable-path", "", param.ExecutablePath, "Path to the executable")
	Command.Flags().StringVarP(&param.HandlerType, "handler-type", "", param.HandlerType, "Handler type name")
	Command.Flags().StringVarP(&param.Config, "config", "", param.Config, "Filepath to Handlers additional configuration file")
}

func run(*cobra.Command, []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	handler := localexec.NewHandler(param.ListenAddr, param.Config, param.ExecutablePath, param.HandlerType, flags.NewLogger())
	return handlers.Serve(ctx, handler)
}
