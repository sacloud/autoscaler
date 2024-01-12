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

package direct

import (
	"context"
	"fmt"
	"os"

	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	sacloudotel "github.com/sacloud/autoscaler/otel"
	"github.com/sacloud/autoscaler/request"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/go-otelsetup"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
			if err := validate.Struct(param); err != nil {
				return err
			}
			return flags.ValidateInputsConfigFlags(cmd, args)
		},
	),
	RunE: run,
}

const (
	// ExitCodeDoneWithNoop up/downリクエストを受け取ったが処理なしの場合(JOB_DONE_NOOP)
	ExitCodeDoneWithNoop = 129
	// ExitCodeUnacceptableState up/downリクエストが受け入れられない状態の場合(処理の実行中やcooldown期間中、シャットダウン中の場合など)
	ExitCodeUnacceptableState = 130
)

type parameter struct {
	Source           string `name:"--source" validate:"required,printascii,max=1024"`
	ResourceName     string `name:"--resource-name" validate:"required,printascii,max=1024"`
	DesiredStateName string `name:"--desired-state-name" validate:"omitempty,printascii,max=1024"`
	Sync             bool   `name:"--sync"`
}

var param = &parameter{
	Source:           defaults.SourceName,
	ResourceName:     defaults.ResourceName,
	DesiredStateName: "",
	Sync:             false,
}

func init() {
	flags.SetDestinationFlag(Command)
	flags.SetInputsConfigFlag(Command)
	Command.Flags().StringVarP(&param.ResourceName, "resource-name", "", param.ResourceName, "Name of the target resource")
	Command.Flags().StringVarP(&param.Source, "source", "", param.Source, "A string representing the request source, passed to AutoScaler Core")
	Command.Flags().StringVarP(&param.DesiredStateName, "desired-state-name", "", param.DesiredStateName, "Name of the desired state defined in Core's configuration file")
	Command.Flags().BoolVarP(&param.Sync, "sync", "", param.Sync, "Flag for synchronous handling")
}

func run(_ *cobra.Command, args []string) error {
	var exitCode int
	ctx, span := sacloudotel.Tracer().Start(otelsetup.ContextForTrace(context.Background()), "inputs.direct",
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attribute.StringSlice("args", args)),
	)

	opts := &grpcutil.DialOption{
		Destination: flags.Destination(),
	}

	conn, cleanup, err := grpcutil.DialContext(ctx, opts)
	if err != nil {
		return err
	}
	defer func() {
		cleanup()
		if exitCode != 0 {
			span.End()
			os.Exit(exitCode)
		}
	}()

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
		Source:           param.Source,
		ResourceName:     param.ResourceName,
		DesiredStateName: param.DesiredStateName,
		Sync:             param.Sync,
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

	// 何らかの理由で処理されなかった場合の終了コードを設定
	switch {
	case res.Status == request.ScalingJobStatus_JOB_DONE_NOOP:
		exitCode = ExitCodeDoneWithNoop
	case res.Message != "":
		exitCode = ExitCodeUnacceptableState
	}

	span.End()
	return nil
}
