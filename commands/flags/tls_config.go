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

type tlsConfigFlags struct {
	TLSConfig string `name:"--tls-config" validate:"omitempty,file"`
}

var tlsConfig = &tlsConfigFlags{
	TLSConfig: "",
}

func SetTLSConfigFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&tlsConfig.TLSConfig, "tls-config", "", tlsConfig.TLSConfig, "File path of TLS config")
}

func ValidateTLSConfigFlags(*cobra.Command, []string) error {
	return validate.Struct(tlsConfig)
}

func TLSConfig() string {
	return tlsConfig.TLSConfig
}
