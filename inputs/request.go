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

package inputs

import (
	"github.com/sacloud/autoscaler/validate"
)

// ScalingRequest Inputsからのリクエストを表す
type ScalingRequest struct {
	Source           string `name:"source" validate:"omitempty,printascii,max=1024"`
	ResourceName     string `name:"resource-name" validate:"omitempty,printascii,max=1024"`
	RequestType      string `name:"request-type" validate:"required,oneof=up down"`
	DesiredStateName string `name:"desired-state-name" validate:"omitempty,printascii,max=1024"`
}

func (r *ScalingRequest) Validate() error {
	return validate.Struct(r)
}
