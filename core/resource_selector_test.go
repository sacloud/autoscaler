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

package core

import (
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/iaas-api-go"
	"github.com/sacloud/iaas-api-go/search"
	"github.com/sacloud/iaas-api-go/types"
	"github.com/stretchr/testify/require"
)

func TestResourceSelector_Validate(t *testing.T) {
	type fields struct {
		ID    types.ID
		Tags  []string
		Names []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// invalid
		{
			name:    "empty",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "ID with Names",
			fields: fields{
				ID:    types.ID(1),
				Names: []string{"names"},
			},
			wantErr: true,
		},
		{
			name: "ID with Tags",
			fields: fields{
				ID:   types.ID(1),
				Tags: []string{"tags"},
			},
			wantErr: true,
		},
		{
			name: "ID with Names and Tags",
			fields: fields{
				ID:    types.ID(1),
				Names: []string{"names"},
				Tags:  []string{"tags"},
			},
			wantErr: true,
		},
		// valid
		{
			name: "only ID",
			fields: fields{
				ID: types.ID(1),
			},
			wantErr: false,
		},
		{
			name: "only Names",
			fields: fields{
				Names: []string{"names"},
			},
			wantErr: false,
		},
		{
			name: "only Tags",
			fields: fields{
				Tags: []string{"tags"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ResourceSelector{
				ID:    tt.fields.ID,
				Tags:  tt.fields.Tags,
				Names: tt.fields.Names,
			}
			if err := rs.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceSelector_findCondition(t *testing.T) {
	type fields struct {
		ID    types.ID
		Tags  []string
		Names []string
	}
	tests := []struct {
		name   string
		fields fields
		want   *iaas.FindCondition
	}{
		{
			name: "ID",
			fields: fields{
				ID: types.ID(1),
			},
			want: &iaas.FindCondition{
				Filter: search.Filter{
					search.Key("ID"): search.ExactMatch(types.ID(1).String()),
				},
			},
		},
		{
			name: "Names",
			fields: fields{
				Names: []string{"names"},
			},
			want: &iaas.FindCondition{
				Filter: search.Filter{
					search.Key("Name"): search.PartialMatch("names"),
				},
			},
		},
		{
			name: "Tags",
			fields: fields{
				Tags: []string{"tags"},
			},
			want: &iaas.FindCondition{
				Filter: search.Filter{
					search.Key("Tags.Name"): search.PartialMatch("tags"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &ResourceSelector{
				ID:    tt.fields.ID,
				Tags:  tt.fields.Tags,
				Names: tt.fields.Names,
			}
			got := rs.findCondition()
			require.EqualValues(t, tt.want, got)
		})
	}
}

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

func TestIdOrNameOrSelector_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want IdOrNameOrSelector
	}{
		{
			name: "empty",
			data: []byte(``),
			want: IdOrNameOrSelector{ResourceSelector{}},
		},
		{
			name: "id",
			data: []byte(`123456789012`),
			want: IdOrNameOrSelector{ResourceSelector{ID: 123456789012}},
		},
		{
			name: "name",
			data: []byte(`foobar`),
			want: IdOrNameOrSelector{ResourceSelector{Names: []string{"foobar"}}},
		},
		{
			name: "selector",
			data: []byte(`names: ["foobar1", "foobar2"]`),
			want: IdOrNameOrSelector{ResourceSelector{Names: []string{"foobar1", "foobar2"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v IdOrNameOrSelector
			if err := yaml.UnmarshalWithOptions(tt.data, &v, yaml.Strict()); err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			require.EqualValues(t, tt.want, v)
		})
	}
}
