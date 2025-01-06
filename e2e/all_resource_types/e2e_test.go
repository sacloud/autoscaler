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

package all_resource_types

import (
	"log"
	"os/exec"
	"testing"

	"github.com/sacloud/packages-go/e2e"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	defer teardown()
	setup()

	m.Run()
}

func TestE2E_AllResourceTypes(t *testing.T) {
	cmd := exec.Command("autoscaler", "validate")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatal(err)
	}
	require.Equal(t, "OK\n", string(output))
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
	if err := e2e.TerraformDestroy(); err != nil {
		log.Fatal(err)
	}
}
