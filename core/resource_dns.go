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

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceDNS struct {
	*ResourceBase

	apiClient sacloud.APICaller
	dns       *sacloud.DNS
	def       *ResourceDefDNS
}

func NewResourceDNS(ctx *RequestContext, apiClient sacloud.APICaller, def *ResourceDefDNS, dns *sacloud.DNS) (*ResourceDNS, error) {
	resource := &ResourceDNS{
		ResourceBase: &ResourceBase{resourceType: ResourceTypeDNS},
		apiClient:    apiClient,
		dns:          dns,
		def:          def,
	}
	if err := resource.setResourceIDTag(ctx); err != nil {
		return nil, err
	}
	return resource, nil
}

func (r *ResourceDNS) String() string {
	if r == nil || r.dns == nil {
		return "(empty)"
	}
	return fmt.Sprintf("{Type: %s, ID: %s, Name: %s}", r.Type(), r.dns.ID, r.dns.Name)
}

func (r *ResourceDNS) Compute(ctx *RequestContext, refresh bool) (Computed, error) {
	if refresh {
		if err := r.refresh(ctx); err != nil {
			return nil, err
		}
	}

	computed := &computedDNS{
		instruction: handler.ResourceInstructions_NOOP,
		dns:         &sacloud.DNS{},
		resource:    r,
	}
	if err := mapconvDecoder.ConvertTo(r.dns, computed.dns); err != nil {
		return nil, fmt.Errorf("computing desired state failed: %s", err)
	}

	return computed, nil
}

func (r *ResourceDNS) setResourceIDTag(ctx *RequestContext) error {
	tags, changed := SetupTagsWithResourceID(r.dns.Tags, r.dns.ID)
	if changed {
		dnsOp := sacloud.NewDNSOp(r.apiClient)
		updated, err := dnsOp.Update(ctx, r.dns.ID, &sacloud.DNSUpdateRequest{
			Description:  r.dns.Description,
			Tags:         tags,
			IconID:       r.dns.IconID,
			Records:      r.dns.Records,
			SettingsHash: r.dns.SettingsHash,
		})
		if err != nil {
			return err
		}
		r.dns = updated
	}
	return nil
}

func (r *ResourceDNS) refresh(ctx *RequestContext) error {
	dnsOp := sacloud.NewDNSOp(r.apiClient)

	// まずキャッシュしているリソースのIDで検索
	dns, err := dnsOp.Read(ctx, r.dns.ID)
	if err != nil {
		if sacloud.IsNotFoundError(err) {
			// 見つからなかったらIDマーカータグを元に検索
			found, err := dnsOp.Find(ctx, FindConditionWithResourceIDTag(r.dns.ID))
			if err != nil {
				return err
			}
			if len(found.DNS) == 0 {
				return fmt.Errorf("dns not found with: Filter='%s'", resourceIDMarkerTag(r.dns.ID))
			}
			if len(found.DNS) > 1 {
				return fmt.Errorf("invalid state: found multiple dns with: Filter='%s'", resourceIDMarkerTag(r.dns.ID))
			}
			dns = found.DNS[0]
		} else {
			return err
		}
	}
	r.dns = dns
	return r.setResourceIDTag(ctx)
}
