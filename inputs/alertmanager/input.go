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

package alertmanager

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/sacloud/autoscaler/version"
)

type Input struct{}

func (in *Input) Name() string {
	return "alertmanager"
}

func (in *Input) Version() string {
	return version.FullVersion()
}

func (in *Input) ShouldAccept(req *http.Request) (bool, error) {
	if req.Method == http.MethodPost {
		reqData, err := io.ReadAll(req.Body)
		if err != nil {
			return false, err
		}
		var received alertManagerWebhookBody
		if err := json.Unmarshal(reqData, &received); err != nil {
			return false, err
		}
		if received.Status == "firing" {
			return true, nil
		}
	}
	return false, nil
}

type alertManagerWebhookBody struct {
	Status string `json:"status"`
}