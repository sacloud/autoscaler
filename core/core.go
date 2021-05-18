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

	"github.com/sacloud/autoscaler/request"
)

// Core AutoScaler Coreのインスタンス
type Core struct {
	Credential *Credential
}

func (c *Core) Up(ctx context.Context, eventSource *EventSource) *Job {
	// TODO 未実装
	return &Job{
		ID:     c.generateJobID(JobTypeUp, eventSource),
		Type:   JobTypeUp,
		Status: request.ScalingJobStatus_JOB_DONE,
	}
}

func (c *Core) Down(ctx context.Context, eventSource *EventSource) *Job {
	// TODO 未実装
	return &Job{
		ID:     c.generateJobID(JobTypeDown, eventSource),
		Type:   JobTypeDown,
		Status: request.ScalingJobStatus_JOB_DONE,
	}
}

func (c *Core) Status(ctx context.Context, jobID string) *Job {
	// TODO 未実装
	return &Job{
		ID:     jobID,
		Type:   JobTypeDown,
		Status: request.ScalingJobStatus_JOB_DONE,
	}
}

func (c *Core) generateJobID(tp JobTypes, eventSource *EventSource) string {
	return fmt.Sprintf("%s_%s", tp.String(), eventSource.String())
}

type EventSource struct {
	Source            string
	Action            string
	ResourceGroupName string
}

func (es *EventSource) String() string {
	return fmt.Sprintf("%s-%s-%s", es.Source, es.Action, es.ResourceGroupName)
}
