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

package core

import (
	"os"

	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/libsacloud/v2/helper/api"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

var (
	testZone   = "is1a"
	testLogger = log.NewLogger(&log.LoggerOption{
		Writer:    os.Stderr,
		JSON:      false,
		TimeStamp: true,
		Caller:    true,
		Level:     log.LevelDebug,
	})
)

func testAPIClient() sacloud.APICaller {
	return api.NewCaller(&api.CallerOptions{
		AccessToken:       "fake",
		AccessTokenSecret: "fake",
		UserAgent:         "sacloud/autoscaler/fake",
		TraceAPI:          os.Getenv("SAKURACLOUD_TRACE") != "",
		TraceHTTP:         os.Getenv("SAKURACLOUD_TRACE") != "",
		FakeMode:          true,
	})
}
