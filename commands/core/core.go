// Copyright 2021-2022 The sacloud/autoscaler Authors
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

package core

import (
	"github.com/sacloud/autoscaler/commands/core/example"
	"github.com/sacloud/autoscaler/commands/core/handlers"
	"github.com/sacloud/autoscaler/commands/core/resources"
	"github.com/sacloud/autoscaler/commands/core/start"
	"github.com/sacloud/autoscaler/commands/core/validate"
	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:           "core",
	Aliases:       []string{"server"},
	Short:         "A set of sub commands to manage autoscaler's core server",
	SilenceErrors: true,

	// Note: 互換性維持用(since v0.5)
	// coreのサブコマンド群はroot直下に配置するが、互換性維持のためcoreサブコマンド配下にも配置する。
	// このときにcoreサブコマンド自体を非表示にするようにHiddenを設定しておく
	Hidden: true,
}

var subCommands = []*cobra.Command{
	example.Command,
	start.Command,
	handlers.Command,
	validate.Command,
	resources.Command,
}

func init() {
	AddSubCommandsTo(Command)
}

// AddSubCommandsTo 指定のコマンドに対しサブコマンドを登録する
func AddSubCommandsTo(cmd *cobra.Command) {
	cmd.AddCommand(subCommands...)
}
