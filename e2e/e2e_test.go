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

// +build e2e

package e2e

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2"
	"github.com/sacloud/libsacloud/v2/helper/api"
	"github.com/sacloud/libsacloud/v2/sacloud"
	"github.com/sacloud/libsacloud/v2/sacloud/search"
)

const (
	coreReadyMarker        = `message="autoscaler core started" address=autoscaler.sock`
	inputsReadyMarker      = `message=started address=127.0.0.1:8080`
	upJobDoneMarker        = `request=Up source=default resource=server status=JOB_DONE`
	downJobDoneMarker      = `request=Down source=default resource=server status=JOB_DONE`
	inCoolDownTimeMarker   = `job-message="job is in an unacceptable state"`
	inCoolDownTimeResponse = `"message":"job is in an unacceptable state"`
)

var (
	coreCmd    = exec.Command("autoscaler", "server", "start")
	inputCmd   = exec.Command("autoscaler", "inputs", "grafana", "--addr", "127.0.0.1:8080")
	refreshCmd = exec.Command("terraform", "apply", "-refresh-only", "-auto-approve")
	outputs    []string
	mu         sync.Mutex

	proxyLBReadyTimeout = 5 * time.Minute
	e2eTestTimeout      = 20 * time.Minute

	//go:embed webhook.json
	grafanaWebhookBody []byte
)

var apiCaller = api.NewCaller(&api.CallerOptions{
	AccessToken:       os.Getenv("SAKURACLOUD_ACCESS_TOKEN"),
	AccessTokenSecret: os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET"),
	UserAgent: fmt.Sprintf(
		"sacloud/autoscaler/v%s/e2e-test (%s/%s; +https://github.com/sacloud/autoscaler) libsacloud/%s",
		version.Version,
		runtime.GOOS,
		runtime.GOARCH,
		libsacloud.Version,
	),
	HTTPRequestTimeout:   300,
	HTTPRequestRateLimit: 10,
	RetryMax:             10,
	TraceAPI:             os.Getenv("SAKURACLOUD_TRACE") != "",
	TraceHTTP:            os.Getenv("SAKURACLOUD_TRACE") != "",
})

func TestMain(m *testing.M) {
	setup()
	defer teardown()

	m.Run()
}

func TestE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), e2eTestTimeout)
	defer cancel()

	/**************************************************************************
	 * Step 0: 現在のクラウド上のリソースの確認/ポーリング開始
	 *************************************************************************/

	// ProxyLBへのHTTPリクエストが通るようになるまで待ち & ポーリング開始
	if err := waitProxyLBAndStartHTTPRequestLoop(ctx, t); err != nil {
		t.Fatal(err)
	}

	// サーバプランの確認(前提)
	server, err := fetchSakuraCloudServer()
	if err != nil {
		t.Fatal(err)
	}
	if server.CPU != 1 || server.GetMemoryGB() != 1 {
		fatalWithStderrOutputs(t,
			fmt.Sprintf(
				"server has unexpected initial plan: expected: {CPU:1, Memory:1} actual: {CPU:%d, Memory:%d}",
				server.CPU,
				server.GetMemoryGB(),
			),
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

	/**************************************************************************
	 * Step 1-1: スケールアップ
	 *************************************************************************/
	// Grafana InputsにWebhookでUpリクエストを送信
	resp, err := http.Post("http://127.0.0.1:8080/up?resource-name=server", "text/plain", bytes.NewReader(grafanaWebhookBody))
	if err != nil {
		fatalWithStderrOutputs(t, err)
	}
	if resp.StatusCode != http.StatusOK {
		fatalWithStderrOutputs(t,
			fmt.Sprintf("Grafana Inputs returns unexpected status code: expected: 200 actual: %d", resp.StatusCode))
	}

	// Coreのジョブ完了まで待機
	if err := waitOutput(upJobDoneMarker, 10*time.Minute); err != nil {
		fatalWithStderrOutputs(t, err)
	}

	/**************************************************************************
	 * Step 1-2: スケールアップ結果の確認
	 *************************************************************************/
	server, err = fetchSakuraCloudServer()
	if err != nil {
		t.Fatal(err)
	}
	if server.CPU != 2 || server.GetMemoryGB() != 2 {
		fatalWithStderrOutputs(t,
			fmt.Sprintf(
				"server has unexpected plan: expected: {CPU:2, Memory:2} actual: {CPU:%d, Memory:%d}",
				server.CPU,
				server.GetMemoryGB(),
			),
		)
	}

	/**************************************************************************
	 * Step 1-3: 冷却期間の確認
	 *************************************************************************/
	resp, err = http.Post("http://127.0.0.1:8080/up?resource-name=server", "text/plain", bytes.NewReader(grafanaWebhookBody))
	if err != nil {
		fatalWithStderrOutputs(t, err)
	}
	if resp.StatusCode != http.StatusOK {
		fatalWithStderrOutputs(t,
			fmt.Sprintf("Grafana Inputs returns unexpected status code: expected: 200 actual: %d", resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fatalWithStderrOutputs(t, err)
	}
	if !strings.Contains(string(body), inCoolDownTimeResponse) {
		fatalWithStderrOutputs(t,
			fmt.Sprintf("Grafana Inputs returns unexpected response: expected: %s actual: %s", inCoolDownTimeResponse, string(body)))
	}
	// 冷却期間中である事のメッセージを受け取っているはず
	if err := waitOutput(inCoolDownTimeMarker, 10*time.Second); err != nil {
		fatalWithStderrOutputs(t, err)
	}

	// 冷却期間待機
	time.Sleep(30 * time.Second)

	// Terraformステートのリフレッシュ(複数回IDが変更されるため毎回リフレッシュしておく)
	refreshCmd.Run() // nolint

	/**************************************************************************
	 * Step 2-1: スケールダウン
	 *************************************************************************/
	// Grafana InputsにWebhookでDownリクエストを送信
	resp, err = http.Post("http://127.0.0.1:8080/down?resource-name=server", "text/plain", bytes.NewReader(grafanaWebhookBody))
	if err != nil {
		fatalWithStderrOutputs(t, err)
	}
	if resp.StatusCode != http.StatusOK {
		fatalWithStderrOutputs(t,
			fmt.Sprintf("Grafana Inputs returns unexpected status code: expected: 200 actual: %d", resp.StatusCode))
	}

	// Coreのジョブ完了まで待機
	if err := waitOutput(downJobDoneMarker, 10*time.Minute); err != nil {
		fatalWithStderrOutputs(t, err)
	}

	/**************************************************************************
	 * Step 2-2: スケールダウン結果の確認
	 *************************************************************************/
	server, err = fetchSakuraCloudServer()
	if err != nil {
		t.Fatal(err)
	}
	if server.CPU != 1 || server.GetMemoryGB() != 1 {
		fatalWithStderrOutputs(t,
			fmt.Sprintf(
				"server has unexpected plan: expected: {CPU:1, Memory:1} actual: {CPU:%d, Memory:%d}",
				server.CPU,
				server.GetMemoryGB(),
			),
		)
	}
	// Terraformステートのリフレッシュ(複数回IDが変更されるため毎回リフレッシュしておく)
	refreshCmd.Run() // nolint
}

func setup() {
	coreOutputs, err := coreCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	inputOutputs, err := inputCmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := coreCmd.Start(); err != nil {
		log.Fatal(err)
	}
	if err := inputCmd.Start(); err != nil {
		log.Fatal(err)
	}

	go collectOutputs("[Core]", coreOutputs)
	go collectOutputs("[Grafana Inputs]", inputOutputs)

	if err := waitOutput(coreReadyMarker, 3*time.Second); err != nil {
		logOutputs()
		log.Fatal(err)
	}
	if err := waitOutput(inputsReadyMarker, 3*time.Second); err != nil {
		logOutputs()
		log.Fatal(err)
	}
}

func teardown() {
	// shutdown inputs and core
	if err := inputCmd.Process.Signal(syscall.SIGINT); err != nil {
		log.Println(err)
	}
	if err := inputCmd.Wait(); err != nil {
		log.Println(err)
	}

	if err := coreCmd.Process.Signal(syscall.SIGINT); err != nil {
		log.Println(err)
	}
	if err := coreCmd.Wait(); err != nil {
		log.Println(err)
	}
}

func collectOutputs(prefix string, reader io.ReadCloser) {
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		mu.Lock()
		outputs = append(outputs, prefix+" "+line)
		mu.Unlock()
	}
}

func waitOutput(marker string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	doneCh := make(chan error)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				doneCh <- ctx.Err()
			case <-ticker.C:
				if isMarkerExistInOutputs(marker) {
					doneCh <- nil
				}
			}
		}
	}()

	return <-doneCh
}

