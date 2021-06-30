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

	"github.com/sacloud/libsacloud/v2/sacloud"

	"github.com/c-robinson/iplib"
	"github.com/goccy/go-yaml"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ServerGroupInstanceTemplate struct {
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`

	IconID          string                 `yaml:"icon_id"`
	CDROMID         string                 `yaml:"cdrom_id"`
	PrivateHostID   string                 `yaml:"private_host_id"`
	InterfaceDriver types.EInterfaceDriver `yaml:"interface_driver"`

	Plan              *ServerGroupInstancePlan     `yaml:"plan"`
	Disks             []*ServerGroupDiskTemplate   `yaml:"disks"`
	EditParameter     *ServerGroupDiskEditTemplate `yaml:"edit_parameter"`
	NetworkInterfaces []*ServerGroupNICTemplate    `yaml:"network_interfaces"`
}

func (s *ServerGroupInstanceTemplate) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	// TODO 実装
	return nil
}

type ServerGroupInstancePlan struct {
	Core         int  `yaml:"core"`
	Memory       int  `yaml:"memory"`
	DedicatedCPU bool `yaml:"dedicated_cpu"`
}

type ServerGroupDiskTemplate struct {
	NamePrefix  string   `yaml:"name_prefix"` // {{.ServerName}}{{.Name}}{{.Number}}
	Tags        []string `yaml:"tags"`
	Description string   `yaml:"description"`
	IconID      string   `yaml:"icon_id"`

	SourceArchiveSelector *ResourceSelector `yaml:"source_archive"`
	SourceDiskSelector    *ResourceSelector `yaml:"source_disk"`
	OSType                string

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

type ServerGroupDiskEditTemplate struct {
	Disabled            bool     `yaml:"disabled"`         // ディスクの修正を行わない場合にtrue
	HostNamePrefix      string   `yaml:"host_name_prefix"` // からの場合は{{ .ServerName }} 、そうでなければ {{ .HostNamePrefix }}{{ .Number }}
	Password            string   `yaml:"password"`         // グループ内のサーバは全部一緒になるが良いか??
	DisablePasswordAuth bool     `yaml:"disable_pw_auth"`
	EnableDHCP          bool     `yaml:"enable_dhcp"`
	ChangePartitionUUID bool     `yaml:"change_partition_uuid"`
	StartupScripts      []string `yaml:"startup_scripts"`

	SSHKeys   []string `yaml:"ssh_keys"`
	SSHKeyIDs []string `yaml:"ssh_key_ids"`
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

type ServerGroupNICTemplate struct {
	Upstream         *ServerGroupNICUpstream `yaml:"upstream"`           // "shared" or *ResourceSelector
	AssignCidrBlock  string                  `yaml:"assign_cidr_block"`  // 上流がスイッチの場合(ルータ含む)に割り当てるIPアドレスのCIDRブロック
	AssignNetMaskLen int                     `yaml:"assign_netmask_len"` // 上流がスイッチの場合(ルータ含む)に割り当てるサブネットマスク長
	DefaultRoute     string                  `yaml:"default_route"`
	PacketFilterID   string                  `yaml:"packet_filter_id"`
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
