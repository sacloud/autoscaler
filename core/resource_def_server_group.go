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
	"sort"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/size"
)

type ResourceDefServerGroup struct {
	*ResourceDefBase `yaml:",inline" validate:"required"`

	ServerNamePrefix string `yaml:"server_name_prefix"` // NameまたはServerNamePrefixが必須
	ServerNameFormat string `yaml:"server_name_format"`

	Zone  string   `yaml:"zone" validate:"required_without=Zones,omitempty,zone"`
	Zones []string `yaml:"zones" validate:"required_without=Zone,omitempty,unique,dive,required,zone"`

	MinSize int `yaml:"min_size" validate:"min=0,ltefield=MaxSize"`
	MaxSize int `yaml:"max_size" validate:"min=0,gtecsfield=MinSize"`

	Plans []*ServerGroupPlan `yaml:"plans"`

	Template      *ServerGroupInstanceTemplate `yaml:"template" validate:"required"`
	ShutdownForce bool                         `yaml:"shutdown_force"`

	ParentDef *ParentResourceDef `yaml:"parent"`
}

func (d *ResourceDefServerGroup) String() string {
	return fmt.Sprintf("Zones: %s, Name: %s", d.Zones, d.Name())
}

func (d *ResourceDefServerGroup) namePrefix() string {
	if d.ServerNamePrefix != "" {
		return d.ServerNamePrefix
	}
	return d.Name()
}