func isMarkerExistInOutputs(marker string) bool {
	mu.Lock()
	defer mu.Unlock()
	for _, line := range outputs {
		if strings.Contains(line, marker) {
			return true
		}
	}
	return false
}

func waitForProxyLBReady(url string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	doneCh := make(chan error)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				doneCh <- ctx.Err()
				return
			case <-ticker.C:
				if err := httpRequestToProxyLB(url); err != nil {
					continue
				}
				doneCh <- nil
				return
			}
		}
	}()

	return <-doneCh
}

func httpRequestToProxyLB(url string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("got unexpected status code: %d", res.StatusCode)
	}
	return nil
}

func logOutputs() {
	mu.Lock()
	defer mu.Unlock()
	log.Println("Outputs:::\n" + strings.Join(outputs, "\n"))
}

func fatalWithStderrOutputs(t *testing.T, args ...interface{}) {
	logOutputs()
	t.Fatal(args...)
}

func fetchSakuraCloudServer() (*sacloud.Server, error) {
	serverOp := sacloud.NewServerOp(apiCaller)

	found, err := serverOp.Find(context.Background(), "is1a", &sacloud.FindCondition{
		Count: 1,
		Filter: search.Filter{
			search.Key("Name"): search.ExactMatch("autoscaler-e2e-test"),
		},
	})
	if err != nil {
		return nil, err
	}

	if len(found.Servers) == 0 {
		return nil, fmt.Errorf("server 'autoscaler-e2e-test' not found on is1a zone")
	}
	return found.Servers[0], nil
}

func waitProxyLBAndStartHTTPRequestLoop(ctx context.Context, t *testing.T) error {
	elbOp := sacloud.NewProxyLBOp(apiCaller)
	found, err := elbOp.Find(context.Background(), &sacloud.FindCondition{
		Count: 1,
		Filter: search.Filter{
			search.Key("Name"): search.ExactMatch("autoscaler-e2e-test"),
		},
	})
	if err != nil {
		return err
	}

	if len(found.ProxyLBs) == 0 {
		return fmt.Errorf("proxylb 'autoscaler-e2e-test' not found")
	}
	elb := found.ProxyLBs[0]

	// vip宛にリクエストが通るまで待機
	url := fmt.Sprintf("http://%s", elb.VirtualIPAddress)
	if err := waitForProxyLBReady(url, proxyLBReadyTimeout); err != nil {
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
				if err := httpRequestToProxyLB(url); err != nil {
					t.Error(err)
					return
				}
			}
		}
	}()
	return nil
}
