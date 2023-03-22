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

package core

import "fmt"

type RequestTypes int

const (
	requestTypeUnknown RequestTypes = iota
	requestTypeUp                   // スケールアップ or スケールアウト
	requestTypeDown                 // スケールダウン or スケールイン
)

func (r RequestTypes) String() string {
	switch r {
	case requestTypeUp:
		return "Up"
	case requestTypeDown:
		return "Down"
	default:
		return "unknown request type"
	}
}

type requestInfo struct {
	requestType      RequestTypes
	source           string
	resourceName     string
	desiredStateName string
	sync             bool
}

func (r *requestInfo) String() string {
	return fmt.Sprintf("%#v", r)
}

// ID リクエストのパラメータから一意に決まるID
//
// RequestTypesのUp/Downの違いやDesiredStateNameの違いを問わずに値を決めるため、同一リソースに対するアクションが実行中かの判定に利用できる
func (r *requestInfo) ID() string {
	return r.resourceName
}
