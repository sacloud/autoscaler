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

package e2e

import (
	"log"
	"os"
	"os/exec"
)

func TerraformInit() error {
	stateFile := "terraform.tfstate"
	if _, err := os.Stat(stateFile); err == nil {
		if err := os.Remove(stateFile); err != nil {
			return err
		}
	}
	return exec.Command("terraform", "init").Run()
}

func TerraformApply() error {
	return exec.Command("terraform", "apply", "-auto-approve").Run()
}

func TerraformRefresh() error {
	return exec.Command("terraform", "apply", "-refresh-only", "-auto-approve").Run()
}

func TerraformDestroy() error {
	if os.Getenv("SKIP_CLEANUP") != "" {
		log.Println("Cleanup skipped")
		return nil
	}
	return exec.Command("terraform", "destroy", "-auto-approve").Run()
}

func CleanupAutoScalerSocketFile() error {
	socketFile := "autoscaler.sock"
	if _, err := os.Stat(socketFile); err == nil {
		if err := os.Remove(socketFile); err != nil {
			return err
		}
	}
	return nil
}
