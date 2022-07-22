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

package defaults

import "time"

const (
	CoreSocketAddr   = "unix:autoscaler.sock" // CoreのデフォルトgRPCエンドポイント(Inputsから呼ばれる)
	CoreConfigPath   = "autoscaler.yaml"      // CoreのConfigurationのファイルパス
	CoreExporterAddr = ":8081"                // CoreのExporterがリッスンするデフォルトのアドレス
	ListenAddress    = ":8080"                // Inputsがリッスンするデフォルトのアドレス

	ResourceName     = "default"
	SourceName       = "default"
	DesiredStateName = "default"

	CoolDownTime        = 10 * time.Minute // 同一ジョブの実行制御のための冷却期間
	ShutdownGracePeriod = 10 * time.Minute
)

var (
	// CoreSocketAddrCandidates CoreのgRPCエンドポイントの候補値のリスト
	//
	// InputsでCoreのエンドポイントアドレスが省略された場合にこれらを上から順に確認し、ファイルの存在が確認できたものから利用される
	CoreSocketAddrCandidates = []string{
		CoreSocketAddr,
		"unix:/var/run/autoscaler.sock",
		"unix:/var/run/autoscaler/autoscaler.sock",
	}

	// SetupGracePeriods リソース定義種別ごとのセットアップのための猶予時間(秒数)
	SetupGracePeriods = map[string]int{
		"Server": 60,
	}
)
