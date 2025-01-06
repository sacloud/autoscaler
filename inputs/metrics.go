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
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sacloud/autoscaler/metrics"
)

var (
	counter     *prometheus.CounterVec
	upCounter   *prometheus.CounterVec
	downCounter *prometheus.CounterVec
)

func initMetrics() {
	metrics.InitErrorCount("inputs_to_core")

	counter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sacloud_autoscaler_webhook_requests_total",
			Help: "A counter for requests to the webhooks",
		},
		[]string{"code"},
	)

	upCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sacloud_autoscaler_webhook_requests_up",
			Help: "A counter for requests to the /up webhook",
		},
		[]string{"code"},
	)

	downCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sacloud_autoscaler_webhook_requests_down",
			Help: "A counter for requests to the /down webhook",
		},
		[]string{"code"},
	)

	counter.WithLabelValues("200")
	counter.WithLabelValues("400")
	counter.WithLabelValues("500")

	upCounter.WithLabelValues("200")
	upCounter.WithLabelValues("400")
	upCounter.WithLabelValues("500")

	downCounter.WithLabelValues("200")
	downCounter.WithLabelValues("400")
	downCounter.WithLabelValues("500")
}
