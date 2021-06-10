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
	"fmt"
	"reflect"
	"testing"
)

func TestActions_Validate(t *testing.T) {
	type args struct {
		ctx      context.Context
		handlers Handlers
	}
	tests := []struct {
		name string
		a    Actions
		args args
		want []error
	}{
		{
			name: "basic",
			a: Actions{
				"foobar": []string{"handler1", "handler2"},
			},
			args: args{
				ctx: context.Background(),
				handlers: Handlers{
					{Name: "handler1"},
					{Name: "handler2"},
				},
			},
			want: nil,
		},
		{
			name: "not exists",
			a: Actions{
				"foobar": []string{"not-exists1", "not-exists2"},
			},
			args: args{
				ctx: context.Background(),
				handlers: Handlers{
					{Name: "handler1"},
					{Name: "handler2"},
				},
			},
			want: []error{
				fmt.Errorf("handler %q is not defined", "not-exists1"),
				fmt.Errorf("handler %q is not defined", "not-exists2"),
			},
		},
		{
			name: "mixed",
			a: Actions{
				"foobar": []string{"handler1", "not-exists2", "handler2"},
			},
			args: args{
				ctx: context.Background(),
				handlers: Handlers{
					{Name: "handler1"},
					{Name: "handler2"},
				},
			},
			want: []error{
				fmt.Errorf("handler %q is not defined", "not-exists2"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Validate(tt.args.ctx, tt.args.handlers); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Validate() = %v, want %v", got, tt.want)
			}
		})
	}
}
