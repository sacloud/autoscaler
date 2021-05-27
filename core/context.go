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
)

type Context struct {
	ctx     context.Context
	request *requestInfo
}

func NewContext(parent context.Context, request *requestInfo) *Context {
	return &Context{
		ctx:     parent,
		request: request,
	}
}

func (c *Context) Request() *requestInfo {
	return c.request
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
