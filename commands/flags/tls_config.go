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
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

type inputsTLSConfigFlags struct {
	TLSConfig string `name:"--tls-config" validate:"omitempty,file"`
}

var inputsTLSConfig = &inputsTLSConfigFlags{
	TLSConfig: "",
}

func SetInputsTLSConfigFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&inputsTLSConfig.TLSConfig, "tls-config", "", inputsTLSConfig.TLSConfig, "File path of input server TLS config")
}

func ValidateInputsTLSConfigFlags(cmd *cobra.Command, args []string) error {
	return validate.Struct(inputsTLSConfig)
}

func InputsTLSConfig() string {
	return inputsTLSConfig.TLSConfig
}
