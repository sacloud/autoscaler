#!/bin/bash
# Copyright 2021-2025 The sacloud/autoscaler Authors
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

#******************************************************************************
# local-execハンドラの利用例
#******************************************************************************
# 以下のように起動する
# $ autoscaler handlers local-exec --executable-path handler.sh --handler-type handle
#
# Coreのコンフィギュレーションで以下のようにカスタムハンドラを登録しておく
#
# handlers:
#   - name: "local-exec"
#     endpoint: "unix:autoscaler-handlers-local-exec.sock
#
#******************************************************************************
#
# local-execハンドラへは標準入力経由でパラメータが渡される
#   パラメータの詳細: protos/handler.proto
#
# - 終了コードに0以外を返すとエラーとみなす
# - 標準出力に何か書き込むとCoreのログに出力される
# - 標準エラーへの書き込みは無視される
#******************************************************************************

# 例: 標準入力に渡された内容をそのまま標準出力へ出力する
# 出力した内容はCoreとlocal-execハンドラのログに出力される
cat
