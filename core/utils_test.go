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
	"testing"

	"github.com/sacloud/autoscaler/test"
	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
)

func testContext() *RequestContext {
	return NewRequestContext(context.Background(), &requestInfo{
		requestType:       requestTypeUp,
		source:            "default",
		action:            "default",
		resourceGroupName: "web",
	}, test.Logger)
}

func initTestServer(t *testing.T) func() {
	serverOp := sacloud.NewServerOp(test.APIClient)
	server, err := serverOp.Create(context.Background(), test.Zone, &sacloud.ServerCreateRequest{
		CPU:                  2,
		MemoryMB:             4 * size.GiB,
		ServerPlanCommitment: types.Commitments.Standard,
		ServerPlanGeneration: types.PlanGenerations.Default,
		ConnectedSwitches:    nil,
		InterfaceDriver:      types.InterfaceDrivers.VirtIO,
		Name:                 "test-server",
	})
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		if err := serverOp.Delete(context.Background(), test.Zone, server.ID); err != nil {
			t.Logf("[WARN] deleting server failed: %s", err)
		}
	}
}

func initTestDNS(t *testing.T) func() {
	dnsOp := sacloud.NewDNSOp(test.APIClient)
	dns, err := dnsOp.Create(context.Background(), &sacloud.DNSCreateRequest{
		Name: "test-dns.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		if err := dnsOp.Delete(context.Background(), dns.ID); err != nil {
			t.Logf("[WARN] deleting dns failed: %s", err)
		}
	}
}
