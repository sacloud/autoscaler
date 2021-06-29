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

	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/search"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type MultiZoneSelector struct {
	*ResourceSelector `yaml:",inline"`
	Zones             []string `yaml:"zones"`
}

func (rs *MultiZoneSelector) String() string {
	return rs.ResourceSelector.String() + fmt.Sprintf(", Zones: %s", rs.Zones)
}

func (rs *MultiZoneSelector) Validate() error {
	if rs == nil || rs.ResourceSelector == nil {
		return fmt.Errorf("selector: required")
	}
	if err := rs.ResourceSelector.Validate(); err != nil {
		return err
	}
	if len(rs.Zones) == 0 {
		return fmt.Errorf("selector.Zones: required")
	}

	for _, zone := range rs.Zones {
		exist := false
		for _, z := range sacloud.SakuraCloudZones {
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

type SingleZoneSelector struct {
	*ResourceSelector `yaml:",inline"`
	Zone              string `yaml:"zone"`
}

func (rs *SingleZoneSelector) String() string {
	return rs.ResourceSelector.String() + fmt.Sprintf(", Zone: %s", rs.Zone)
}

func (rs *SingleZoneSelector) Validate() error {
	if err := rs.ResourceSelector.Validate(); err != nil {
		return err
	}
	if len(rs.Zone) == 0 {
		return fmt.Errorf("selector.Zone: required")
	}
	return nil
}

// ResourceSelector さくらのクラウド上で対象リソースを特定するための情報を提供する
type ResourceSelector struct {
	ID    types.ID `yaml:"id"`
	Names []string `yaml:"names"`
}

func (rs *ResourceSelector) String() string {
	if rs != nil {
		return fmt.Sprintf("ID: %s, Names: %s", rs.ID, rs.Names)
	}
	return ""
}

func (rs *ResourceSelector) findCondition() *sacloud.FindCondition {
	fc := &sacloud.FindCondition{
		Filter: search.Filter{},
	}
	if !rs.ID.IsEmpty() {
		fc.Filter[search.Key("ID")] = search.ExactMatch(rs.ID.String())
	}
	if len(rs.Names) != 0 {
		fc.Filter[search.Key("Name")] = search.PartialMatch(rs.Names...)
	}
	return fc
}

func (rs *ResourceSelector) Validate() error {
	if rs == nil {
		return fmt.Errorf("selector: required")
	}

	if rs.ID.IsEmpty() && len(rs.Names) == 0 {
		return fmt.Errorf("selector.ID or selector.Names: required")
	}

	if !rs.ID.IsEmpty() && len(rs.Names) > 0 {
		return fmt.Errorf("selector.ID and selector.Names: cannot specify both")
	}
	return nil
}
