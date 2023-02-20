// Copyright 2021-2023 The sacloud/autoscaler Authors
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
	"time"

	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
)

// ParentResourceDef サーバやサーバグループの親リソースを示すResourceDefinition実装
type ParentResourceDef struct {
	TypeName string          `yaml:"type" validate:"required,oneof=DNS EnhancedLoadBalancer ELB GSLB LoadBalancer"`
	Selector *NameOrSelector `yaml:"selector" validate:"required"`
}

func (d *ParentResourceDef) Type() ResourceTypes {
	switch d.TypeName {
	case ResourceTypeELB.String(), "ELB":
		return ResourceTypeELB
	case ResourceTypeGSLB.String():
		return ResourceTypeGSLB
	case ResourceTypeDNS.String():
		return ResourceTypeDNS
	case ResourceTypeLoadBalancer.String():
		return ResourceTypeLoadBalancer
	}
	return ResourceTypeUnknown
}

func (d *ParentResourceDef) UnmarshalYAML(ctx context.Context, data []byte) error {
	type alias ParentResourceDef
	var v alias
	if err := yaml.UnmarshalContext(ctx, data, &v); err != nil {
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

func (d *ParentResourceDef) Validate(ctx context.Context, apiClient iaas.APICaller, zone string) []error {
	errors := &multierror.Error{}

	if _, err := d.findCloudResources(ctx, apiClient, zone); err != nil {
		errors = multierror.Append(errors, err)
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ParentResourceDef) Compute(ctx *RequestContext, apiClient iaas.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient, ctx.zone)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, resource := range cloudResources {
		r, err := NewParentResource(ctx, apiClient, d, resource, ctx.zone)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

// LastModifiedAt この定義が対象とするリソース(群)の最終更新日時を返す
func (d *ParentResourceDef) LastModifiedAt(ctx *RequestContext, apiClient iaas.APICaller) (time.Time, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient, ctx.zone)
	if err != nil {
		return time.Time{}, err
	}
	last := time.Time{}
	for _, r := range cloudResources {
		if r.GetModifiedAt().After(last) {
			last = r.GetModifiedAt()
		}
	}
	return last, nil
}

type SakuraCloudResource interface {
	GetID() types.ID
	GetName() string
	GetModifiedAt() time.Time
}

func (d *ParentResourceDef) findCloudResources(ctx context.Context, apiClient iaas.APICaller, zone string) ([]SakuraCloudResource, error) {
	selector := d.Selector
	var results []SakuraCloudResource

	switch d.Type() {
	case ResourceTypeELB:
		op := iaas.NewProxyLBOp(apiClient)
		found, err := op.Find(ctx, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.ProxyLBs {
			results = append(results, v)
		}
	case ResourceTypeGSLB:
		op := iaas.NewGSLBOp(apiClient)
		found, err := op.Find(ctx, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.GSLBs {
			results = append(results, v)
		}
	case ResourceTypeDNS:
		op := iaas.NewDNSOp(apiClient)
		found, err := op.Find(ctx, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.DNS {
			results = append(results, v)
		}
	case ResourceTypeRouter:
		op := iaas.NewInternetOp(apiClient)
		found, err := op.Find(ctx, zone, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.Internet {
			results = append(results, &sakuraCloudRouter{Internet: v, zone: zone})
		}
	case ResourceTypeLoadBalancer:
		op := iaas.NewLoadBalancerOp(apiClient)
		found, err := op.Find(ctx, zone, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}
		for _, v := range found.LoadBalancers {
			results = append(results, v)
		}
	}

	switch len(results) {
	case 0:
		return nil, validate.Errorf("resource not found with selector: %s", selector.String())
	case 1:
		return results, nil
	default:
		var names []string
		for _, r := range results {
			names = append(names, fmt.Sprintf("{ID:%s, Name:%s}", r.GetID(), r.GetName()))
		}
		return nil, validate.Errorf("A parent resource definition must return one resource, but got multiple resources: definition: {Type:%s, Selector:%s}, got: %s",
			d.Type(), d.Selector, fmt.Sprintf("[%s]", strings.Join(names, ",")),
		)
	}
}
