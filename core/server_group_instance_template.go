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
	"net"
	"strings"

	"github.com/c-robinson/iplib"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/query"
	"github.com/sacloud/iaas-api-go/ostype"
	"github.com/sacloud/iaas-api-go/types"
)

type ServerGroupInstanceTemplate struct {
	Tags        []string `yaml:"tags" validate:"unique,max=10,dive,max=32"`
	UseGroupTag bool     `yaml:"use_group_tag"`

	Description string `yaml:"description" validate:"max=512"`

	IconId string              `yaml:"icon_id"`
	Icon   *IdOrNameOrSelector `yaml:"icon"`

	CDROMId string              `yaml:"cdrom_id"`
	CDROM   *IdOrNameOrSelector `yaml:"cdrom"`

	PrivateHostId string              `yaml:"private_host_id"`
	PrivateHost   *IdOrNameOrSelector `yaml:"private_host"`

	InterfaceDriver types.EInterfaceDriver `yaml:"interface_driver" validate:"omitempty,oneof=virtio e1000"`

	Plan              *ServerGroupInstancePlan     `yaml:"plan" validate:"required"`
	Disks             []*ServerGroupDiskTemplate   `yaml:"disks" validate:"max=4"`
	EditParameter     *ServerGroupDiskEditTemplate `yaml:"edit_parameter"`
	CloudConfig       ServerGroupCloudConfig       `yaml:",inline"`
	NetworkInterfaces []*ServerGroupNICTemplate    `yaml:"network_interfaces" validate:"max=10"`
}

// Validate .
func (s *ServerGroupInstanceTemplate) Validate(ctx context.Context, apiClient iaas.APICaller, def *ResourceDefServerGroup) []error {
	if errs := validate.StructWithMultiError(s); len(errs) > 0 {
		return errs
	}

	errors := &multierror.Error{}
	if s.IconId != "" && s.Icon != nil {
		errors = multierror.Append(errors, validate.Errorf("only one of icon and icon_id can be specified"))
	}
	if s.IconId != "" {
		loadCtx, ok := ctx.(*config.LoadConfigContext)
		if ok {
			loadCtx.Logger().Warn("icon_id is deprecated. use icon instead")
		}
	}

	if s.CDROMId != "" && s.CDROM != nil {
		errors = multierror.Append(errors, validate.Errorf("only one of cdrom and cdrom_id can be specified"))
	}
	if s.CDROMId != "" {
		loadCtx, ok := ctx.(*config.LoadConfigContext)
		if ok {
			loadCtx.Logger().Warn("cdrom_id is deprecated. use cdrom instead")
		}
	}

	if s.PrivateHostId != "" && s.PrivateHost != nil {
		errors = multierror.Append(errors, validate.Errorf("only one of private_host and private_host_id can be specified"))
	}
	if s.PrivateHostId != "" {
		loadCtx, ok := ctx.(*config.LoadConfigContext)
		if ok {
			loadCtx.Logger().Warn("private_host_id is deprecated. use private_host instead")
		}
	}

	for _, zone := range def.Zones {
		if err := s.Plan.Validate(ctx, apiClient, zone); err != nil {
			errors = multierror.Append(errors, err)
		}
		for _, disk := range s.Disks {
			errors = multierror.Append(errors, disk.Validate(ctx, apiClient, zone)...)
		}
	}

	// TODO EditParameter/CloudConfigそれぞれにおいて、Disks[0]が存在&対応していることを検証
	//  https://github.com/sacloud/autoscaler/issues/255 の対応時に合わせて対応する。

	switch {
	case s.EditParameter != nil && !s.CloudConfig.Empty():
		errors = multierror.Append(errors, fmt.Errorf("only one of edit_parameter and cloud_config can be specified"))
	case s.EditParameter != nil:
		errors = multierror.Append(errors, s.EditParameter.Validate()...)
	case !s.CloudConfig.Empty():
		errors = multierror.Append(errors, s.CloudConfig.Validate()...)
	}

	for i, nic := range s.NetworkInterfaces {
		errors = multierror.Append(errors, nic.Validate(ctx, def.ParentDef, def.MaxSize, i)...)
	}

	return errors.Errors
}

