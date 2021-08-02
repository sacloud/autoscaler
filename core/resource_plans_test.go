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
	"context"
	"reflect"
	"testing"
)

type stubResourcePlan struct {
	name       string
	memorySize int
}

func (p *stubResourcePlan) PlanName() string {
	return p.name
}

func (p *stubResourcePlan) Equals(resource interface{}) bool {
	size := resource.(int)
	return p.memorySize == size
}

func (p *stubResourcePlan) LessThan(resource interface{}) bool {
	size := resource.(int)
	return p.memorySize < size
}

func (p *stubResourcePlan) GreaterThan(resource interface{}) bool {
	size := resource.(int)
	return size < p.memorySize
}

func (p *stubResourcePlan) LessThanPlan(plan ResourcePlan) bool {
	target, ok := plan.(*stubResourcePlan)
	if !ok {
		return false
	}
	return p.memorySize < target.memorySize
}

func Test_desiredPlan(t *testing.T) {
	type args struct {
		ctx     *RequestContext
		current interface{}
		plans   ResourcePlans
	}
	tests := []struct {
		name    string
		args    args
		want    ResourcePlan
		wantErr bool
	}{
		{
			name: "Up returns next plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType: requestTypeUp,
					},
				},
				current: 1,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 1},
				},
			},
			want:    &stubResourcePlan{memorySize: 2},
			wantErr: false,
		},
		{
			name: "Up returns nil if resource has greater plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType: requestTypeUp,
					},
				},
				current: 2,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 1},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Up returns next plan if resource has unknown and lesser plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType: requestTypeUp,
					},
				},
				current: 2,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 1},
					&stubResourcePlan{memorySize: 3},
					&stubResourcePlan{memorySize: 4},
				},
			},
			want:    &stubResourcePlan{memorySize: 3},
			wantErr: false,
		},
		{
			name: "Down returns prev plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType: requestTypeDown,
					},
				},
				current: 2,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 1},
				},
			},
			want:    &stubResourcePlan{memorySize: 1},
			wantErr: false,
		},
		{
			name: "Down returns nil if resource has lesser plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType: requestTypeDown,
					},
				},
				current: 1,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 1},
					&stubResourcePlan{memorySize: 2},
				},
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "Down returns prev plan if resource has unknown and larger plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType: requestTypeDown,
					},
				},
				current: 3,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 1},
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 4},
					&stubResourcePlan{memorySize: 5},
				},
			},
			want:    &stubResourcePlan{memorySize: 2},
			wantErr: false,
		},
		{
			name: "Up returns named plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType:      requestTypeUp,
						desiredStateName: "named",
					},
				},
				current: 1,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 3, name: "named"},
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 1},
				},
			},
			want:    &stubResourcePlan{memorySize: 3, name: "named"},
			wantErr: false,
		},
		{
			name: "Up returns error when greater plan not exists",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType:      requestTypeUp,
						desiredStateName: "named",
					},
				},
				current: 5,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 3, name: "named"},
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 1},
				},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Down returns named plan",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType:      requestTypeDown,
						desiredStateName: "named",
					},
				},
				current: 3,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 3},
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 1, name: "named"},
				},
			},
			want:    &stubResourcePlan{memorySize: 1, name: "named"},
			wantErr: false,
		},
		{
			name: "Down returns error when greater plan not exists",
			args: args{
				ctx: &RequestContext{
					ctx: context.Background(),
					request: &requestInfo{
						requestType:      requestTypeDown,
						desiredStateName: "named",
					},
				},
				current: 2,
				plans: ResourcePlans{
					&stubResourcePlan{memorySize: 3, name: "named"},
					&stubResourcePlan{memorySize: 2},
					&stubResourcePlan{memorySize: 1},
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := desiredPlan(tt.args.ctx, tt.args.current, tt.args.plans)
			if (err != nil) != tt.wantErr {
				t.Errorf("desiredPlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("desiredPlan() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourcePlans_within(t *testing.T) {
	type args struct {
		resource interface{}
	}
	tests := []struct {
		name string
		p    ResourcePlans
		args args
		want bool
	}{
		{
			name: "empty plans",
			p:    ResourcePlans{},
			args: args{
				resource: 1,
			},
			want: false,
		},
		{
			name: "single plan: less than",
			p: ResourcePlans{
				&stubResourcePlan{memorySize: 1},
			},
			args: args{
				resource: 0,
			},
			want: false,
		},
		{
			name: "single plan: equals",
			p: ResourcePlans{
				&stubResourcePlan{memorySize: 1},
			},
			args: args{
				resource: 1,
			},
			want: true,
		},
		{
			name: "single plan: greater than",
			p: ResourcePlans{
				&stubResourcePlan{memorySize: 1},
			},
			args: args{
				resource: 2,
			},
			want: false,
		},
		{
			name: "plans: less than",
			p: ResourcePlans{
				&stubResourcePlan{memorySize: 1},
				&stubResourcePlan{memorySize: 2},
			},
			args: args{
				resource: 0,
			},
			want: false,
		},
		{
			name: "plans: within 1",
			p: ResourcePlans{
				&stubResourcePlan{memorySize: 1},
				&stubResourcePlan{memorySize: 2},
			},
			args: args{
				resource: 1,
			},
			want: true,
		},
		{
			name: "plans: within 2",
			p: ResourcePlans{
				&stubResourcePlan{memorySize: 1},
				&stubResourcePlan{memorySize: 2},
			},
			args: args{
				resource: 2,
			},
			want: true,
		},
		{
			name: "plans: out: greater than",
			p: ResourcePlans{
				&stubResourcePlan{memorySize: 1},
				&stubResourcePlan{memorySize: 2},
			},
			args: args{
				resource: 3,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.within(tt.args.resource); got != tt.want {
				t.Errorf("within() = %v, want %v", got, tt.want)
			}
		})
	}
}
