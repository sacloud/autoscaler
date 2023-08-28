// Copyright 2021-2023 The sacloud/autoscaler Authors
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

package zabbix

import (
	"log/slog"
	"net/http"

	"github.com/sacloud/autoscaler/version"
)

type Input struct {
	dest       string
	addr       string
	configPath string
	logger     *slog.Logger
}

func NewInput(dest, addr, configPath string, logger *slog.Logger) *Input {
	return &Input{
		dest:       dest,
		addr:       addr,
		configPath: configPath,
		logger:     logger,
	}
}

func (in *Input) Name() string {
	return "zabbix"
}

func (in *Input) Version() string {
	return version.FullVersion()
}

func (in *Input) Destination() string {
	return in.dest
}

func (in *Input) ListenAddress() string {
	return in.addr
}

func (in *Input) ConfigPath() string {
	return in.configPath
}

func (in *Input) GetLogger() *slog.Logger {
	return in.logger
}

func (in *Input) ShouldAccept(req *http.Request) (bool, error) {
	if req.Method == http.MethodPost {
		// POSTでリクエストされたらtrueを返す(Bodyの中は読まない)
		return true, nil
	}
	return false, nil
}
