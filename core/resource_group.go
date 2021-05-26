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

	"github.com/goccy/go-yaml"
)

// ResourceGroups 一意な名前をキーとするリソースのリスト
type ResourceGroups struct {
	groups map[string]Resources
}

func newResourceGroups() *ResourceGroups {
	return &ResourceGroups{
		groups: make(map[string]Resources),
	}
}

func (rg *ResourceGroups) Get(key string) Resources {
	v, _ := rg.GetOk(key)
	return v
}

func (rg *ResourceGroups) GetOk(key string) (Resources, bool) {
	v, ok := rg.groups[key]
	return v, ok
}

func (rg *ResourceGroups) Set(key string, resources Resources) {
	rg.groups[key] = resources
}

func (rg *ResourceGroups) appendResource(key string, r Resource) {
	rs, ok := rg.GetOk(key)
	if !ok {
		rs = Resources{}
	}
	rs = append(rs, r)
	rg.Set(key, rs)
}

func (rg *ResourceGroups) UnmarshalYAML(data []byte) error {
	var rawMap map[string][]map[string]interface{}
	if err := yaml.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	resourceGroups := newResourceGroups()
	for k, v := range rawMap {
		for _, v := range v {
			rawTypeName, ok := v["type"]
			if !ok {
				return fmt.Errorf("'type' field required: %v", v)
			}
			typeName, ok := rawTypeName.(string)
			if !ok {
				return fmt.Errorf("'type' is not string: %v", v)
			}

			remarshelded, err := yaml.Marshal(v)
			if err != nil {
				return fmt.Errorf("yaml.Marshal failed with key:%s, element: %v", k, v)
			}

			var resource Resource
			switch typeName {
			case "Server":
				resource = &Server{}
			case "ServerGroup":
				resource = &ServerGroup{}
			case "EnhancedLoadBalancer", "ELB":
				resource = &EnhancedLoadBalancer{}
			case "GSLB":
				resource = &GSLB{}
			case "DNS":
				resource = &DNS{}
			default:
				return fmt.Errorf("received unexpected type: %s", typeName)
			}

			if err := yaml.Unmarshal(remarshelded, resource); err != nil {
				return fmt.Errorf("yaml.Unmarshal failed with key:%s, element: %v", k, v)
			}

			// TypeNameのエイリアスを正規化
			if elb, ok := resource.(*EnhancedLoadBalancer); ok {
				elb.TypeName = "EnhancedLoadBalancer"
			}

			resourceGroups.appendResource(k, resource)
		}
	}
	*rg = *resourceGroups
	return nil
}
