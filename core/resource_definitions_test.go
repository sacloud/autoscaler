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
	"testing"

	"github.com/sacloud/autoscaler/handler"
	"github.com/sacloud/autoscaler/handlers"

	"github.com/sacloud/autoscaler/handlers/stub"
	"github.com/sacloud/autoscaler/test"
	"github.com/stretchr/testify/require"
)

func TestResourceDefinitions_HandleAll(t *testing.T) {
	ctx := testContext()
	defer initTestServer(t)()
	defer initTestDNS(t)()

	server := &ResourceDefServer{
		ResourceDefBase: &ResourceDefBase{
			TypeName: "Server",
			TargetSelector: &ResourceSelector{
				Names: []string{"test-server"},
				Zone:  test.Zone,
			},
		},
	}
	dns := &ResourceDefDNS{
		ResourceDefBase: &ResourceDefBase{
			TypeName: "DNS",
			TargetSelector: &ResourceSelector{
				Names: []string{"test-dns.com"},
			},
			children: ResourceDefinitions{server},
		},
	}
	defs := ResourceDefinitions{dns}

	var called []string
	stubHandler := &Handler{
		Name: "stub",
		BuiltinHandler: &stub.Handler{
			Logger: test.Logger,
			HandleFunc: func(request *handler.HandleRequest, sender handlers.ResponseSender) error {
				if server := request.Desired.GetServer(); server != nil {
					// HandleAllの中でParentが設定されているか
					require.NotNil(t, server.Parent.GetDns())

					called = append(called, "server")
				} else if dns := request.Desired.GetDns(); dns != nil {
					called = append(called, "dns")
				}
				return nil
			},
		},
	}

	err := defs.HandleAll(ctx, test.APIClient, Handlers{stubHandler})
	require.NoError(t, err)
	// 子から先にHandleされているか?
	require.Equal(t, []string{"server", "dns"}, called)
}
