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

package webhook

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/inputs"
	"github.com/sacloud/autoscaler/inputs/webhook"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "webhook",
	Short: "Start web server for handle webhooks with command",
	PreRunE: flags.ValidateMultiFunc(true,
		flags.ValidateDestinationFlags,
		flags.ValidateListenerFlags,
		flags.ValidateInputsConfigFlags,
		func(*cobra.Command, []string) error {
			return validate.Struct(param)
		},
	),
	RunE: run,
}

type parameter struct {
	AcceptHTTPMethods []string `name:"--accept-http-methods" validate:"required,dive,oneof=GET POST PUT DELETE HEAD"`
	ExecutablePath    string   `name:"--executable-path" validate:"required,file"`
}

var param = &parameter{
	AcceptHTTPMethods: []string{http.MethodPost, http.MethodPut},
}

func init() {
	flags.SetDestinationFlag(Command)
	flags.SetInputsConfigFlag(Command)
	flags.SetListenerFlag(Command, defaults.ListenAddress)

	Command.Flags().StringSliceVarP(&param.AcceptHTTPMethods, "accept-http-methods", "", param.AcceptHTTPMethods, "List of HTTP methods to accept")
	Command.Flags().StringVarP(&param.ExecutablePath, "executable-path", "", param.ExecutablePath, "Path to the executable to determine if webhooks should be accepted")
}

func run(*cobra.Command, []string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	in, err := webhook.NewInput(flags.Destination(), flags.ListenAddr(), flags.InputsConfig(), flags.NewLogger(), param.AcceptHTTPMethods, param.ExecutablePath)
	if err != nil {
		return err
	}
	return inputs.Serve(ctx, in)
}
