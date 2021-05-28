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
	groups map[string]*ResourceGroup
}

type ResourceGroup struct {
	Handlers  []*ResourceHandlerConfig `yaml:"handlers"`
	Resources Resources                `yaml:"resources"`
	Name      string
}

type ResourceHandlerConfig struct {
	Name string `yaml:"name"`
	// TODO 未実装
	//Selector *ResourceSelector `yaml:"selector"`
}

func newResourceGroups() *ResourceGroups {
	return &ResourceGroups{
		groups: make(map[string]*ResourceGroup),
	}
}

func (rg *ResourceGroups) Get(key string) *ResourceGroup {
	v, _ := rg.GetOk(key)
	return v
}

func (rg *ResourceGroups) GetOk(key string) (*ResourceGroup, bool) {
	v, ok := rg.groups[key]
	return v, ok
}

func (rg *ResourceGroups) Set(key string, group *ResourceGroup) {
	group.Name = key
	rg.groups[key] = group
}

func (rg *ResourceGroups) UnmarshalYAML(data []byte) error {
	var loaded map[string]*ResourceGroup
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		return err
	}
	for k, v := range loaded {
		v.Name = k
	}
	*rg = ResourceGroups{groups: loaded}
	return nil
}

func (rg *ResourceGroup) UnmarshalYAML(data []byte) error {
	var rawMap map[string]interface{}
	if err := yaml.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	resourceGroup := &ResourceGroup{}
	resources := rawMap["resources"].([]interface{})
	for _, rawResource := range resources {
		v := rawResource.(map[string]interface{})
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
			return fmt.Errorf("yaml.Marshal failed with %v", v)
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
			return fmt.Errorf("yaml.Unmarshal failed with %v", v)
		}

		// TypeNameのエイリアスを正規化
		if elb, ok := resource.(*EnhancedLoadBalancer); ok {
			elb.TypeName = "EnhancedLoadBalancer"
		}

		resourceGroup.Resources = append(resourceGroup.Resources, resource)
	}

	if rawHandlers, ok := rawMap["handlers"]; ok {
		handlers := rawHandlers.([]interface{})
		for _, name := range handlers {
			if n, ok := name.(string); ok {
				resourceGroup.Handlers = append(resourceGroup.Handlers, &ResourceHandlerConfig{Name: n})
			}
		}
	}

	*rg = *resourceGroup
	return nil
}