var groupSpecialTags = []string{"@group=a", "@group=b", "@group=c", "@group=d"}

func (s *ServerGroupInstanceTemplate) CalculateTagsByIndex(serverIndexWithinGroup int, zoneLength int) []string {
	if !s.UseGroupTag {
		return s.Tags
	}
	if serverIndexWithinGroup < 0 || zoneLength <= 0 {
		return s.Tags
	}
	for _, t := range s.Tags {
		if strings.HasPrefix(t, "@group=") {
			return s.Tags // 既に@groupタグを持っていたらテンプレートのまま返す
		}
	}
	indexWithinZones := serverIndexWithinGroup / zoneLength // ゾーン内でのインデックス
	return append(s.Tags, groupSpecialTags[indexWithinZones%len(groupSpecialTags)])
}

func (s *ServerGroupInstanceTemplate) FindIconId(ctx context.Context, apiClient iaas.APICaller) (string, error) {
	if s.IconId != "" {
		return s.IconId, nil
	}
	if s.Icon != nil {
		found, err := iaas.NewIconOp(apiClient).Find(ctx, s.Icon.findCondition())
		if err != nil {
			return "", err
		}
		if len(found.Icons) == 0 {
			return "", nil
		}
		return found.Icons[0].ID.String(), nil
	}
	return "", nil
}

func (s *ServerGroupInstanceTemplate) FindCDROMId(ctx context.Context, apiClient iaas.APICaller, zone string) (string, error) {
	if s.CDROMId != "" {
		return s.CDROMId, nil
	}
	if s.CDROM != nil {
		found, err := iaas.NewCDROMOp(apiClient).Find(ctx, zone, s.CDROM.findCondition())
		if err != nil {
			return "", err
		}
		if len(found.CDROMs) == 0 {
			return "", nil
		}
		return found.CDROMs[0].ID.String(), nil
	}
	return "", nil
}

func (s *ServerGroupInstanceTemplate) FindPrivateHostId(ctx context.Context, apiClient iaas.APICaller, zone string) (string, error) {
	if s.PrivateHostId != "" {
		return s.PrivateHostId, nil
	}
	if s.PrivateHost != nil {
		found, err := iaas.NewPrivateHostOp(apiClient).Find(ctx, zone, s.PrivateHost.findCondition())
		if err != nil {
			return "", err
		}
		if len(found.PrivateHosts) == 0 {
			return "", nil
		}
		return found.PrivateHosts[0].ID.String(), nil
	}
	return "", nil
}

type ServerGroupInstancePlan struct {
	Core         int    `yaml:"core"`
	Memory       int    `yaml:"memory"`
	GPU          int    `yaml:"gpu"`
	CPUModel     string `yaml:"cpu_model"`
	DedicatedCPU bool   `yaml:"dedicated_cpu"`
}

func (p *ServerGroupInstancePlan) Validate(ctx context.Context, apiClient iaas.APICaller, zone string) error {
	_, err := query.FindServerPlan(ctx, iaas.NewServerPlanOp(apiClient), zone, &query.FindServerPlanRequest{
		CPU:        p.Core,
		MemoryGB:   p.Memory,
		GPU:        p.GPU,
		CPUModel:   p.CPUModel,
		Commitment: boolToCommitment(p.DedicatedCPU),
		Generation: types.PlanGenerations.Default,
	})
	if err != nil {
		// NOTE: 本来はエラー内容に応じてValidationErrorを返すべきだが、
		//       query.FindServerPlanの実装がエラー内容の判定を行えるようになっていないためここでは判定していない
		return validate.Errorf("plan {%s} not found: %s", p.String(), err)
	}
	return nil
}

func (p *ServerGroupInstancePlan) String() string {
	return fmt.Sprintf("Core:%d, Memory:%d, DedicatedCPU:%t, GPU:%d, CPUModel:%s",
		p.Core, p.Memory, p.DedicatedCPU, p.GPU, p.CPUModel)
}

