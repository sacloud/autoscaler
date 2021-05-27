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

// AutoScaler Core
//
// Usage:
//   autoscaler-handlers-fake [flags]
//
// Flags:
//   -address: (optional) URL of gRPC endpoint of the handler. default:`unix:autoscaler-handlers-fake.sock`
package main

import (
	"github.com/sacloud/autoscaler/handlers"
	"github.com/sacloud/autoscaler/handlers/fake"
)

func main() {
	handlers.Serve(&fake.Handler{})
}
