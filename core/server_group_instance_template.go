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
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/libsacloud/v2/helper/query"
	"github.com/sacloud/libsacloud/v2/sacloud/ostype"

	"github.com/sacloud/libsacloud/v2/sacloud"

	"github.com/c-robinson/iplib"
	"github.com/goccy/go-yaml"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ServerGroupInstanceTemplate struct {
	Tags        []string `yaml:"tags" validate:"unique,max=10,dive,max=32"`
	Description string   `yaml:"description" validate:"max=512"`

	IconID          string                 `yaml:"icon_id"`
	CDROMID         string                 `yaml:"cdrom_id"`
	PrivateHostID   string                 `yaml:"private_host_id"`
	InterfaceDriver types.EInterfaceDriver `yaml:"interface_driver" validate:"omitempty,oneof=virtio e1000"`

	Plan              *ServerGroupInstancePlan     `yaml:"plan" validate:"required"`
	Disks             []*ServerGroupDiskTemplate   `yaml:"disks" validate:"max=4"`
	EditParameter     *ServerGroupDiskEditTemplate `yaml:"edit_parameter"`
	NetworkInterfaces []*ServerGroupNICTemplate    `yaml:"network_interfaces" validate:"max=10"`
}

// Validate .
func (s *ServerGroupInstanceTemplate) Validate(ctx context.Context, apiClient sacloud.APICaller, def *ResourceDefServerGroup) []error {
	if errs := validate.StructWithMultiError(s); len(errs) > 0 {
		return errs
	}

	errors := &multierror.Error{}
	if err := s.Plan.Validate(ctx, apiClient, def.Zone); err != nil {
		errors = multierror.Append(errors, err)
	}
	for _, disk := range s.Disks {
		errors = multierror.Append(errors, disk.Validate(ctx, apiClient, def.Zone)...)
	}
	if s.EditParameter != nil {
		errors = multierror.Append(errors, s.EditParameter.Validate()...)
	}
	for _, nic := range s.NetworkInterfaces {
		errors = multierror.Append(errors, nic.Validate(def.MaxSize)...)
	}

	return errors.Errors
}

type ServerGroupInstancePlan struct {
	Core         int  `yaml:"core"`
	Memory       int  `yaml:"memory"`
	DedicatedCPU bool `yaml:"dedicated_cpu"`
}

func (p *ServerGroupInstancePlan) Validate(ctx context.Context, apiClient sacloud.APICaller, zone string) error {
	_, err := query.FindServerPlan(ctx, sacloud.NewServerPlanOp(apiClient), zone, &query.FindServerPlanRequest{
		CPU:        p.Core,
		MemoryGB:   p.Memory,
		Commitment: boolToCommitment(p.DedicatedCPU),
		Generation: types.PlanGenerations.Default,
	})
	if err != nil {
		return fmt.Errorf("plan {%s} not found: %s", p.String(), err)
	}
	return nil
}

func (p *ServerGroupInstancePlan) String() string {
	return fmt.Sprintf("Core:%d, Memory:%d, DedicatedCPU:%t", p.Core, p.Memory, p.DedicatedCPU)
}

type ServerGroupDiskTemplate struct {
	NamePrefix  string   `yaml:"name_prefix"` // {{.ServerName}}{{.Name}}{{.Number}}
	Tags        []string `yaml:"tags" validate:"unique,max=10,dive,max=32"`
	Description string   `yaml:"description" validate:"max=512"`
	IconID      string   `yaml:"icon_id"`

	// ブランクディスクの場合は以下3つをゼロ値にする
	SourceArchiveSelector *NameOrSelector `yaml:"source_archive"`
	SourceDiskSelector    *NameOrSelector `yaml:"source_disk"`
	OSType                string          `yaml:"os_type"`

	Plan       string `yaml:"plan" validate:"omitempty,oneof=ssd hdd"`
	Connection string `yaml:"connection" validate:"omitempty,oneof=virtio ide"`
	Size       int    `yaml:"size"`
}

func (t *ServerGroupDiskTemplate) DiskName(serverName string, index int) string {
	if t.NamePrefix == "" {
		return fmt.Sprintf("%s-disk%03d", serverName, index+1)
	}
	return fmt.Sprintf("%s-%d", t.NamePrefix, index+1)
}

// HostName HostNamePrefixとindexからホスト名を算出する
//
// HostNamePrefixが空の場合はserverNameをそのまま返す
func (t *ServerGroupDiskEditTemplate) HostName(serverName string, index int) string {
	if t.HostNamePrefix == "" {
		return serverName
	}
	return fmt.Sprintf("%s-%03d", t.HostNamePrefix, index+1)
}

func (t *ServerGroupDiskTemplate) Validate(ctx context.Context, apiClient sacloud.APICaller, zone string) []error {
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

	if _, _, err := t.FindDiskSource(ctx, apiClient, zone); err != nil {
		errors = multierror.Append(errors, err)
	}

	return errors.Errors
}

