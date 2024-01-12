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

package commands

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/sacloud/autoscaler/commands/completion"
	"github.com/sacloud/autoscaler/commands/core"
	"github.com/sacloud/autoscaler/commands/flags"
	"github.com/sacloud/autoscaler/commands/handlers"
	"github.com/sacloud/autoscaler/commands/inputs"
	cmdVersion "github.com/sacloud/autoscaler/commands/version"
	sacloudotel "github.com/sacloud/autoscaler/otel"
	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/go-otelsetup"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "autoscaler",
	Short:             "autoscaler is a tool for managing the scale of resources on SAKURA cloud",
	PersistentPreRunE: flags.ValidateLogFlags,
	SilenceUsage:      true,
	SilenceErrors:     false,
	Version:           "v" + version.Version,
}

var subCommands = []*cobra.Command{
	completion.Command,
	inputs.Command,
	core.Command, // Memo: 互換性維持用(since: v0.5) v1.0リリース時に除去する
	cmdVersion.Command,
	handlers.Command,
}

func init() {
	flags.SetLogFlags(rootCmd)
	rootCmd.SetVersionTemplate("{{.Version}}\n")
	rootCmd.AddCommand(subCommands...)
	// coreコマンドをrootCmd直下に追加
	core.AddSubCommandsTo(rootCmd)
}

func Execute() {
	if err := execute(); err != nil {
		os.Exit(1)
	}
}

func execute() error {
	// initialize otel SDK
	otelShutdown, err := otelsetup.Init(context.Background(), sacloudotel.AppName, version.Version)
	if err != nil {
		log.Println("Error in initializing OTel SDK: " + err.Error())
		return err
	}
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
		if err != nil {
			log.Println("Error in initializing OTel SDK: " + err.Error())
		}
	}()

	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
