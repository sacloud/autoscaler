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

package resources

import (
	"context"
	"fmt"

	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/core"
	"github.com/sacloud/autoscaler/defaults"
	sacloudotel "github.com/sacloud/autoscaler/otel"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/go-otelsetup"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/trace"
)

var Command = &cobra.Command{
	Use:   "resources [flags]...",
	Short: "list target resources",
	PreRunE: flags.ValidateMultiFunc(true,
		func(*cobra.Command, []string) error {
			return validate.Struct(param)
		},
		flags.ValidateStrictModeFlags,
	),
	RunE: run,
}

type parameter struct {
	ConfigPath string `name:"--config" validate:"required,file"`
}

var param = &parameter{
	ConfigPath: defaults.CoreConfigPath,
}

func init() {
	Command.Flags().StringVar(&param.ConfigPath, "config", param.ConfigPath, "File path of configuration of AutoScaler Core")
	flags.SetStrictModeFlag(Command)
}
func run(_ *cobra.Command, _ []string) error {
	ctx, span := sacloudotel.Tracer().Start(otelsetup.ContextForTrace(context.Background()), "commands/core/resources#run",
		trace.WithSpanKind(trace.SpanKindClient),
	)
	defer span.End()

	tree, err := core.ResourcesTree(ctx, "", param.ConfigPath, flags.StrictMode(), flags.NewLogger())
	if err != nil {
		return err
	}
	fmt.Println("\n" + tree + "\n")
	return nil
}
