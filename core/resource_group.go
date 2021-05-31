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

	"github.com/goccy/go-yaml"
)

type ResourceGroup struct {
	HandlerConfigs []*ResourceHandlerConfig `yaml:"handlers"`
	Resources      Resources                `yaml:"resources"`
	Name           string
}

type ResourceHandlerConfig struct {
	Name string `yaml:"name"`
	// TODO 未実装
	//Selector *ResourceSelector `yaml:"selector"`
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
				resourceGroup.HandlerConfigs = append(resourceGroup.HandlerConfigs, &ResourceHandlerConfig{Name: n})
			}
		}
	}

	*rg = *resourceGroup
	return nil
}

func (rg *ResourceGroup) ComputeAll(ctx *Context, apiClient sacloud.APICaller) ([]Desired, error) {
	// TODO 並列化
	var allDesired []Desired
	err := rg.Resources.Walk(func(resource Resource) error {
		desired, err := resource.Desired(ctx, apiClient)
		if err != nil {
			return err
		}
		allDesired = append(allDesired, desired)
		return nil
	})
	// TODO 並べ替え
	return allDesired, err
}

// Handlers 引数で指定されたハンドラーのリストをHandlerConfigsに合致するハンドラだけにフィルタして返す
func (rg *ResourceGroup) Handlers(allHandlers Handlers) (Handlers, error) {
	// ビルトイン + configで定義されたハンドラーからHandlerConfigsに定義されたハンドラーを探す
	var handlers Handlers
	for _, conf := range rg.HandlerConfigs {
		var found *Handler
		for _, h := range allHandlers {
			if h.Name == conf.Name {
				found = h
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("handler %q not found", conf.Name)
		}
		handlers = append(handlers, found)
	}
	return handlers, nil
}
