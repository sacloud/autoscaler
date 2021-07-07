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

package inputs

import "testing"

func TestScalingRequest_Validate(t *testing.T) {
	webhookBodyMaxLen = 1

	type fields struct {
		Source           string
		Action           string
		GroupName        string
		RequestType      string
		DesiredStateName string
	}
	tests := []struct {
		name    string
		maxLen  int
		fields  fields
		wantErr bool
	}{
		{
			name: "no error",
			fields: fields{
				Source:           "1",
				Action:           "1",
				GroupName:        "1",
				RequestType:      "up",
				DesiredStateName: "1",
			},
			maxLen:  1,
			wantErr: false,
		},
		{
			name: "source",
			fields: fields{
				Source: "12",
			},
			maxLen:  1,
			wantErr: true,
		},
		{
			name: "action",
			fields: fields{
				Action: "12",
			},
			maxLen:  1,
			wantErr: true,
		},
		{
			name: "group name",
			fields: fields{
				GroupName: "12",
			},
			maxLen:  1,
			wantErr: true,
		},
		{
			name: "desired state name",
			fields: fields{
				DesiredStateName: "12",
			},
			maxLen:  1,
			wantErr: true,
		},
		{
			name: "request type",
			fields: fields{
				RequestType: "foo",
			},
			maxLen:  1,
			wantErr: true,
		},
		{
			name: "invalid char: multibyte",
			fields: fields{
				RequestType: "あ",
			},
			maxLen:  1,
			wantErr: true,
		},
		{
			name: "invalid char: \\n",
			fields: fields{
				RequestType: "\n",
			},
			maxLen:  1,
			wantErr: true,
		},
		{
			name: "valid char",
			fields: fields{
				Action:           ` !"#%$&'()*+,-./@[\]^_{|}~`,
				GroupName:        ` !"#%$&'()*+,-./@[\]^_{|}~`,
				Source:           ` !"#%$&'()*+,-./@[\]^_{|}~`,
				DesiredStateName: ` !"#%$&'()*+,-./@[\]^_{|}~`,
				RequestType:      "up",
			},
			maxLen:  1024,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ScalingRequest{
				Source:           tt.fields.Source,
				Action:           tt.fields.Action,
				ResourceName:     tt.fields.GroupName,
				RequestType:      tt.fields.RequestType,
				DesiredStateName: tt.fields.DesiredStateName,
			}
			webhookBodyMaxLen = int64(tt.maxLen)
			if err := r.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