type ServerGroupDiskTemplate struct {
	NamePrefix  string   `yaml:"name_prefix"` // {{.ServerName}}{{.Name}}{{.Number}}
	NameFormat  string   `yaml:"name_format"`
	Tags        []string `yaml:"tags" validate:"unique,max=10,dive,max=32"`
	Description string   `yaml:"description" validate:"max=512"`

	IconId string              `yaml:"icon_id"`
	Icon   *IdOrNameOrSelector `yaml:"icon"`

	// ブランクディスクの場合は以下3つをゼロ値にする
	SourceArchiveSelector *NameOrSelector `yaml:"source_archive"`
	SourceDiskSelector    *NameOrSelector `yaml:"source_disk"`
	OSType                string          `yaml:"os_type"`

	Plan       string `yaml:"plan" validate:"omitempty,oneof=ssd hdd"`
	Connection string `yaml:"connection" validate:"omitempty,oneof=virtio ide"`
	Size       int    `yaml:"size"`
}

func (t *ServerGroupDiskTemplate) DiskName(serverName string, index int) string {
	nameFormat := t.NameFormat
	if nameFormat == "" {
		nameFormat = "%s-disk%03d"
	}
	if t.NamePrefix != "" {
		nameFormat = "%s-%d"
	}

	return fmt.Sprintf(nameFormat, serverName, index+1)
}

// HostName HostNamePrefixとindexからホスト名を算出する
//
// HostNamePrefixが空の場合はserverNameをそのまま返す
func (t *ServerGroupDiskEditTemplate) HostName(serverName string, index int) string {
	if t.HostNameFormat == "" && t.HostNamePrefix == "" {
		return serverName
	}

	nameFormat := t.HostNameFormat
	if nameFormat == "" {
		nameFormat = "%s-%03d"
	}
	if t.HostNamePrefix != "" {
		nameFormat = "%s-%03d"
	}
	return fmt.Sprintf(nameFormat, serverName, index+1)
}

