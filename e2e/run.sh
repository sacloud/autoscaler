#!/bin/bash
# Copyright 2021 The sacloud Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


set -x

: "Provisioning Infrastructures on SAKURA Cloud..."
rm -rf .terraform*
rm -f terraform.tfstate*
terraform init
terraform apply -auto-approve

: "Setting up..."
rm -f autoscaler.sock

: "Running e2e test..."
go test -v -tags=e2e ./...
RESULT=$?

if [ -n "$SKIP_CLEANUP" ]; then
: "Cleanup skipped"
else
: "Cleaning up Infrastructures..."
terraform destroy -auto-approve
fi

echo "Done: $RESULT"
exit $RESULT