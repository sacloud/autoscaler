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
	"fmt"
	"os"
	"runtime"

	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/libsacloud/v2"
	"github.com/sacloud/libsacloud/v2/helper/api"
)

var SacloudAPICaller = api.NewCaller(&api.CallerOptions{
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