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

package handler

import "encoding/json"

func (x *ServerGroupInstance_NIC) EachIPAndExposedPort(fn func(ip string, port int) error) error {
	if x == nil || x.ExposeInfo == nil || x.AssignedNetwork == nil {
		return nil
	}

	for _, port := range x.ExposeInfo.Ports {
		if err := fn(x.AssignedNetwork.IpAddress, int(port)); err != nil {
			return err
		}
	}
	return nil
}

func (x *HandleRequest) JSON() []byte {
	if x == nil {
		return nil
	}
	data, err := json.Marshal(x)
	if err != nil {
		return nil
	}
	return data
}

func (x *PostHandleRequest) JSON() []byte {
	if x == nil {
		return nil
	}
	data, err := json.Marshal(x)
	if err != nil {
		return nil
	}
	return data
}
