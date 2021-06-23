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

func (rg *ResourceDefGroup) UnmarshalYAML(data []byte) error {
	var rawMap map[string]interface{}
	if err := yaml.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	group := &ResourceDefGroup{}
	resources := rawMap["resources"].([]interface{})
	for _, rawResource := range resources {
		v := rawResource.(map[string]interface{})
		resource, err := rg.unmarshalResourceDefFromMap(v)
		if err != nil {
			return err
		}

		rg.setParentResource(nil, resource)
		group.ResourceDefs = append(group.ResourceDefs, resource)
	}

	if rawActions, ok := rawMap["actions"]; ok {
		group.Actions = Actions{}

		actions := rawActions.(map[string]interface{})
		for k, v := range actions {
			var handlers []string

			v := v.([]interface{})
			for _, v := range v {
				if v, ok := v.(string); ok {
					handlers = append(handlers, v)
				}
			}

			group.Actions[k] = handlers
		}
	}

	*rg = *group
	return nil
}

func (rg *ResourceDefGroup) setParentResource(parent, r ResourceDefinition) {
	if parent != nil {
		if v, ok := r.(ChildResourceDefinition); ok {
			v.SetParent(parent)
		}
	}
	for _, child := range r.Children() {
		rg.setParentResource(r, child)
	}
}

func (rg *ResourceDefGroup) unmarshalResourceDefFromMap(data map[string]interface{}) (ResourceDefinition, error) {
	rawTypeName, ok := data["type"]
	if !ok {
		return nil, fmt.Errorf("'type' field required: %v", data)
	}
	typeName, ok := rawTypeName.(string)
	if !ok {
		return nil, fmt.Errorf("'type' is not string: %v", data)
	}

	remarshelded, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("yaml.Marshal failed with %v", data)
	}

	var defs ResourceDefinitions
	if rawChildren, ok := data["resources"]; ok {
		if children, ok := rawChildren.([]interface{}); ok {
			for _, child := range children {
				if c, ok := child.(map[string]interface{}); ok {
					r, err := rg.unmarshalResourceDefFromMap(c)
					if err != nil {
						return nil, err
					}
					defs = append(defs, r)
				}
			}
		}
	}

	var def ResourceDefinition
	switch typeName {
	case "Server":
		v := &ResourceDefServer{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.children = defs
		def = v
	case "EnhancedLoadBalancer", "ELB":
		v := &ResourceDefELB{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		// TypeNameのエイリアスを正規化
		v.TypeName = "EnhancedLoadBalancer"
		v.children = defs
		def = v
	case "GSLB":
		v := &ResourceDefGSLB{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.children = defs
		def = v
	case "DNS":
		v := &ResourceDefDNS{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.children = defs
		def = v
	case "Router":
		v := &ResourceDefRouter{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.children = defs
		def = v
	default:
		return nil, fmt.Errorf("unexpected type: %s", typeName)
	}

	return def, nil
}