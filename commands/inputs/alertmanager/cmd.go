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

package alertmanager

import (
	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/inputs"
	"github.com/sacloud/autoscaler/inputs/alertmanager"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "alertmanager",
	Short: "Start web server for handle webhooks from AlertManager",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return validate.Struct(param)
	},
	RunE: run,
}

type parameter struct {
	Destination string `name:"--dest" validate:"required"`
	ListenAddr  string `name:"--addr" validate:"required"`
}

var param = &parameter{
	Destination: defaults.CoreSocketAddr,
	ListenAddr:  ":3001",
}

func init() {
	Command.Flags().StringVarP(&param.Destination, "dest", "", param.Destination, "URL of gRPC endpoint of AutoScaler Core")
	Command.Flags().StringVarP(&param.ListenAddr, "addr", "", param.ListenAddr, "the TCP address for the server to listen on")
}

func run(cmd *cobra.Command, args []string) error {
	return inputs.Serve(alertmanager.NewInput(param.Destination, param.ListenAddr, flags.NewLogger()))
}