func (d *ResourceDefServerGroup) Validate(ctx context.Context, apiClient iaas.APICaller) []error {
	errors := &multierror.Error{}
	if d.namePrefix() == "" {
		errors = multierror.Append(errors, validate.Errorf("name or server_name_prefix: required"))
	}

	if err := d.printWarningForServerNamePrefix(ctx); err != nil {
		errors = multierror.Append(errors, err)
	}

	if d.Zone != "" && len(d.Zones) > 0 {
		errors = multierror.Append(errors, validate.Errorf("only one of zone and zones can be specified"))
	}
	// HACK: 値の正規化、処理する適切なタイミングがないため暫定的にここで処理している
	//       ResourceDefinitionレベルで初期化処理インターフェースができたらそちらに移動する
	if d.Zone != "" {
		d.Zones = []string{d.Zone}
	}

	if len(d.Zones) > 1 && d.ParentDef != nil {
		if d.ParentDef.Type() == ResourceTypeLoadBalancer { // 親リソース種別が増えたらここを修正
			errors = multierror.Append(errors, validate.Errorf("multiple zones cannot be specified when the parent is a LoadBalancer"))
		}
	}

	for _, p := range d.Plans {
		if !(d.MinSize <= p.Size && p.Size <= d.MaxSize) {
			errors = multierror.Append(errors, validate.Errorf("plan: plan.size must be between min_size and max_size: size:%d", p.Size))
		}
	}
	errors = multierror.Append(errors, d.Template.Validate(ctx, apiClient, d)...)
	if d.ParentDef != nil {
		errors = multierror.Append(errors, d.ParentDef.Validate(ctx, apiClient, d.Zone)...)
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

// printWarningForServerNamePrefix ServerNamePrefixをNameで代用することへの経過措置、将来のバージョンでは除去される予定
//
// see:https://github.com/sacloud/autoscaler/issues/338
func (d *ResourceDefServerGroup) printWarningForServerNamePrefix(ctx context.Context) error {
	if d.Name() != "" && d.ServerNamePrefix == "" {
		if loggerHolder, ok := ctx.(config.LoggerHolder); ok {
			if err := loggerHolder.Logger().Warn("message", "required: server_name_prefix"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *ResourceDefServerGroup) resourcePlans() ResourcePlans {
	var plans ResourcePlans
	for size := d.MinSize; size <= d.MaxSize; size++ {
		plan := &ServerGroupPlan{Size: size}
		for _, p := range d.Plans {
			if p.Size == plan.Size {
				plan.Name = p.Name
				break
			}
		}
		plans = append(plans, plan)
	}
	return plans
}

func (d *ResourceDefServerGroup) Compute(ctx *RequestContext, apiClient iaas.APICaller) (Resources, error) {
	// 現在のリソースを取得
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	// Min/MaxとUp/Downを考慮してサーバ数を決定
	plan, err := d.desiredPlan(ctx, len(cloudResources))
	if err != nil {
		return nil, err
	}

	var parent Resource
	if d.ParentDef != nil {
		parents, err := d.ParentDef.Compute(ctx.WithZone(d.Zones[0]), apiClient)
		if err != nil {
			return nil, err
		}
		if len(parents) != 1 {
			return nil, fmt.Errorf("got invalid parent resources: %#+v", parents)
		}
		parent = parents[0]
	}

	var resources Resources
	for i := range cloudResources {
		instance := &ResourceServerGroupInstance{
			ResourceBase: &ResourceBase{
				resourceType:     ResourceTypeServerGroupInstance,
				setupGracePeriod: d.SetupGracePeriod(),
			},
			apiClient:   apiClient,
			server:      cloudResources[i],
			zone:        cloudResources[i].Zone.Name,
			def:         d,
			instruction: handler.ResourceInstructions_NOOP,
			parent:      parent,
		}
		instance.indexInGroup = d.resourceIndex(instance)
		if i >= plan.Size {
			instance.instruction = handler.ResourceInstructions_DELETE
		}
		resources = append(resources, instance)
	}

	// セレクタ->ID変換、ゾーン非依存なので先に検索しておく
	iconId, err := d.Template.FindIconId(ctx, apiClient)
	if err != nil {
		return nil, err
	}
	for len(resources) < plan.Size {
		zone := d.determineZone(len(resources))
		ctx := ctx.WithZone(zone)

		// セレクタ->ID変換
		cdromId, err := d.Template.FindCDROMId(ctx, apiClient, zone)
		if err != nil {
			return nil, err
		}
		privateHostId, err := d.Template.FindPrivateHostId(ctx, apiClient, zone)
		if err != nil {
			return nil, err
		}

		commitment := types.Commitments.Standard
		if d.Template.Plan.DedicatedCPU {
			commitment = types.Commitments.DedicatedCPU
		}
		serverName, index := d.determineServerName(resources)

		resources = append(resources, &ResourceServerGroupInstance{
			ResourceBase: &ResourceBase{
				resourceType:     ResourceTypeServerGroupInstance,
				setupGracePeriod: d.SetupGracePeriod(),
			},
			apiClient: apiClient,
			server: &iaas.Server{
				Name:                 serverName,
				Tags:                 d.Template.CalculateTagsByIndex(index, len(d.Zones)),
				Description:          d.Template.Description,
				IconID:               types.StringID(iconId),
				CDROMID:              types.StringID(cdromId),
				PrivateHostID:        types.StringID(privateHostId),
				InterfaceDriver:      d.Template.InterfaceDriver,
				CPU:                  d.Template.Plan.Core,
				MemoryMB:             d.Template.Plan.Memory * size.GiB,
				GPU:                  d.Template.Plan.GPU,
				ServerPlanCPUModel:   d.Template.Plan.CPUModel,
				ServerPlanCommitment: commitment,
			},
			zone:         zone,
			def:          d,
			instruction:  handler.ResourceInstructions_CREATE,
			indexInGroup: index,
			parent:       parent,
		})
	}
	return resources, nil
}

func (d *ResourceDefServerGroup) desiredPlan(ctx *RequestContext, currentCount int) (*ServerGroupPlan, error) {
	if ctx.Request().resourceName != d.Name() {
		return &ServerGroupPlan{Size: currentCount}, nil
	}
	plans := d.resourcePlans()
	plan, err := desiredPlan(ctx, currentCount, plans)
	if err != nil {
		return nil, err
	}
	if plan != nil {
		if v, ok := plan.(*ServerGroupPlan); ok {
			return v, nil
		}
		return nil, fmt.Errorf("invalid plan: %#v", plan)
	}
	return &ServerGroupPlan{Size: currentCount}, nil
}

func (d *ResourceDefServerGroup) resourceIndex(resource Resource) int {
	for i := 0; i < d.MaxSize; i++ {
		name := d.serverNameByIndex(i)
		if resource.(*ResourceServerGroupInstance).server.Name == name {
			return i
		}
	}
	return 0
}

func (d *ResourceDefServerGroup) serverNameByIndex(index int) string {
	nameFormat := d.ServerNameFormat
	if nameFormat == "" {
		nameFormat = "%s-%03d"
	}
	return fmt.Sprintf(nameFormat, d.namePrefix(), index+1)
}

// determineServerName resourcesから次に追加すべきサーバの名前を決定する
//
// resourcesには連番を割り当てるが、途中抜け([*-001,*-003]のようなパターン)があった場合は抜けている番号から割り当てられる(この例だと*-002)
func (d *ResourceDefServerGroup) determineServerName(resources Resources) (string, int) {
	for i := range resources {
		name := d.serverNameByIndex(i)
		exist := false
		for _, r := range resources {
			if r.(*ResourceServerGroupInstance).server.Name == name {
				exist = true
				break
			}
		}
		if !exist { // 連番の途中抜けパターン
			return name, i
		}
	}
	return d.serverNameByIndex(len(resources)), len(resources)
}

// determineZone サーバのインデックスからサーバを配置すべきゾーンを決定する
// 現在はインデックスのみ考慮しており、特定ゾーンへのサーバの偏りがあっても考慮されない
// d.Zonesが空だとpanicする
func (d *ResourceDefServerGroup) determineZone(index int) string {
	switch len(d.Zones) {
	case 0:
		panic("invalid zones")
	case 1:
		return d.Zones[0]
	default:
		return d.Zones[index%len(d.Zones)]
	}
}

func (d *ResourceDefServerGroup) findCloudResources(ctx context.Context, apiClient iaas.APICaller) ([]*iaas.Server, error) {
	serverOp := iaas.NewServerOp(apiClient)
	selector := &ResourceSelector{Names: []string{d.namePrefix()}}

	var servers []*iaas.Server
	for _, zone := range d.Zones {
		found, err := serverOp.Find(ctx, zone, selector.findCondition())
		if err != nil {
			return nil, fmt.Errorf("computing status failed: %s", err)
		}

		// Nameとd.namePrefix()が前方一致するリソースだけに絞る
		servers = append(servers, d.filterCloudServers(found.Servers)...)
	}

	// 名前の昇順にソート
	sort.Slice(servers, func(i, j int) bool {
		return servers[i].Name < servers[j].Name
	})
	return servers, nil
}

func (d *ResourceDefServerGroup) filterCloudServers(servers []*iaas.Server) []*iaas.Server {
	var filtered []*iaas.Server
	for _, server := range servers {
		if strings.HasPrefix(server.Name, d.namePrefix()) {
			filtered = append(filtered, server)
		}
	}
	return filtered
}

// LastModifiedAt この定義が対象とするリソース(群)の最終更新日時を返す
//
// ServerGroupではModifiedAt or Instance.StatusChangedAtの最も遅い時刻を返す
func (d *ResourceDefServerGroup) LastModifiedAt(ctx *RequestContext, apiClient iaas.APICaller) (time.Time, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return time.Time{}, err
	}
	return d.lastModifiedAt(cloudResources), nil
}

func (d *ResourceDefServerGroup) lastModifiedAt(cloudResources []*iaas.Server) time.Time {
	last := time.Time{}
	for _, r := range cloudResources {
		times := []time.Time{
			r.ModifiedAt,
			r.InstanceStatusChangedAt,
		}
		for _, t := range times {
			if t.After(last) {
				last = t
			}
		}
	}
	return last
}
