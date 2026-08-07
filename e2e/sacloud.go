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

package e2e

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sacloud/autoscaler/version"
	"github.com/sacloud/sacloud-sdk-go/api/iaas"
	"github.com/sacloud/sacloud-sdk-go/common/packages/envvar"
	"github.com/sacloud/sacloud-sdk-go/common/saclient"
)

func init() {
	if envvar.StringFromEnvMulti([]string{"SAKURA_ACCESS_TOKEN", "SAKURACLOUD_ACCESS_TOKEN"}, "") == "" {
		panic("SAKURA_ACCESS_TOKEN (or SAKURACLOUD_ACCESS_TOKEN) is required")
	}
	if envvar.StringFromEnvMulti([]string{"SAKURA_ACCESS_TOKEN_SECRET", "SAKURACLOUD_ACCESS_TOKEN_SECRET"}, "") == "" {
		panic("SAKURA_ACCESS_TOKEN_SECRET (or SAKURACLOUD_ACCESS_TOKEN_SECRET) is required")
	}

	var sa saclient.Client
	if err := sa.SetEnviron(os.Environ()); err != nil {
		panic(err)
	}
	if err := sa.SetWith(
		saclient.WithUserAgent(fmt.Sprintf(
			"sacloud/autoscaler@v%s:e2e-test (%s/%s; +https://github.com/sacloud/autoscaler) %s",
			version.Version,
			runtime.GOOS,
			runtime.GOARCH,
			iaas.DefaultUserAgent,
		)),
		saclient.WithDefaultTimeout(300*time.Second),
	); err != nil {
		panic(err)
	}
	if err := sa.Populate(); err != nil {
		panic(err)
	}

	cfg, err := sa.EndpointConfig()
	if err != nil {
		panic(err)
	}
	if cfg.APIRootURL != "" {
		if strings.HasSuffix(cfg.APIRootURL, "/") {
			cfg.APIRootURL = strings.TrimRight(cfg.APIRootURL, "/")
		}
		iaas.SakuraCloudAPIRoot = cfg.APIRootURL
	}
	if len(cfg.Zones) > 0 {
		iaas.SakuraCloudZones = cfg.Zones
	}
	if cfg.DefaultZone != "" {
		iaas.APIDefaultZone = cfg.DefaultZone
	}

	SacloudAPICaller = iaas.NewClientFromSaclient(&sa)
}

var SacloudAPICaller iaas.APICaller
