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

package flags

import (
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

type listenerFlags struct {
	ListenAddr string `name:"--addr" validate:"required,printascii"`
}

var listener = &listenerFlags{}

func SetListenerFlag(cmd *cobra.Command, defaultValue string) {
	listener.ListenAddr = defaultValue
	cmd.Flags().StringVarP(&listener.ListenAddr, "addr", "", listener.ListenAddr, "the address for the server to listen on")
}

func ValidateListenerFlags(*cobra.Command, []string) error {
	return validate.Struct(listener)
}

func ListenAddr() string {
	return listener.ListenAddr
}
