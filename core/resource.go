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

import "github.com/sacloud/libsacloud/v2/sacloud"

// Resource さくらのクラウド上のリソースを表すリソース
//
// ResourceDefinitionにより作成され、あるべき状態(Desired State)を示すComputedを算出する
// オートスケール動作をサポートするためにリソースにメタデータを付与する
type Resource interface {
	// Compute 現在/あるべき姿を算出する
	//
	// さくらのクラウド上の1つのリソースに対し1つのComputedを返す
	// Selector()の値によっては複数のComputedを返しても良い
	// Computeの結果はキャッシュしておき、Computed()で参照可能にしておく
	// キャッシュはClearCache()を呼ぶまで保持しておく
	Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Computed, error)

	// Computed Compute()の結果のキャッシュ、Compute()呼び出し前はnilを返す
	Computed() Computed

	// ClearCache Compute()の結果のキャッシュをクリアする
	ClearCache()
}
