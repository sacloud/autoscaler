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

package main

import (
	"context"
	"log"

	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/commands/version"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/example"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "autoscaler-handlers-example",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: flags.ValidateMultiFunc(true,
		flags.ValidateLogFlags,
		flags.ValidateListenerFlags,
		flags.ValidateInputsConfigFlags,
	),
}

var serveCmd = &cobra.Command{
	Use:           "serve [flags]...",
	Short:         "start example handler",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.NoArgs,
	PreRunE: flags.ValidateMultiFunc(true,
		flags.ValidateListenerFlags,
		flags.ValidateInputsConfigFlags,
	),
	RunE: run,
}

func init() {
	flags.SetLogFlags(rootCmd)

	flags.SetListenerFlag(serveCmd, "unix:example-handler.sock")
	flags.SetInputsConfigFlag(serveCmd)

	rootCmd.AddCommand(
		serveCmd,
		version.Command,
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(_ *cobra.Command, _ []string) error {
	return handlers.Serve(context.Background(), example.NewHandler(flags.ListenAddr(), flags.InputsConfig(), flags.NewLogger()))
}
