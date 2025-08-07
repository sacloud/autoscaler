// Copyright 2021-2025 The sacloud/autoscaler Authors
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
	"net/http"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/autoscaler/validate"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/stretchr/testify/require"
)

func TestServerGroupInstanceTemplate_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantErrText string
		expect      *ServerGroupInstanceTemplate
	}{
		{
			name: "invalid",
			args: args{
				data: []byte("invalid"),
			},
			wantErr: true,
		},
		{
			name: "minimum",
			args: args{
				data: []byte(`
plan:
  core: 1
  memory: 1
`),
			},
			expect: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
			},
		},
		{
			name: "cloud-config",
			args: args{
				data: []byte(`
plan:
  core: 1
  memory: 1
disks:
  - source_archive:
      names: ["Ubuntu", "cloudimg"]
    size: 20
cloud_config: "#cloud-config"
`),
			},
			expect: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
				Disks: []*ServerGroupDiskTemplate{
					{
						SourceArchiveSelector: &NameOrSelector{
							ResourceSelector: ResourceSelector{
								Names: []string{"Ubuntu", "cloudimg"},
							},
						},
						Size: 20,
					},
				},
				CloudConfig: ServerGroupCloudConfig{
					CloudConfig: test.StringOrFilePath(t, "#cloud-config"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s ServerGroupInstanceTemplate
			if err := yaml.UnmarshalWithOptions(tt.args.data, &s, yaml.Strict()); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expect != nil {
				require.EqualValues(t, tt.expect, &s)
			}
		})
	}
}

func TestServerGroupNICUpstream_UnmarshalYAML(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		expect  *ServerGroupNICUpstream
	}{
		{
			name: "shared",
			args: args{
				data: []byte("shared"),
			},
			wantErr: false,
			expect: &ServerGroupNICUpstream{
				raw:      []byte("shared"),
				shared:   true,
				selector: nil,
			},
		},
		{
			name: "shared with quote",
			args: args{
				data: []byte(`"shared"`),
			},
			wantErr: false,
			expect: &ServerGroupNICUpstream{
				raw:      []byte(`"shared"`),
				shared:   true,
				selector: nil,
			},
		},
		{
			name: "shared with single quote",
			args: args{
				data: []byte(`'shared'`),
			},
			wantErr: false,
			expect: &ServerGroupNICUpstream{
				raw:      []byte(`'shared'`),
				shared:   true,
				selector: nil,
			},
		},
		{
			name: "shared with newline",
			args: args{
				data: []byte("shared\n"),
			},
			wantErr: false,
			expect: &ServerGroupNICUpstream{
				raw:      []byte("shared"),
				shared:   true,
				selector: nil,
			},
		},
		{
			name: "invalid 'shared' string",
			args: args{
				data: []byte(`'sha"red'`),
			},
			wantErr: true,
		},
		{
			name: "selector",
			args: args{
				data: []byte(`names: ["test"]`),
			},
			wantErr: false,
			expect: &ServerGroupNICUpstream{
				raw:    []byte(`names: ["test"]`),
				shared: false,
				selector: &ResourceSelector{
					Names: []string{"test"},
				},
			},
		},
		{
			name: "invalid",
			args: args{
				data: []byte("foobar"),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s ServerGroupNICUpstream
			if err := yaml.UnmarshalWithOptions(tt.args.data, &s, yaml.Strict()); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expect != nil {
				require.EqualValues(t, tt.expect, &s)
			}
		})
	}
}

