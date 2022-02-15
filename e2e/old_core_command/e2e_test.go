// Copyright 2021-2022 The sacloud/autoscaler Authors
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
	"log"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sacloud/autoscaler/e2e"
)

const (
	coreReadyMarker = `message=started address=autoscaler.sock`
)

var (
	coreCmd = exec.Command("autoscaler", "core", "start")
	output  = &e2e.Output{}
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
	go output.CollectOutputs("[Core]", coreOutputs)
	if err := output.WaitOutput(coreReadyMarker, 3*time.Second); err != nil {
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

	defer output.OutputLogs()
}

func setup() {
	if err := e2e.TerraformInit(); err != nil {
		log.Fatal(err)
	}
	if err := e2e.TerraformApply(); err != nil {
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
}
