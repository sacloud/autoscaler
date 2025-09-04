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

//go:build e2e
// +build e2e

package horizontal_scaling

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	autoscalerE2E "github.com/sacloud/autoscaler/e2e"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/helper/cleanup"
	"github.com/sacloud/iaas-api-go/search"
	serverService "github.com/sacloud/iaas-service-go/server"
	"github.com/sacloud/packages-go/e2e"
)

const (
	coreReadyMarker   = `msg=started address=autoscaler.sock`
	upJobDoneMarker   = `request=Up source=default resource=autoscaler-e2e-horizontal-scaling status=JOB_DONE`
	downJobDoneMarker = `request=Down source=default resource=autoscaler-e2e-horizontal-scaling status=JOB_DONE`
	keepJobDoneMarker = `request=Keep source=default resource=autoscaler-e2e-horizontal-scaling status=JOB_DONE`
)

var (
	coreCmd       = exec.Command("autoscaler", "start")
	upCmd         = exec.Command("autoscaler", "inputs", "direct", "--resource-name", "autoscaler-e2e-horizontal-scaling", "up")
	upToMediumCmd = exec.Command("autoscaler", "inputs", "direct", "--resource-name", "autoscaler-e2e-horizontal-scaling", "--desired-state-name", "medium", "up")
	downCmd       = exec.Command("autoscaler", "inputs", "direct", "--resource-name", "autoscaler-e2e-horizontal-scaling", "down")
	keepCmd       = exec.Command("autoscaler", "inputs", "direct", "--resource-name", "autoscaler-e2e-horizontal-scaling", "keep")

	zones               = []string{"tk1b", "is1b"}
	proxyLBReadyTimeout = 5 * time.Minute
	e2eTestTimeout      = 20 * time.Minute

	output *e2e.Output
)

func TestMain(m *testing.M) {
	defer teardown()
	setup()

	m.Run()
}

func TestE2E_HorizontalScaling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	/**************************************************************************
	 * Step 0: Coreの起動確認
	 *************************************************************************/
	log.Println("step0: setup")

	// grpc-health-probeでSERVINGになっていることを確認
	out, err := exec.Command("grpc-health-probe", "-addr", "unix:autoscaler.sock").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "status: SERVING") {
		t.Fatalf("grpc-health-prove: unexpected response: %s", string(out))
	}

	/**************************************************************************
	 * Step 1-1: スケールアウト(0 -> 1)
	 *************************************************************************/
	log.Println("step1-1: scale out")

	// Direct InputでUpリクエストを送信
	if err := upCmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Coreのジョブ完了まで待機
	if err := output.WaitOutput(upJobDoneMarker, 10*time.Minute); err != nil {
		output.Fatal(t, err)
	}

	// 以降はProxyLB->ServerへのHTTPリクエストが通るはずなのでポーリングを続ける
	if err := waitProxyLBAndStartHTTPRequestLoop(ctx, t); err != nil {
		t.Fatal(err)
	}

	/**************************************************************************
	 * Step 1-2: スケールアウト結果の確認
	 *************************************************************************/
	log.Println("step1-2: check results")
	servers, err := fetchSakuraCloudServers()
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 1 {
		output.Fatalf(t,
			"got unexpected server count: expected:1 actual:%d",
			len(servers),
		)
	}

	// 冷却期間待機
	time.Sleep(180 * time.Second)

	// Terraformステートのリフレッシュ(複数回IDが変更されるため毎回リフレッシュしておく)
	e2e.TerraformRefresh() // nolint

	/**************************************************************************
	 * Step 2-1: desired state nameを指定してのスケールアウト
	 *************************************************************************/
	log.Println("step2-1: scale out with desired state name")
	if err := upToMediumCmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Coreのジョブ完了まで待機
	if err := output.WaitOutput(upJobDoneMarker, 10*time.Minute); err != nil {
		output.Fatal(t, err)
	}

	/**************************************************************************
	 * Step 2-2: スケールアウト結果の確認
	 *************************************************************************/
	log.Println("step2-2: check results")
	servers, err = fetchSakuraCloudServers()
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 3 {
		output.Fatalf(t,
			"got unexpected server count: expected:3 actual:%d",
			len(servers),
		)
	}

	// 冷却期間待機
	time.Sleep(180 * time.Second)

	// Terraformステートのリフレッシュ(複数回IDが変更されるため毎回リフレッシュしておく)
	e2e.TerraformRefresh() // nolint

	/**************************************************************************
	 * Step 3-1: 台数維持
	 *************************************************************************/
	log.Println("step3-1: keep")
	// 一台消してからKeepリクエストを送り、台数維持が働くことを確認する
	if err := deleteServerFromELB(ctx, t, servers[0]); err != nil {
		t.Fatal(err)
	}
	if err := cleanup.DeleteServer(ctx, autoscalerE2E.SacloudAPICaller, servers[0].Zone.Name, servers[0].ID, true); err != nil {
		t.Fatal(err)
	}

	if err := keepCmd.Run(); err != nil {
		t.Fatal(err)
	}
	// Coreのジョブ完了まで待機
	if err := output.WaitOutput(keepJobDoneMarker, 10*time.Minute); err != nil {
		output.Fatal(t, err)
	}
	/**************************************************************************
	 * Step 3-2: 台数維持結果の確認
	 *************************************************************************/
	log.Println("step3-2: check results")
	servers, err = fetchSakuraCloudServers()
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 3 {
		output.Fatalf(t,
			"got unexpected server count: expected:3 actual:%d",
			len(servers),
		)
	}
	// 冷却期間待機
	time.Sleep(180 * time.Second)

	/**************************************************************************
	 * Step 4-1: スケールイン
	 *************************************************************************/
	log.Println("step4-1: scale in")
	if err := downCmd.Run(); err != nil {
		t.Fatal(err)
	}

	// Coreのジョブ完了まで待機
	if err := output.WaitOutput(downJobDoneMarker, 10*time.Minute); err != nil {
		output.Fatal(t, err)
	}
	/**************************************************************************
	 * Step 4-2: スケールイン結果の確認
	 *************************************************************************/
	log.Println("step4-2: check results")
	servers, err = fetchSakuraCloudServers()
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 2 {
		output.Fatalf(t,
			"got unexpected server count: expected:2 actual:%d",
			len(servers),
		)
	}
	// Terraformステートのリフレッシュ(複数回IDが変更されるため毎回リフレッシュしておく)
	e2e.TerraformRefresh() // nolint

	output.Output()
}