func TestServerGroupNICTemplate_IPAddressByIndexFromCidrBlock(t1 *testing.T) {
	type fields struct {
		Upstream        *ServerGroupNICUpstream
		AssignCidrBlock string
	}
	type args struct {
		index int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		want1   int
		wantErr bool
	}{
		{
			name: "shared",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{shared: true},
			},
			args: args{
				index: 0,
			},
			want:    "",
			want1:   -1,
			wantErr: false,
		},
		{
			name: "basic switched network",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/24",
			},
			args: args{
				index: 0,
			},
			want:    "192.168.0.1",
			want1:   24,
			wantErr: false,
		},
		{
			name: "basic switched network second address",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/24",
			},
			args: args{
				index: 1,
			},
			want:    "192.168.0.2",
			want1:   24,
			wantErr: false,
		},
		{
			name: "basic switched network with carry error",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/24",
			},
			args: args{
				index: 256,
			},
			want:    "",
			want1:   -1,
			wantErr: true,
		},
		{
			name: "basic switched network without carry error",
			fields: fields{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock: "192.168.0.0/16",
			},
			args: args{
				index: 256,
			},
			want:    "192.168.1.1",
			want1:   16,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ServerGroupNICTemplate{
				Upstream:        tt.fields.Upstream,
				AssignCidrBlock: tt.fields.AssignCidrBlock,
			}
			got, got1, err := t.IPAddressByIndexFromCidrBlock(tt.args.index)
			if (err != nil) != tt.wantErr {
				t1.Errorf("IPAddressByIndexFromCidrBlock() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t1.Errorf("IPAddressByIndexFromCidrBlock() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t1.Errorf("IPAddressByIndexFromCidrBlock() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestServerGroupInstanceTemplate_Validate(t *testing.T) {
	tests := []struct {
		name     string
		template *ServerGroupInstanceTemplate
		want     []error
	}{
		{
			name:     "empty",
			template: &ServerGroupInstanceTemplate{},
			want:     []error{validate.Errorf("plan: required")},
		},
		{
			name: "minimum",
			template: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
			},
			want: nil,
		},
		{
			name: "field validation",
			template: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
				Tags:            []string{"duplicate", "duplicate"},
				InterfaceDriver: types.EInterfaceDriver("foobar"),
			},
			want: []error{
				validate.Errorf("tags: unique"),
				validate.Errorf("interface_driver: oneof=virtio e1000"),
			},
		},
		{
			name: "icon",
			template: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
				Icon:   &IdOrNameOrSelector{ResourceSelector{Names: []string{"test"}}},
				IconId: "123456789012",
			},
			want: []error{
				validate.Errorf("only one of icon and icon_id can be specified"),
			},
		},
		{
			name: "cdrom",
			template: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
				CDROM:   &IdOrNameOrSelector{ResourceSelector{Names: []string{"test"}}},
				CDROMId: "123456789012",
			},
			want: []error{
				validate.Errorf("only one of cdrom and cdrom_id can be specified"),
			},
		},
		{
			name: "private_host",
			template: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
				PrivateHost:   &IdOrNameOrSelector{ResourceSelector{Names: []string{"test"}}},
				PrivateHostId: "123456789012",
			},
			want: []error{
				validate.Errorf("only one of private_host and private_host_id can be specified"),
			},
		},
		{
			name: "cloud-config",
			template: &ServerGroupInstanceTemplate{
				Plan: &ServerGroupInstancePlan{
					Core:   1,
					Memory: 1,
				},
				// Note: 本来はcloudimgを指定すべきだが、fakeデータには含まれないため代わりにUbuntuを指定しておく
				Disks: []*ServerGroupDiskTemplate{
					{
						OSType: "ubuntu",
						Size:   20,
					},
				},
				EditParameter: &ServerGroupDiskEditTemplate{Disabled: true},
				CloudConfig:   ServerGroupCloudConfig{CloudConfig: test.StringOrFilePath(t, "#cloud-config")},
			},
			want: []error{
				fmt.Errorf("only one of edit_parameter and cloud_config can be specified"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.template.Validate(testContext(), test.APIClient, &ResourceDefServerGroup{
				ResourceDefBase: &ResourceDefBase{
					DefName: "test",
				},
				Zone: test.Zone,
			})
			require.EqualValues(t, tt.want, errs)
		})
	}
}

