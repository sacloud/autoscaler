// Copyright 2021-2025 The sacloud/autoscaler Authors
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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCoolDown_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *CoolDown
		wantErr bool
	}{
		{
			name: "struct",
			data: []byte(`
up: 1
down: 2
`),
			want: &CoolDown{
				Up:   1,
				Down: 2,
			},
			wantErr: false,
		},
		{
			name: "int",
			data: []byte(`1`),
			want: &CoolDown{
				Up:   1,
				Down: 1,
				Keep: 1,
			},
			wantErr: false,
		},
		{
			name:    "invalid char",
			data:    []byte(`a`),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "empty",
			data:    []byte(``),
			want:    &CoolDown{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CoolDown{}
			if err := c.UnmarshalYAML(context.Background(), tt.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				require.EqualValues(t, tt.want, c)
			}
		})
	}
}
