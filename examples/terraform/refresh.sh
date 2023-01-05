#!/bin/bash
# Copyright 2021-2023 The sacloud/autoscaler Authors
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


# 標準入力経由で渡されるJSONの例
#{
#  "source": "example",
#  "resource_name": example"
#  "scaling_job_id": "example",
#  "result": 1, # CREATED:1 | UPDATED:2 | DELETED:3
#  "current": {
#    "Resource": ...
#  }
#}

# CREATED/UPDATED/DELETEDの場合はterraformのステートをリフレッシュする
SHOULD_HANDLE=$(jq '[.] | any(.result == 1 or .result == 2 or .result == 3)')

if [ "$SHOULD_HANDLE" == "true" ]; then
  WORKING_DIR=$(cd $(dirname $0); pwd)
  terraform -chdir="$WORKING_DIR" apply -refresh-only -auto-approve 2>&1 1> /dev/null
  exit $?
else
  echo "ignored"
  exit 0
fi
