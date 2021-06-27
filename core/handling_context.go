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

import "github.com/sacloud/autoscaler/handler"

// HandlingContext 1リクエスト中の1リソースに対するハンドリングのスコープに対応するコンテキスト
//
// context.Contextを実装し、core.Contextに加えて現在処理中のComputedを保持する
type HandlingContext struct {
	*RequestContext
	cachedComputed Computed
}

func NewHandlingContext(parent *RequestContext, computed Computed) *HandlingContext {
	return &HandlingContext{
		RequestContext: parent,
		cachedComputed: computed,
	}
}

func (c *HandlingContext) WithLogger(keyvals ...interface{}) *HandlingContext {
	ctx := NewHandlingContext(&RequestContext{
		ctx:       c.RequestContext.ctx,
		request:   c.RequestContext.request,
		job:       c.RequestContext.job,
		logger:    c.RequestContext.logger,
		tlsConfig: c.RequestContext.tlsConfig,
	}, c.cachedComputed)
	ctx.logger = ctx.logger.With(keyvals...)
	return ctx
}

// CurrentComputed 現在処理中の[]Computedを返す
func (c *HandlingContext) CurrentComputed() Computed {
	return c.cachedComputed
}

// ComputeResult コンテキストに保持しているComputedと渡されたComputedを比較しHandleの結果を算出する
func (c *HandlingContext) ComputeResult(computed Computed) handler.PostHandleRequest_ResourceHandleResults {
	if computed.Instruction() != handler.ResourceInstructions_NOOP {
		return handler.PostHandleRequest_UNKNOWN
	}

	switch {
	case computed.ID() == "": // deleted?
		return handler.PostHandleRequest_DELETED
	case computed.ID() == c.cachedComputed.ID(): // in-place update?
		if c.cachedComputed.Instruction() == handler.ResourceInstructions_UPDATE {
			return handler.PostHandleRequest_UPDATED
		}
	case c.cachedComputed.ID() == "" && computed.ID() != "": // created?
		return handler.PostHandleRequest_CREATED
	case c.cachedComputed.ID() != computed.ID(): // plan changed?
		if c.cachedComputed.Instruction() == handler.ResourceInstructions_UPDATE {
			return handler.PostHandleRequest_UPDATED
		}
	}

	return handler.PostHandleRequest_UNKNOWN
}
