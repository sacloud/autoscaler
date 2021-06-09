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

package core

import (
	"context"
	"time"

	"github.com/sacloud/autoscaler/log"
)

// Context 1リクエストのスコープに対応するコンテキスト、context.Contextを実装し、リクエスト情報や現在のジョブの情報を保持する
type Context struct {
	ctx     context.Context
	request *requestInfo
	job     *JobStatus
	logger  *log.Logger
}

func NewContext(parent context.Context, request *requestInfo, logger *log.Logger) *Context {
	logger = logger.With("request-type", request.requestType, "scaling-job-id", request.ID())
	return &Context{
		ctx:     parent,
		request: request,
		logger:  logger,
	}
}

// WithJobStatus JobStatusを持つContextを現在のContextを元に作成して返す
//
// 現在のContextが親Contextとなる
func (c *Context) WithJobStatus(job *JobStatus) *Context {
	return &Context{
		ctx: c,
		request: &requestInfo{
			requestType:       c.request.requestType,
			source:            c.request.source,
			action:            c.request.action,
			resourceGroupName: c.request.resourceGroupName,
			desiredStateName:  c.request.desiredStateName,
		},
		logger: c.logger,
		job:    job,
	}
}

// ForRefresh リフレッシュのためのContextを現在のContextを元に作成して返す
//
// 現在のContextが親Contextとなる
func (c *Context) ForRefresh() *Context {
	return NewContext(c, &requestInfo{
		requestType:       c.request.requestType,
		source:            c.request.source,
		action:            c.request.action,
		resourceGroupName: c.request.resourceGroupName,
		desiredStateName:  c.request.desiredStateName,
		refresh:           true,
	}, c.logger)
}

// Request 現在のコンテキストで受けたリクエストの情報を返す
func (c *Context) Request() *requestInfo {
	return c.request
}

// Logger 現在のコンテキストのロガーを返す
func (c *Context) Logger() *log.Logger {
	return c.logger
}

// JobID 現在のコンテキストでのJobのIDを返す
//
// まだJobの実行決定が行われていない場合でも値を返す
func (c *Context) JobID() string {
	return c.request.ID()
}

// Job 現在のコンテキストで実行中のJobを返す
//
// まだJobの実行決定が行われていない場合はnilを返す
func (c *Context) Job() *JobStatus {
	return c.job
}

func (c *Context) init() {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
}

// Deadline context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *Context) Deadline() (deadline time.Time, ok bool) {
	c.init()
	return c.ctx.Deadline()
}

// Done context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *Context) Done() <-chan struct{} {
	c.init()
	return c.ctx.Done()
}

// Err context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *Context) Err() error {
	c.init()
	return c.ctx.Err()
}

// Value context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *Context) Value(key interface{}) interface{} {
	c.init()
	return c.ctx.Value(key)
}
