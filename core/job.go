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
	"fmt"
	"sync"
	"time"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/sacloud/autoscaler/request"
)

// JobStatus スケールアウト/イン/アップ/ダウンなどの各種ジョブを表す
//
// Inputsからのリクエストパラメータ ResourceNameごとに作成される
type JobStatus struct {
	requestType   RequestTypes
	id            string
	status        request.ScalingJobStatus
	statusChanged time.Time
	coolDownTime  time.Duration
	mu            sync.Mutex
}

func NewJobStatus(req *requestInfo, coolDownTime time.Duration) *JobStatus {
	return &JobStatus{
		requestType:   req.requestType,
		id:            req.ID(),
		status:        request.ScalingJobStatus_JOB_UNKNOWN,
		statusChanged: time.Now(),
		coolDownTime:  coolDownTime,
	}
}

func (j *JobStatus) Type() RequestTypes {
	return j.requestType
}

func (j *JobStatus) ID() string {
	return j.id
}

func (j *JobStatus) Status() request.ScalingJobStatus {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.status
}

func (j *JobStatus) SetStatus(status request.ScalingJobStatus) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.status = status
	j.statusChanged = time.Now()
}

func (j *JobStatus) String() string {
	return fmt.Sprintf("Type: %s ID: %s Status: %s StatusChanged: %s", j.Type(), j.ID(), j.Status(), j.statusChanged)
}

// Acceptable このジョブが新規に受け入れ可能(新たに起動できる)状態の場合true
func (j *JobStatus) Acceptable() bool {
	switch j.Status() {
	case request.ScalingJobStatus_JOB_ACCEPTED, request.ScalingJobStatus_JOB_RUNNING:
		// すでに受け入れ済み or 実行中
		return false
	default:
		// 以外は冷却期間でなければtrue
		return !j.inCoolDownTime()
	}
}

// inCoolDownTime StatusがDONE、かつ冷却期間内であればtrue
func (j *JobStatus) inCoolDownTime() bool {
	if j.coolDownTime == 0 {
		j.coolDownTime = defaults.CoolDownTime
	}
	return j.Status() == request.ScalingJobStatus_JOB_DONE &&
		j.statusChanged.After(time.Now().Add(-1*j.coolDownTime))
}
