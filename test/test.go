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

package test

import (
	"context"
	"os"
	"testing"

	"github.com/sacloud/libsacloud/v2/pkg/size"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/types"

	"github.com/sacloud/autoscaler/log"

	"github.com/sacloud/libsacloud/v2/helper/api"
)

var (
	Zone      = "is1a"
	APIClient = api.NewCaller(&api.CallerOptions{
		AccessToken:       "fake",
		AccessTokenSecret: "fake",
		UserAgent:         "sacloud/autoscaler/fake/test",
		TraceAPI:          os.Getenv("SAKURACLOUD_TRACE") != "",
		TraceHTTP:         os.Getenv("SAKURACLOUD_TRACE") != "",
		FakeMode:          true,
	})
	Logger = log.NewLogger(&log.LoggerOption{
		Writer:    os.Stderr,
		JSON:      false,
		TimeStamp: true,
		Caller:    true,
		Level:     log.LevelDebug,
	})
)

func AddTestServer(t *testing.T, name string) (*sacloud.Server, func()) {
	serverOp := sacloud.NewServerOp(APIClient)
	server, err := serverOp.Create(context.Background(), Zone, &sacloud.ServerCreateRequest{
		CPU:                  2,
		MemoryMB:             4 * size.GiB,
		ServerPlanCommitment: types.Commitments.Standard,
		ServerPlanGeneration: types.PlanGenerations.Default,
		ConnectedSwitches:    nil,
		InterfaceDriver:      types.InterfaceDrivers.VirtIO,
		Name:                 name,
	})
	if err != nil {
		t.Fatal(err)
	}

	return server, func() {
		if err := serverOp.Delete(context.Background(), Zone, server.ID); err != nil {
			t.Logf("[WARN] deleting server failed: %s", err)
		}
	}
}

func AddTestDNS(t *testing.T, name string) (*sacloud.DNS, func()) {
	dnsOp := sacloud.NewDNSOp(APIClient)
	dns, err := dnsOp.Create(context.Background(), &sacloud.DNSCreateRequest{
		Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}

	return dns, func() {
		if err := dnsOp.Delete(context.Background(), dns.ID); err != nil {
			t.Logf("[WARN] deleting dns failed: %s", err)
		}
	}
}

func AddTestSwitch(t *testing.T, name string) (*sacloud.Switch, func()) {
	swOp := sacloud.NewSwitchOp(APIClient)
	sw, err := swOp.Create(context.Background(), Zone, &sacloud.SwitchCreateRequest{
		Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}

	return sw, func() {
		if err := swOp.Delete(context.Background(), Zone, sw.ID); err != nil {
			t.Logf("[WARN] deleting switch failed: %s", err)
		}
	}
}
