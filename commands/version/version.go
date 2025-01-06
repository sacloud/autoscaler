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

package version

import (
	"fmt"

	"github.com/sacloud/autoscaler/version"
	"github.com/spf13/cobra"
)

var all bool

func init() {
	Command.Flags().BoolVarP(&all, "all", "a", false, "show full version info")
}

var Command = &cobra.Command{
	Use:   "version",
	Short: "show version",
	RunE: func(*cobra.Command, []string) error {
		v := "v" + version.Version
		if all {
			v = version.FullVersion()
		}
		fmt.Println(v)
		return nil
	},
}
