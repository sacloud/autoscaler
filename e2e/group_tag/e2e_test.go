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

package group_tag

import (
	"context"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	autoscalerE2E "github.com/sacloud/autoscaler/e2e"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/search"
	serverService "github.com/sacloud/iaas-service-go/server"
	"github.com/sacloud/packages-go/e2e"
)

const (
	coreReadyMarker = `message=started address=autoscaler.sock`
	upJobDoneMarker = `request=Up source=default resource=autoscaler-e2e-group-tag status=JOB_DONE`
)

var (
	coreCmd = exec.Command("autoscaler", "start")
	upCmd   = exec.Command("autoscaler", "inputs", "direct",
		"--desired-state-name", "largest",
		"--resource-name", "autoscaler-e2e-group-tag",
		"up")

	zones          = []string{"tk1b", "is1b"}
	e2eTestTimeout = 20 * time.Minute

	output *e2e.Output
)

func TestMain(m *testing.M) {
	defer teardown()
	setup()

	m.Run()
}

func TestE2E_GroupTag(t *testing.T) {
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

	/**************************************************************************
	 * Step 1-2: スケールアウト結果の確認
	 *************************************************************************/
	log.Println("step1-2: check results")
	servers, err := fetchSakuraCloudServers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(servers) != 10 {
		output.Fatalf(t,
			"got unexpected server count: expected:10 actual:%d",
			len(servers),
		)
	}
	zoneAndGroupTag := make(map[string]struct{})
	for _, server := range servers {
		if len(server.Tags) == 0 {
			output.Fatalf(t, "got unexpected server tag: %s", server.Tags)
		}
		zoneAndGroupTag[server.Zone.Name+server.Tags[0]] = struct{}{}
	}
	if len(zoneAndGroupTag) != 8 {
		output.Fatalf(t, "got unexpected group tag")
	}

	output.Output()
}

func setup() {
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

	// サーバはTerraform管理外のためここでクリーンアップする
	servers, err := fetchSakuraCloudServers(context.Background())
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

func fetchSakuraCloudServers(ctx context.Context) ([]*iaas.Server, error) {
	serverOp := iaas.NewServerOp(autoscalerE2E.SacloudAPICaller)

	var servers []*iaas.Server
	for _, zone := range zones {
		found, err := serverOp.Find(ctx, zone, &iaas.FindCondition{
			Filter: search.Filter{
				search.Key("Name"): search.PartialMatch("autoscaler-e2e-group-tag"),
			},
		})
		if err != nil {
			return nil, err
		}
		servers = append(servers, found.Servers...)
	}

	return servers, nil
}
