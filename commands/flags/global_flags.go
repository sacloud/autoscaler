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
	"os"

	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

type flags struct {
	LogLevel  string `name:"--log-level" validate:"required,oneof=error warn info debug"`
	LogFormat string `name:"--log-format" validate:"required,oneof=logfmt json"`
}

var global = &flags{
	LogLevel:  "info",
	LogFormat: "logfmt",
}

func SetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&global.LogLevel, "log-level", "", global.LogLevel, "Level of logging to be output. options: [ error | warn | info | debug ]")
	cmd.PersistentFlags().StringVarP(&global.LogFormat, "log-format", "", global.LogFormat, "Format of logging to be output. options: [ logfmt | json ]")
}

func ValidateFlags(cmd *cobra.Command, args []string) error {
	return validate.Struct(global)
}

func NewLogger() *log.Logger {
	return log.NewLogger(&log.LoggerOption{
		Writer:    os.Stderr,
		JSON:      global.LogFormat == "json",
		TimeStamp: true,
		Caller:    false,
		Level:     log.Level(global.LogLevel),
	})
}
