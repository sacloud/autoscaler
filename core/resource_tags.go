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
	"fmt"
	"strings"

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/search"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

const resourceIDMarkerTagName = "@autoscaler.id"

// resourceIDMarkerTag さくらのクラウド上のリソースに付与するIDマーカータグ
//
// プラン変更時のリソースID変更に追随するために現在のIDをタグとして保持しておくためのもの
func resourceIDMarkerTag(id types.ID) string {
	return fmt.Sprintf("%s=%s", resourceIDMarkerTagName, id.String())
}

// SetupTagsWithResourceID さくらのクラウド上のリソースのタグをセットアップする
//
// 古いIDマーカータグがあれば削除し、指定のIDのIDマーカータグが付与されていなければ追加する
// タグに変更があった場合は2番目の戻り値としてtrueを返す
func SetupTagsWithResourceID(current types.Tags, id types.ID) (tags types.Tags, changed bool) {
	idTag := resourceIDMarkerTag(id)

	for _, t := range current {
		// 古いタグが残っていたら削除
		if strings.HasPrefix(t, resourceIDMarkerTagName) && t != idTag {
			changed = true
		} else {
			tags = append(tags, t)
		}
	}

	if !existTag(tags, idTag) {
		tags = append(tags, idTag)
		changed = true
	}

	return tags, changed
}

// FindConditionWithResourceIDTag IDマーカータグで検索するためのAPIパラメータを作成する
func FindConditionWithResourceIDTag(id types.ID) *sacloud.FindCondition {
	return &sacloud.FindCondition{
		Filter: search.Filter{
			search.Key("Tags.Name"): search.TagsAndEqual(resourceIDMarkerTag(id)),
		},
	}
}

func existTag(tags types.Tags, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}
