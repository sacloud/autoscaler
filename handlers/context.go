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

package handlers

import (
	"context"
	"time"

	"github.com/sacloud/autoscaler/handler"
)

type HandlerContext struct {
	ctx          context.Context
	scalingJobID string
	sender       ResponseSender
	reporter     reportFn
}

func NewHandlerContext(scalingJobID string, sender ResponseSender) *HandlerContext {
	return &HandlerContext{
		ctx:          context.Background(),
		scalingJobID: scalingJobID,
		sender:       sender,
		reporter:     reporter(sender, scalingJobID),
	}
}

func (c *HandlerContext) Report(status handler.HandleResponse_Status, formatAndArgs ...interface{}) error {
	return c.reporter(status, formatAndArgs...)
}

func (c *HandlerContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *HandlerContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *HandlerContext) Err() error {
	return c.ctx.Err()
}

func (c *HandlerContext) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}
