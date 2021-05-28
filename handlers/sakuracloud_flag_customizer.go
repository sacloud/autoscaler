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

package handlers

import (
	"flag"
	"os"

	"github.com/sacloud/libsacloud/v2/helper/api"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type SakuraCloudFlagCustomizer struct {
	token  string
	secret string
}

func (c *SakuraCloudFlagCustomizer) CustomizeFlags(fs *flag.FlagSet) {
	fs.StringVar(&c.token, "token", "", "API token for SAKURA Cloud. This value can also be set via 'SAKURACLOUD_ACCESS_TOKEN' environment variable")
	fs.StringVar(&c.secret, "secret", "", "API secret for SAKURA Cloud. This value can also be set via 'SAKURACLOUD_ACCESS_TOKEN_SECRET' environment variable")
}

func (c *SakuraCloudFlagCustomizer) Token() string {
	if c.token != "" {
		return c.token
	}
	return os.Getenv("SAKURACLOUD_ACCESS_TOKEN")
}

func (c *SakuraCloudFlagCustomizer) Secret() string {
	if c.secret != "" {
		return c.secret
	}
	return os.Getenv("SAKURACLOUD_ACCESS_TOKEN_SECRET")
}
func (c *SakuraCloudFlagCustomizer) APIClient() sacloud.APICaller {
	return api.NewCaller(&api.CallerOptions{
		AccessToken:       c.Token(),
		AccessTokenSecret: c.Secret(),
		//APIRootURL:           "",
		//DefaultZone:          "",
		//AcceptLanguage:       "",
		//HTTPClient:           nil,
		//HTTPRequestTimeout:   0,
		//HTTPRequestRateLimit: 0,
		//RetryMax:             0,
		//RetryWaitMax:         0,
		//RetryWaitMin:         0,
		UserAgent: "sacloud/autoscaler-handlers", // TODO カスタマイズ可能にしたい
		TraceAPI:  os.Getenv("SAKURACLOUD_TRACE") != "",
		TraceHTTP: os.Getenv("SAKURACLOUD_TRACE") != "",
		FakeMode:  os.Getenv("FAKE_MODE") != "",
	})
}
