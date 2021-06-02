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

// AutoScaler Inputs: Direct
//
// Usage:
//   autoscaler-inputs-direct [flags] up|down|status
//
// Arguments:
//   up: run the Up func
//   down: run the Down func
//
// Flags:
//   -dest: (optional) URL of gRPC endpoint of AutoScaler Core. default:`unix:autoscaler.sock`
//   -action: (optional) Name of the action to perform. default:`default`
//   -group: (optional) Name of the target resource group. default:`default`
//   -source: (optional) A string representing the request source, passed to AutoScaler Core. default:`default`
package main

import (
	"github.com/sacloud/autoscaler/inputs"
	"github.com/sacloud/autoscaler/inputs/alertmanager"
)

func main() {
	inputs.Serve(&alertmanager.Input{})
}
