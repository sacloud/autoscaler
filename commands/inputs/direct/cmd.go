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

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/request"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var Command = &cobra.Command{
	Use:       "direct {up | down} [flags]...",
	Short:     "Send Up/Down request directly to Core server",
	RunE:      run,
	ValidArgs: []string{"up", "down"},
	Args:      cobra.ExactValidArgs(1),
}

var (
	dest             string
	action           string
	group            string
	source           string
	desiredStateName string
)

func init() {
	Command.Flags().StringVarP(&dest, "dest", "", defaults.CoreSocketAddr, "URL of gRPC endpoint of AutoScaler Core")
	Command.Flags().StringVarP(&action, "action", "", defaults.ActionName, "Name of the action to perform")
	Command.Flags().StringVarP(&group, "group", "", defaults.ResourceGroupName, "Name of the target resource group")
	Command.Flags().StringVarP(&source, "source", "", defaults.SourceName, "A string representing the request source, passed to AutoScaler Core")
	Command.Flags().StringVarP(&desiredStateName, "desired-state-name", "", defaults.DesiredStateName, "Name of the desired state defined in Core's configuration file")
}

func run(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// TODO 簡易的な実装、後ほど整理&切り出し
	conn, err := grpc.DialContext(ctx, dest, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

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
		Source:            source,
		Action:            action,
		ResourceGroupName: group,
		DesiredStateName:  desiredStateName,
	})
	if err != nil {
		return err
	}

	fmt.Printf("status: %s, job-id: %s", res.Status, res.ScalingJobId)
	if res.Message != "" {
		fmt.Printf(", message: %s", res.Message)
	}
	fmt.Println()
	return nil
}