func TestServerGroupNICTemplate_Validate(t *testing.T) {
	type args struct {
		maxServerNum int
	}
	tests := []struct {
		name     string
		template *ServerGroupNICTemplate
		args     args
		want     []error
	}{
		{
			name:     "shared",
			template: &ServerGroupNICTemplate{Upstream: &ServerGroupNICUpstream{shared: true}},
			args:     args{maxServerNum: 1},
			want:     nil,
		},
		{
			name: "packet_filter",
			template: &ServerGroupNICTemplate{
				Upstream:       &ServerGroupNICUpstream{shared: true},
				PacketFilter:   &IdOrNameOrSelector{ResourceSelector{Names: []string{"test"}}},
				PacketFilterId: "123456789012",
			},
			want: []error{
				validate.Errorf("only one of packet_filter and packet_filter_id can be specified"),
			},
		},
		{
			name: "shared with network settings",
			template: &ServerGroupNICTemplate{
				Upstream:        &ServerGroupNICUpstream{shared: true},
				AssignCidrBlock: "192.0.2.0/24",
			},
			args: args{maxServerNum: 1},
			want: []error{validate.Errorf("upstream=shared but network settings are specified")},
		},
		{
			name: "network settings",
			template: &ServerGroupNICTemplate{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock:  "192.0.2.0/24",
				AssignNetMaskLen: 24,
				DefaultRoute:     "192.0.2.1",
			},
			args: args{maxServerNum: 1},
			want: nil,
		},
		{
			name: "network settings with smaller assign cidr block",
			template: &ServerGroupNICTemplate{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock:  "192.0.2.16/28",
				AssignNetMaskLen: 24,
				DefaultRoute:     "192.0.2.1",
			},
			args: args{maxServerNum: 1},
			want: nil,
		},
		{
			name: "invalid cidr block",
			template: &ServerGroupNICTemplate{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock:  "192.0.2.0/111",
				AssignNetMaskLen: 24,
				DefaultRoute:     "192.0.2.1",
			},
			args: args{maxServerNum: 5},
			want: []error{validate.Errorf("assign_cidr_block: cidrv4")},
		},
		{
			name: "invalid network settings",
			template: &ServerGroupNICTemplate{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock:  "192.0.2.0/30",
				AssignNetMaskLen: 24,
				DefaultRoute:     "192.0.2.1",
			},
			args: args{maxServerNum: 5},
			want: []error{validate.Errorf("assign_cidr_block is too small")},
		},
		{
			name: "invalid default route",
			template: &ServerGroupNICTemplate{
				Upstream: &ServerGroupNICUpstream{
					selector: &ResourceSelector{Names: []string{"test"}},
				},
				AssignCidrBlock:  "192.0.2.0/24",
				AssignNetMaskLen: 24,
				DefaultRoute:     "10.0.0.1",
			},
			args: args{maxServerNum: 1},
			want: []error{validate.Errorf("default_route and assigned_address must be in the same network: assign_cidr_block:192.0.2.0/24, assign_netmask_len:24, default_route:10.0.0.1")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			got := tt.template.Validate(context.Background(), nil, tt.args.maxServerNum, 0)
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestServerGroupNICMetadata_Validate(t *testing.T) {
	type args struct {
		parent   *ParentResourceDef
		nicIndex int
	}
	tests := []struct {
		name   string
		expose *ServerGroupNICMetadata
		args   args
		want   []error
	}{
		{
			name:   "minimum",
			expose: &ServerGroupNICMetadata{},
			args:   args{nil, 0},
			want:   nil,
		},
		{
			name: "global nic metadata with nixIndex == 0",
			expose: &ServerGroupNICMetadata{
				Ports:           []int{8080},
				ServerGroupName: "foobar",
				Weight:          1,
				VIPs:            []string{},
				HealthCheck:     nil,
				RecordName:      "www",
				RecordTTL:       10,
			},
			args: args{nil, 0},
			want: nil,
		},
		{
			name: "global nic metadata with nixIndex > 0",
			expose: &ServerGroupNICMetadata{
				Ports:           []int{8080},
				ServerGroupName: "foobar",
				Weight:          1,
				VIPs:            []string{},
				HealthCheck:     nil,
				RecordName:      "www",
				RecordTTL:       10,
			},
			args: args{nil, 1},
			want: []error{
				validate.Errorf("server_group_name: can only be specified for the first NIC"),
				validate.Errorf("weight: can only be specified for the first NIC"),
				validate.Errorf("record_name: can only be specified for the first NIC"),
				validate.Errorf("record_ttl: can only be specified for the first NIC"),
			},
		},
		{
			name:   "minimum with ELB",
			expose: &ServerGroupNICMetadata{},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeELB.String(),
				},
				nicIndex: 0,
			},
			want: []error{
				validate.Errorf("ports: required when parent is EnhancedLoadBalancer"),
			},
		},
		{
			name: "full with ELB",
			expose: &ServerGroupNICMetadata{
				Ports:           []int{80},
				ServerGroupName: "foobar",
				Weight:          1,
				VIPs:            []string{"192.168.0.1"},
				HealthCheck: &ServerGroupNICMetadataHealthCheck{
					Protocol:   "http",
					Path:       "/healthz",
					StatusCode: http.StatusOK,
				},
				RecordName: "www",
				RecordTTL:  10,
			},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeELB.String(),
				},
				nicIndex: 0,
			},
			want: []error{
				validate.Errorf("weight: can't specify if parent resource type is EnhancedLoadBalancer"),
				validate.Errorf("vips: can't specify if parent resource type is EnhancedLoadBalancer"),
				validate.Errorf("health_check: can't specify if parent resource type is EnhancedLoadBalancer"),
				validate.Errorf("record_name: can't specify if parent resource type is EnhancedLoadBalancer"),
				validate.Errorf("record_ttl: can't specify if parent resource type is EnhancedLoadBalancer"),
			},
		},
		{
			name:   "minimum with GSLB",
			expose: &ServerGroupNICMetadata{},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeGSLB.String(),
				},
				nicIndex: 0,
			},
			want: nil,
		},
		{
			name: "full with GSLB",
			expose: &ServerGroupNICMetadata{
				Ports:           []int{80},
				ServerGroupName: "foobar",
				Weight:          1,
				VIPs:            []string{"192.168.0.1"},
				HealthCheck: &ServerGroupNICMetadataHealthCheck{
					Protocol:   "http",
					Path:       "/healthz",
					StatusCode: http.StatusOK,
				},
				RecordName: "www",
				RecordTTL:  10,
			},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeGSLB.String(),
				},
				nicIndex: 0,
			},
			want: []error{
				validate.Errorf("server_group_name: can't specify if parent resource type is GSLB"),
				validate.Errorf("vips: can't specify if parent resource type is GSLB"),
				validate.Errorf("health_check: can't specify if parent resource type is GSLB"),
				validate.Errorf("record_name: can't specify if parent resource type is GSLB"),
				validate.Errorf("record_ttl: can't specify if parent resource type is GSLB"),
			},
		},
		{
			name:   "minimum with LB",
			expose: &ServerGroupNICMetadata{},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeLoadBalancer.String(),
				},
				nicIndex: 0,
			},
			want: []error{
				validate.Errorf("ports: required when parent is LoadBalancer"),
				validate.Errorf("health_check: required when parent is LoadBalancer"),
			},
		},
		{
			name: "full with LB",
			expose: &ServerGroupNICMetadata{
				Ports:           []int{80},
				ServerGroupName: "foobar",
				Weight:          1,
				VIPs:            []string{"192.168.0.1"},
				HealthCheck: &ServerGroupNICMetadataHealthCheck{
					Protocol:   "http",
					Path:       "/healthz",
					StatusCode: http.StatusOK,
				},
				RecordName: "www",
				RecordTTL:  10,
			},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeLoadBalancer.String(),
				},
				nicIndex: 0,
			},
			want: []error{
				validate.Errorf("server_group_name: can't specify if parent resource type is LoadBalancer"),
				validate.Errorf("weight: can't specify if parent resource type is LoadBalancer"),
				validate.Errorf("record_name: can't specify if parent resource type is LoadBalancer"),
				validate.Errorf("record_ttl: can't specify if parent resource type is LoadBalancer"),
			},
		},
		{
			name:   "minimum with DNS",
			expose: &ServerGroupNICMetadata{},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeDNS.String(),
				},
				nicIndex: 0,
			},
			want: nil,
		},
		{
			name: "full with DNS",
			expose: &ServerGroupNICMetadata{
				Ports:           []int{80},
				ServerGroupName: "foobar",
				Weight:          1,
				VIPs:            []string{"192.168.0.1"},
				HealthCheck: &ServerGroupNICMetadataHealthCheck{
					Protocol:   "http",
					Path:       "/healthz",
					StatusCode: http.StatusOK,
				},
				RecordName: "www",
				RecordTTL:  10,
			},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeDNS.String(),
				},
				nicIndex: 0,
			},
			want: []error{
				validate.Errorf("server_group_name: can't specify if parent resource type is DNS"),
				validate.Errorf("weight: can't specify if parent resource type is DNS"),
				validate.Errorf("vips: can't specify if parent resource type is DNS"),
				validate.Errorf("health_check: can't specify if parent resource type is DNS"),
			},
		},
		{
			name: "with invalid health_check",
			expose: &ServerGroupNICMetadata{
				Ports: []int{80},
				VIPs:  []string{"192.168.0.1"},
				HealthCheck: &ServerGroupNICMetadataHealthCheck{
					Protocol:   "ping",
					Path:       "/healthz",
					StatusCode: http.StatusOK,
				},
			},
			args: args{
				parent: &ParentResourceDef{
					TypeName: ResourceTypeLoadBalancer.String(),
				},
				nicIndex: 0,
			},
			want: []error{
				validate.Errorf("path: can not be specified if protocol is not http or https"),
				validate.Errorf("status_code: can not be specified if protocol is not http or https"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.expose.Validate(tt.args.parent, tt.args.nicIndex)
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestServerGroupNICMetadataHealthCheck_Validate(t *testing.T) {
	tests := []struct {
		name string
		hc   *ServerGroupNICMetadataHealthCheck
		want []error
	}{
		{
			name: "minimum",
			hc:   &ServerGroupNICMetadataHealthCheck{},
			want: []error{
				validate.Errorf("protocol: required"),
			},
		},
		{
			name: "http with path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol:   "http",
				Path:       "/",
				StatusCode: http.StatusOK,
			},
			want: nil,
		},
		{
			name: "http without path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol: "http",
			},
			want: []error{
				validate.Errorf("path: required if protocol is http or https"),
				validate.Errorf("status_code: required if protocol is http or https"),
			},
		},
		{
			name: "https with path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol:   "https",
				Path:       "/",
				StatusCode: http.StatusOK,
			},
			want: nil,
		},
		{
			name: "https without path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol: "https",
			},
			want: []error{
				validate.Errorf("path: required if protocol is http or https"),
				validate.Errorf("status_code: required if protocol is http or https"),
			},
		},
		{
			name: "ping with path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol:   "ping",
				Path:       "/",
				StatusCode: http.StatusOK,
			},
			want: []error{
				validate.Errorf("path: can not be specified if protocol is not http or https"),
				validate.Errorf("status_code: can not be specified if protocol is not http or https"),
			},
		},
		{
			name: "ping without path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol: "ping",
			},
			want: nil,
		},
		{
			name: "tcp with path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol:   "tcp",
				Path:       "/",
				StatusCode: http.StatusOK,
			},
			want: []error{
				validate.Errorf("path: can not be specified if protocol is not http or https"),
				validate.Errorf("status_code: can not be specified if protocol is not http or https"),
			},
		},
		{
			name: "tcp without path and code",
			hc: &ServerGroupNICMetadataHealthCheck{
				Protocol: "ping",
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.hc.Validate()
			require.EqualValues(t, tt.want, got)
		})
	}
}

