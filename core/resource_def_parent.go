// Copyright 2021-2022 The sacloud/autoscaler Authors
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
	"strings"

	"github.com/goccy/go-yaml"

	"github.com/sacloud/libsacloud/v2/sacloud/types"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

// ParentResourceDef サーバやサーバグループの親リソースを示すResourceDefinition実装
type ParentResourceDef struct {
	TypeName string          `yaml:"type" validate:"required,oneof=DNS EnhancedLoadBalancer ELB GSLB LoadBalancer Router"`
	Selector *NameOrSelector `yaml:"selector" validate:"required"`

	// zone Compute前に親リソース側で設定される
	zone string `yaml:"-" validate:"omitempty,zone"`
}

func (d *ParentResourceDef) Name() string {
	return "" // Note: 常に空文字列なためこのリソースが直接処理対象になることがない
}

func (d *ParentResourceDef) Type() ResourceTypes {
	switch d.TypeName {
	case ResourceTypeELB.String(), "ELB":
		return ResourceTypeELB
	case ResourceTypeGSLB.String():
		return ResourceTypeGSLB
	case ResourceTypeDNS.String():
		return ResourceTypeDNS
	case ResourceTypeRouter.String():
		return ResourceTypeRouter
	case ResourceTypeLoadBalancer.String():
		return ResourceTypeLoadBalancer
	}
	return ResourceTypeUnknown
}

func (d *ParentResourceDef) UnmarshalYAML(ctx context.Context, data []byte) error {
	type alias ParentResourceDef
	var v alias
	if err := yaml.UnmarshalWithContext(ctx, data, &v); err != nil {
		return err
	}
	*d = ParentResourceDef(v)

	// 正規化
	d.TypeName = d.Type().String()
	return nil
}

func (d *ParentResourceDef) String() string {
	return fmt.Sprintf("Type: %s, %s", d.Type().String(), d.Selector.String())
}

func (d *ParentResourceDef) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}

	if _, err := d.findCloudResources(ctx, apiClient); err != nil {
		errors = multierror.Append(errors, err)
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ParentResourceDef) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, resource := range cloudResources {
		r, err := NewParentResource(ctx, apiClient, d, resource)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

type SakuraCloudResource interface {
	GetID() types.ID
	GetName() string
}

func (d *ParentResourceDef) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]SakuraCloudResource, error) {
	selector := d.Selector
	var results []SakuraCloudResource

	switch d.Type() {
	case ResourceTypeELB:
		op := sacloud.NewProxyLBOp(apiClient)
		found, err := op.Find(ctx, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.ProxyLBs {
			results = append(results, v)
		}
	case ResourceTypeGSLB:
		op := sacloud.NewGSLBOp(apiClient)
		found, err := op.Find(ctx, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.GSLBs {
			results = append(results, v)
		}
	case ResourceTypeDNS:
		op := sacloud.NewDNSOp(apiClient)
		found, err := op.Find(ctx, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.DNS {
			results = append(results, v)
		}
	case ResourceTypeRouter:
		op := sacloud.NewInternetOp(apiClient)
		found, err := op.Find(ctx, d.zone, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.Internet {
			results = append(results, v)
		}
	case ResourceTypeLoadBalancer:
		op := sacloud.NewLoadBalancerOp(apiClient)
		found, err := op.Find(ctx, d.zone, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.LoadBalancers {
			results = append(results, v)
		}
	}

	switch len(results) {
	case 0:
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	case 1:
		return results, nil
	default:
		var names []string
		for _, r := range results {
			names = append(names, fmt.Sprintf("{ID:%s, Name:%s}", r.GetID(), r.GetName()))
		}
		return nil, fmt.Errorf("A parent resource definition must return one resource, but got multiple resources: definition: {Type:%s, Selector:%s}, got: %s",
			d.Type(), d.Selector, fmt.Sprintf("[%s]", strings.Join(names, ",")),
		)
	}
}