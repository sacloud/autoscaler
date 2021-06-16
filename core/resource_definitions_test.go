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
		forwardFn  ResourceDefWalkFunc
		backwardFn ResourceDefWalkFunc
	}
	var results []string
	tests := []struct {
		name     string
		r        ResourceDefinitions
		args     args
		expected []string
		wantErr  bool
	}{
		{
			name: "order",
			r: ResourceDefinitions{
				&EnhancedLoadBalancer{
					ResourceBase: &ResourceBase{
						TypeName: "ELB",
						TargetSelector: &ResourceSelector{
							ID: 1,
						},
						children: ResourceDefinitions{
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 2,
									},
								},
							},
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 3,
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
							ID: 4,
						},
						children: ResourceDefinitions{
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 5,
									},
									children: ResourceDefinitions{
										&Server{
											ResourceBase: &ResourceBase{
												TypeName: "Server",
												TargetSelector: &ResourceSelector{
													ID: 6,
												},
											},
										},
										&Server{
											ResourceBase: &ResourceBase{
												TypeName: "Server",
												TargetSelector: &ResourceSelector{
													ID: 7,
												},
											},
										},
									},
								},
							},
							&Server{
								ResourceBase: &ResourceBase{
									TypeName: "Server",
									TargetSelector: &ResourceSelector{
										ID: 8,
									},
								},
							},
						},
					},
				},
			},
			args: args{
				forwardFn: func(r ResourceDefinition) error {
					results = append(results, "forward"+r.Selector().ID.String())
					return nil
				},
				backwardFn: func(r ResourceDefinition) error {
					results = append(results, "backward"+r.Selector().ID.String())
					return nil
				},
			},
			expected: []string{
				// forwardは親から、backwardは子の処理後
				"forward1",

				// 末端なためforward/backwardが順に呼び出し
				"forward2",
				"backward2",

				// 末端なためforward/backwardが順に呼び出し
				"forward3",
				"backward3",

				// 子のbackwardの後で親のbackward
				"backward1",

				// ネストが深いパターン
				"forward4", // 4

				"forward5", // 4-5

				"forward6", // 4-5-6
				"backward6",

				"forward7", // 4-5-7
				"backward7",

				"backward5", // 4-5

				"forward8", // 4-8
				"backward8",

				"backward4", // 4
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results = []string{}
			if err := tt.r.Walk(tt.args.forwardFn, tt.args.backwardFn); (err != nil) != tt.wantErr {
				t.Errorf("Walk() error = %v, wantErr %v", err, tt.wantErr)
			}
			require.EqualValues(t, tt.expected, results)
		})
	}
}
