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

package old_core_command

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	autoscalerE2E "github.com/sacloud/autoscaler/e2e"
	"github.com/sacloud/sacloud-sdk-go/api/iaas"
	"github.com/sacloud/sacloud-sdk-go/api/iaas/types"
	sdkE2E "github.com/sacloud/sacloud-sdk-go/common/packages/e2e"
)

const (
	coreReadyMarker = `msg=started address=autoscaler.sock`
)

var (
	coreCmd    = exec.Command("autoscaler", "core", "start")
	setupELBID types.ID
)

func TestMain(m *testing.M) {
	defer teardown()
	setup()

	m.Run()
}

func TestE2E_OldCoreCommand(t *testing.T) {

	/**************************************************************************
	 * Step 1: 古いコマンド(autoscaler core start)でのCoreの起動確認
	 *************************************************************************/
	log.Println("step0: setup")
	coreOutputs, err := coreCmd.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := coreCmd.Start(); err != nil {
		t.Fatal(err)
	}

	output := sdkE2E.NewOutput(coreOutputs, "[Core]")
	defer output.Output()

	if err := output.WaitOutput(coreReadyMarker, time.Minute); err != nil {
		t.Fatal(err)
	}

	// grpc-health-probeでSERVINGになっていることを確認
	out, err := exec.Command("grpc-health-probe", "-addr", "unix:autoscaler.sock").CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), "status: SERVING") {
		t.Fatalf("grpc-health-prove: unexpected response: %s", string(out))
	}

}

func setup() {
	log.SetOutput(io.Discard)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	ctx := context.Background()
	elbOp := iaas.NewProxyLBOp(autoscalerE2E.SacloudAPICaller)

	elb, err := elbOp.Create(ctx, &iaas.ProxyLBCreateRequest{
		Plan: types.ProxyLBPlans.CPS100,
		HealthCheck: &iaas.ProxyLBHealthCheck{
			Protocol:  "http",
			Path:      "/",
			DelayLoop: 10,
		},
		BindPorts: []*iaas.ProxyLBBindPort{
			{
				ProxyMode: "http",
				Port:      80,
			},
		},
		Timeout: &iaas.ProxyLBTimeout{InactiveSec: 10},
		Region:  "is1",
		Name:    "autoscaler-e2e-old-core-command",
	})
	if err != nil {
		panic(err)
	}
	setupELBID = elb.ID
}

func teardown() {
	if setupELBID != 0 {
		_ = iaas.NewProxyLBOp(autoscalerE2E.SacloudAPICaller).Delete(context.Background(), setupELBID)
	}

	if coreCmd.Process != nil {
		if err := coreCmd.Process.Signal(syscall.SIGINT); err != nil {
			log.Println(err)
		}
		if err := coreCmd.Wait(); err != nil {
			log.Println(err)
		}
	}
}
