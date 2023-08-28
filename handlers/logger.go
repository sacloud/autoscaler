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

package handlers

import (
	"log/slog"

	"github.com/sacloud/autoscaler/log"
)

type HandlerLogger struct {
	Logger *slog.Logger
}

func (l *HandlerLogger) GetLogger() *slog.Logger {
	if l.Logger == nil {
		l.Logger = log.NewLogger(&log.LoggerOption{
			Writer:    nil,
			JSON:      false,
			TimeStamp: true,
			Caller:    false,
			Level:     slog.LevelInfo,
		})
	}
	return l.Logger
}

func (l *HandlerLogger) SetLogger(logger *slog.Logger) {
	l.Logger = logger
}