func setup() {
	if err := e2e.TerraformInit(); err != nil {
		log.Fatal(err)
	}
	if err := e2e.TerraformApply(); err != nil {
		log.Fatal(err)
	}

	coreOutputs, err := coreCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := coreCmd.Start(); err != nil {
		log.Fatal(err)
	}

	output = e2e.NewOutput(coreOutputs, "[Core]")

	if err := output.WaitOutput(coreReadyMarker, 3*time.Second); err != nil {
		output.Output()
		log.Fatal(err)
	}
}

func teardown() {
	if coreCmd.Process != nil {
		if err := coreCmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Println(err)
		}
		if err := coreCmd.Wait(); err != nil {
			log.Println(err)
		}
	}

	if err := e2e.TerraformDestroy(); err != nil {
		log.Println(err)
	}

	// サーバはTerraform管理外のためここでクリーンアップする
	servers, err := fetchSakuraCloudServers()
	if err != nil {
		log.Println(err)
	} else {
		svc := serverService.New(autoscalerE2E.SacloudAPICaller)
		for _, zone := range zones {
			for _, server := range servers {
				err := svc.Delete(&serverService.DeleteRequest{
					Zone:           zone,
					ID:             server.ID,
					WithDisks:      true,
					FailIfNotFound: false,
					Force:          true,
				})
				if err != nil {
					log.Println(err)
				}
			}
		}
	}
}

func fetchSakuraCloudServers() ([]*iaas.Server, error) {
	serverOp := iaas.NewServerOp(autoscalerE2E.SacloudAPICaller)

	var servers []*iaas.Server
	for _, zone := range zones {
		found, err := serverOp.Find(context.Background(), zone, &iaas.FindCondition{
			Filter: search.Filter{
				search.Key("Name"): search.PartialMatch("autoscaler-e2e-horizontal-scaling"),
			},
		})
		if err != nil {
			return nil, err
		}
		servers = append(servers, found.Servers...)
	}

	return servers, nil
}

func deleteServerFromELB(ctx context.Context, t *testing.T, server *iaas.Server) error {
	elbOp := iaas.NewProxyLBOp(autoscalerE2E.SacloudAPICaller)
	found, err := elbOp.Find(context.Background(), &iaas.FindCondition{
		Count: 1,
		Filter: search.Filter{
			search.Key("Name"): search.ExactMatch("autoscaler-e2e-horizontal-scaling"),
		},
	})
	if err != nil {
		return err
	}

	if len(found.ProxyLBs) == 0 {
		return fmt.Errorf("proxylb 'autoscaler-e2e-horizontal-scaling' not found")
	}
	elb := found.ProxyLBs[0]

	var servers []*iaas.ProxyLBServer
	for _, s := range elb.Servers {
		if s.IPAddress != server.Interfaces[0].IPAddress {
			servers = append(servers, s)
		}
		elb.Servers = servers
	}
	_, err = elbOp.UpdateSettings(ctx, elb.ID, &iaas.ProxyLBUpdateSettingsRequest{
		HealthCheck:          elb.HealthCheck,
		SorryServer:          elb.SorryServer,
		BindPorts:            elb.BindPorts,
		Servers:              elb.Servers,
		Rules:                elb.Rules,
		LetsEncrypt:          elb.LetsEncrypt,
		StickySession:        elb.StickySession,
		Timeout:              elb.Timeout,
		Gzip:                 elb.Gzip,
		BackendHttpKeepAlive: elb.BackendHttpKeepAlive,
		ProxyProtocol:        elb.ProxyProtocol,
		Syslog:               elb.Syslog,
		SettingsHash:         elb.SettingsHash,
	})
	return err
}

func waitProxyLBAndStartHTTPRequestLoop(ctx context.Context, t *testing.T) error {
	elbOp := iaas.NewProxyLBOp(autoscalerE2E.SacloudAPICaller)
	found, err := elbOp.Find(context.Background(), &iaas.FindCondition{
		Count: 1,
		Filter: search.Filter{
			search.Key("Name"): search.ExactMatch("autoscaler-e2e-horizontal-scaling"),
		},
	})
	if err != nil {
		return err
	}

	if len(found.ProxyLBs) == 0 {
		return fmt.Errorf("proxylb 'autoscaler-e2e-horizontal-scaling' not found")
	}
	elb := found.ProxyLBs[0]

	// vip宛にリクエストが通るまで待機
	url := fmt.Sprintf("http://%s", elb.VirtualIPAddress)
	if err := e2e.HttpRequestUntilSuccess(url, proxyLBReadyTimeout); err != nil {
		t.Fatal(err)
	}

	// リクエストが通るようになったら定期的にリクエストを送り、正常なレスポンス(StatusCode==200)を得られていることを確認し続ける
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := e2e.HttpGet(url); err != nil {
					log.Println("[ERROR]", err)
					t.Error(err)
					return
				}
			}
		}
	}()
	return nil
}
