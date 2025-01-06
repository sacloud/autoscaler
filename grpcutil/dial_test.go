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

package grpcutil

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/sacloud/autoscaler/defaults"
	"github.com/stretchr/testify/require"
)

func TestDialContext(t *testing.T) {
	tests := []struct {
		name string
		opt  *DialOption
		err  error
	}{
		{
			name: "empty destination",
			opt: &DialOption{
				Destination: "",
			},
			err: fmt.Errorf(
				"default socket file not found in [%s]",
				strings.Join(defaults.CoreSocketAddrCandidates, ", "),
			),
		},
		{
			name: "non-empty destination",
			opt: &DialOption{
				Destination: "unix:invalid.sock", // この段階ではエラーにしない
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := DialContext(context.Background(), tt.opt)
			require.EqualValues(t, tt.err, err)
		})
	}
}
