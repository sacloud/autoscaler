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

package version

import (
	"fmt"
	"runtime"
)

var (
	// Version app version
	Version = "0.16.2"
	// Revision git commit short commithash
	Revision = "xxxxxx" // set on build time
)

// FullVersion return full version text
func FullVersion() string {
	return fmt.Sprintf("v%s %s/%s, build %s", Version, runtime.GOOS, runtime.GOARCH, Revision)
}
