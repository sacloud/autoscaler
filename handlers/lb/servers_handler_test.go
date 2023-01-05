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

package lb

import (
	"testing"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/iaas-api-go"
	"github.com/stretchr/testify/require"
)

func TestServersHandler_filteredVIPs(t *testing.T) {
	type args struct {
		vips       iaas.LoadBalancerVirtualIPAddresses
		exposeInfo *handler.ServerGroupInstance_ExposeInfo
	}
	tests := []struct {
		name string
		args args
		want iaas.LoadBalancerVirtualIPAddresses
	}{
		{
			name: "without vips",
			args: args{
				vips: iaas.LoadBalancerVirtualIPAddresses{
					{
						VirtualIPAddress: "192.168.0.1",
						Port:             80,
					},
					{
						VirtualIPAddress: "192.168.0.1",
						Port:             443,
					},
					{
						VirtualIPAddress: "192.168.0.2",
						Port:             8080,
					},
				},
				exposeInfo: &handler.ServerGroupInstance_ExposeInfo{},
			},
			want: iaas.LoadBalancerVirtualIPAddresses{
				{
					VirtualIPAddress: "192.168.0.1",
					Port:             80,
				},
				{
					VirtualIPAddress: "192.168.0.1",
					Port:             443,
				},
				{
					VirtualIPAddress: "192.168.0.2",
					Port:             8080,
				},
			},
		},
		{
			name: "with vips",
			args: args{
				vips: iaas.LoadBalancerVirtualIPAddresses{
					{
						VirtualIPAddress: "192.168.0.1",
						Port:             80,
					},
					{
						VirtualIPAddress: "192.168.0.1",
						Port:             443,
					},
					{
						VirtualIPAddress: "192.168.0.2",
						Port:             8080,
					},
				},
				exposeInfo: &handler.ServerGroupInstance_ExposeInfo{
					Vips: []string{"192.168.0.1"},
				},
			},
			want: iaas.LoadBalancerVirtualIPAddresses{
				{
					VirtualIPAddress: "192.168.0.1",
					Port:             80,
				},
				{
					VirtualIPAddress: "192.168.0.1",
					Port:             443,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ServersHandler{}
			got := h.filteredVIPs(tt.args.vips, tt.args.exposeInfo)
			require.EqualValues(t, tt.want, got)
		})
	}
}
