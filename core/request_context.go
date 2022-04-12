// Copyright 2021-2022 The sacloud/autoscaler Authors
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

	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/log"
)

// RequestContext 1リクエストのスコープに対応するコンテキスト、context.Contextを実装し、リクエスト情報や現在のジョブの情報を保持する
type RequestContext struct {
	ctx       context.Context
	request   *requestInfo
	job       *JobStatus
	logger    *log.Logger
	tlsConfig *config.TLSStruct
	zone      string

	handled bool
}

// NewRequestContext 新しいリクエストコンテキストを生成する
func NewRequestContext(parent context.Context, request *requestInfo, tlsConfig *config.TLSStruct, logger *log.Logger) *RequestContext {
	logger = logger.With("request", request.requestType, "source", request.source, "resource", request.resourceName)
	return &RequestContext{
		ctx:       parent,
		request:   request,
		logger:    logger,
		tlsConfig: tlsConfig,
	}
}

// WithJobStatus JobStatusを持つContextを現在のContextを元に作成して返す
//
// 現在のContextが親Contextとなる
func (c *RequestContext) WithJobStatus(job *JobStatus) *RequestContext {
	return &RequestContext{
		ctx: c,
		request: &requestInfo{
			requestType:      c.request.requestType,
			source:           c.request.source,
			resourceName:     c.request.resourceName,
			desiredStateName: c.request.desiredStateName,
		},
		logger:    c.logger,
		job:       job,
		tlsConfig: c.tlsConfig,
	}
}

// WithZone Zoneを保持するContextを現在のContextを元に作成して返す
//
// 現在のContextが親Contextとなる
func (c *RequestContext) WithZone(zone string) *RequestContext {
	return &RequestContext{
		ctx:       c,
		request:   c.request,
		logger:    c.logger,
		job:       c.job,
		tlsConfig: c.tlsConfig,
		zone:      zone,
	}
}

// Request 現在のコンテキストで受けたリクエストの情報を返す
func (c *RequestContext) Request() *requestInfo {
	return c.request
}

// Logger 現在のコンテキストのロガーを返す
func (c *RequestContext) Logger() *log.Logger {
	return c.logger
}

// JobID 現在のコンテキストでのJobのIDを返す
//
// まだJobの実行決定が行われていない場合でも値を返す
func (c *RequestContext) JobID() string {
	return c.request.ID()
}

// Job 現在のコンテキストで実行中のJobを返す
//
// まだJobの実行決定が行われていない場合はnilを返す
func (c *RequestContext) Job() *JobStatus {
	return c.job
}

func (c *RequestContext) init() {
	if c.ctx == nil {
		c.ctx = context.Background()
	}
}

// Deadline context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *RequestContext) Deadline() (deadline time.Time, ok bool) {
	c.init()
	return c.ctx.Deadline()
}

// Done context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *RequestContext) Done() <-chan struct{} {
	c.init()
	return c.ctx.Done()
}

// Err context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *RequestContext) Err() error {
	c.init()
	return c.ctx.Err()
}

// Value context.Contextの実装、内部で保持しているcontextに処理を委譲している
func (c *RequestContext) Value(key interface{}) interface{} {
	c.init()
	return c.ctx.Value(key)
}
