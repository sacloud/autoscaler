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

package main

import (
	"context"
	"log"

	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/commands/version"
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/example"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "autoscaler-handlers-example",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: flags.ValidateMultiFunc(true,
		flags.ValidateLogFlags,
		func(cmd *cobra.Command, args []string) error {
			return validate.Struct(param)
		},
	),
}

var serveCmd = &cobra.Command{
	Use:           "serve [flags]...",
	Short:         "start example handler",
	SilenceErrors: true,
	SilenceUsage:  true,
	Args:          cobra.NoArgs,
	RunE:          run,
}

func init() {
	flags.SetLogFlags(rootCmd)
	serveCmd.Flags().StringVar(&param.ListenAddr, "addr", param.ListenAddr, "Address of the gRPC endpoint to listen to")

	rootCmd.AddCommand(
		serveCmd,
		version.Command,
	)
}

type parameter struct {
	ListenAddr string
}

var param = &parameter{
	ListenAddr: "unix:autoscaler-handlers-example.sock",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(_ *cobra.Command, _ []string) error {
	return handlers.Serve(context.Background(), example.NewHandler(param.ListenAddr, flags.NewLogger()))
}
