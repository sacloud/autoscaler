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

//go:build !windows
// +build !windows

package webhook

import (
	_ "embed"
	"testing"

	"github.com/sacloud/autoscaler/test"
	"github.com/stretchr/testify/require"
)

//go:embed test/webhook.json
var webhookBody []byte

func TestInput_execCommand(t *testing.T) {
	tests := []struct {
		name           string
		executablePath string
		want           bool
		err            string
	}{
		{
			name:           "not found",
			executablePath: "test/not-found.sh",
			want:           false,
			err:            `exec: "test/not-found.sh": stat test/not-found.sh: no such file or directory`,
		},
		{
			name:           "alerting",
			executablePath: "test/alerting.sh",
			want:           true,
			err:            "",
		},
		{
			name:           "not alerting",
			executablePath: "test/not-alerting.sh",
			want:           false,
			err:            `command "test/not-alerting.sh" returned non zero status: exit status 4`,
		},
		{
			name:           "non executable",
			executablePath: "test/non-executable.sh",
			want:           false,
			err:            `exec: "test/non-executable.sh": permission denied`,
		},
		{
			name:           "invalid executable",
			executablePath: "test/invalid.txt",
			want:           false,
			err:            `command "test/invalid.txt" returned non zero status: fork/exec test/invalid.txt: exec format error`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in, err := NewInput("", "", "", test.Logger, []string{"POST"}, tt.executablePath)
			if err != nil {
				require.EqualError(t, err, tt.err)
				return
			}

			got, err := in.execCommand(webhookBody)
			if err != nil {
				require.EqualError(t, err, tt.err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
