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
	"io"
	"os"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/libsacloud/v2/sacloud/types"
	"github.com/stretchr/testify/require"
)

func TestNameOrSelector_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want NameOrSelector
	}{
		{
			name: "empty",
			data: []byte(``),
			want: NameOrSelector{ResourceSelector{}},
		},
		{
			name: "name",
			data: []byte(`foobar`),
			want: NameOrSelector{ResourceSelector{Names: []string{"foobar"}}},
		},
		{
			name: "selector",
			data: []byte(`names: ["foobar1", "foobar2"]`),
			want: NameOrSelector{ResourceSelector{Names: []string{"foobar1", "foobar2"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v NameOrSelector
			if err := yaml.UnmarshalWithOptions(tt.data, &v, yaml.Strict()); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			require.EqualValues(t, tt.want, v)
		})
	}
}

func TestIDOrSelector_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    IDOrSelector
		wantErr bool
	}{
		{
			name: "empty",
			data: []byte(``),
			want: IDOrSelector{ResourceSelector{}},
		},
		{
			name:    "invalid",
			data:    []byte(`foobar`),
			want:    IDOrSelector{},
			wantErr: true,
		},
		{
			name: "id",
			data: []byte(`123456789012`),
			want: IDOrSelector{ResourceSelector{ID: types.ID(123456789012)}},
		},
		{
			name: "selector",
			data: []byte(`names: ["foobar1", "foobar2"]`),
			want: IDOrSelector{ResourceSelector{Names: []string{"foobar1", "foobar2"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v IDOrSelector
			if err := yaml.UnmarshalWithOptions(tt.data, &v, yaml.Strict()); tt.wantErr != (err != nil) {
				t.Fatalf("unexpected error: %s", err)
			}
			require.EqualValues(t, tt.want, v)
		})
	}
}

func TestStringOrFilePath_UnmarshalYAML(t *testing.T) {
	file, err := os.CreateTemp("", "TestStringOrFilePath")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	if _, err := io.WriteString(file, "dummy-file"); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		data []byte
		want StringOrFilePath
	}{
		{
			name: "empty",
			data: []byte(``),
			want: StringOrFilePath(""),
		},
		{
			name: "string",
			data: []byte(`foobar`),
			want: StringOrFilePath("foobar"),
		},
		{
			name: "file",
			data: []byte(file.Name()),
			want: StringOrFilePath("dummy-file"),
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
