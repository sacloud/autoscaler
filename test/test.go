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
	"strings"
	"testing"
	"time"

	"github.com/sacloud/autoscaler/config"
	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/sacloud-sdk-go/api/iaas"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/defaults"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/fake"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/types"
	"github.com/sacloud/sacloud-sdk-go/common/packages/size"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
)

var (
	Zone      = "is1a"
	APIClient iaas.APICaller
	Logger    = log.NewLogger(&log.LoggerOption{
		Writer:    os.Stderr,
		JSON:      false,
		TimeStamp: true,
		Caller:    true,
		Level:     slog.LevelDebug,
	})
)

func init() {
	os.Setenv("SAKURACLOUD_FAKE_MODE", "1") //nolint:errcheck
	fake.SwitchFactoryFuncToFake()

	var originalStatePollingInterval = defaults.DefaultStatePollingInterval
	var originalDBStatusPollingInterval = defaults.DefaultDBStatusPollingInterval

	defaults.DefaultStatePollingInterval = 10 * time.Millisecond
	defaults.DefaultDBStatusPollingInterval = 10 * time.Millisecond

	fake.DiskCopyDuration = time.Millisecond
	fake.PowerOnDuration = time.Millisecond
	fake.PowerOffDuration = time.Millisecond

	// core.SakuraCloud と同じ条件で saclient.Client を初期化するため
	// 環境変数をベースに、fake モードとテスト用トークンを設定する
	env := os.Environ()
	env = append(env, "SAKURACLOUD_FAKE_MODE=1")
	env = append(env, "SAKURA_ACCESS_TOKEN=fake")
	env = append(env, "SAKURA_ACCESS_TOKEN_SECRET=fake")

	var sa saclient.Client
	if err := sa.SetWith(saclient.WithUserAgent("sacloud/autoscaler/test")); err != nil {
		defaults.DefaultStatePollingInterval = originalStatePollingInterval
		defaults.DefaultDBStatusPollingInterval = originalDBStatusPollingInterval
		panic(err)
	}
	if err := sa.SetEnviron(env); err != nil {
		defaults.DefaultStatePollingInterval = originalStatePollingInterval
		defaults.DefaultDBStatusPollingInterval = originalDBStatusPollingInterval
		panic(err)
	}
	if err := sa.Populate(); err != nil {
		defaults.DefaultStatePollingInterval = originalStatePollingInterval
		defaults.DefaultDBStatusPollingInterval = originalDBStatusPollingInterval
		panic(err)
	}

	cfg, err := sa.EndpointConfig()
	if err != nil {
		panic(err)
	}
	if cfg.APIRootURL != "" {
		if strings.HasSuffix(cfg.APIRootURL, "/") {
			cfg.APIRootURL = strings.TrimRight(cfg.APIRootURL, "/")
		}
		iaas.SakuraCloudAPIRoot = cfg.APIRootURL
	}
	if len(cfg.Zones) > 0 {
		iaas.SakuraCloudZones = cfg.Zones
	}
	if cfg.DefaultZone != "" {
		iaas.APIDefaultZone = cfg.DefaultZone
	}
	APIClient = iaas.NewClientFromSaclient(&sa)
}

func AddTestServer(t *testing.T, name string) (*iaas.Server, func()) {
	serverOp := iaas.NewServerOp(APIClient)
	server, err := serverOp.Create(context.Background(), Zone, &iaas.ServerCreateRequest{
		CPU:               2,
		MemoryMB:          4 * size.GiB,
		Commitment:        types.Commitments.Standard,
		Generation:        types.PlanGenerations.Default,
		ConnectedSwitches: nil,
		InterfaceDriver:   types.InterfaceDrivers.VirtIO,
		Name:              name,
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
