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
	"github.com/spf13/cobra"
)

var (
	address    string
	configPath string
)

func init() {
	Command.Flags().StringVar(&address, "address", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")
	Command.Flags().StringVar(&configPath, "config", defaults.CoreConfigPath, "File path of configuration of AutoScaler Core")
}

var Command = &cobra.Command{
	Use:   "start [flags]...",
	Short: "start autoscaler's core server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		return core.Start(ctx, configPath, flags.NewLogger())
	},
}
