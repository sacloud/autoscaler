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

package core

import (
	"testing"
	"time"

	"github.com/sacloud/autoscaler/request"
)

func TestJobStatus_Acceptable(t *testing.T) {
	type fields struct {
		status        request.ScalingJobStatus
		statusChanged time.Time
		coolDownTime  time.Duration
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "returns true if status is DONE and is not in cooling down time",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_DONE,
				statusChanged: time.Now().Add(-2 * time.Second),
				coolDownTime:  time.Second,
			},
			want: true,
		},
		{
			name: "returns false if is in cooling down time",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_DONE,
				statusChanged: time.Now(),
				coolDownTime:  10 * time.Second,
			},
			want: false,
		},
		{
			name: "returns false if status is RUNNING",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_RUNNING,
				statusChanged: time.Now().Add(-2 * time.Second),
				coolDownTime:  time.Second,
			},
			want: false,
		},
		{
			name: "returns true if status is UNKNOWN",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_UNKNOWN,
				statusChanged: time.Now().Add(-2 * time.Second),
				coolDownTime:  time.Second,
			},
			want: true,
		},
		{
			name: "returns true if status is CANCELED",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_CANCELED,
				statusChanged: time.Now().Add(-2 * time.Second),
				coolDownTime:  time.Second,
			},
			want: true,
		},
		{
			name: "returns true if status is DONE_NOOP",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_DONE_NOOP,
				statusChanged: time.Now().Add(-2 * time.Second),
				coolDownTime:  time.Second,
			},
			want: true,
		},
		{
			name: "returns false if status is ACCEPTED",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_ACCEPTED,
				statusChanged: time.Now().Add(-2 * time.Second),
				coolDownTime:  time.Second,
			},
			want: false,
		},
		{
			name: "returns true if status is FAILED",
			fields: fields{
				status:        request.ScalingJobStatus_JOB_FAILED,
				statusChanged: time.Now().Add(-2 * time.Second),
				coolDownTime:  time.Second,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			j := &JobStatus{
				status:        tt.fields.status,
				statusChanged: tt.fields.statusChanged,
				coolDownTime:  tt.fields.coolDownTime,
			}
			if got := j.Acceptable(); got != tt.want {
				t.Errorf("Acceptable() = %v, want %v", got, tt.want)
			}
		})
	}
}
