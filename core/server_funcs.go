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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func assignedNetwork(nic *sacloud.InterfaceView, index int) *handler.NetworkInfo {
	var ipAddress string
	if nic.UpstreamType == types.UpstreamNetworkTypes.Shared {
		ipAddress = nic.IPAddress
	} else {
		ipAddress = nic.UserIPAddress
	}
	return &handler.NetworkInfo{
		IpAddress: ipAddress,
		Netmask:   uint32(nic.UserSubnetNetworkMaskLen),
		Gateway:   nic.UserSubnetDefaultRoute,
		Index:     uint32(index),
	}
}
