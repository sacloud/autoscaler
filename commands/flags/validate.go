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

package flags

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
)

// ValidateMultiFunc 指定のfuncを順次適用するfuncを返す、mergeがfalseの場合、funcがerrorを返したら即時リターンする
func ValidateMultiFunc(merge bool, funcs ...func(*cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		errors := &multierror.Error{}
		for _, fn := range funcs {
			if err := fn(cmd, args); err != nil {
				if !merge {
					return err
				}
				errors = multierror.Append(errors, err)
			}
		}
		return errors.ErrorOrNil()
	}
}
