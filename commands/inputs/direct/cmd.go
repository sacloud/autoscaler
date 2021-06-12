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

package direct

import (
	"context"
	"fmt"

	"github.com/sacloud/autoscaler/commands/flags"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/request"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var Command = &cobra.Command{
	Use:       "direct {up | down} [flags]...",
	Short:     "Send Up/Down request directly to Core server",
	ValidArgs: []string{"up", "down"},
	Args:      cobra.ExactValidArgs(1),
	PreRunE: flags.ValidateMultiFunc(true,
		flags.ValidateDestinationFlags,
		func(cmd *cobra.Command, args []string) error {
			return validate.Struct(param)
		},
	),
	RunE: run,
}

type parameter struct {
	Source            string `name:"--source" validate:"required,printascii,max=1024"`
	Action            string `name:"--action" validate:"required,printascii,max=1024"`
	ResourceGroupname string `name:"--resource-group-name" validate:"required,printascii,max=1024"`
	DesiredStateName  string `name:"--desired-state-name" validate:"omitempty,printascii,max=1024"`
}

var param = &parameter{
	Source:            defaults.SourceName,
	Action:            defaults.ActionName,
	ResourceGroupname: defaults.ResourceGroupName,
	DesiredStateName:  "",
}

func init() {
	flags.SetDestinationFlag(Command)
	Command.Flags().StringVarP(&param.Action, "action", "", param.Action, "Name of the action to perform")
	Command.Flags().StringVarP(&param.ResourceGroupname, "resource-group-name", "", param.ResourceGroupname, "Name of the target resource group")
	Command.Flags().StringVarP(&param.Source, "source", "", param.Source, "A string representing the request source, passed to AutoScaler Core")
	Command.Flags().StringVarP(&param.DesiredStateName, "desired-state-name", "", param.DesiredStateName, "Name of the desired state defined in Core's configuration file")
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	conn, cleanup, err := grpcutil.DialContext(ctx, &grpcutil.DialOption{Destination: flags.Destination()})
	if err != nil {
		return err
	}
	defer cleanup()

	req := request.NewScalingServiceClient(conn)
	var f func(ctx context.Context, in *request.ScalingRequest, opts ...grpc.CallOption) (*request.ScalingResponse, error)

	switch args[0] {
	case "up":
		f = req.Up
	case "down":
		f = req.Down
	default:
		return fmt.Errorf("invalid args: %v", args)
	}
	res, err := f(ctx, &request.ScalingRequest{
		Source:            param.Source,
		Action:            param.Action,
		ResourceGroupName: param.ResourceGroupname,
		DesiredStateName:  param.DesiredStateName,
	})
	if err != nil {
		return err
	}

	// 単発の出力のためlog(標準エラー)ではなく標準出力に書く
	fmt.Printf("status: %s, job-id: %s", res.Status, res.ScalingJobId)
	if res.Message != "" {
		fmt.Printf(", message: %s", res.Message)
	}
	fmt.Println()
	return nil
}
