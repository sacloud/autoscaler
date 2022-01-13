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

package flags

import (
	"fmt"
	"os"
	"strings"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/grpcutil"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

type destinationFlags struct {
	Destination string `name:"--dest" validate:"omitempty,printascii"`
}

var (
	destination     = &destinationFlags{}
	destinationDesc = fmt.Sprintf(
		`Address of the gRPC endpoint of AutoScaler Core. 
If no value is specified, it will search for a valid value among the following values and use it.
[%s]`,
		strings.Join(defaults.CoreSocketAddrCandidates, ", "),
	)
)

func SetDestinationFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&destination.Destination, "dest", "", destination.Destination, destinationDesc)
}

func ValidateDestinationFlags(*cobra.Command, []string) error {
	if err := validate.Struct(destination); err != nil {
		return err
	}
	if defaultDestination() == "" {
		return fmt.Errorf(
			"--dest: Core's socket file is not found in [%s]",
			strings.Join(defaults.CoreSocketAddrCandidates, ", "),
		)
	}
	return nil
}

func Destination() string {
	if destination.Destination != "" {
		return destination.Destination
	}
	return defaultDestination()
}

func defaultDestination() string {
	for _, dest := range defaults.CoreSocketAddrCandidates {
		_, endpoint, err := grpcutil.ParseTarget(dest)
		if err != nil {
			panic(err) // defaultsでの定義誤り
		}
		if _, err := os.Stat(endpoint); err == nil {
			return dest
		}
	}
	return ""
}
