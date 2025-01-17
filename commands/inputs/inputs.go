// Copyright 2021-2025 The sacloud/autoscaler Authors
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

package inputs

import (
	"github.com/sacloud/autoscaler/commands/inputs/alertmanager"
	"github.com/sacloud/autoscaler/commands/inputs/direct"
	"github.com/sacloud/autoscaler/commands/inputs/grafana"
	"github.com/sacloud/autoscaler/commands/inputs/webhook"
	"github.com/sacloud/autoscaler/commands/inputs/zabbix"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:           "inputs",
	Short:         "A set of sub commands to manage autoscaler's inputs",
	SilenceErrors: true,
}

var subCommands = []*cobra.Command{
	alertmanager.Command,
	direct.Command,
	grafana.Command,
	webhook.Command,
	zabbix.Command,
}

func init() {
	Command.AddCommand(subCommands...)
}
