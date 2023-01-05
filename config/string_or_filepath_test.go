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

package config

import (
	"bytes"
	"context"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/autoscaler/log"
	"github.com/stretchr/testify/require"
)

func TestStringOrFilePath_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		strictMode bool
		warning    string
		want       StringOrFilePath
	}{
		{
			name: "empty",
			data: []byte(``),
			want: StringOrFilePath{
				content:    "",
				isFilePath: false,
			},
		},
		{
			name: "string",
			data: []byte(`foobar`),
			want: StringOrFilePath{
				content:    "foobar",
				isFilePath: false,
			},
		},
		{
			name: "file",
			data: []byte("dummy.txt"),
			want: StringOrFilePath{
				content:    "dummy",
				isFilePath: true,
			},
		},
		{
			name:       "strict",
			data:       []byte("dummy.txt"),
			strictMode: true,
			want: StringOrFilePath{
				content:    "dummy.txt",
				isFilePath: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf := bytes.NewBufferString("")
			ctx := NewLoadConfigContext(context.Background(), tt.strictMode, log.NewLogger(&log.LoggerOption{
				Writer: logBuf,
			}))
			var v StringOrFilePath
			if err := yaml.UnmarshalContext(ctx, tt.data, &v, yaml.Strict()); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			require.EqualValues(t, tt.want, v)
			require.Equal(t, tt.warning, logBuf.String())
		})
	}
}
