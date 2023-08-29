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

package config

import (
	"context"
	"log/slog"
	"time"

	"github.com/sacloud/autoscaler/log"
)

var (
	_ LoadConfigHolder = &LoadConfigContext{}
	_ LoggerHolder     = &LoadConfigContext{}
	_ context.Context  = &LoadConfigContext{}
)

// LoadConfigContext コンフィグのロードオプションを保持するcontext.Context実装
type LoadConfigContext struct {
	parent context.Context
	strict bool
	logger *slog.Logger
}

// LoadConfigHolder コンフィグのロードオプションを保持しているかを示すインターフェース
type LoadConfigHolder interface {
	StrictMode() bool
}

type LoggerHolder interface {
	Logger() *slog.Logger
}

func NewLoadConfigContext(ctx context.Context, strict bool, logger *slog.Logger) context.Context {
	if logger == nil {
		logger = log.NewLogger(nil)
	}
	return &LoadConfigContext{parent: ctx, strict: strict, logger: logger}
}

func (c *LoadConfigContext) StrictMode() bool {
	return c.strict
}

func (c *LoadConfigContext) Logger() *slog.Logger {
	return c.logger
}

// Deadline context.Context実装
func (c *LoadConfigContext) Deadline() (time.Time, bool) {
	return c.parent.Deadline()
}

// Done context.Context実装
func (c *LoadConfigContext) Done() <-chan struct{} {
	return c.parent.Done()
}

// Err context.Context実装
func (c *LoadConfigContext) Err() error {
	return c.parent.Err()
}

// Value context.Context実装
func (c *LoadConfigContext) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}
