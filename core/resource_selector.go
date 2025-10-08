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
	"context"
	"fmt"
	"strconv"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/search"
	"github.com/sacloud/iaas-api-go/types"
)

type MultiZoneSelector struct {
	*ResourceSelector `yaml:",inline" validate:"required"`
	Zones             []string `yaml:"zones" validate:"required,zones"`
}

func (rs *MultiZoneSelector) String() string {
	return rs.ResourceSelector.String() + fmt.Sprintf(", Zones: %s", rs.Zones)
}

func (rs *MultiZoneSelector) Validate() error {
	if rs == nil || rs.ResourceSelector == nil {
		return validate.Errorf("selector: required")
	}
	if err := rs.ResourceSelector.Validate(); err != nil {
		return err
	}
	if len(rs.Zones) == 0 {
		return fmt.Errorf("selector.Zones: required")
	}

	for _, zone := range rs.Zones {
		exist := false
		for _, z := range iaas.SakuraCloudZones {
			if z == zone {
				exist = true
				break
			}
		}
		if !exist {
			return fmt.Errorf("selector.Zones: invalid zone: %s", zone)
		}
	}
	return nil
}

// ResourceSelector さくらのクラウド上で対象リソースを特定するための情報を提供する
type ResourceSelector struct {
	ID    types.ID `yaml:"id" validate:"required_without_all=Tags Names"`
	Tags  []string `yaml:"tags" validate:"required_without_all=ID Names"`
	Names []string `yaml:"names" validate:"required_without_all=ID Tags"`
}

func (rs *ResourceSelector) String() string {
	if rs != nil {
		return fmt.Sprintf("ID: %s, Names: %s, Tags: %s", rs.ID, rs.Names, rs.Tags)
	}
	return ""
}

func (rs *ResourceSelector) findCondition() *iaas.FindCondition {
	fc := &iaas.FindCondition{
		Filter: search.Filter{},
	}
	if !rs.ID.IsEmpty() {
		fc.Filter[search.Key("ID")] = search.ExactMatch(rs.ID.String())
	}
	if len(rs.Names) != 0 {
		fc.Filter[search.Key("Name")] = search.PartialMatch(rs.Names...)
	}
	if len(rs.Tags) != 0 {
		fc.Filter[search.Key("Tags.Name")] = search.TagsAndEqual(rs.Tags...)
	}
	return fc
}

func (rs *ResourceSelector) Validate() error {
	if err := validate.Struct(rs); err != nil {
		return err
	}

	if !rs.ID.IsEmpty() && (len(rs.Names) > 0 || len(rs.Tags) > 0) {
		return validate.Errorf("selector.ID and (selector.Names or selector.Tags): cannot specify both")
	}
	return nil
}

// NameOrSelector 名前(文字列)、もしくはResourceSelectorを表すstruct
type NameOrSelector struct {
	ResourceSelector
}

func (v *NameOrSelector) UnmarshalYAML(ctx context.Context, data []byte) error {
	// セレクタとしてUnmarshalしてみてエラーだったら文字列と見なす
	var selector ResourceSelector
	if err := yaml.UnmarshalWithOptions(data, &selector, yaml.Strict()); err != nil {
		selector = ResourceSelector{
			Names: []string{string(data)},
		}
	}
	*v = NameOrSelector{ResourceSelector: selector}
	return nil
}

// IdOrNameOrSelector ID、名前、またはResourceSelectorを表すstruct
type IdOrNameOrSelector struct {
	ResourceSelector
}

func (v *IdOrNameOrSelector) UnmarshalYAML(ctx context.Context, data []byte) error {
	var selector ResourceSelector
	s := string(data)

	// 数値だけの文字列だったら(ParseUintが成功したら)IDが指定されたとみなす
	if _, err := strconv.ParseUint(s, 10, 64); err == nil {
		selector = ResourceSelector{ID: types.StringID(s)}
	} else {
		// セレクタとしてUnmarshalしてみてエラーだったらNamesと見なす
		if err := yaml.UnmarshalWithOptions(data, &selector, yaml.Strict()); err != nil {
			selector = ResourceSelector{
				Names: []string{s},
			}
		}
	}

	*v = IdOrNameOrSelector{ResourceSelector: selector}
	return nil
}
