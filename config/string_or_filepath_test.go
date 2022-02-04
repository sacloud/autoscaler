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

package config

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
)

func TestStringOrFilePath_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want StringOrFilePath
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v StringOrFilePath
			if err := yaml.UnmarshalWithOptions(tt.data, &v, yaml.Strict()); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			require.EqualValues(t, tt.want, v)
		})
	}
}