func (t *ServerGroupDiskTemplate) FindDiskSource(ctx context.Context, apiClient sacloud.APICaller, zone string) (sourceArchiveID, sourceDiskID string, retErr error) {
	switch {
	case t.OSType != "":
		archive, err := query.FindArchiveByOSType(ctx, sacloud.NewArchiveOp(apiClient), zone, ostype.StrToOSType(t.OSType))
		if err != nil {
			retErr = err
			return
		}
		sourceArchiveID = archive.ID.String()
		return
	case t.SourceArchiveSelector != nil:
		found, err := sacloud.NewArchiveOp(apiClient).Find(ctx, zone, t.SourceArchiveSelector.findCondition())
		if err != nil {
			retErr = err
			return
		}
		if len(found.Archives) == 0 {
			retErr = fmt.Errorf("source archive not found with: %s", t.SourceArchiveSelector)
			return
		}
		if len(found.Archives) > 1 {
			retErr = fmt.Errorf("multiple source archive found with: %s, archives: %#v", t.SourceArchiveSelector, found.Archives)
			return
		}
		sourceArchiveID = found.Archives[0].ID.String()
		return
	case t.SourceDiskSelector != nil:
		found, err := sacloud.NewDiskOp(apiClient).Find(ctx, zone, t.SourceDiskSelector.findCondition())
		if err != nil {
			retErr = err
			return
		}
		if len(found.Disks) == 0 {
			retErr = fmt.Errorf("source disk not found with: %s", t.SourceArchiveSelector)
			return
		}
		if len(found.Disks) > 1 {
			retErr = fmt.Errorf("multiple source disk found with: %s, archives: %#v", t.SourceArchiveSelector, found.Disks)
			return
		}
		sourceDiskID = found.Disks[0].ID.String()
		return
	}
	// blank disk: 2番目以降のディスクや別途Tinkerbellなどのベアメタルプロビジョニングを行う場合などに到達し得る
	return "", "", nil
}

type ServerGroupDiskEditTemplate struct {
	Disabled            bool               `yaml:"disabled"`         // ディスクの修正を行わない場合にtrue
	HostNamePrefix      string             `yaml:"host_name_prefix"` // からの場合は{{ .ServerName }} 、そうでなければ {{ .HostNamePrefix }}{{ .Number }}
	Password            string             `yaml:"password"`         // グループ内のサーバは全部一緒になるが良いか??
	DisablePasswordAuth bool               `yaml:"disable_pw_auth"`
	EnableDHCP          bool               `yaml:"enable_dhcp"`
	ChangePartitionUUID bool               `yaml:"change_partition_uuid"`
	StartupScripts      []StringOrFilePath `yaml:"startup_scripts"`

	SSHKeys   []StringOrFilePath `yaml:"ssh_keys"`
	SSHKeyIDs []string           `yaml:"ssh_key_ids"`
}

func (t *ServerGroupDiskEditTemplate) Validate() []error {
	hasValue := t.HostNamePrefix != "" ||
		t.Password != "" ||
		len(t.StartupScripts) > 0 ||
		len(t.SSHKeys) > 0 ||
		len(t.SSHKeyIDs) > 0

	if t.Disabled && hasValue {
		return []error{fmt.Errorf("disabled=true but a value is specified")}
	}
	return nil
}

type ServerGroupNICTemplate struct {
	Upstream         *ServerGroupNICUpstream `yaml:"upstream" validate:"required"`                         // "shared" or *ResourceSelector
	AssignCidrBlock  string                  `yaml:"assign_cidr_block" validate:"omitempty,cidrv4"`        // 上流がスイッチの場合(ルータ含む)に割り当てるIPアドレスのCIDRブロック
	AssignNetMaskLen int                     `yaml:"assign_netmask_len" validate:"omitempty,min=1,max=32"` // 上流がスイッチの場合(ルータ含む)に割り当てるサブネットマスク長
	DefaultRoute     string                  `yaml:"default_route" validate:"omitempty,ipv4"`
	PacketFilterID   string                  `yaml:"packet_filter_id"`
}

func (t *ServerGroupNICTemplate) Validate(maxServerNum int) []error {
	if errs := validate.StructWithMultiError(t); len(errs) > 0 {
		return errs
	}

	errors := &multierror.Error{}
	hasNetworkSettings := t.AssignCidrBlock != "" || t.AssignNetMaskLen > 0 || t.DefaultRoute != ""
	if t.Upstream.UpstreamShared() && hasNetworkSettings {
		return []error{fmt.Errorf("upstream=shared but network settings are specified")}
	}

	if hasNetworkSettings {
		ip, ipNet, err := net.ParseCIDR(t.AssignCidrBlock)
		if err != nil {
			return []error{fmt.Errorf("invalid cidr block")}
		}
		maskLen, _ := ipNet.Mask.Size()
		if iplib.NewNet(ip, maskLen).Count4() < uint32(maxServerNum) {
			errors = multierror.Append(errors, fmt.Errorf("assign_cidr_block is too small"))
		}
		if t.DefaultRoute != "" {
			assignedNet := iplib.NewNet(ip, maskLen)
			if !assignedNet.Contains(net.ParseIP(t.DefaultRoute)) {
				errors = multierror.Append(errors,
					fmt.Errorf(
						"default_route and assigned_address must be in the same network: assign_cidr_block:%s, assign_netmask_len:%d, default_route:%s",
						t.AssignCidrBlock,
						t.AssignNetMaskLen,
						t.DefaultRoute,
					))
			}
		}
	}
	return errors.Errors
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

func (s *ServerGroupNICUpstream) UnmarshalYAML(data []byte) error {
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
