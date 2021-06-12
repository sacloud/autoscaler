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

package builtins

import "github.com/sacloud/libsacloud/v2/sacloud"

// SakuraCloudAPICaller さくらのクラウドAPIを利用するビルトインハンドラ向けのインターフェース
type SakuraCloudAPICaller interface {
	APICaller() sacloud.APICaller
	SetAPICaller(caller sacloud.APICaller)
}

// SakuraCloudAPIClient SakuraCloudAPICallerの実装、各ハンドラーに埋め込んで利用する
type SakuraCloudAPIClient struct {
	caller sacloud.APICaller
}

func (c *SakuraCloudAPIClient) APICaller() sacloud.APICaller {
	return c.caller
}

func (c *SakuraCloudAPIClient) SetAPICaller(caller sacloud.APICaller) {
	c.caller = caller
}