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
	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

type ServerPlan struct {
	Core   int // コア数
	Memory int // メモリサイズ(GiB)
}

type Server struct {
	DedicatedCPU  bool
	PrivateHostID types.ID
	Zone          string
	Plans         []ServerPlan
	Wrappers      Resources
}

type CurrentServer struct {
	ResourceState handler.ResourceStatus
}

func (s *Server) Type() ResourceTypes {
	return ResourceTypeServer
}

func (s *Server) Selector() *ResourceSelector {
	// TODO 実装
	return nil
}

func (s *Server) Current() CurrentResource {
	// TODO 実装
	return nil
}

func (s *Server) Desired() Desired {
	// TODO 実装
	return nil
}
