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

package commands

import (
	"os"

	"github.com/sacloud/autoscaler/commands/completion"
	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/commands/inputs"
	"github.com/sacloud/autoscaler/commands/server"
	"github.com/sacloud/autoscaler/commands/version"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "autoscaler",
	Short:             "autoscaler is a tool for managing the scale of resources on SAKURA cloud",
	PersistentPreRunE: flags.ValidateLogFlags,
	SilenceUsage:      true,
	SilenceErrors:     false,
}

var subCommands = []*cobra.Command{
	completion.Command,
	inputs.Command,
	server.Command,
	version.Command,
}

func init() {
	flags.SetLogFlags(rootCmd)
	rootCmd.AddCommand(subCommands...)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