func TestServerGroupInstanceTemplate_CalculateTagsByIndex(t *testing.T) {
	type args struct {
		serverIndexWithinGroup int
		zoneLength             int
	}
	tests := []struct {
		name string
		tags []string
		args args
		want []string
	}{
		{
			name: "minimum",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: 0,
				zoneLength:             1,
			},
			want: []string{groupSpecialTags[0]},
		},
		{
			name: "return as-is with invalid index",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: -1,
				zoneLength:             1,
			},
			want: []string{},
		},
		{
			name: "return as-is with invalid zoneLength",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: 0,
				zoneLength:             0,
			},
			want: []string{},
		},
		{
			name: "return as-is when template already has a group tag",
			tags: []string{"tag", groupSpecialTags[0]},
			args: args{
				serverIndexWithinGroup: 0,
				zoneLength:             0,
			},
			want: []string{"tag", groupSpecialTags[0]},
		},
		{
			name: "added a group tag after template tags",
			tags: []string{"tag"},
			args: args{
				serverIndexWithinGroup: 0,
				zoneLength:             1,
			},
			want: []string{"tag", groupSpecialTags[0]},
		},
		{
			name: "single zone",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: 3,
				zoneLength:             1,
			},
			want: []string{groupSpecialTags[3]},
		},
		{
			name: "single zone",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: 4,
				zoneLength:             1,
			},
			want: []string{groupSpecialTags[0]},
		},
		{
			name: "multiple zones",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: 3,
				zoneLength:             2,
			},
			want: []string{groupSpecialTags[1]},
		},
		{
			name: "multiple zones",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: 8,
				zoneLength:             4,
			},
			want: []string{groupSpecialTags[2]},
		},
		{
			name: "multiple zones",
			tags: []string{},
			args: args{
				serverIndexWithinGroup: 17,
				zoneLength:             2,
			},
			want: []string{groupSpecialTags[0]},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ServerGroupInstanceTemplate{
				Tags:        tt.tags,
				UseGroupTag: true,
			}
			got := s.CalculateTagsByIndex(tt.args.serverIndexWithinGroup, tt.args.zoneLength)
			require.EqualValues(t, tt.want, got)
		})
	}
}
