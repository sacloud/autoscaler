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

func (rds *ResourceDefinitions) UnmarshalYAML(data []byte) error {
	var rawResources []interface{}
	if err := yaml.UnmarshalWithOptions(data, &rawResources, yaml.Strict()); err != nil {
		return err
	}

	resourceDefs := ResourceDefinitions{}
	for _, rawResource := range rawResources {
		v, ok := rawResource.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid value: resource: %s", rawResource)
		}
		resource, err := rds.unmarshalResourceDefFromMap(v)
		if err != nil {
			return err
		}

		rds.setParentResource(nil, resource)
		resourceDefs = append(resourceDefs, resource)
	}

	*rds = resourceDefs
	return nil
}

func (rds *ResourceDefinitions) setParentResource(parent, r ResourceDefinition) {
	if parent != nil {
		if v, ok := r.(ChildResourceDefinition); ok {
			v.SetParent(parent)
		}
	}
	for _, child := range r.Children() {
		rds.setParentResource(r, child)
	}
}

func (rds *ResourceDefinitions) unmarshalResourceDefFromMap(data map[string]interface{}) (ResourceDefinition, error) {
	rawTypeName, ok := data["type"]
	if !ok {
		return nil, fmt.Errorf("'type' field required: %v", data)
	}
	typeName, ok := rawTypeName.(string)
	if !ok {
		return nil, fmt.Errorf("'type' is not string: %v", data)
	}

	var defs ResourceDefinitions
	if rawChildren, ok := data["resources"]; ok {
		if children, ok := rawChildren.([]interface{}); ok {
			for _, child := range children {
				if c, ok := child.(map[string]interface{}); ok {
					r, err := rds.unmarshalResourceDefFromMap(c)
					if err != nil {
						return nil, err
					}
					defs = append(defs, r)
				}
			}
		}
		delete(data, "resources")
	}

	remarshelded, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("yaml.Marshal failed with %v", data)
	}

	var def ResourceDefinition
	switch typeName {
	case "Server":
		v := &ResourceDefServer{}
		if err := yaml.UnmarshalWithOptions(remarshelded, v, yaml.Strict()); err != nil {
			return nil, err
		}
		v.children = defs
		def = v
	case "ServerGroup":
		v := &ResourceDefServerGroup{}
		if err := yaml.UnmarshalWithOptions(remarshelded, v, yaml.Strict()); err != nil {
			return nil, err
		}
		v.children = defs
		def = v
	case "EnhancedLoadBalancer", "ELB":
		v := &ResourceDefELB{}
		if err := yaml.UnmarshalWithOptions(remarshelded, v, yaml.Strict()); err != nil {
			return nil, err
		}
		// TypeNameのエイリアスを正規化
		v.TypeName = "EnhancedLoadBalancer"
		v.children = defs
		def = v
	case "GSLB":
		v := &ResourceDefGSLB{}
		if err := yaml.UnmarshalWithOptions(remarshelded, v, yaml.Strict()); err != nil {
			return nil, err
		}
		v.children = defs
		def = v
	case "DNS":
		v := &ResourceDefDNS{}
		if err := yaml.UnmarshalWithOptions(remarshelded, v, yaml.Strict()); err != nil {
			return nil, err
		}
		v.children = defs
		def = v
	case "Router":
		v := &ResourceDefRouter{}
		if err := yaml.UnmarshalWithOptions(remarshelded, v, yaml.Strict()); err != nil {
			return nil, err
		}
		v.children = defs
		def = v
	case "LoadBalancer", "LB":
		v := &ResourceDefLoadBalancer{}
		if err := yaml.UnmarshalWithOptions(remarshelded, v, yaml.Strict()); err != nil {
			return nil, err
		}
		// TypeNameのエイリアスを正規化
		v.TypeName = "LoadBalancer"
		v.children = defs
		def = v
	default:
		return nil, fmt.Errorf("unexpected type: %s", typeName)
	}

	return def, nil
}