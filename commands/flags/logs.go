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

package flags

import (
	"log/slog"
	"os"

	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/validate"
	"github.com/spf13/cobra"
)

type logFlags struct {
	LogLevel  string `name:"--log-level" validate:"required,oneof=error warn info debug"`
	LogFormat string `name:"--log-format" validate:"required,oneof=logfmt json"`
}

func (l *logFlags) slogLevel() slog.Level {
	switch l.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

var logs = &logFlags{
	LogLevel:  "info",
	LogFormat: "logfmt",
}

func SetLogFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&logs.LogLevel, "log-level", "", logs.LogLevel, "Level of logging to be output. options: [ error | warn | info | debug ]")
	cmd.PersistentFlags().StringVarP(&logs.LogFormat, "log-format", "", logs.LogFormat, "Format of logging to be output. options: [ logfmt | json ]")
}

func ValidateLogFlags(*cobra.Command, []string) error {
	return validate.Struct(logs)
}

func NewLogger() *slog.Logger {
	return log.NewLogger(&log.LoggerOption{
		Writer:    os.Stderr,
		JSON:      logs.LogFormat == "json",
		TimeStamp: true,
		Caller:    false,
		Level:     logs.slogLevel(),
	})
}