func (t *ServerGroupDiskTemplate) Validate(ctx context.Context, apiClient iaas.APICaller, zone string) []error {
	if errs := validate.StructWithMultiError(t); len(errs) > 0 {
		return errs
	}

	errors := &multierror.Error{}
	if t.SourceArchiveSelector != nil {
		if err := t.SourceArchiveSelector.Validate(); err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	if t.SourceDiskSelector != nil {
		if err := t.SourceDiskSelector.Validate(); err != nil {
			errors = multierror.Append(errors, err)
		}
	}
	if t.NamePrefix != "" && t.NameFormat != "" {
		errors = multierror.Append(errors, validate.Errorf("only one of name_prefix and name_format can be specified"))
	}

	if t.IconId != "" && t.Icon != nil {
		errors = multierror.Append(errors, validate.Errorf("only one of icon and icon_id can be specified"))
	}
	if t.IconId != "" {
		loadCtx, ok := ctx.(*config.LoadConfigContext)
		if ok {
			loadCtx.Logger().Warn("icon_id is deprecated. use icon instead")
		}
	}

	if _, _, err := t.FindDiskSource(ctx, apiClient, zone); err != nil {
		errors = multierror.Append(errors, err)
	}

	// TODO プラン/サイズがクラウド上で有効な値になっているか検証

	return errors.Errors
}

func (t *ServerGroupDiskTemplate) FindIconID(ctx context.Context, apiClient iaas.APICaller) (string, error) {
	if t.IconId != "" {
		return t.IconId, nil
	}
	if t.Icon != nil {
		found, err := iaas.NewIconOp(apiClient).Find(ctx, t.Icon.findCondition())
		if err != nil {
			return "", err
		}
		if len(found.Icons) == 0 {
			return "", nil
		}
		return found.Icons[0].ID.String(), nil
	}
	return "", nil
}

func (t *ServerGroupDiskTemplate) FindDiskSource(ctx context.Context, apiClient iaas.APICaller, zone string) (sourceArchiveID, sourceDiskID string, retErr error) {
	switch {
	case t.OSType != "":
		archive, err := query.FindArchiveByOSType(ctx, iaas.NewArchiveOp(apiClient), zone, ostype.StrToOSType(t.OSType))
		if err != nil {
			// NOTE: 本来はエラー内容に応じてValidationErrorを返すべきだが、
			//       query.FindArchiveByOSTypeの実装がエラー内容の判定を行えるようになっていないためここでは判定していない
			retErr = validate.New(err)
			return
		}
		sourceArchiveID = archive.ID.String()
		return
	case t.SourceArchiveSelector != nil:
		found, err := iaas.NewArchiveOp(apiClient).Find(ctx, zone, t.SourceArchiveSelector.findCondition())
		if err != nil {
			retErr = err
			return
		}
		if len(found.Archives) == 0 {
			retErr = validate.Errorf("source archive not found with: {zone: %s, %v}", zone, t.SourceArchiveSelector)
			return
		}
		if len(found.Archives) > 1 {
			retErr = validate.Errorf("multiple source archive found with: {zone: %s, %v}, archives: %v", zone, t.SourceArchiveSelector, found.Archives)
			return
		}
		sourceArchiveID = found.Archives[0].ID.String()
		return
	case t.SourceDiskSelector != nil:
		found, err := iaas.NewDiskOp(apiClient).Find(ctx, zone, t.SourceDiskSelector.findCondition())
		if err != nil {
			retErr = err
			return
		}
		if len(found.Disks) == 0 {
			retErr = validate.Errorf("source disk not found with: {zone: %s, %v}", zone, t.SourceDiskSelector)
			return
		}
		if len(found.Disks) > 1 {
			retErr = validate.Errorf("multiple source disk found with: {zone: %s, %v}, disks: %v", zone, t.SourceDiskSelector, found.Disks)
			return
		}
		sourceDiskID = found.Disks[0].ID.String()
		return
	}
	// blank disk: 2番目以降のディスクや別途Tinkerbellなどのベアメタルプロビジョニングを行う場合などに到達し得る
	return "", "", nil
}

type ServerGroupDiskEditTemplate struct {
	Disabled bool `yaml:"disabled"` // ディスクの修正を行わない場合にtrue

	HostNamePrefix string `yaml:"host_name_prefix"` // からの場合は{{ .ServerName }} 、そうでなければ {{ .HostNamePrefix }}{{ .Number }}
	HostNameFormat string `yaml:"host_name_format"`

	Password            string                    `yaml:"password"` // グループ内のサーバは全部一緒になるが良いか??
	DisablePasswordAuth bool                      `yaml:"disable_pw_auth"`
	EnableDHCP          bool                      `yaml:"enable_dhcp"`
	ChangePartitionUUID bool                      `yaml:"change_partition_uuid"`
	StartupScripts      []config.StringOrFilePath `yaml:"startup_scripts"`

	SSHKeys []config.StringOrFilePath `yaml:"ssh_keys"`
}

func (t *ServerGroupDiskEditTemplate) Validate() []error {
	hasValue := t.HostNamePrefix != "" ||
		t.Password != "" ||
		len(t.StartupScripts) > 0 ||
		len(t.SSHKeys) > 0

	if t.Disabled && hasValue {
		return []error{validate.Errorf("disabled=true but a value is specified")}
	}

	if t.HostNamePrefix != "" && t.HostNameFormat != "" {
		return []error{validate.Errorf("only one of host_name_prefix and host_name_format can be specified")}
	}
	return nil
}

type ServerGroupCloudConfig struct {
	CloudConfig config.StringOrFilePath `yaml:"cloud_config"`
}

func (c ServerGroupCloudConfig) String() string {
	return c.CloudConfig.String()
}

func (c ServerGroupCloudConfig) Empty() bool {
	return c.CloudConfig.String() == ""
}

func (c ServerGroupCloudConfig) Validate() []error {
	var m map[string]interface{}
	opts := []yaml.DecodeOption{yaml.Strict(), yaml.DisallowDuplicateKey()}
	if err := yaml.UnmarshalWithOptions([]byte(c.CloudConfig.String()), &m, opts...); err != nil {
		return []error{validate.Errorf("invalid cloud-config: %s", err)}
	}
	return nil
}

type ServerGroupNICTemplate struct {
	Upstream         *ServerGroupNICUpstream `yaml:"upstream" validate:"required"`                         // "shared" or *ResourceSelector
	AssignCidrBlock  string                  `yaml:"assign_cidr_block" validate:"omitempty,cidrv4"`        // 上流がスイッチの場合(ルータ含む)に割り当てるIPアドレスのCIDRブロック
	AssignNetMaskLen int                     `yaml:"assign_netmask_len" validate:"omitempty,min=1,max=32"` // 上流がスイッチの場合(ルータ含む)に割り当てるサブネットマスク長
	DefaultRoute     string                  `yaml:"default_route" validate:"omitempty,ipv4"`

	PacketFilterId string              `yaml:"packet_filter_id"`
	PacketFilter   *IdOrNameOrSelector `yaml:"packet_filter"`

	ExposeInfo *ServerGroupNICMetadata `yaml:"expose"`
}

func (t *ServerGroupNICTemplate) Validate(ctx context.Context, parent *ParentResourceDef, maxServerNum, nicIndex int) []error {
	if errs := validate.StructWithMultiError(t); len(errs) > 0 {
		return errs
	}

	errors := &multierror.Error{}
	if t.PacketFilterId != "" && t.PacketFilter != nil {
		errors = multierror.Append(errors, validate.Errorf("only one of packet_filter and packet_filter_id can be specified"))
	}
	if t.PacketFilterId != "" {
		loadCtx, ok := ctx.(*config.LoadConfigContext)
		if ok {
			loadCtx.Logger().Warn("icon_id is deprecated. use icon instead")
		}
	}

	hasNetworkSettings := t.AssignCidrBlock != "" || t.AssignNetMaskLen > 0 || t.DefaultRoute != ""
	if t.Upstream.UpstreamShared() && hasNetworkSettings {
		return []error{validate.Errorf("upstream=shared but network settings are specified")}
	}

	if hasNetworkSettings {
		ip, ipNet, err := net.ParseCIDR(t.AssignCidrBlock)
		if err != nil {
			return []error{validate.Errorf("invalid cidr block")}
		}
		maskLen, _ := ipNet.Mask.Size()
		if iplib.NewNet4(ip, maskLen).Count() < uint32(maxServerNum) {
			errors = multierror.Append(errors, validate.Errorf("assign_cidr_block is too small"))
		}
		if t.AssignNetMaskLen != 0 {
			maskLen = t.AssignNetMaskLen
		}
		if t.DefaultRoute != "" {
			assignedNet := iplib.NewNet(ipNet.IP, maskLen)
			if !assignedNet.Contains(net.ParseIP(t.DefaultRoute)) {
				errors = multierror.Append(errors,
					validate.Errorf(
						"default_route and assigned_address must be in the same network: assign_cidr_block:%s, assign_netmask_len:%d, default_route:%s",
						t.AssignCidrBlock,
						t.AssignNetMaskLen,
						t.DefaultRoute,
					))
			}
		}
	}

	if t.ExposeInfo != nil {
		errors = multierror.Append(errors, t.ExposeInfo.Validate(parent, nicIndex)...)
	}
	return errors.Errors
}

func (t *ServerGroupNICTemplate) FindPacketFilterId(ctx context.Context, apiClient iaas.APICaller, zone string) (string, error) {
	if t.PacketFilterId != "" {
		return t.PacketFilterId, nil
	}
	if t.PacketFilter != nil {
		found, err := iaas.NewPacketFilterOp(apiClient).Find(ctx, zone, t.PacketFilter.findCondition())
		if err != nil {
			return "", err
		}
		if len(found.PacketFilters) == 0 {
			return "", nil
		}
		return found.PacketFilters[0].ID.String(), nil
	}
	return "", nil
}

// IPAddressByIndexFromCidrBlock AssignCidrBlockからindexに対応するIPアドレスを返す
//
// 戻り値: IPアドレス, マスク長, エラー
func (t *ServerGroupNICTemplate) IPAddressByIndexFromCidrBlock(index int) (string, int, error) {
	if t.Upstream == nil || t.Upstream.UpstreamShared() || t.AssignCidrBlock == "" {
		return "", -1, nil
	}

	ip, ipNet, err := net.ParseCIDR(t.AssignCidrBlock)
	if err != nil {
		return "", -1, err
	}
	mask, _ := ipNet.Mask.Size()
	if t.AssignNetMaskLen > 0 {
		mask = t.AssignNetMaskLen
	}
	newIP := iplib.IncrementIP4By(ip, uint32(index+1))

	if !ipNet.Contains(newIP) {
		return "", -1, fmt.Errorf("next ip %s is not in cidr block: %s", newIP.String(), t.AssignCidrBlock)
	}

	return newIP.String(), mask, nil
}

type ServerGroupNICUpstream struct {
	raw      []byte
	shared   bool
	selector *ResourceSelector
}

func (s *ServerGroupNICUpstream) UnmarshalYAML(ctx context.Context, data []byte) error {
	if string(data) == "shared" {
		*s = ServerGroupNICUpstream{raw: data, shared: true}
		return nil
	}
	var selector ResourceSelector
	if err := yaml.UnmarshalWithOptions(data, &selector, yaml.Strict()); err != nil {
		return err
	}
	*s = ServerGroupNICUpstream{raw: data, shared: false, selector: &selector}
	return nil
}

func (s *ServerGroupNICUpstream) UpstreamShared() bool {
	return s.shared
}

func (s *ServerGroupNICUpstream) Selector() *ResourceSelector {
	if s.shared {
		return nil
	}
	return s.selector
}

// ServerGroupNICMetadata 上流リソースの操作時に参照されるメタデータ
type ServerGroupNICMetadata struct {
	// Ports 公開するポート
	Ports []int `yaml:"ports" validate:"unique,dive,min=1,max=65535"`

	// ServerGroupName ELBの実サーバとして登録する場合のサーバグループ名
	ServerGroupName string `yaml:"server_group_name"`

	// Weight GSLBに実サーバとして登録する場合の重み値
	Weight int `yaml:"weight"`

	// VIP LBに実サーバとして登録する場合の対象VIPリスト
	//
	// 省略した場合はこのメタデータがつけられたNICと同じネットワーク内に存在するVIP全てが対象となる
	VIPs []string `yaml:"vips" validate:"omitempty,dive,ipv4"`

	// HealthCheck LBに実サーバとして登録する場合のヘルスチェック
	HealthCheck *ServerGroupNICMetadataHealthCheck `yaml:"health_check"`

	// RecordName DNSにレコード登録する場合のレコード名
	RecordName string `yaml:"record_name"`

	// RecordTTL DNSにレコード登録する場合のTTL
	RecordTTL int `yaml:"record_ttl" validate:"omitempty,min=10,max=3600000"`
}

func (m *ServerGroupNICMetadata) Validate(parent *ParentResourceDef, nicIndex int) []error {
	if errs := validate.StructWithMultiError(m); len(errs) > 0 {
		return errs
	}
	errors := &multierror.Error{}
	if nicIndex > 0 {
		// グローバルIPを要求する項目がNIC[0]以外で指定されていたらエラーとする
		format := "%s: can only be specified for the first NIC"
		if m.ServerGroupName != "" {
			errors = multierror.Append(errors, validate.Errorf(format, "server_group_name"))
		}
		if m.Weight > 0 {
			errors = multierror.Append(errors, validate.Errorf(format, "weight"))
		}
		if m.RecordName != "" {
			errors = multierror.Append(errors, validate.Errorf(format, "record_name"))
		}
		if m.RecordTTL > 0 {
			errors = multierror.Append(errors, validate.Errorf(format, "record_ttl"))
		}
	}

	if parent != nil {
		format := "%s: can't specify if parent resource type is %s"
		requiredFormat := "%s: required when parent is %s"

		switch parent.Type() {
		case ResourceTypeELB:
			if len(m.Ports) == 0 {
				errors = multierror.Append(errors, validate.Errorf(requiredFormat, "ports", ResourceTypeELB))
			}
			if m.Weight > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "weight", ResourceTypeELB))
			}
			if len(m.VIPs) > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "vips", ResourceTypeELB))
			}
			if m.HealthCheck != nil {
				errors = multierror.Append(errors, validate.Errorf(format, "health_check", ResourceTypeELB))
			}
			if m.RecordName != "" {
				errors = multierror.Append(errors, validate.Errorf(format, "record_name", ResourceTypeELB))
			}
			if m.RecordTTL > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "record_ttl", ResourceTypeELB))
			}
		case ResourceTypeGSLB:
			if m.ServerGroupName != "" {
				errors = multierror.Append(errors, validate.Errorf(format, "server_group_name", ResourceTypeGSLB))
			}
			if len(m.VIPs) > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "vips", ResourceTypeGSLB))
			}
			if m.HealthCheck != nil {
				errors = multierror.Append(errors, validate.Errorf(format, "health_check", ResourceTypeGSLB))
			}
			if m.RecordName != "" {
				errors = multierror.Append(errors, validate.Errorf(format, "record_name", ResourceTypeGSLB))
			}
			if m.RecordTTL > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "record_ttl", ResourceTypeGSLB))
			}
		case ResourceTypeLoadBalancer:
			if len(m.Ports) == 0 {
				errors = multierror.Append(errors, validate.Errorf(requiredFormat, "ports", ResourceTypeLoadBalancer))
			}
			if m.HealthCheck == nil {
				errors = multierror.Append(errors, validate.Errorf(requiredFormat, "health_check", ResourceTypeLoadBalancer))
			}
			if m.ServerGroupName != "" {
				errors = multierror.Append(errors, validate.Errorf(format, "server_group_name", ResourceTypeLoadBalancer))
			}
			if m.Weight > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "weight", ResourceTypeLoadBalancer))
			}
			if m.RecordName != "" {
				errors = multierror.Append(errors, validate.Errorf(format, "record_name", ResourceTypeLoadBalancer))
			}
			if m.RecordTTL > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "record_ttl", ResourceTypeLoadBalancer))
			}
		case ResourceTypeDNS:
			if m.ServerGroupName != "" {
				errors = multierror.Append(errors, validate.Errorf(format, "server_group_name", ResourceTypeDNS))
			}
			if m.Weight > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "weight", ResourceTypeDNS))
			}
			if len(m.VIPs) > 0 {
				errors = multierror.Append(errors, validate.Errorf(format, "vips", ResourceTypeDNS))
			}
			if m.HealthCheck != nil {
				errors = multierror.Append(errors, validate.Errorf(format, "health_check", ResourceTypeDNS))
			}
		}
	}

	if m.HealthCheck != nil {
		errors = multierror.Append(errors, m.HealthCheck.Validate()...)
	}

	return errors.Errors
}

