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
	"fmt"
	"sync"
	"time"

	"github.com/sacloud/autoscaler/request"
)

// JobStatus スケールアウト/イン/アップ/ダウンなどの各種ジョブを表す
//
// Inputsからのリクエストパラメータ ResourceNameごとに作成される
type JobStatus struct {
	id       string
	status   request.ScalingJobStatus
	coolDown *CoolDown
	mu       sync.Mutex
}

func NewJobStatus(req *requestInfo, coolDown *CoolDown) *JobStatus {
	if coolDown == nil {
		coolDown = &CoolDown{}
	}
	return &JobStatus{
		id:       req.ID(),
		status:   request.ScalingJobStatus_JOB_DONE, // 完了状態 == ジョブ受け入れ可能ということで初期値にしておく
		coolDown: coolDown,
	}
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
}

func (j *JobStatus) String() string {
	return fmt.Sprintf("ID: %s Status: %s", j.ID(), j.Status())
}

// Acceptable このジョブが新規に受け入れ可能(新たに起動できる)状態の場合true
func (j *JobStatus) Acceptable(requestType RequestTypes, lastModifiedAt time.Time) bool {
	switch j.Status() {
	case request.ScalingJobStatus_JOB_ACCEPTED, request.ScalingJobStatus_JOB_RUNNING:
		// すでに受け入れ済み or 実行中
		return false
	default:
		// 以外は冷却期間でなければtrue
		return !j.inCoolDownTime(requestType, lastModifiedAt)
	}
}

// inCoolDownTime StatusがDONE、かつ冷却期間内であればtrue
func (j *JobStatus) inCoolDownTime(requestType RequestTypes, lastModifiedAt time.Time) bool {
	coolDownTime := j.coolDown.Duration(requestType)
	return j.Status() == request.ScalingJobStatus_JOB_DONE &&
		lastModifiedAt.After(time.Now().Add(-1*coolDownTime))
}
