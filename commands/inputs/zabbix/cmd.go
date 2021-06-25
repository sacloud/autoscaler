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

package zabbix

import (
	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/inputs"
	"github.com/sacloud/autoscaler/inputs/zabbix"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "zabbix",
	Short: "Start web server for handle webhooks from Zabbix",
	PreRunE: flags.ValidateMultiFunc(true,
		flags.ValidateDestinationFlags,
		flags.ValidateListenerFlags,
	),
	RunE: run,
}

func init() {
	flags.SetDestinationFlag(Command)
	flags.SetListenerFlag(Command)
}

func run(cmd *cobra.Command, args []string) error {
	return inputs.Serve(zabbix.NewInput(flags.Destination(), flags.ListenAddr(), flags.NewLogger()))
}