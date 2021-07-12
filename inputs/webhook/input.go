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

package webhook

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os/exec"

	"github.com/sacloud/autoscaler/log"
	"github.com/sacloud/autoscaler/version"
)

type Input struct {
	dest       string
	addr       string
	configPath string
	logger     *log.Logger

	acceptHTTPMethods []string
	executablePath    string
}

func NewInput(dest, addr, configPath string, logger *log.Logger, acceptHTTPMethods []string, executablePath string) (*Input, error) {
	if len(acceptHTTPMethods) == 0 {
		return nil, fmt.Errorf("acceptHTTPMethod: required")
	}
	if len(executablePath) == 0 {
		return nil, fmt.Errorf("executablePath: required")
	}

	execPath, err := exec.LookPath(executablePath)
	if err != nil {
		return nil, err
	}

	return &Input{
		dest:       dest,
		addr:       addr,
		configPath: configPath,
		logger:     logger,

		acceptHTTPMethods: acceptHTTPMethods,
		executablePath:    execPath,
	}, nil
}
func (in *Input) Name() string {
	return "webhook"
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

func (in *Input) GetLogger() *log.Logger {
	return in.logger
}

func (in *Input) ShouldAccept(req *http.Request) (bool, error) {
	if !in.allowedMethod(req) {
		return false, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return false, err
	}

	return in.execCommand(body)
}

func (in *Input) allowedMethod(req *http.Request) bool {
	for _, m := range in.acceptHTTPMethods {
		if m == req.Method {
			return true
		}
	}
	return false
}

func (in *Input) execCommand(body []byte) (bool, error) {
	cmd := exec.Command(in.executablePath, string(body))
	bodyReader := bytes.NewReader(body)
	cmd.Stdin = bodyReader

	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("command %q returned non zero status: %s", in.executablePath, err)
	}
	return cmd.ProcessState.Success(), nil
}
