// Copyright 2021-2025 The sacloud/autoscaler Authors
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
	"github.com/sacloud/autoscaler/handler"
)

// Computed 各リソースが算出した希望する状態(Desired State)を示すインターフェース
//
// 1つのComputedに対し1つのさくらのクラウド上のリソースが対応する
type Computed interface {
	// Type このComputedが表すさくらのクラウド上のリソースの種別
	Type() ResourceTypes
	// ID このComputedが表すさくらのクラウド上のリソースのID、まだ存在しないリソースの場合は空文字を返す
	ID() string
	// Name このComputedが表すさくらのクラウド上のリソースのName、ログ出力用
	Name() string
	// Zone このComputedが表すさくらのクラウド上のリソースが属するゾーン名、グローバルリソースの場合は空文字を返す
	Zone() string

	// Instruction 現在のリソースの状態から算出されたハンドラーへの指示の種類
	Instruction() handler.ResourceInstructions
	// SetupGracePeriod セットアップのための猶予時間(秒数)
	//
	// Handleされた後、セットアップの完了を待つためにこの秒数分待つ
	SetupGracePeriod() int

	// Current ハンドラーに渡すパラメータ、現在の状態を示す 現在存在しないリソースの場合はnilを返す
	Current() *handler.Resource
	// Desired ハンドラーに渡すパラメータ、InstructionがNOOPやDELETEの場合はnilを返す
	Desired() *handler.Resource
}
