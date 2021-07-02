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

	"github.com/sacloud/autoscaler/test"
)

func TestResourceDefDNS_Validate(t *testing.T) {
	_, cleanup1 := test.AddTestDNS(t, "test1.com")
	defer cleanup1()

	_, cleanup2 := test.AddTestDNS(t, "test2.com")
	defer cleanup2()

	_, cleanup3 := test.AddTestServer(t, "test")
	defer cleanup3()

	type fields struct {
		Selector        *ResourceSelector
		ResourceDefBase *ResourceDefBase
	}
	tests := []struct {
		name      string
		fields    fields
		wantError bool
	}{
		{
			name: "returns error when having children and returns multiple resource",
			fields: fields{
				ResourceDefBase: &ResourceDefBase{
					TypeName: ResourceTypeDNS.String(),
					children: ResourceDefinitions{
						&ResourceDefServer{
							ResourceDefBase: &ResourceDefBase{
								TypeName: ResourceTypeServer.String(),
							},
							Selector: &MultiZoneSelector{
								ResourceSelector: &ResourceSelector{
									Names: []string{"test"},
								},
							},
						},
					},
				},
				Selector: &ResourceSelector{
					Names: []string{"test"},
				},
			},
			wantError: true,
		},
		{
			name: "returns no error",
			fields: fields{
				Selector: &ResourceSelector{
					Names: []string{"test1.com"},
				},
				ResourceDefBase: &ResourceDefBase{
					TypeName: ResourceTypeDNS.String(),
					children: ResourceDefinitions{
						&ResourceDefServer{
							ResourceDefBase: &ResourceDefBase{
								TypeName: ResourceTypeServer.String(),
							},
							Selector: &MultiZoneSelector{
								ResourceSelector: &ResourceSelector{
									Names: []string{"test"},
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &ResourceDefDNS{
				ResourceDefBase: tt.fields.ResourceDefBase,
				Selector:        tt.fields.Selector,
			}
			if got := d.Validate(testContext(), test.APIClient); tt.wantError != (len(got) > 0) {
				t.Errorf("Validate() = %v, wantError %t", got, tt.wantError)
			}
		})
	}
}