func (m *ServerGroupNICMetadata) ToExposeInfo() *handler.ServerGroupInstance_ExposeInfo {
	var ports []uint32
	for _, p := range m.Ports {
		ports = append(ports, uint32(p))
	}

	var healthCheck *handler.ServerGroupInstance_HealthCheck
	if m.HealthCheck != nil {
		healthCheck = &handler.ServerGroupInstance_HealthCheck{
			Protocol:   m.HealthCheck.Protocol,
			Path:       m.HealthCheck.Path,
			StatusCode: uint32(m.HealthCheck.StatusCode),
		}
	}

	return &handler.ServerGroupInstance_ExposeInfo{
		Ports:           ports,
		ServerGroupName: m.ServerGroupName,
		Weight:          uint32(m.Weight),
		Vips:            m.VIPs,
		HealthCheck:     healthCheck,
		RecordName:      m.RecordName,
		Ttl:             uint32(m.RecordTTL),
	}
}

type ServerGroupNICMetadataHealthCheck struct {
	Protocol   string `yaml:"protocol" validate:"required,oneof=http https ping tcp"`
	Path       string `yaml:"path"`
	StatusCode int    `yaml:"status_code"`
}

func (h *ServerGroupNICMetadataHealthCheck) Validate() []error {
	if errs := validate.StructWithMultiError(h); len(errs) > 0 {
		return errs
	}

	errors := &multierror.Error{}

	switch h.Protocol {
	case "http", "https":
		if h.Path == "" {
			errors = multierror.Append(errors, validate.Errorf("path: required if protocol is http or https"))
		}
		if h.StatusCode == 0 {
			errors = multierror.Append(errors, validate.Errorf("status_code: required if protocol is http or https"))
		}
	default:
		if h.Path != "" {
			errors = multierror.Append(errors, validate.Errorf("path: can not be specified if protocol is not http or https"))
		}
		if h.StatusCode > 0 {
			errors = multierror.Append(errors, validate.Errorf("status_code: can not be specified if protocol is not http or https"))
		}
	}

	return errors.Errors
}
