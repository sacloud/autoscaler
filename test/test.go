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

package test

import (
	"context"
	"log/slog"
	"os"
	"testing"

	client "github.com/sacloud/api-client-go"
	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/api"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/sacloud/packages-go/size"
)

var (
	Zone      = "is1a"
	APIClient = api.NewCallerWithOptions(&api.CallerOptions{
		Options: &client.Options{
			AccessToken:       "fake",
			AccessTokenSecret: "fake",
			UserAgent:         "sacloud/autoscaler/fake/test",
			Trace:             os.Getenv("SAKURACLOUD_TRACE") != "",
		},
		TraceAPI: os.Getenv("SAKURACLOUD_TRACE") != "",
		FakeMode: true,
	})
	Logger = log.NewLogger(&log.LoggerOption{
		Writer:    os.Stderr,
		JSON:      false,
		TimeStamp: true,
		Caller:    true,
		Level:     slog.LevelDebug,
	})
)

func AddTestServer(t *testing.T, name string) (*iaas.Server, func()) {
	serverOp := iaas.NewServerOp(APIClient)
	server, err := serverOp.Create(context.Background(), Zone, &iaas.ServerCreateRequest{
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

func AddTestDNS(t *testing.T, name string) (*iaas.DNS, func()) {
	dnsOp := iaas.NewDNSOp(APIClient)
	dns, err := dnsOp.Create(context.Background(), &iaas.DNSCreateRequest{
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

func AddTestSwitch(t *testing.T, name string) (*iaas.Switch, func()) {
	swOp := iaas.NewSwitchOp(APIClient)
	sw, err := swOp.Create(context.Background(), Zone, &iaas.SwitchCreateRequest{
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

func AddTestELB(t *testing.T, name string) func() {
	ctx := context.Background()
	client := iaas.NewProxyLBOp(APIClient)
	elb, err := client.Create(ctx, &iaas.ProxyLBCreateRequest{
		Plan: types.ProxyLBPlans.CPS100,
		Name: name,
	})
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		if err := client.Delete(ctx, elb.ID); err != nil {
			t.Logf("[WARN] deleting ELB failed: %s", err)
		}
	}
}

func StringOrFilePath(t *testing.T, s string) config.StringOrFilePath {
	v, err := config.NewStringOrFilePath(context.Background(), s)
	if err != nil {
		t.Logf("[WARN] invaild StringOrFilePath value: %s", s)
		return config.StringOrFilePath{}
	}
	return *v
}
