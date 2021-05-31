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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResources_Walk(t *testing.T) {
	type args struct {
		fn ResourceWalkFunc
	}
	var results []string
	tests := []struct {
		name    string
		r       Resources
		args    args
		wantErr bool
	}{
		{
			name: "order",
			r: Resources{
				&EnhancedLoadBalancer{
					ResourceBase: &ResourceBase{
						TypeName: "ELB",
						TargetSelector: &ResourceSelector{
							ID: 3,
						},
						Children: Resources{
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 1,
									},
								},
							},
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 2,
									},
								},
							},
						},
					},
				},
				&EnhancedLoadBalancer{
					ResourceBase: &ResourceBase{
						TypeName: "ELB",
						TargetSelector: &ResourceSelector{
							ID: 6,
						},
						Children: Resources{
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 4,
									},
								},
							},
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 5,
									},
								},
							},
						},
					},
				},
			},
			args: args{
				fn: func(r Resource) error {
					results = append(results, r.Selector().ID.String())
					return nil
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.r.Walk(tt.args.fn); (err != nil) != tt.wantErr {
				t.Errorf("Walk() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.EqualValues(t, []string{"1", "2", "3", "4", "5", "6"}, results)
		})
	}
}
