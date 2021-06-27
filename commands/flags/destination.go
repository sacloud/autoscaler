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

package flags

import (
	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

type destinationFlags struct {
	Destination string `name:"--dest" validate:"required,printascii"`
}

var destination = &destinationFlags{
	Destination: defaults.CoreSocketAddr,
}

func SetDestinationFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&destination.Destination, "dest", "", destination.Destination, "Address of the gRPC endpoint of AutoScaler Core")
}

func ValidateDestinationFlags(*cobra.Command, []string) error {
	return validate.Struct(destination)
}

func Destination() string {
	return destination.Destination
}
