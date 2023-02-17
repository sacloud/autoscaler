// Copyright 2021-2023 The sacloud/autoscaler Authors
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

package vertical_scaling

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	autoscalerE2E "github.com/sacloud/autoscaler/e2e"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/search"
	"github.com/sacloud/packages-go/e2e"
)

const (
	coreReadyMarker        = `message=started address=autoscaler.sock`
	inputsReadyMarker      = `message=started address=127.0.0.1:8080`
	upJobDoneMarker        = `request=Up source=default resource=server status=JOB_DONE`
	downJobDoneMarker      = `request=Down source=default resource=server status=JOB_DONE`
	inCoolDownTimeMarker   = `job-message="job is in an unacceptable state"`
	inCoolDownTimeResponse = `"message":"job is in an unacceptable state"`
)

var (
	coreCmd  = exec.Command("autoscaler", "start", "--log-level", "debug")
	inputCmd = exec.Command("autoscaler", "inputs", "grafana", "--addr", "127.0.0.1:8080")

	proxyLBReadyTimeout = 5 * time.Minute
	e2eTestTimeout      = 30 * time.Minute

	//go:embed webhook.json
	grafanaWebhookBody []byte

	coreOutput   *e2e.Output
	inputsOutput *e2e.Output
)

func TestMain(m *testing.M) {
	defer teardown()
	setup()

	m.Run()
}

func TestE2E_VerticalScaling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	/**************************************************************************
	 * Step 0: 現在のクラウド上のリソースの確認/ポーリング開始
	 *************************************************************************/
	log.Println("step0: setup")
	// ProxyLBへのHTTPリクエストが通るようになるまで待ち & ポーリング開始
	if err := waitResourceReadyLAndStartHTTPRequestLoop(ctx, t); err != nil {
		t.Fatal(err)
	}

	// サーバプランの確認(前提)
	server, err := fetchSakuraCloudServer()
	if err != nil {
		t.Fatal(err)
	}
	if server.CPU != 1 || server.GetMemoryGB() != 2 {
		coreOutput.Fatalf(t,
			"server has unexpected initial plan: expected: {CPU:1, Memory:2} actual: {CPU:%d, Memory:%d}",
			server.CPU,
			server.GetMemoryGB(),
		)
	}

	// grpc-health-probeでSERVINGになっていることを確認
	out, err := exec.Command("grpc-health-probe", "-addr", "unix:autoscaler.sock").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "status: SERVING") {
		t.Fatalf("grpc-health-prove: unexpected response: %s", string(out))
	}

	// 冷却期間待機(リソース更新日時が基準のため初回のスケールアップも影響を受ける)
	time.Sleep(180 * time.Second)

	/**************************************************************************
	 * Step 1-1: スケールアップ
	 *************************************************************************/
	log.Println("step1-1: scale up")

	// Grafana InputsにWebhookでUpリクエストを送信
	resp, err := http.Post("http://127.0.0.1:8080/up?resource-name=server", "text/plain", bytes.NewReader(grafanaWebhookBody))
	if err != nil {
		coreOutput.Fatal(t, err)
	}
	if resp.StatusCode != http.StatusOK {
		coreOutput.Fatalf(t,
			"Grafana Inputs returns unexpected status code: expected: 200 actual: %d",
			resp.StatusCode,
		)
	}

	// Coreのジョブ完了まで待機
	if err := coreOutput.WaitOutput(upJobDoneMarker, 10*time.Minute); err != nil {
		coreOutput.Fatal(t, err)
	}

	/**************************************************************************
	 * Step 1-2: スケールアップ結果の確認
	 *************************************************************************/
	log.Println("step1-2: check results")
	server, err = fetchSakuraCloudServer()
	if err != nil {
		t.Fatal(err)
	}
	if server.CPU != 2 || server.GetMemoryGB() != 4 {
		coreOutput.Fatalf(t,
			"server has unexpected plan: expected: {CPU:2, Memory:4} actual: {CPU:%d, Memory:%d}",
			server.CPU,
			server.GetMemoryGB(),
		)
	}

	/**************************************************************************
	 * Step 1-3: 冷却期間の確認
	 *************************************************************************/
	log.Println("step1-3: cooling down")
	resp, err = http.Post("http://127.0.0.1:8080/up?resource-name=server", "text/plain", bytes.NewReader(grafanaWebhookBody))
	if err != nil {
		coreOutput.Fatal(t, err)
	}
	if resp.StatusCode != http.StatusOK {
		coreOutput.Fatalf(t,
			"Grafana Inputs returns unexpected status code: expected: 200 actual: %d",
			resp.StatusCode,
		)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		coreOutput.Fatal(t, err)
	}
	if !strings.Contains(string(body), inCoolDownTimeResponse) {
		coreOutput.Fatalf(t,
			"Grafana Inputs returns unexpected response: expected: %s actual: %s",
			inCoolDownTimeResponse,
			string(body),
		)
	}
	// 冷却期間中である事のメッセージを受け取っているはず
	if err := inputsOutput.WaitOutput(inCoolDownTimeMarker, 10*time.Second); err != nil {
		coreOutput.Fatal(t, err)
	}

	// 冷却期間待機
	time.Sleep(180 * time.Second)

	// Terraformステートのリフレッシュ(複数回IDが変更されるため毎回リフレッシュしておく)
	e2e.TerraformRefresh() // nolint

	/**************************************************************************
	 * Step 2-1: スケールダウン
	 *************************************************************************/
	log.Println("step2-1: scale down")
	// Grafana InputsにWebhookでDownリクエストを送信
	resp, err = http.Post("http://127.0.0.1:8080/down?resource-name=server", "text/plain", bytes.NewReader(grafanaWebhookBody))
	if err != nil {
		coreOutput.Fatal(t, err)
	}
	if resp.StatusCode != http.StatusOK {
		coreOutput.Fatalf(t,
			"Grafana Inputs returns unexpected status code: expected: 200 actual: %d",
			resp.StatusCode,
		)
	}

	// Coreのジョブ完了まで待機
	if err := coreOutput.WaitOutput(downJobDoneMarker, 10*time.Minute); err != nil {
		coreOutput.Fatal(t, err)
	}

	/**************************************************************************
	 * Step 2-2: スケールダウン結果の確認
	 *************************************************************************/
	log.Println("step2-2: check results")
	server, err = fetchSakuraCloudServer()
	if err != nil {
		t.Fatal(err)
	}
	if server.CPU != 1 || server.GetMemoryGB() != 2 {
		coreOutput.Fatalf(t,
			"server has unexpected plan: expected: {CPU:1, Memory:2} actual: {CPU:%d, Memory:%d}",
			server.CPU,
			server.GetMemoryGB(),
		)
	}
	// Terraformステートのリフレッシュ(複数回IDが変更されるため毎回リフレッシュしておく)
	e2e.TerraformRefresh() // nolint
	coreOutput.Output()
	inputsOutput.Output()
}

