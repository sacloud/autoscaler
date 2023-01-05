// Copyright 2021-2023 The sacloud/autoscaler Authors
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

package handlers

import (
	"fmt"

	"github.com/sacloud/autoscaler/handler"
)

func report(sender ResponseSender, id string, status handler.HandleResponse_Status, formatAndArgs ...interface{}) error {
	var log string

	switch len(formatAndArgs) {
	case 0:
		log = ""
	case 1:
		log = fmt.Sprintf("%s", formatAndArgs[0])
	default:
		log = fmt.Sprintf(formatAndArgs[0].(string), formatAndArgs[1:]...)
	}

	return sender.Send(&handler.HandleResponse{
		ScalingJobId: id,
		Status:       status,
		Log:          log,
	})
}

type reportFn func(handler.HandleResponse_Status, ...interface{}) error

// reporter reportをカリー化する
func reporter(sender ResponseSender, id string) reportFn {
	return func(status handler.HandleResponse_Status, formatAndArgs ...interface{}) error {
		return report(sender, id, status, formatAndArgs...)
	}
}
