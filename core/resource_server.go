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
	"errors"
	"fmt"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ServerPlan struct {
	Core   int // コア数
	Memory int // メモリサイズ(GiB)
}

type Server struct {
	*ResourceBase `yaml:",inline"`
	DedicatedCPU  bool         `yaml:"dedicated_cpu"`
	PrivateHostID types.ID     `yaml:"private_host_id"`
	Zone          string       `yaml:"zone"`
	Plans         []ServerPlan `yaml:"plans"`
	Wrappers      Resources    `yaml:"wrappers"`

	current *currentServer
	desired *desiredServer
}

func (s *Server) Validate() error {
	selector := s.Selector()
	if selector == nil {
		return errors.New("selector: required")
	}
	if len(selector.Zones) == 0 {
		return errors.New("selector.Zones: least one value required")
	}
	return nil
}

func (s *Server) Calculate(ctx *Context, apiClient sacloud.APICaller) (CurrentResource, Desired, error) {
	if err := s.Validate(); err != nil {
		return nil, nil, err
	}

	serverOp := sacloud.NewServerOp(apiClient)
	selector := s.Selector()
	var server *sacloud.Server
	for _, zone := range selector.Zones {
		fc := selector.FindCondition()
		found, err := serverOp.Find(ctx, zone, fc)
		if err != nil {
			return nil, nil, fmt.Errorf("calculating server status failed: %s", err)
		}
		if len(found.Servers) > 0 {
			server = found.Servers[0]
		}
	}

	if server == nil {
		s.current = &currentServer{
			status: handler.ResourceStatus_NOT_EXISTS,
			server: nil,
		}
		return s.current, nil, nil
	}

	desired := &desiredServer{
		raw: &sacloud.Server{},
	}
	if err := mapconvDecoder.ConvertTo(server, desired.raw); err != nil {
		return nil, nil, fmt.Errorf("calculating desired state failed: %s", err)
	}

	plan := s.desiredPlan(ctx, server)

	if plan != nil {
		desired.raw.CPU = plan.Core
		desired.raw.MemoryMB = plan.Memory * size.GiB
	}

	s.current = &currentServer{
		status: handler.ResourceStatus_RUNNING, // TODO RUNNINGだけで良いか? 作成途中などを示すステータスは不要か?
		server: server,
	}
	s.desired = desired

	return s.current, s.desired, nil
}

func (s *Server) desiredPlan(ctx *Context, current *sacloud.Server) *ServerPlan {
	var fn func(i int) *ServerPlan
	// TODO s.Plansがない場合のデフォルト値を考慮する
	// TODO s.Plansの並べ替えを考慮する

	switch ctx.Request().requestType {
	case requestTypeUp:
		fn = func(i int) *ServerPlan {
			if i < len(s.Plans) {
				return &ServerPlan{
					Core:   s.Plans[i+1].Core,
					Memory: s.Plans[i+1].Memory,
				}
			}
			return nil
		}
	case requestTypeDown:
		fn = func(i int) *ServerPlan {
			if i > 0 {
				return &ServerPlan{
					Core:   s.Plans[i-1].Core,
					Memory: s.Plans[i-1].Memory,
				}
			}
			return nil
		}
	default:
		return nil // 到達しないはず
	}

	for i, plan := range s.Plans {
		if plan.Core == current.CPU && plan.Memory == current.GetMemoryGB() {
			return fn(i)
		}
	}
	return nil
}

// currentServer CurrentResourceを実装した、現在のサーバの状態を表すためのデータ構造
type currentServer struct {
	status handler.ResourceStatus
	server *sacloud.Server
}

func (c *currentServer) Status() handler.ResourceStatus {
	return c.status
}

func (c *currentServer) Raw() interface{} {
	return c.server
}

type desiredServer struct {
	raw *sacloud.Server
}

func (d *desiredServer) Raw() interface{} {
	return d.raw
}