func setup() {
	if err := e2e.TerraformInit(); err != nil {
		log.Fatal(err)
	}
	if err := e2e.TerraformApply(); err != nil {
		log.Fatal(err)
	}

	coreCmdOut, err := coreCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	inputCmdOut, err := inputCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := coreCmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := inputCmd.Start(); err != nil {
		log.Fatal(err)
	}

	coreOutput = e2e.NewOutput(coreCmdOut, "[Core]")
	inputsOutput = e2e.NewOutput(inputCmdOut, "[Grafana Inputs]")

	if err := coreOutput.WaitOutput(coreReadyMarker, 3*time.Second); err != nil {
		coreOutput.Output()
		log.Fatal(err)
	}
	if err := inputsOutput.WaitOutput(inputsReadyMarker, 3*time.Second); err != nil {
		inputsOutput.Output()
		log.Fatal(err)
	}
}

func teardown() {
	// shutdown inputs and core
	if inputCmd.Process != nil {
		if err := inputCmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Println(err)
		}
		if err := inputCmd.Wait(); err != nil {
			log.Println(err)
		}
	}

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
}

func fetchSakuraCloudServer() (*iaas.Server, error) {
	serverOp := iaas.NewServerOp(autoscalerE2E.SacloudAPICaller)

	found, err := serverOp.Find(context.Background(), "tk1b", &iaas.FindCondition{
		Filter: search.Filter{
			search.Key("Name"): search.PartialMatch("autoscaler-e2e-vertical-scaling"),
		},
	})
	if err != nil {
		return nil, err
	}

	if len(found.Servers) == 0 {
		return nil, fmt.Errorf("server 'autoscaler-e2e-vertical-scaling' not found on tk1b zone")
	}
	return found.Servers[0], nil
}

func waitResourceReadyLAndStartHTTPRequestLoop(ctx context.Context, t *testing.T) error {
	// 対象サーバごとにHTTPリクエストが通るようになるまで待つ
	serverOp := iaas.NewServerOp(autoscalerE2E.SacloudAPICaller)
	found, err := serverOp.Find(context.Background(), "tk1b", &iaas.FindCondition{
		Filter: search.Filter{
			search.Key("Name"): search.PartialMatch("autoscaler-e2e-vertical-scaling"),
		},
	})
	if err != nil {
		return err
	}
	if len(found.Servers) == 0 {
		return fmt.Errorf("server 'autoscaler-e2e-vertical-scaling-*' not found")
	}

	for _, server := range found.Servers {
		url := fmt.Sprintf("http://%s", server.Interfaces[0].IPAddress)
		if err := e2e.HttpRequestUntilSuccess(url, proxyLBReadyTimeout); err != nil {
			t.Fatal(err)
		}
	}

	// ELB
	elbOp := iaas.NewProxyLBOp(autoscalerE2E.SacloudAPICaller)
	foundElb, err := elbOp.Find(context.Background(), &iaas.FindCondition{
		Filter: search.Filter{
			search.Key("Name"): search.PartialMatch("autoscaler-e2e-vertical-scaling"),
		},
	})
	if err != nil {
		return err
	}

	if len(foundElb.ProxyLBs) == 0 {
		return fmt.Errorf("proxylb 'autoscaler-e2e-vertical-scaling' not found")
	}
	elb := foundElb.ProxyLBs[0]

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
