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

	"github.com/sacloud/libsacloud/v2/sacloud"
)

type stubResourceDef struct {
	*ResourceDefBase
	computeFunc func(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error)
}

func (d *stubResourceDef) Validate(_ context.Context, _ sacloud.APICaller) []error {
	return nil
}

func (d *stubResourceDef) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources2, error) {
	if d.computeFunc != nil {
		return d.computeFunc(ctx, apiClient)
	}
	return nil, nil
}

// TODO リソース切り替え時に名前変更
type stubResource2 struct {
	*ResourceBase2
	computeFunc func(ctx *RequestContext, refresh bool) (Computed, error)
}

func (r *stubResource2) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if r.computeFunc != nil {
		return r.computeFunc(ctx, refresh)
	}
	return nil, nil
}
